package qmsetapi

import (
	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/rpc.v1"
)

// ------------------------------------------------------------------------

type Config struct {
	Host      string `json:"host"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

type Client struct {
	Host string
	Conn rpc.Client
}

func New(cfg *Config) Client {

	mac := &digest.Mac{
		AccessKey: cfg.AccessKey,
		SecretKey: []byte(cfg.SecretKey),
	}
	mgrc := digest.NewClient(mac, nil)

	return Client{
		Host: cfg.Host,
		Conn: rpc.Client{mgrc},
	}
}

// ------------------------------------------------------------------------

func (p Client) Adds(l rpc.Logger, grp string, kvs []string) (err error) {

	return p.Conn.CallWithForm(l, nil, p.Host+"/adds", map[string][]string{
		"c":  {grp},
		"kv": kvs,
	})
}

func (p Client) Get(l rpc.Logger, grp string, key string) (values []string, err error) {

	err = p.Conn.CallWithForm(l, &values, p.Host+"/get", map[string][]string{
		"c": {grp},
		"k": {key},
	})
	return
}

// ------------------------------------------------------------------------

func (p Client) Badd(l rpc.Logger, grp string, encodedVals []string) (err error) {

	return p.Conn.CallWithForm(l, nil, p.Host+"/badd", map[string][]string{
		"c": {grp},
		"v": encodedVals,
	})
}

func (p Client) Bchk(l rpc.Logger, grp string, encodedVals []string) (idxExists []int, err error) {

	err = p.Conn.CallWithForm(l, &idxExists, p.Host+"/bchk", map[string][]string{
		"c": {grp},
		"v": encodedVals,
	})
	return
}

// ------------------------------------------------------------------------
