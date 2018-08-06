package account

import "qbox.us/api/account.v2"

func convertFromAccUserInfo(accUserInfo account.UserInfo) *UserInfo {
	utype := UserType(accUserInfo.UserType)
	userInfo := &UserInfo{
		Uid:                   accUserInfo.Uid,
		Username:              accUserInfo.Username,
		Email:                 accUserInfo.Email,
		UserType:              utype,
		ParentUid:             accUserInfo.ParentUid,
		IsActivated:           accUserInfo.IsActivated,
		IsDisabled:            utype.IsDisabled(),
		LastParentOperationAt: accUserInfo.LastParentOperationAt,
	}
	return userInfo
}
