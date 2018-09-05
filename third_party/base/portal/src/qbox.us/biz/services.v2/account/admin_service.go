package account

import (
	"net/http"
	"net/url"
	"strconv"

	"qbox.us/oauth"

	"github.com/qiniu/rpc.v1"
	"qbox.us/admin_api/account.v2"

	"qbox.us/biz/utils.v2/log"
)

type AdminAccountService interface {
	FindInfoByEmail(email string) (*Info, error)
	FindInfoByUsername(username string) (*Info, error)
	FindInfoByUid(uid uint32) (*Info, error)
	UserSetUsername(uid uint32, username string) (info *Info, err error)
	ListUsers(offset, limit int) (infos []*Info, err error)
	ListUsersByUtype(utype uint32, offset, limit int) (infos []*Info, err error)
	ListUsersByUids(uids []uint32) (infos []*Info, err error)
	TokenCreate(uid uint32) (*oauth.Token, error)
	UserSetPassword(uin uint32, password string) (*Info, error)
	UserSetEmail(uin uint32, email string) (*Info, error)
	UserSetUserType(userName string, utype uint32) (*Info, error)
	UserCreateByPassword(email, password string) (*Info, error)
	UserDisable(uid uint32, reason string, disabledType DisabledType) (*Info, error)
	UserAutoEnable(uid uint32) (*Info, error)
	UserForceEnable(uid uint32) (*Info, error)
	UserSetCustomerGroup(uid uint32, cg CustomerGroup) (*Info, error)
	UserCreateByVendor(vendor, vendorId, vendorEmail string) (*Info, error)
	UserBindAccount(uid uint32, vendor, vendorId, vendorEmail string) (*Info, error)
	UserUnbindAccount(uid uint32, vendor string) (*Info, error)
	UserUpdate(uid uint32, param url.Values) (info *Info, err error)

	ListUsersByMarker(marker uint32, limit int) (infos []*Info, err error)
	ListChildrenByMarker(parent, marker uint32, limit int) (infos []*Info, err error)
	GetChildrenCount(parent uint32) (count int, err error)
}

type adminAccountService struct {
	logger log.ReqLogger

	service   *account.Service
	rpcLogger rpc.Logger
}

var _ AdminAccountService = new(adminAccountService)

func NewAdminAccountService(host string, adminOAuth http.RoundTripper, logger log.ReqLogger) AdminAccountService {
	acc := new(adminAccountService)
	acc.logger = logger

	acc.rpcLogger = log.NewRpcWrapper(logger)
	acc.service = account.New(host, adminOAuth)
	return acc
}

func (a *adminAccountService) FindInfoByEmail(email string) (info *Info, err error) {
	userInfo, err := a.service.UserInfoById(email, a.rpcLogger)
	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}

func (a *adminAccountService) FindInfoByUsername(username string) (*Info, error) {
	// FindInfoByEmail 可以直接用 username 作为参数
	return a.FindInfoByEmail(username)
}

func (a *adminAccountService) FindInfoByUid(uid uint32) (info *Info, err error) {
	userInfo, err := a.service.UserInfoByUid(uid, a.rpcLogger)
	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}

func (a *adminAccountService) TokenCreate(uid uint32) (token *oauth.Token, err error) {
	tk, err := a.service.TokenCreate(uid, a.rpcLogger)
	if err != nil {
		return
	}
	token = &tk
	return
}

func (a *adminAccountService) UserSetPassword(uid uint32, password string) (info *Info, err error) {
	userInfo, err := a.service.UserSetPassword(uid, password, a.rpcLogger)
	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}

func (a *adminAccountService) UserSetEmail(uid uint32, email string) (info *Info, err error) {
	userInfo, err := a.service.UserSetEmail(uid, email, a.rpcLogger)
	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}

func (a *adminAccountService) UserSetUserType(userName string, utype uint32) (info *Info, err error) {
	userInfo, err := a.service.UserSetUserType(userName, utype, a.rpcLogger)
	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}

