package mq

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/api"
	"qbox.us/auditlog2"
	"qbox.us/errors"
	"qbox.us/net/httputil"
	"qbox.us/net/httputil/flag"
	"qbox.us/net/webroute"
	"qbox.us/servend/account"
	"qbox.us/servestk"
	"qiniupkg.com/trace.v1"

	. "github.com/qiniu/ctype"
	cioutil "qbox.us/cc/ioutil"
)

// ---------------------------------------------------------------

type Config struct {
	Account  account.InterfaceEx
	DataPath string
	auditlog2.Config
	ChunkBits     uint
	Expires       uint
	SaveHours     int
	CheckInterval int64
}

// ---------------------------------------------------------------

type Service struct {
	Config

	mutex sync.RWMutex
	mqs   map[string]*Instance
}

func Open(cfg *Config) (r *Service, err error) {

	_, err = os.Stat(cfg.DataPath)
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(cfg.DataPath, 0755)
		if err != nil {
			err = errors.Info(err, "mkdir mq data path failed").Detail(err)
			return
		}
	}
	if cfg.CheckInterval == int64(0) {
		cfg.CheckInterval = 3600 //1hour
	}

	if cfg.SaveHours == 0 {
		cfg.SaveHours = 72 //3days
	}
	if cfg.SaveHours <= 24 {
		log.Fatalln("should not set saveHours less than 24 hours for security")
	}
	fis, err := cioutil.ReadDir(cfg.DataPath)
	if err != nil {
		err = errors.Info(err, "read mqs from datapath failed").Detail(err)
		return
	}

	mqs := make(map[string]*Instance)
	r = &Service{Config: *cfg, mqs: mqs}
	r.DataPath += "/"

	for _, fi := range fis {
		name := fi.Name()
		if checkValidMQ(name) == nil {
			mq, err2 := OpenInstance(r.DataPath+name, r.ChunkBits, uint32(r.Expires))
			if err2 != nil {
				errors.Info(err2, "open mq failed:", name).Detail(err2).Warn()
				continue
			}
			mqs[name] = mq
		}
	}
	return
}

func checkValidMQ(name string) (err error) {

	pos := strings.Index(name, "-")
	if pos <= 0 {
		err = errors.Info(errors.EINVAL, "invalid mq name")
		return
	}

	if !IsType(DIGIT|ALPHA, name[:pos]) {
		err = errors.Info(errors.EINVAL, "invalid mq name - invalid uid part")
		return
	}

	if !IsType(XMLSYMBOL_NEXT_CHAR, name[pos+1:]) {
		err = errors.Info(errors.EINVAL, "invalid mq name - invalid mqid part")
	}
	return
}

func (r *Service) Close() (err error) {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, mq := range r.mqs {
		mq.Close()
	}
	r.mqs = make(map[string]*Instance)

	return
}

func (r *Service) DoMake_(w http.ResponseWriter, req *http.Request) {

	httputil.ReplyWithCode(w, 596) // 未实现
	return
}

func (r *Service) getMQ(req *http.Request) (mq *Instance, err error) {

	user, err := account.GetAuthExt(r.Account, req)
	if err != nil {
		err = errors.Info(api.EBadToken, "bad token").Detail(err)
		return
	}

	var mqid string
	flag.Parse(req.URL.Path[1:], flag.SW{Val: &mqid})
	if mqid == "" {
		err = errors.Info(errors.EINVAL, "invalid arguments: mqid not provided")
		return
	}

	name := strconv.FormatUint(uint64(user.Uid), 36) + "-" + mqid
	return r.adminGetMQ(name)
}

func (r *Service) adminGetMQ(name string) (mq *Instance, err error) {

	r.mutex.RLock()
	mq, ok := r.mqs[name]
	r.mutex.RUnlock()

	if !ok {
		err = errors.Info(errors.ENOENT, "mq not found")
	}
	return
}

