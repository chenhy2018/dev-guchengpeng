package account

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"labix.org/v2/mgo/bson"
	"qbox.us/api/account"
	"qbox.us/oauth"
)

const (
	HOST                    = "http://222.73.201.152:9100"
	MOCK_INIT_USER_ID       = "init@qbox.net"
	MOCK_INIT_USER_PASSWORD = "qboxtest123"
	MOCK_CLIENT_ID          = "abcd0c7edcdf914228ed8aa7c6cee2f2bc6155e2"
	MOCK_CLIENT_SECRET      = "fc9ef8b171a74e197b17f85ba23799860ddf3b9c"
)

func TestBsonForVendor(t *testing.T) {
	v := Vendor{
		Vendor:      "a",
		VendorId:    "b",
		VendorEmail: "c",
	}
	bytes, err := bson.Marshal(v)
	assert.NoError(t, err)
	v2 := Vendor{}
	err = bson.Unmarshal(bytes, &v2)
	assert.NoError(t, err)
	assert.Equal(t, v, v2)
}

func TestBsonForInfo(t *testing.T) {
	v := Info{
		Id:        "a",
		Email:     "b",
		CreatedAt: 1,
		Vendors:   []Vendor{},
	}
	bytes, err := bson.Marshal(v)
	assert.NoError(t, err)
	v2 := Info{}
	err = bson.Unmarshal(bytes, &v2)
	assert.NoError(t, err)
	assert.Equal(t, v, v2)
}

func TestInfo_UserInfo(t *testing.T) {
	assert.Equal(t, Info{}.UserInfo(), account.UserInfo{})
}

func TestInfo_Disabled(t *testing.T) {
	i := Info{Utype: 4}
	assert.False(t, i.IsDisabled())
	i.Disable()
	assert.True(t, i.IsDisabled())
	i.Disable()
	assert.True(t, i.IsDisabled())
	i.Enable()
	assert.False(t, i.IsDisabled())
	i.Enable()
	assert.False(t, i.IsDisabled())
	assert.Equal(t, i.Utype, uint32(4))
}

func TestInfo_CustomGroup(t *testing.T) {
	assert.Equal(t, Info{}.GetCustomerGroup(), CUSTOMER_GROUP_INVALID)
	assert.Equal(t, Info{Utype: 4}.GetCustomerGroup(), CUSTOMER_GROUP_NORMAL)
	assert.Equal(t, Info{Utype: 8}.GetCustomerGroup(), CUSTOMER_GROUP_NORMAL)
	assert.Equal(t, Info{Utype: 12}.GetCustomerGroup(), CUSTOMER_GROUP_NORMAL)
	assert.Equal(t, Info{Utype: 10}.GetCustomerGroup(), CUSTOMER_GROUP_VIP)
	assert.Equal(t, Info{Utype: 6}.GetCustomerGroup(), CUSTOMER_GROUP_VIP)
	assert.Equal(t, Info{Utype: 20}.GetCustomerGroup(), CUSTOMER_GROUP_EXP)
	assert.Equal(t, Info{Utype: 22}.GetCustomerGroup(), CUSTOMER_GROUP_EXP)
	assert.Equal(t, Info{Utype: 18}.GetCustomerGroup(), CUSTOMER_GROUP_EXP)
}

func getService(t *testing.T) *Service {
	a, err := NewService(
		HOST,
		MOCK_CLIENT_ID,
		MOCK_CLIENT_SECRET,
		MOCK_INIT_USER_ID,
		MOCK_INIT_USER_PASSWORD)
	assert.NoError(t, err)
	return a

}

func createUser(t *testing.T, l *xlog.Logger) (info Info, password string) {
	name := bson.NewObjectId().Hex()
	email := strings.ToUpper(name + "@example.com")
	id := name + "@example.com"
	password = name

	info, err := getService(t).UserCreateByPassword(email, password, l)
	assert.NoError(t, err)
	assert.Equal(t, info.Id, id)
	return
}

func verifyPassword(email, password string) (err error) {
	transport := &oauth.Transport{
		Config: &oauth.Config{
			ClientId:     MOCK_CLIENT_ID,
			ClientSecret: MOCK_CLIENT_SECRET,
			Scope:        "Scope",
			AuthURL:      "",
			TokenURL:     HOST + "/oauth2/token",
			RedirectURL:  "",
		},
		Transport: http.DefaultTransport, // it is default
	}
	_, _, err = transport.ExchangeByPassword(email, password)
	return
}

func TestSDK_UserSetUserType(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	a := getService(t)
	l := xlog.NewDummy()

	user, _ := createUser(t, l)
	user2, err := a.UserSetUserType(user.Id, 20, l)
	assert.NoError(t, err)
	assert.Equal(t, user2.Uid, user.Uid)
	assert.Equal(t, user2.Utype, uint32(0x14))

	user3, err := a.UserInfoByUid(user.Uid, l)
	assert.NoError(t, err)
	assert.Equal(t, user3.Utype, uint32(0x14))
}

