package mockable

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/qiniu/http/restrpc.v1"
	"qiniupkg.com/mockable.v1/net"
	mockhttp "qiniupkg.com/mockable.v1/net/http"
	"qiniupkg.com/x/log.v7"
)

/* ---------------------------------------------------------------------------

{
    "idcs": [
        {
            "name": <IdcName1>,
            "nodes": [
                {
                    "name": <NodeName1>,
                    "ips": {
                        "tel": "<LocalIP11>",
                        "bgp": "<LocalIP12>"
                        ...
                    },
                    "defaultIsp": "tel", # for net.Dial
					"procs": [
						{# 第一个需要提供测量带宽的端口
							"name": "<Name>",
							"workdir": <WorkDir1>,
							"exec": [<App1>, <Arg11>, ..., <Arg1N>]
						},
						...
					]
                },
                ...
            ]
        },
        ...
    ],
    "speeds": [
        {
            "from": <IdcNameFrom:Isp>,
            "to": <IdcNameTo:Isp>,
            "speed": [<SpeedItem1>, ...] #详细见net.Speed的定义，在len(speed)=0时表示两个Idc不连通
        },
        ...
    ],
    "defaultSpeed": [<SpeedItem1>, ...]
}

// -------------------------------------------------------------------------*/

type Topology struct {
	Idcs         []IdcInfo   `json:"idcs"`
	Speeds       []SpeedInfo `json:"speeds"`
	DefaultSpeed net.Speed   `json:"defaultSpeed"`
}

type IdcInfo struct {
	Name  string     `json:"name"`
	Nodes []NodeInfo `json:"nodes"`
}

type NodeInfo struct {
	Name       string            `json:"name"`
	Ips        map[string]string `json:"ips"`
	DefaultIsp string            `json:"defaultIsp"`
	Procs      []ProcInfo        `json:"procs"`
}

type ProcInfo struct {
	Name    string   `json:"name"`
	WorkDir string   `json:"workdir,omitempty"`
	Exec    []string `json:"exec"`
}

type SpeedInfo struct {
	From  string    `json:"from"`
	To    string    `json:"to"`
	Speed net.Speed `json:"speed,omitempty"`
}

var (
	statPort       = flag.String("m:statPort", "", "stat port")
	nodeName       = flag.String("m:node", "", "node name")
	topologyBase64 = flag.String("m:top", "", "urlsafe base64 encoded topology")
)

var (
	NodeName string
)

func Init() { // 由节点进程(Node)的main函数调用，用来初始化mockable/net、mockable/http、etc。

	flag.Parse()
	if *nodeName == "" { // 不指定 -node 说明不是测试环境
		return
	}
	NodeName = *nodeName

	if *topologyBase64 == "" {
		log.Fatalln("Please specify -node <NodeName> -top <Topology> parameters")
	}

	initWithArgs(NodeName, *topologyBase64)

	if *statPort != "" {
		router := restrpc.Router{}
		mux := router.Register(&statService{})
		go func() {
			err := mockhttp.MockListenAndServe("0.0.0.0:"+*statPort, mux)
			if err != nil {
				log.Fatal("MockListenAndServe", err)
			}
		}()
	}
}

func initWithArgs(nodeName string, topologyBase64 string) {
	topologyData, err := base64.URLEncoding.DecodeString(topologyBase64)
	if err != nil {
		log.Fatalln("invalid topology: not urlsafe base64 encoded -", err)
	}

	var top Topology
	err = json.Unmarshal(topologyData, &top)
	if err != nil {
		log.Fatalln("invalid topology: json.Unmarshal failed -", err)
	}

	idcSpeeds := make(map[string]net.Speed)
	for _, t := range top.Speeds {
		idcSpeeds[t.From+"/"+t.To] = t.Speed
		idcSpeeds[t.To+"/"+t.From] = t.Speed
	}

	for _, idc := range top.Idcs {
		for _, node := range idc.Nodes {
			if node.Name == nodeName {
				ips := make([]string, 1, 2)
				isps := make(map[string]int, len(node.Ips))
				ips[0] = node.Ips[node.DefaultIsp]
				isps[node.DefaultIsp] = 0
				if ips[0] == "" {
					log.Fatalln("invalid default isp:", node.DefaultIsp, "of node:", node.Name)
				}
				for isp, ip := range node.Ips {
					if isp == node.DefaultIsp {
						continue
					}
					isps[isp] = len(ips)
					ips = append(ips, ip)
				}

				nodeSpeeds := make(map[string][]net.Speed)
				for _, idcTo := range top.Idcs {
					for _, nodeTo := range idcTo.Nodes {
						for ispTo, ipTo := range nodeTo.Ips {
							routeTo := "/" + idcTo.Name + ":" + ispTo
							ispSpeeds := make([]net.Speed, len(isps))
							if idc.Name == idcTo.Name {
								for i := range ispSpeeds {
									ispSpeeds[i] = net.SpeedNotLimit
								}
							} else {
								for isp, idx := range isps {
									route := idc.Name + ":" + isp + routeTo
									if ispSpeed, ok := idcSpeeds[route]; ok {
										ispSpeeds[idx] = ispSpeed
									} else {
										ispSpeeds[idx] = top.DefaultSpeed
									}
								}
							}
							nodeSpeeds[ipTo] = ispSpeeds
						}
					}
				}
				net.Init(&net.InitConfig{
					MockingIPs: ips,
					Speeds:     nodeSpeeds,
				})
				return
			}
		}
	}

	log.Fatalln("unknown node name:", nodeName)
}

