package account

import (
	"os"
	"strings"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
	"labix.org/v2/mgo/bson"
	"qbox.us/api/account"
)

const (
	HOST                    = "http://222.73.201.152:9100"
	MOCK_INIT_USER_ID       = "init@qbox.net"
	MOCK_INIT_USER_PASSWORD = "qboxtest123"
	MOCK_CLIENT_ID          = "abcd0c7edcdf914228ed8aa7c6cee2f2bc6155e2"
	MOCK_CLIENT_SECRET      = "fc9ef8b171a74e197b17f85ba23799860ddf3b9c"
)

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

func TestSDK(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	a, err := NewService(
		HOST,
		MOCK_CLIENT_ID,
		MOCK_CLIENT_SECRET,
		MOCK_INIT_USER_ID,
		MOCK_INIT_USER_PASSWORD)
	assert.NoError(t, err)

	name := bson.NewObjectId().Hex()
	email := strings.ToUpper(name + "@example.com")
	id := name + "@example.com"

	info, code, err := a.UserCreate("github", name, email)
	assert.NoError(t, err)
	assert.Equal(t, code, 200)
	assert.Equal(t, info.Id, id)
	assert.Equal(t, len(info.Vendors), 1)

	info, code, err = a.Info(info.Id)
	assert.NoError(t, err)
	assert.Equal(t, code, 200)
	assert.Equal(t, info.Id, id)
	assert.Equal(t, len(info.Vendors), 1)

	info, code, err = a.UserInfoById(info.Id)
	assert.NoError(t, err)
	assert.Equal(t, code, 200)
	assert.Equal(t, info.Id, id)
	assert.Equal(t, len(info.Vendors), 1)

	info, code, err = a.UserInfoByUid(info.Uid)
	assert.NoError(t, err)
	assert.Equal(t, code, 200)
	assert.Equal(t, info.Id, id)
	assert.Equal(t, len(info.Vendors), 1)

	info, code, err = a.UserInfoByVendor("github", name)
	assert.NoError(t, err)
	assert.Equal(t, code, 200)
	assert.Equal(t, info.Id, id)
	assert.Equal(t, len(info.Vendors), 1)

	info, code, err = a.UserBindAccount(info.Uid, "csdn", name, email)
	assert.NoError(t, err)
	assert.Equal(t, code, 200)
	assert.Equal(t, info.Id, id)
	assert.Equal(t, len(info.Vendors), 2)

	// bind again: failed
	_, code, err = a.UserBindAccount(info.Uid, "csdn", name, email)
	assert.Equal(t, err.Error(), "vendor_id_exist")
	assert.Equal(t, code, 400)

	info, code, err = a.UserUnbindAccount(info.Uid, "csdn")
	assert.NoError(t, err)
	assert.Equal(t, code, 200)
	assert.Equal(t, info.Id, id)
	assert.Equal(t, len(info.Vendors), 1)
	assert.Equal(t, info.Vendors[0].VendorEmail, email)
	assert.False(t, info.Vendors[0].CreatedAt.IsZero())

	// unbind again: failed
	_, code, err = a.UserUnbindAccount(info.Uid, "csdn")
	assert.Equal(t, err.Error(), "vendor_not_exist")
	assert.Equal(t, code, 400)

	info, code, err = a.UserDisable(info.Uid, "reason 1", DISABLED_TYPE_AUTO)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x8014))
	assert.Equal(t, info.DisabledReason, "reason 1")

	info, code, err = a.UserInfoByUid(info.Uid)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x8014))

	info, code, err = a.UserEnable(info.Uid)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x14))
	assert.Equal(t, info.DisabledReason, "")

	info, code, err = a.UserDisable(info.Uid, "reason 1", DISABLED_TYPE_MANUAL)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x8014))
	assert.Equal(t, info.DisabledReason, "reason 1")

	info, code, err = a.UserInfoByUid(info.Uid)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x8014))

	_, code, err = a.UserEnable(info.Uid)
	assert.Equal(t, code, 400)
	assert.Equal(t, err.Error(), "need_manunal_enable")

	info, code, err = a.UserInfoByUid(info.Uid)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x8014))

	info, code, err = a.UserForceEnable(info.Uid)
	assert.NoError(t, err)
	assert.Equal(t, code, 200)
	assert.Equal(t, info.Utype, uint32(0x14))

	info, code, err = a.UserInfoByUid(info.Uid)
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(0x14))

	info, code, err = a.UserSetCustomerGroup(info.Uid, 2) //vip
	assert.NoError(t, err)
	assert.Equal(t, info.Utype, uint32(6))

	infos, code, err := a.ListUsers(0, 20)
	assert.NoError(t, err)
	assert.Equal(t, len(infos), 20)

	infos, code, err = a.ListUsersByUtype(4, 0, 20)
	assert.NoError(t, err)
	assert.Equal(t, len(infos), 20)

	infoMap, code, err := a.ListUsersByUids([]uint32{info.Uid})
	assert.NoError(t, err)
	assert.Equal(t, len(infoMap), 1)

	token, code, err := a.TokenCreate(info.Uid)
	assert.NoError(t, err)
	assert.Equal(t, code, 200)
	assert.NotEmpty(t, token.AccessToken)
	assert.NotEmpty(t, token.RefreshToken)
	assert.Equal(t, token.TokenExpiry, 3600)
}