func TestSDK_ListUsersByLastLogintTime(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	a := getService(t)
	l := xlog.NewDummy()

	res, err := a.ListUsersByLastLoginTime(time.Now().Add(-time.Hour*24*14), time.Now().Add(-time.Hour*24*7), 0, 20, l)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestSDK_UserSetPasssword(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	a := getService(t)
	l := xlog.NewDummy()

	user, password := createUser(t, l)
	assert.NoError(t, verifyPassword(user.Email, password))
	user2, err := a.UserSetPassword(user.Uid, "new_password", l)
	assert.NoError(t, err)
	assert.Equal(t, user2.Uid, user.Uid)

	assert.NoError(t, verifyPassword(user.Email, "new_password"))
	err = verifyPassword(user.Email, password)
	assert.Equal(t, err.Error(), "failed_authentication")
}

func TestSDK_UserSetEmail(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	a := getService(t)
	l := xlog.NewDummy()

	user, password := createUser(t, l)
	assert.NoError(t, verifyPassword(user.Email, password))
	newEmail := bson.NewObjectId().Hex() + "@gmail.com"
	user2, err := a.UserSetEmail(user.Uid, newEmail, l)
	assert.NoError(t, err)
	assert.Equal(t, user2.Uid, user.Uid)

	assert.NoError(t, verifyPassword(newEmail, password))
	err = verifyPassword(user.Email, password)
	assert.Equal(t, err.Error(), "failed_authentication")
}

func TestSDK_UserCreateByPassword(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	l := xlog.NewDummy()
	createUser(t, l)
}

func TestSDK_UserUpdate(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	l := xlog.NewDummy()
	a := getService(t)
	user, _ := createUser(t, l)
	param := url.Values{}
	param.Set("utype", "36")
	param.Set("child_email_domain", "example.com")
	info, err := a.UserUpdate(user.Uid, param, l)
	assert.NoError(t, err)
	info, err = a.UserInfoByUid(user.Uid, l)
	assert.NoError(t, err)
	assert.Equal(t, info.ChildEmailDomain, "example.com")
	assert.Equal(t, info.Utype, uint32(0x24))
}

func TestSDK_UserChildren(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	l := xlog.NewDummy()
	a := getService(t)
	user, _ := createUser(t, l)
	param := url.Values{}
	param.Set("utype", "36")
	param.Set("child_email_domain", "example.com")
	_, err := a.UserUpdate(user.Uid, param, l)
	assert.NoError(t, err)

	infos, err := a.UserChildren(user.Uid, 0, 0, l)
	assert.NoError(t, err)
	assert.Equal(t, len(infos), 0)
}

func TestSDK(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	a := getService(t)
	l := xlog.NewDummy()

	name := bson.NewObjectId().Hex()
	email := strings.ToUpper(name + "@example.com")
	id := name + "@example.com"

	info, err := a.UserCreateByVendor("github", name, email, l)
	assert.NoError(t, err)
	assert.Equal(t, info.Id, id)
	assert.Equal(t, len(info.Vendors), 1)

	info, err = a.UserInfoById(info.Id, l)
	assert.NoError(t, err)
	assert.Equal(t, info.Id, id)
	assert.Equal(t, len(info.Vendors), 1)

	info, err = a.UserInfoByUid(info.Uid, l)
	assert.NoError(t, err)
	assert.Equal(t, info.Id, id)
	assert.Equal(t, len(info.Vendors), 1)

	info, err = a.UserInfoByVendor("github", name, l)
	assert.NoError(t, err)
	assert.Equal(t, info.Id, id)
	assert.Equal(t, len(info.Vendors), 1)

	info, err = a.UserBindAccount(info.Uid, "csdn", name, email, l)
	assert.NoError(t, err)
	assert.Equal(t, info.Id, id)
	assert.Equal(t, len(info.Vendors), 2)

	// bind again: failed
	_, err = a.UserBindAccount(info.Uid, "csdn", name, email, l)
	assert.Equal(t, err.(*rpc.ErrorInfo).Err, "vendor_id_exist")
	assert.Equal(t, err.(*rpc.ErrorInfo).Code, 400)

	info, err = a.UserUnbindAccount(info.Uid, "csdn", l)
	assert.NoError(t, err)
	assert.Equal(t, info.Id, id)
	assert.Equal(t, len(info.Vendors), 1)
	assert.Equal(t, info.Vendors[0].VendorEmail, email)
	assert.False(t, info.Vendors[0].CreatedAt.IsZero())

	// unbind again: failed
	_, err = a.UserUnbindAccount(info.Uid, "csdn", l)
	assert.Equal(t, err.(*rpc.ErrorInfo).Err, "vendor_not_exist")

	info, err = a.UserDisable(info.Uid, "reason 1", DISABLED_TYPE_AUTO, l)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x8014))
	assert.Equal(t, info.DisabledReason, "reason 1")

	info, err = a.UserInfoByUid(info.Uid, l)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x8014))

	info, err = a.UserAutoEnable(info.Uid, l)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x14))
	assert.Equal(t, info.DisabledReason, "")

	info, err = a.UserDisable(info.Uid, "reason 1", DISABLED_TYPE_MANUAL, l)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x8014))
	assert.Equal(t, info.DisabledReason, "reason 1")

	info, err = a.UserInfoByUid(info.Uid, l)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x8014))

	_, err = a.UserAutoEnable(info.Uid, l)
	assert.Equal(t, err.(*rpc.ErrorInfo).Err, "need_manunal_enable")

	info, err = a.UserInfoByUid(info.Uid, l)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x8014))

	info, err = a.UserForceEnable(info.Uid, l)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x14))

	info, err = a.UserInfoByUid(info.Uid, l)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x14))

	info, err = a.UserSetCustomerGroup(info.Uid, 2, l) //vip
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(6))

	infos, err := a.ListUsers(0, 20, l)
	assert.NoError(t, err)
	assert.Equal(t, len(infos), 20)

	infos, err = a.ListUsersByUtype(4, 0, 20, l)
	assert.NoError(t, err)
	assert.Equal(t, len(infos), 20)

	infoMap, err := a.ListUsersByUids([]uint32{info.Uid}, l)
	assert.NoError(t, err)
	assert.Equal(t, len(infoMap), 1)

	token, err := a.TokenCreate(info.Uid, l)
	assert.NoError(t, err)
	assert.NotEmpty(t, token.AccessToken)
	assert.NotEmpty(t, token.RefreshToken)
	assert.Equal(t, token.TokenExpiry, 3600)
}
