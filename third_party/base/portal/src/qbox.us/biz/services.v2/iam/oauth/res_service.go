package oauth

import (
	"fmt"
	"net/http"

	"qbox.us/iam/api"

	"github.com/qiniu/rpc.v1"

	"qbox.us/iam/entity"
)

// ResService 使用 IAM OAuth 鉴权，获取与当前用户相关的资源
type ResService interface {
	Profile() (user *entity.User, err error)
	Keypairs() (keypairs []entity.Keypair, err error)
}

type resService struct {
	host   string
	client rpc.Client
	l      rpc.Logger
}

var _ ResService = &resService{}

// NewResService returns a implementation of ResService.
func NewResService(host string, userOAuth *Transport, l rpc.Logger) ResService {
	return &resService{
		host: host,
		client: rpc.Client{
			Client: &http.Client{Transport: userOAuth},
		},
		l: l,
	}
}

func (s *resService) Profile() (user *entity.User, err error) {
	var output struct {
		api.CommonResponse
		Data *entity.User `json:"data"`
	}
	err = s.client.GetCall(s.l, &output, fmt.Sprintf("%s/iam/api/profile", s.host))
	if err != nil {
		return
	}

	user = output.Data
	return
}

func (s *resService) Keypairs() (keypairs []entity.Keypair, err error) {
	var output struct {
		api.CommonResponse
		Data []entity.Keypair `json:"data"`
	}
	err = s.client.GetCall(s.l, &output, fmt.Sprintf("%s/iam/api/keypairs", s.host))
	if err != nil {
		return
	}

	keypairs = output.Data
	return
}
