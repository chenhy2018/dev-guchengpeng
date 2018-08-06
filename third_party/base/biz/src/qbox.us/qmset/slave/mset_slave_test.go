package slave

import (
	"fmt"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/ts"
	"qbox.us/http/account.v1/acctest"
	"qbox.us/mockacc"
	. "qbox.us/qmset/proto"
	"testing"
)

// ------------------------------------------------------------------------

type notifyItem struct {
	id  string
	kvs []string
}

type addNotifier struct {
	items []notifyItem
}

func notEq(a, b []string) bool {
	if len(a) != len(b) {
		return true
	}
	for i, v := range a {
		if b[i] != v {
			return true
		}
	}
	return false
}

func (p *addNotifier) itemsVerify(t *testing.T, id string, kvs []string) {
	if kvs == nil {
		if len(p.items) != 0 {
			ts.Fatal(t, "itemsVerify:", p.items, id, kvs)
		}
	} else if len(p.items) != 1 || p.items[0].id != id || notEq(p.items[0].kvs, kvs) {
		ts.Fatal(t, "itemsVerify:", p.items, id, kvs)
	}
	p.items = nil
}

func (p *addNotifier) AddNotify(l rpc.Logger, id string, kvs []string) {
	p.items = append(p.items, notifyItem{id, kvs})
}

// ------------------------------------------------------------------------

func TestMset(t *testing.T) {

	flipCfgs := []*FlipConfig{
		&FlipConfig{
			Msets: []MsetGrpCfg{
				{Id: "aaa", Max: 3},
			},
			Expires: 100,
		},
	}

	an := new(addNotifier)
	cfg := &Config{
		AddNotifier: an, // 向 master 发送通知
		FlipCfgs:    flipCfgs,
	}
	cfg.AuthParser = mockacc.Parser

	p, err := New(cfg)
	if err != nil {
		ts.Fatal(t, "qmset/master.New failed:", err)
	}

	env := acctest.NewAdminEnv(nil)

	err = p.WspAdds(&addsArgs{"aaa", []string{"a:1", "a:hello"}}, env)
	if err != nil {
		ts.Fatal(t, "qmset/master.WspAdds failed:", err)
	}
	fmt.Println(an.items)
	an.itemsVerify(t, "aaa", []string{"a:1", "a:hello"})

	err = p.WspAdds(&addsArgs{"aaa", []string{"a:hello", "a:12", "a:world", "b:32"}}, env)
	if err != nil {
		ts.Fatal(t, "qmset/master.WspAdds failed:", err)
	}
	fmt.Println(an.items)
	an.itemsVerify(t, "aaa", []string{"a:12", "b:32"})
}

// ------------------------------------------------------------------------
