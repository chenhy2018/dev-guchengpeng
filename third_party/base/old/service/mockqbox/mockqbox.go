package main

import (
	"flag"
	"fmt"
	"http"
	"os"
	"qbox.us/mi"
	"qbox.us/mockacc"
	"qbox.us/mockfs"
	"qbox.us/mockio"
	"qbox.us/mockrs"
	"strconv"
	"strings"
)

var (
	useTest     = flag.Bool("t", false, "Use test.qbox.us:9876")
	host        = flag.String("h", "127.0.0.1", "<BindingHost>")
	google      = flag.Bool("g", false, "Use google account or not")
	userAndPass = flag.String("a", "", "<User>:<Password>")
	root        = flag.String("d", "", "<RootDirectory>")
	port        = flag.Int("p", 9876, "<Port>")
)

const usageMsg = `
The mockqbox application mocks three services: account, fs and io system. They use the same port. Default port is 9876.
When you specify -g flag, we use GoogleAccount policy to mock account system. Default is using SimpleAccount.
When you specify -a <User>:<Passwd> flag, we use <User>:<Passwd> as user account. Default is qboxtest:qboxtest123.
When you specify -d <RootDirectory> flag, we use that directory as data path. Default is $HOME/mockQBoxData.
`

func main() {
	fmt.Fprint(os.Stderr, "Usage:\n\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, usageMsg)
	flag.Parse()

	if *root == "" {
		home := os.Getenv("HOME")
		*root = home + "/mockQBoxData"
	}

	fsRoot := *root + "/fs"
	err := os.MkdirAll(fsRoot, 0755)
	if err != nil {
		fmt.Fprintln(os.Stderr, "\nError:", err)
		return
	}

	rsRoot := *root + "/rs"
	err = os.MkdirAll(rsRoot, 0755)
	if err != nil {
		fmt.Fprintln(os.Stderr, "\nError:", err)
		return
	}

	var accountPolicy mockacc.Interface
	if *google {
		fmt.Fprint(os.Stderr, "\nInfo: account policy is using GoogleAccount\n")
	} else {
		sa := mockacc.SimpleAccount{
			"qboxtest":  "qboxtest123",
			"qboxtest1": "qboxtest123",
			"qboxtest2": "qboxtest123",
			"qboxtest3": "qboxtest123",
		}
		if *userAndPass != "" {
			t := strings.Split(*userAndPass, ":", 2)
			if len(t) != 2 {
				fmt.Fprint(os.Stderr, "\nError: -a flag parameter is <User>:<Passwd>\n")
				return
			}
			sa[t[0]] = t[1]
			fmt.Fprintf(os.Stderr, "\nInfo: add user - %s:%s\n", t[0], t[1])
		}
		fmt.Fprintf(os.Stderr, "\nInfo: account policy is using SimpleAccount\n")
		accountPolicy = sa
	}

	fmt.Fprintf(os.Stderr, "\nInfo: root path is %v\n", *root)

	var ioHost string
	if *useTest {
		ioHost = "http://test.qbox.us:"
	} else {
		ioHost = "http://" + *host + ":"
	}

	acc := mockacc.Account{}
	port1 := strconv.Itoa(*port)
	ioCfg := mockio.Config{Root: *root, Account: acc}
	fsCfg := mockfs.Config{Root: fsRoot, IoHost: ioHost + port1, Account: acc}
	rsCfg := mockrs.Config{Root: rsRoot, IoHost: ioHost + port1, Account: acc}

	miCfgStr :=
		`{
	"local": {
		"account": "http://127.0.0.1:9876",
		"fs": "http://127.0.0.1:9876",
		"rs": "http://127.0.0.1:9875",
		"es": "http://127.0.0.1:9874",
		"io": "http://127.0.0.1:9876",
		"ios": ["http://127.0.0.1:9876"]
	},
	"aio_local": {
		"account": "http://127.0.0.1:9876",
		"fs": "http://127.0.0.1:9876",
		"rs": "http://127.0.0.1:9875",
		"es": "http://127.0.0.1:9874",
		"io": "http://127.0.0.1:9876",
		"ios": ["http://127.0.0.1:9876"]
	},
}`
	miCfg := &mi.Config{Settings: nil, Account: acc}
	miCfg.LoadFromString(miCfgStr)

	mux := http.NewServeMux()
	mockacc.RegisterHandlers(mux, accountPolicy)
	mockfs.RegisterHandlers(mux, fsCfg)
	mi.RegisterHandlers(mux, miCfg)
	err = mockio.RegisterHandlers(mux, ioCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError: init fail - %v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "\nInfo: run mock services at %s ...\n", ioHost+port1)

	go mockrs.Run(":"+strconv.Itoa(*port-1), rsCfg)
	http.ListenAndServe(":"+port1, mux)
}
