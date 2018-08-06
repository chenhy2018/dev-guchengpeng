package master

import (
	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
	"net/http"
)

// ------------------------------------------------------------------------

type slavesNotifier struct {
	mgrc       *http.Client
	slaveHosts []string // [idc1_slave, idc2_slave, ...]
}

func NewFlipsNotifier(
	mgrAccessKey, mgrSecretKey string, slaveHosts []string) *slavesNotifier {

	mac := &digest.Mac{
		AccessKey: mgrAccessKey, // 向 slave 发送指令时的帐号
		SecretKey: []byte(mgrSecretKey),
	}
	mgrc := digest.NewClient(mac, nil)
	return &slavesNotifier{
		mgrc: mgrc, slaveHosts: slaveHosts,
	}
}

var stringSliceOne = []string{"1"}

func (p *slavesNotifier) FlipsNotify(grpIds []string, clear bool) {

	mgrc := rpc.Client{p.mgrc}

	for _, host := range p.slaveHosts {
		params := map[string][]string{
			"c": grpIds,
		}
		if clear {
			params["clear"] = stringSliceOne
		}
		err := mgrc.CallWithForm(nil, nil, host+"/flips", params)
		if err != nil {
			log.Warn("master.FlipsNotify failed:", host, err)
		}
	}
}

// ------------------------------------------------------------------------

type nilFlipsNotifier struct{}

func (r nilFlipsNotifier) FlipsNotify(grpIds []string, clear bool) {}

var NilFlipsNotifier nilFlipsNotifier

// ------------------------------------------------------------------------