/*
 * put (向MQ添加新消息)
 *
请求包：
	POST /put/<MQID>
	Content-Type: application/octet-stream
	Body: <Message>
返回包：
	200 OK
	X-Id: <MessageId>
*/
func (r *Service) DoPut_(w http.ResponseWriter, req *http.Request) {

	mq, err := r.getMQ(req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	msg, err := ioutil.ReadAll(req.Body)
	if err != nil {
		err = errors.Info(err, "read message failed").Detail(err)
		httputil.Error(w, err)
		return
	}

	msgId, err := mq.Put(msg)
	if err != nil {
		err = errors.Info(err, "put message to mq failed").Detail(err)
		httputil.Error(w, err)
		return
	}

	//TODO: msgId 需要防篡改

	w.Header().Set("X-Id", strconv.FormatInt(msgId, 36))
	httputil.ReplyWithCode(w, 200)
}

/*
 * get (从MQ提取消息)
 *
请求包：
	POST /get/<MQID>
返回包：
	200 OK
	Content-Type: application/octet-stream
	X-Id: <MessageId>
	Body: <Message>
*/
func (r *Service) DoGet_(w http.ResponseWriter, req *http.Request) {

	mq, err := r.getMQ(req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	msgId, msg, err := mq.Get()
	if err != nil {
		err = errors.Info(err, "get message from mq failed").Detail(err)
		httputil.Error(w, err)
		return
	}

	w.Header().Set("X-Id", strconv.FormatInt(msgId, 36))
	httputil.ReplyWith(w, 200, "application/octet-stream", msg)
}

/*
 * stat (获取MQ状态)
 *
请求包：
	POST /stat/<MQID>
返回包：
	200 OK
	Content-Type: application/json
	{
		"todo" : <todoLen> ,
		"doing" : <processingLen>
	}
*/
func (r *Service) DoStat_(w http.ResponseWriter, req *http.Request) {

	mq, err := r.getMQ(req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	statInfo := mq.Stat()
	httputil.Reply(w, 200, statInfo)
}

/*
 * delete (从MQ删除消息)
 *
请求包：
	POST /delete/<MQID>
	X-Id: <MessageId>
返回包：
	200 OK
*/
func (r *Service) DoDelete_(w http.ResponseWriter, req *http.Request) {

	mq, err := r.getMQ(req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	xId := req.Header.Get("X-Id")
	if xId == "" {
		err = errors.Info(errors.EINVAL, "invalid http header: X-Id required")
		httputil.Error(w, err)
		return
	}

	msgId, err := strconv.ParseInt(xId, 36, 64)
	if err != nil {
		err = errors.Info(errors.EINVAL, "invalid arguments: invalid X-Id").Detail(err)
		httputil.Error(w, err)
		return
	}

	err = mq.Delete(msgId)
	if err != nil {
		err = errors.Info(err, "delete message from mq failed").Detail(err)
		httputil.Error(w, err)
		return
	}

	httputil.ReplyWithCode(w, 200)
}

func (r *Service) RegisterHandlers(mux1 *http.ServeMux) error {

	lh, err := auditlog2.OpenExt("MQ", &r.Config.Config, r.Account)
	if err != nil {
		return err
	}
	mux := trace.NewServeMuxWith(servestk.New(mux1, lh.Handler()))
	return webroute.Register(mux, r, nil)
}

func Run(addr string, cfg *Config) error {

	r, err := Open(cfg)
	if err != nil {
		err = errors.Info(err, "mq.Run: mq.New failed").Detail(err).Warn()
		return err
	}
	defer r.Close()

	//deleteDataFiles
	go r.loopDeleteDataFiles()

	mux := http.NewServeMux()
	err = r.RegisterHandlers(mux)
	if err != nil {
		err = errors.Info(err, "mq.Run: mq.RegisterHandlers failed").Detail(err).Warn()
		return err
	}

	return http.ListenAndServe(addr, mux)
}

func (r *Service) loopDeleteDataFiles() {
	chunkSize := 1 << r.ChunkBits
	log.Infof("mq data file path: %v, size of every dataFile: %v, file will be delete after consumed %v hours\n", r.DataPath, chunkSize, r.SaveHours)

	for {
		delCountTotal := r.deleteDataFiles(int64(chunkSize))
		log.Infof("deleteDataFiles delCountTotal: %v, sleep: %v\n", delCountTotal, r.CheckInterval)
		time.Sleep(time.Duration(r.CheckInterval) * time.Second)
	}
}

var testSaveHours = func() {}

// ---------------------------------------------------------------

/*index里面的offTimeout之前的数据都被处理过，可以被删除，对于已经被处理过的数据保留一定的时间，saveHours作为可配置参数，是指数据被处理之后多久之后被删除
INDEX文件刷新不及时，可能会有较大延迟问题，使用内存中的offTimeout
*/
func (r *Service) deleteDataFiles(chunkSize int64) int {
	xl := xlog.NewDummy()
	delCountTotal := 0
	testSaveHours()

	saveTime := time.Duration(r.SaveHours) * time.Hour
	var names []string
	var mqs []*Instance
	r.mutex.RLock()
	for name, mq := range r.mqs {
		names = append(names, name)
		mqs = append(mqs, mq)
	}
	r.mutex.RUnlock()
	for i, name := range names {
		mqDataPath := path.Join(r.DataPath, name)
		xl.Info("instance: ", name, "dataPath: ", mqDataPath)
		offTimeOut := atomic.LoadInt64(&mqs[i].offTimeout)
		delCount := deleteMqDataFile(xl, mqDataPath, chunkSize, saveTime, offTimeOut)
		xl.Infof("instance: %v, delete %v files\n", name, delCount)
		delCountTotal += delCount
	}
	return delCountTotal
}

func deleteMqDataFile(xl *xlog.Logger, filePath string, chunkSize int64, saveTime time.Duration, offTimeout int64) int {
	count := 0
	nameStr := strconv.FormatInt(offTimeout/chunkSize, 36)
	xl.Printf("index OffTimeout: %v, fileName: %v\n", offTimeout, nameStr)
	fileList, err := getFiles(xl, filePath, offTimeout, chunkSize, saveTime)
	if err != nil {
		xl.Fatalln("getFiles", err)
	}
	for _, fileName := range fileList {
		err = deleteLocalFile(xl, fileName)
		if err != nil {
			xl.Errorf("deleteLocalFile, err: %v, will try to delete it nextTime", err)
		} else {
			count++
		}
	}
	return count
}

func deleteLocalFile(xl *xlog.Logger, fname string) (err error) {
	err = os.Remove(fname)
	if err != nil {
		xl.Errorf("os.Remove fileName :%s, err :%s\n", fname, err)
		return
	}
	xl.Infof("delete file :%s success\n", fname)
	return nil
}

//文件在偏移在offTimeout之前,表示已经被消费;对于这样的文件可以保留saveDays之后删除
func getFiles(xl *xlog.Logger, pathStr string, offTimeout, fileSize int64, saveTime time.Duration) ([]string, error) {
	var fileNames []string
	fi, err := os.Stat(pathStr)
	if err != nil {
		err = errors.Info(err, "os.Stat", pathStr).Detail(err)
		return nil, err
	}
	if !fi.IsDir() {
		err = errors.New("not dir")
		err = errors.Info(err, pathStr).Detail(err)
		return nil, err
	}
	dataFiles, err := ioutil.ReadDir(pathStr)
	if err != nil {
		err = errors.Info(err, "ioutil.ReadDir", pathStr).Detail(err)
		return nil, err
	}
	for _, dateFile := range dataFiles {
		fileName := dateFile.Name()
		if fileName == "INDEX" {
			continue
		}
		fileName2, err := strconv.ParseInt(fileName, 36, 64)
		if err != nil {
			err = errors.Info(err, "strconv.ParseInt", fileName).Detail(err)
			return nil, err
		}

		//找到该文件末尾的offset
		fileOffset := (fileName2 + 1) * fileSize
		if fileOffset < offTimeout {
			if time.Since(dateFile.ModTime()) <= saveTime {
				xl.Printf("no delete for file %v now\n", dateFile.Name())
				continue
			}
			xl.Printf("file: %v, the end offset: %v, offTimeout: %v\n", fileName, fileOffset, offTimeout)
			fileNames = append(fileNames, path.Join(pathStr, fileName))
		}
	}
	return fileNames, nil
}
