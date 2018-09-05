package mockfs

import (
	"fmt"
	"net/http"
	"os"
	"path"
	fss "qbox.us/api/fs"
	"qbox.us/fstest"
	"qbox.us/mockacc"
	"qbox.us/oauth"
	"qbox.us/ts"
	"testing"
	"time"
)

type Root string

func (root Root) CheckDir(dir string, t *testing.T) {
	dir = path.Join(string(root), dir)
	fi, err := os.Lstat(dir)
	if err != nil || !fi.IsDir() {
		ts.Fatal(t, fi, err)
	}
}

func (root Root) IsDeleted(file string) bool {
	file = path.Join(string(root), file)
	return isDeleted(file)
}

func TestMockFS(t *testing.T) {

	home := os.Getenv("HOME")
	root := home + "/mockfsRoot"
	fmt.Println("ROOT:", root)

	os.RemoveAll(root)
	err := os.Mkdir(root, DefaultPerm)
	if err != nil {
		ts.Fatal(t, "Mkdir:", err)
	}

	cfg := Config{Root: root, IoHost: "http://localhost:7888", Account: mockacc.Account{}}
	sa := mockacc.SaInstance
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
		ts.Fatal(t, "ExchangeByPassword:", err)
	}

	fmt.Println(token)

	fs := fss.New("http://localhost:7888", transport)
	root1 := Root(root + "/qboxtest")
	fstest.Do(fs, t, root1)
	os.RemoveAll(root)
}
