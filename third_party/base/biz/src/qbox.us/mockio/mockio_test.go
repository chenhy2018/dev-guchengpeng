// +build ignore

package mockio

import (
	"fmt"
	"net/http"
	"os"
	fss "qbox.us/api/fs"
	//"qbox.us/api/ios"
	"io"
	"io/ioutil"
	"launchpad.net/mgo"
	"log"
	api "qbox.us/api/rs"
	"qbox.us/audit/file"
	"qbox.us/auditlog"
	"qbox.us/mockacc"
	"qbox.us/mockfs"
	"qbox.us/multipart"
	"qbox.us/oauth"
	"qbox.us/rfs"
	"qbox.us/rsplus"
	"qbox.us/servend/account"
	"github.com/qiniu/ts"
	"testing"
	"time"
	stime "time"
)

type M map[string]interface{}

func TestMockIO1(t *testing.T) {

	home := os.Getenv("HOME")
	rootIo := home + "/ioMockRoot"
	rootFs := home + "/fsMockRoot"
	fmt.Println("ROOT:", rootIo, rootFs)

	os.RemoveAll(rootIo)
	os.RemoveAll(rootFs)
	err := os.Mkdir(rootIo, DefaultPerm)
	if err != nil {
		t.Fatal("Mkdir:", err)
	}
	err = os.Mkdir(rootFs, DefaultPerm)
	if err != nil {
		t.Fatal("Mkdir:", err)
	}

	cfgIo := Config{Root: rootIo, Account: mockacc.Account{}}
	cfgFs := mockfs.Config{Root: rootFs, IoHost: "http://localhost:7779", Account: mockacc.Account{}}
	sa := &mockacc.SingleAccount{"qboxtest", "qboxtest123"}
	muxFs := http.NewServeMux()
	mockacc.RegisterHandlers(muxFs, sa)
	mockfs.RegisterHandlers(muxFs, cfgFs)
	go http.ListenAndServe(":7778", muxFs)
	go Run(":7779", cfgIo)
	time.Sleep(1e9)

	cfg1 := &oauth.Config{
		ClientId:     "<ClientId>",
		ClientSecret: "<ClientSecret>",
		Scope:        "<Scope>",
		AuthURL:      "<AuthURL>",
		TokenURL:     "http://localhost:7778/oauth2/token",
		RedirectURL:  "<RedirectURL>",
	}

	transport := &oauth.Transport{Config: cfg1}

	token, _, err := transport.ExchangeByPassword("qboxtest", "qboxtest123")
	if err != nil {
		t.Fatal("ExchangeByPassword:", err)
	}

	fmt.Println(token)

	fs := fss.New("http://localhost:7778", transport)
	{
		code, err := fs.Init()
		if err != nil || code != 200 {
			t.Fatal("Init:", err, code)
		}

		fmt.Println("Init OK!")
	}

	//io := ios.New("http://localhost:7779", transport)

	session, err := mgo.Dial("localhost")
	if err != nil {
		ts.Fatal(t, err)
	}
	defer session.Close()

	c := session.DB("qbox_rsplusTest").C("rsplus")
	c.RemoveAll(M{})

	filesystem, err := rfs.New(c)
	if err != nil {
		ts.Log(t, err)
	}
	acc1 := mockacc.Account{}
	cfgRS := &rsplus.Config{
		Fs:      filesystem,
		IoHost:  "http://localhost:7779",
		Account: acc1,
		Config: auditlog.Config{
			Logger: file.Stderr,
		},
	}
	muxRS := http.NewServeMux()
	mockacc.UserType = account.USER_TYPE_ENTERPRISE_VUSER
	mockacc.RegisterHandlers(muxRS, sa)
	rsplus.RegisterHandlers(muxRS, cfgRS)
	go http.ListenAndServe(":7780", muxRS)
	stime.Sleep(1e9)

	rs := api.New("http://localhost:7780", transport)
	url := cfgRS.IoHost + "/put-auth/"
	var ret map[string]interface{}
	rs.Conn.Call(&ret, url)
	log.Println(ret)

	file := "ioTest.tmp"
	ioutil.WriteFile(file, []byte("file content"), 0777)
	defer os.Remove(file)
	r, ct, err := multipart.Open(map[string][]string{
		"action": {"/rs-put/!" + file},
		"file":   {"@" + file},
	})
	if err != nil {
		ts.Fatal(t, "Open failed:", err)
	}
	defer r.Close()

	fmt.Println("\nContent-Type:", ct)

	req, err := http.NewRequest("POST", ret["url"].(string), r)
	if err != nil {
		ts.Fatal(t, "NewRequest failed:", err)
	}

	req.Header.Set("Content-Type", ct)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ts.Fatal(t, "http.Client.Do failed:", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		ts.Fatal(t, "http.Client.Do status code:", resp.StatusCode)
	}

	io.Copy(os.Stdout, resp.Body)

	os.RemoveAll(rootIo)
	os.RemoveAll(rootFs)
}