// ---------------------------------------------------------------------------

type Cluster struct {
	procs map[string]ProcHandle
}

type ProcHandle struct {
	cmd     *exec.Cmd
	logfile *os.File
}

func shutdown(cmd *exec.Cmd, wait bool) (err error) {
	if cmd.Process == nil {
		return
	}
	cmd.Process.Kill()
	if wait {
		ch := make(chan error, 1)
		go func() {
			errExit := cmd.Wait()
			if errExit != nil {
				log.Infof("node `%s`(%v) exit: %v\n", cmd.Args[0], cmd.Process.Pid, errExit)
			}
			ch <- errExit
		}()
		return <-ch
	}
	return nil
}

func (p *Cluster) Close() (err error) {

	var wg sync.WaitGroup
	wg.Add(len(p.procs))
	for _, proc := range p.procs {
		cmd := proc.cmd
		defer proc.logfile.Close()
		if cmd.Process == nil {
			continue
		}
		shutdown(cmd, false)
		go func() {
			errExit := cmd.Wait()
			if errExit != nil {
				log.Infof("node `%s`(%v) exit: %v\n", cmd.Args[0], cmd.Process.Pid, errExit)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return nil
}

func (p *Cluster) Shutdown(node, procName string, wait bool) (err error) {

	if proc, ok := p.procs[node+"."+procName]; ok {
		err = shutdown(proc.cmd, wait)
		proc.logfile.Close()
		return
	}
	return syscall.ENOENT
}

func RunCluster(topology string) (cluster *Cluster) {

	var top Topology
	err := json.NewDecoder(strings.NewReader(topology)).Decode(&top)
	if err != nil {
		log.Fatalln("RunCluster failed: invalid topology(json.Unmarshal failed) -", err)
	}

	b, err := json.Marshal(&top)
	if err != nil {
		log.Fatalln("RunCluster failed:", err)
		return
	}
	topologyBase64 := base64.URLEncoding.EncodeToString(b)

	procs := make(map[string]ProcHandle)
	for _, idc := range top.Idcs {
		for _, node := range idc.Nodes {
			if node.Name == "" {
				initWithArgs("", topologyBase64)
				continue
			}
			for _, proc := range node.Procs {
				if len(proc.Exec) == 0 {
					log.Fatalln("RunCluster failed: invalid topology - `exec` attribute is empty")
				}
				proc.Exec[0], err = filepath.Abs(proc.Exec[0])
				if err != nil {
					log.Fatal(proc.Exec[0], "is invalid")
				}
				procName := filepath.Base(proc.Exec[0])
				logfile, err := os.Create(proc.WorkDir + "/" + procName + ".log")
				if err != nil {
					log.Fatalln("Create log file failed -", err)
				}
				log.Infof("run %s with: %v", node.Name, proc.Exec)
				proc.Exec = append(proc.Exec, "-m:node", node.Name, "-m:top", topologyBase64)
				if procName == "pili-noded" {
					proc.Exec = append(proc.Exec, "-m:statPort", "199")
				}
				cmd := exec.Command(proc.Exec[0], proc.Exec[1:]...)
				cmd.Dir = proc.WorkDir
				cmd.Stdout = logfile
				cmd.Stderr = logfile
				errStart := cmd.Start()
				if errStart != nil {
					log.Fatalf("RunCluster failed: start node `%s` failed - %v", node.Name, errStart)
				}
				procs[node.Name+"."+proc.Name] = ProcHandle{cmd, logfile}
			}
		}
	}
	return &Cluster{procs: procs}
}

// -------------------------------------------------------------

type prefixWriter struct {
	prefix     string
	writeTo    io.Writer
	needPrefix bool
}

func newPrefixWriter(prefix string, writer io.Writer) io.Writer {

	return &prefixWriter{
		prefix:     prefix,
		writeTo:    writer,
		needPrefix: true,
	}
}

func (pw *prefixWriter) Write(b []byte) (n int, err error) {

	nb := b
	for {
		if len(nb) == 0 {
			return len(b), nil
		}
		if pw.needPrefix {
			_, err := pw.writeTo.Write([]byte(pw.prefix))
			if err != nil {
				return 0, err
			}
			pw.needPrefix = false
		}
		idx := bytes.IndexByte(nb, '\n')
		if idx != -1 {
			pw.needPrefix = true
			_, err := pw.writeTo.Write(nb[:idx+1])
			if err != nil {
				return 0, err
			}
			nb = nb[idx+1:]
		}
	}
}

// ---------------------------------------------------------------------------
