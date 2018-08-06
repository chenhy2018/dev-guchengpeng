package configs

import (
	"strings"

	"qbox.us/net/httputil"

	"github.com/qiniu/http/rpcutil.v1"
)

const (
	UidGPrefix = "uidg"
	AkSkPrefix = "ak"
)

var (
	ErrInvalidIdPrefix = httputil.NewError(400, "invalid id prefix")
)

type getbArg struct {
	Id string `json:"id"`
}

type GetB struct {
}

func (p *GetB) WbrpcGetb(args *getbArg, env *rpcutil.Env) (doc map[string]interface{}, err error) {

	var id = args.Id
	var prefix, key string

	if idx := strings.Index(id, ":"); idx != -1 {
		prefix, key = id[:idx], id[idx+1:]
	} else {
		return nil, ErrInvalidIdPrefix
	}

	switch prefix {
	case AkSkPrefix:
		doc, err = p.accessInfoGetb(key)
	case UidGPrefix:
		doc, err = p.utypeInfoGetb(key)
	default:
		err = ErrInvalidIdPrefix
		return
	}
	return
}

func (p *GetB) accessInfoGetb(key string) (doc map[string]interface{}, err error) {
	info, err := Repo.AccessRepo(key)
	if err != nil {
		httputil.NewError(401, err.Error())
		return
	}

	doc = map[string]interface{}{
		"uid":    0,
		"appId":  0,
		"secret": info.SecretKey,
	}
	return
}

func (p *GetB) utypeInfoGetb(key string) (doc map[string]interface{}, err error) {
	doc = map[string]interface{}{
		"utype": 0x4,
	}
	return
}
