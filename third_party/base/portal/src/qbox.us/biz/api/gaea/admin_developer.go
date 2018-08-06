package gaea

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/qiniu/rpc.v1"
)

type AdminDeveloperService struct {
	host   string
	client rpc.Client
}

func NewAdminDeveloperService(host string, t http.RoundTripper) *AdminDeveloperService {
	return &AdminDeveloperService{
		host: host,
		client: rpc.Client{
			&http.Client{Transport: t},
		},
	}
}

// AdminDeveloperService.InfoByUid provides detailed developer info from `db.developers` table, queried by UID.
// Some fields are omitted for security reasons.
func (s *AdminDeveloperService) InfoByUid(l rpc.Logger, uid uint32) (info *DeveloperInfo, err error) {
	var resp developerInfoResp

	err = s.client.GetCall(l, &resp, s.host+"/admin/developer/by-uid/"+strconv.FormatUint(uint64(uid), 10))
	if err != nil {
		return
	}

	if resp.Code != CodeOK {
		err = errors.New(fmt.Sprintf("AdminDeveloperService::InfoByUid() failed with code: %d", resp.Code))
		return
	}

	info = resp.Info

	return
}

// AdminDeveloperService.InfoByEmail provides detailed developer info from `db.developers` table, queried by E-mail.
// Some fields are omitted for security reasons.
func (s *AdminDeveloperService) InfoByEmail(l rpc.Logger, email string) (info *DeveloperInfo, err error) {
	var resp developerInfoResp

	err = s.client.GetCall(l, &resp, s.host+"/admin/developer/by-email/"+email)
	if err != nil {
		return
	}

	if resp.Code != CodeOK {
		err = errors.New(fmt.Sprintf("AdminDeveloperService::InfoByEmail() failed with code: %d", resp.Code))
		return
	}

	info = resp.Info

	return
}
