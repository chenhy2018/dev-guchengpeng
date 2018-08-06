package fs

import (
	"net/http"
	"qbox.us/api/fs"
	"qbox.us/rpc"
	"strconv"
)

// -----------------------------------------------------------

type Info fs.ExtInfo

type Service struct {
	Host string
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{host, rpc.Client{client}}
}

// -----------------------------------------------------------

func (r *Service) Info(uid uint32) (info Info, code int, err error) {

	code, err = r.Conn.CallWithForm(&info, r.Host+"/admin/info", map[string][]string{
		"uid": {strconv.FormatUint(uint64(uid), 10)},
	})
	return
}

// -----------------------------------------------------------

func (r *Service) SetVipSpace(
	userName string, uid uint32, pkey string, pspaceInM int, pexp int64, preason string) (code int, err error) {

	code, err = r.Conn.CallWithForm(nil, r.Host+"/admin/set", map[string][]string{
		"user": {userName},
		"uid":  {strconv.FormatUint(uint64(uid), 10)},
		"pkey": {pkey},
		"pval": {strconv.Itoa(pspaceInM)},
		"pexp": {strconv.FormatInt(pexp, 10)},
		"pmsg": {preason},
	})
	return
}

// -----------------------------------------------------------

func (r *Service) SetPermanentSpace(
	userName string, uid uint32, pkey string, pspaceInM int, preason string) (code int, err error) {

	code, err = r.Conn.CallWithForm(nil, r.Host+"/admin/set", map[string][]string{
		"user": {userName},
		"uid":  {strconv.FormatUint(uint64(uid), 10)},
		"pkey": {pkey},
		"pval": {strconv.Itoa(pspaceInM)},
		"pexp": {expPermanent},
		"pmsg": {preason},
	})
	return
}

const (
	PermanentExp = (1 << 63) - 1
)

var expPermanent = strconv.FormatInt(PermanentExp, 10)

// -----------------------------------------------------------