func (a *adminAccountService) UserCreateByPassword(email, password string) (info *Info, err error) {
	userInfo, err := a.service.UserCreateByPassword(email, password, a.rpcLogger)
	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}

func (a *adminAccountService) UserDisable(uid uint32, reason string, disabledType DisabledType) (info *Info, err error) {
	dType := account.DisabledType(disabledType)
	userInfo, err := a.service.UserDisable(uid, reason, dType, a.rpcLogger)
	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}

func (a *adminAccountService) UserAutoEnable(uid uint32) (info *Info, err error) {
	userInfo, err := a.service.UserAutoEnable(uid, a.rpcLogger)
	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}

func (a *adminAccountService) UserForceEnable(uid uint32) (info *Info, err error) {
	userInfo, err := a.service.UserForceEnable(uid, a.rpcLogger)
	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}

func (a *adminAccountService) UserSetCustomerGroup(uid uint32, cg CustomerGroup) (info *Info, err error) {
	userInfo, err := a.service.UserSetCustomerGroup(uid, account.CustomerGroup(cg), a.rpcLogger)
	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}

func (a *adminAccountService) UserSetUsername(uid uint32, username string) (info *Info, err error) {
	var userInfo account.Info
	err = a.service.Conn.CallWithForm(a.rpcLogger, &userInfo, a.service.Host+"/admin/user/set_username", map[string][]string{
		"uid":      {strconv.FormatUint(uint64(uid), 10)},
		"username": {username},
	})

	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}

func (a *adminAccountService) UserUpdate(uid uint32, param url.Values) (info *Info, err error) {
	userInfo, err := a.service.UserUpdate(uid, param, a.rpcLogger)

	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}

func (a *adminAccountService) ListUsers(offset, limit int) (infos []*Info, err error) {
	userInfos, err := a.service.ListUsers(offset, limit, a.rpcLogger)
	if err != nil {
		return
	}
	infos = make([]*Info, 0, len(userInfos))
	for _, info := range userInfos {
		infos = append(infos, convertFromAccInfo(info))
	}
	return
}

func (a *adminAccountService) ListUsersByUtype(utype uint32, offset, limit int) (infos []*Info, err error) {
	userInfos, err := a.service.ListUsersByUtype(utype, offset, limit, a.rpcLogger)
	if err != nil {
		return
	}
	infos = make([]*Info, 0, len(userInfos))
	for _, info := range userInfos {
		infos = append(infos, convertFromAccInfo(info))
	}
	return
}

func (a *adminAccountService) ListUsersByUids(uids []uint32) (infos []*Info, err error) {
	userInfos, err := a.service.ListUsersByUids(uids, a.rpcLogger)
	if err != nil {
		return
	}
	infos = make([]*Info, 0, len(userInfos))
	for _, info := range userInfos {
		infos = append(infos, convertFromAccInfo(info))
	}
	return
}

func (a *adminAccountService) ListUsersByMarker(marker uint32, limit int) (infos []*Info, err error) {
	userInfos, err := a.service.ListUsersByMarker(marker, limit, a.rpcLogger)
	if err != nil {
		return
	}
	infos = make([]*Info, 0, len(userInfos))
	for _, info := range userInfos {
		infos = append(infos, convertFromAccInfo(info))
	}
	return
}

func (a *adminAccountService) ListChildrenByMarker(parent, marker uint32, limit int) (infos []*Info, err error) {
	userInfos, err := a.service.ListChildrenByMarker(parent, marker, limit, a.rpcLogger)
	if err != nil {
		return
	}
	infos = make([]*Info, 0, len(userInfos))
	for _, info := range userInfos {
		infos = append(infos, convertFromAccInfo(info))
	}
	return
}

func (a *adminAccountService) GetChildrenCount(parent uint32) (count int, err error) {
	count, err = a.service.CountChildren(parent, a.rpcLogger)
	if err != nil {
		return
	}
	return
}
