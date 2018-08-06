package mockrs

import (
	"encoding/base64"
	"fmt"
	"http"
	"os"
	"qbox.us/api/rs"
	"qbox.us/mockacc"
	"qbox.us/oauth"
	"qbox.us/rpc"
	"qbox.us/ts"
	"testing"
	"time"
)

type putRet struct {
	Hash string "hash"
}

// POST /put/<EncodedEntryURI>/base/<BaseHash>/editTime/<EditTime>/fsize/<FileSize>

func doTestPut(fs *rs.Service, t *testing.T, root string) {
	{
		file := "/t.txt"
		url := fs.Host + "/put/" + rpc.EncodeURI(file) + "/fsize/67"
		fh := []byte("<FileHandle>")

		var ret putRet
		code, err := fs.Conn.CallWithParam(&ret, url, "application/octet-stream", fh)
		if err != nil || code != 200 {
			ts.Fatal(t, "Put:", file, err, code)
		}
		fmt.Println("Put ret:", ret)
		if ret.Hash != base64.URLEncoding.EncodeToString(fh) {
			ts.Fatal(t, "Put fail:", file, ret)
		}

		getRet, code, err := fs.Get(file)
		if err != nil || code != 200 {
			ts.Fatal(t, "Get:", file, err, code)
		}
		fmt.Println("Get ret:", getRet)
		if getRet.Fsize != 67 || getRet.Hash != base64.URLEncoding.EncodeToString(fh) {
			ts.Fatal(t, "Get fail:", file, getRet)
		}
	}
}

func isDeleted(path string) bool {
	_, err := os.Lstat(path)
	return err != nil
}

func doTestDelete(fs *rs.Service, t *testing.T, root string) {
}

func doTestMove(fs *rs.Service, t *testing.T, root string) {
}

func doTestBatch(fs *rs.Service, t *testing.T, root string) {
}

func TestMockRS(t *testing.T) {

	home := os.Getenv("HOME")
	root := home + "/mockrsRoot"
	fmt.Println("ROOT:", root)

	os.RemoveAll(root)
	err := os.Mkdir(root, DefaultPerm)
	if err != nil {
		t.Fatal("Mkdir:", err)
	}

	cfg := Config{Root: root, IoHost: "http://localhost:7888", Account: mockacc.Account{}}
	sa := &mockacc.SingleAccount{"qboxtest", "qboxtest123"}
	mockacc.RegisterHandlers(http.DefaultServeMux, sa)
	go Run(":7888", cfg)
	time.Sleep(1e9)

	cfg1 := &oauth.Config{
		ClientId:     "<ClientId>",
		ClientSecret: "<ClientSecret>",
		Scope:        "<Scope>",
		AuthURL:      "<AuthURL>",
		TokenURL:     "http://localhost:7888/oauth2/token",
		RedirectURL:  "<RedirectURL>",
	}

	transport := &oauth.Transport{Config: cfg1}

	token, _, err := transport.ExchangeByPassword("qboxtest", "qboxtest123")
	if err != nil {
		t.Fatal("ExchangeByPassword:", err)
	}

	fmt.Println(token)

	fs := rs.New("http://localhost:7888", transport)
	{
		code, err := fs.Init()
		if err != nil || code != 200 {
			t.Fatal("Init:", err, code)
		}
		fmt.Println("Init OK!")
	}
	root1 := root + "/qboxtest"
	doTestPut(fs, t, root1)

	os.RemoveAll(root)
}
