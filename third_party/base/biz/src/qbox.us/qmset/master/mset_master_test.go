package master

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/qiniu/ts"
	"qbox.us/http/account.v1/acctest"
	"qbox.us/mockacc"
	"qbox.us/qmset/master/mbloom"
	. "qbox.us/qmset/proto"
	"testing"
)

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

	cfg := &Config{
		FlipsNotifier: NilFlipsNotifier, // 向 slaves 发送通知
		FlipCfgs:      flipCfgs,
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

	avals, err := p.WspGet(&getArgs{"aaa", "a"}, env)
	fmt.Println("WpsGet a:", avals, err)
	if err != nil || len(avals) != 2 || avals[0] != "1" || avals[1] != "hello" {
		ts.Fatal(t, "qmset/master.WpsGet failed:", avals, err)
	}

	err = p.WspAdds(&addsArgs{"aaa", []string{"a:hello", "a:12", "a:world", "b:32"}}, env)
	if err != nil {
		ts.Fatal(t, "qmset/master.WspAdds failed:", err)
	}

	avals, err = p.WspGet(&getArgs{"aaa", "a"}, env)
	fmt.Println("WpsGet a:", avals, err)
	if err != nil || len(avals) != 3 || avals[0] != "1" || avals[1] != "hello" || avals[2] != "12" {
		ts.Fatal(t, "qmset/master.WpsGet failed:", avals, err)
	}

	bvals, err := p.WspGet(&getArgs{"aaa", "b"}, env)
	fmt.Println("WpsGet b:", bvals, err)
	if err != nil || len(bvals) != 1 || bvals[0] != "32" {
		ts.Fatal(t, "qmset/master.WpsGet failed:", avals, err)
	}

	p.grps["aaa"].Flip()

	avals, err = p.WspGet(&getArgs{"aaa", "a"}, env)
	fmt.Println("WpsGet a:", avals, err)
	if err != nil || len(avals) != 3 || avals[0] != "1" || avals[1] != "hello" || avals[2] != "12" {
		ts.Fatal(t, "qmset/master.WpsGet failed:", avals, err)
	}

	bvals, err = p.WspGet(&getArgs{"aaa", "b"}, env)
	fmt.Println("WpsGet b:", bvals, err)
	if err != nil || len(bvals) != 1 || bvals[0] != "32" {
		ts.Fatal(t, "qmset/master.WpsGet failed:", avals, err)
	}

	p.grps["aaa"].Flip()

	avals, err = p.WspGet(&getArgs{"aaa", "a"}, env)
	fmt.Println("WpsGet a:", avals, err)
	if err != nil || len(avals) != 0 {
		ts.Fatal(t, "qmset/master.WpsGet failed:", avals, err)
	}

	bvals, err = p.WspGet(&getArgs{"aaa", "b"}, env)
	fmt.Println("WpsGet b:", bvals, err)
	if err != nil || len(bvals) != 0 {
		ts.Fatal(t, "qmset/master.WpsGet failed:", avals, err)
	}
}

// ------------------------------------------------------------------------

func mbloomKeys(vs ...string) []string {

	h := md5.New()
	keys := make([]string, len(vs))
	for i, v := range vs {
		h.Reset()
		h.Write([]byte(v))
		key := base64.URLEncoding.EncodeToString(h.Sum(nil))[:22]
		keys[i] = key
	}
	return keys
}

func TestMbloom(t *testing.T) {

	flipCfgs := []*FlipConfig{
		&FlipConfig{
			Mblooms: []MbloomGrpCfg{
				{Id: "aaa", Max: 5, Fp: 1e-4},
			},
			Expires: 100,
		},
	}

	cfg := &Config{
		FlipsNotifier: NilFlipsNotifier, // 向 slaves 发送通知
		FlipCfgs:      flipCfgs,
	}
	cfg.AuthParser = mockacc.Parser

	p, err := New(cfg)
	if err != nil {
		ts.Fatal(t, "qmset/master.New failed:", err)
	}

	env := acctest.NewAdminEnv(nil)

	err = p.WspBadd(&mbloom.BaddArgs{"aaa", mbloomKeys("1", "hello")}, env)
	if err != nil {
		ts.Fatal(t, "qmset/master.WspBadd failed:", err)
	}

	idxs, err := p.WspBchk(&mbloom.BchkArgs{"aaa", mbloomKeys("1", "2", "3", "hello")}, env)
	fmt.Println("WspBchk:", idxs, err)
	if err != nil || len(idxs) != 2 || idxs[0] != 0 || idxs[1] != 3 {
		ts.Fatal(t, "qmset/master.WspBchk failed:", idxs, err)
	}

	err = p.WspBadd(&mbloom.BaddArgs{"aaa", mbloomKeys("hello", "12", "world", "32")}, env)
	if err != nil {
		ts.Fatal(t, "qmset/master.WspBadd failed:", err)
	}

	idxs, err = p.WspBchk(&mbloom.BchkArgs{"aaa", mbloomKeys("1", "12", "3", "hello")}, env)
	fmt.Println("WspBchk a:", idxs, err)
	if err != nil || len(idxs) != 3 || idxs[0] != 0 || idxs[1] != 1 || idxs[2] != 3 {
		ts.Fatal(t, "qmset/master.WspBchk failed:", idxs, err)
	}

	p.grps["aaa"].Flip()

	idxs, err = p.WspBchk(&mbloom.BchkArgs{"aaa", mbloomKeys("1", "12", "3", "hello")}, env)
	fmt.Println("WspBchk a:", idxs, err)
	if err != nil || len(idxs) != 3 || idxs[0] != 0 || idxs[1] != 1 || idxs[2] != 3 {
		ts.Fatal(t, "qmset/master.WspBchk failed:", idxs, err)
	}

	p.grps["aaa"].Flip()

	idxs, err = p.WspBchk(&mbloom.BchkArgs{"aaa", mbloomKeys("1", "12", "3", "hello")}, env)
	fmt.Println("WspBchk a:", idxs, err)
	if err != nil || len(idxs) != 0 {
		ts.Fatal(t, "qmset/master.WspBchk failed:", idxs, err)
	}
}

// ------------------------------------------------------------------------
