package logh

import (
	"net/http"
	"qbox.us/audit/logh"
	"qbox.us/servend/account"
)

// ------------------------------------------------------------------------------------------
// [[DEPRECATED]] 这个包已经迁移到 qbox.us/http/audit/jsonlog

type Decoder struct {
	logh.BaseDecoder
	Account account.Interface
}

func (r Decoder) DecodeRequest(req *http.Request) (url string, header, params logh.M) {

	url, header, params = r.BaseDecoder.DecodeRequest(req)
	user, err := account.GetAuth(r.Account, req)
	if err != nil {
		return
	}

	header["Token"] = logh.M{
		"uid":     user.Uid,
		"utype":   user.Utype,
		"appid":   user.Appid,
		"devid":   user.Devid,
		"sudoer":  user.Sudoer,
		"utypesu": user.UtypeSu,
	}
	return
}

// ------------------------------------------------------------------------------------------
// [[DEPRECATED]] 这个包已经迁移到 qbox.us/http/audit/jsonlog

type ExtDecoder struct {
	logh.BaseDecoder
	Account account.InterfaceEx
}

func (r ExtDecoder) DecodeRequest(req *http.Request) (url string, header, params logh.M) {

	url, header, params = r.BaseDecoder.DecodeRequest(req)
	user, err := account.GetAuthExt(r.Account, req)
	if err != nil {
		return
	}

	token := logh.M{
		"uid":   user.Uid,
		"utype": user.Utype,
	}
	if user.UtypeSu != 0 {
		token["sudoer"] = user.Sudoer
		token["utypesu"] = user.UtypeSu
	}
	if user.Appid != 0 {
		token["appid"] = user.Appid
	}
	if user.Devid != 0 {
		token["devid"] = user.Devid
	}
	header["Token"] = token
	return
}

// ------------------------------------------------------------------------------------------
