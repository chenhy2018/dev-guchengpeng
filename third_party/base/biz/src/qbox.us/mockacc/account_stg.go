package mockacc

import (
	"hash/adler32"

	"qbox.us/cc/config"
	"qbox.us/errors"
	"qbox.us/servend/account"
)

// ---------------------------------------------------------------------------------------

type UserInfo struct {
	Id        string `json:"user"`
	Passwd    string `json:"passwd"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Uid       uint32 `json:"uid"`
	Utype     uint32 `json:"utype"`
	Appid     uint32 `json:"appid"`
}

type SimpleAccount []UserInfo

func (sa SimpleAccount) InfoById(user string) (ui account.UserInfo, passwd string, err error) {

	for _, v := range sa {
		if v.Id == user {
			return account.UserInfo{Uid: v.Uid, Utype: v.Utype, Appid: v.Appid}, v.Passwd, nil
		}
	}
	err = errors.New("Invalid account")
	return
}

func (sa SimpleAccount) InfoByUid(uid uint32) (ui account.UserInfo, passwd string, err error) {
	for _, v := range sa {
		if v.Uid == uid {
			return account.UserInfo{Uid: v.Uid, Utype: v.Utype, Appid: v.Appid}, v.Passwd, nil
		}
	}
	err = errors.New("Invalid account")
	return
}

// ---------------------------------------------------------------------------------------

func GetUid(user string) uint32 {
	return adler32.Checksum([]byte(user))
}

var SaInstance = SimpleAccount{
	UserInfo{
		Id:        "root",
		Uid:       GetUid("root"), // 74121669
		Passwd:    "root",
		AccessKey: "4_odedBxmrAHiu4Y0Qp0HPG0NANCf6VAsAjWL_k9",
		SecretKey: "SrRuUVfDX6drVRvpyN8mv8Vcm9XnMZzlbDfvVfMe",
		Utype:     account.USER_TYPE_ADMIN | account.USER_TYPE_ENTERPRISE,
		Appid:     1,
	},
	UserInfo{
		Id:        "qboxtest",
		Uid:       GetUid("qboxtest"), // 260637563
		Passwd:    "qboxtest123",
		AccessKey: "PjFtQJWfvKrSLYkSlV-keCKWzmXzSK1Zp3R9S5MV",
		SecretKey: "Q48lAPnTPVxq20dUmfux9HVCrtC3h-p3MCTgMyXf",
		Utype:     account.USER_TYPE_ENTERPRISE,
		Appid:     2,
	},
}

func GetSa() (sa SimpleAccount) {

	err := config.LoadEx(&sa, "mockacc.conf")
	if err != nil {
		sa = SaInstance
	}
	return
}

// ---------------------------------------------------------------------------------------
