package slave

import (
	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
	"net/http"
)

// ------------------------------------------------------------------------

type masterNotifier struct {
	mgrc       *http.Client
	masterHost string
}

func NewAddNotifier(
	mgrAccessKey, mgrSecretKey string, masterHost string) *masterNotifier {

	mac := &digest.Mac{
		AccessKey: mgrAccessKey, // 向 master 发送指令时的帐号
		SecretKey: []byte(mgrSecretKey),
	}
	mgrc := digest.NewClient(mac, nil)
	return &masterNotifier{
		mgrc: mgrc, masterHost: masterHost,
	}
}

func (p *masterNotifier) AddNotify(l rpc.Logger, id string, kvs []string) {

	mgrc := rpc.Client{p.mgrc}

	err := mgrc.CallWithForm(l, nil, p.masterHost+"/adds", map[string][]string{
		"c":  {id},
		"kv": kvs,
	})
	if err != nil {
		log.Warn("slave.AddNotify failed:", p.masterHost, err)
	}
}

// ------------------------------------------------------------------------

type nilAddNotifier struct{}

func (r nilAddNotifier) AddNotify(l rpc.Logger, id string, kvs []string) {}

var NilAddNotifier nilAddNotifier

// ------------------------------------------------------------------------
