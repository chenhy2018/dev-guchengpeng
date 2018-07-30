package account

func (a *accountService) CreateChild(email, password string) (uinfo *UserInfo, err error) {
	accUserInfo, err := a.service.UserCreateChild(email, password, a.rpcLogger)
	if err != nil {
		return
	}

	uinfo = convertFromAccUserInfo(accUserInfo)
	return
}

func (a *accountService) DisableChild(uid uint32, reason string) (uinfo *UserInfo, err error) {
	accUserInfo, err := a.service.UserDisableChild(uid, reason, a.rpcLogger)
	if err != nil {
		return
	}

	uinfo = convertFromAccUserInfo(accUserInfo)
	return
}

func (a *accountService) EnableChild(uid uint32) (uinfo *UserInfo, err error) {
	accUserInfo, err := a.service.UserEnableChild(uid, a.rpcLogger)
	if err != nil {
		return
	}

	uinfo = convertFromAccUserInfo(accUserInfo)
	return
}
