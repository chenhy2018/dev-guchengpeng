package account

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"labix.org/v2/mgo/bson"
	"qbox.us/admin_api/v2/account"
)

const (
	HOST                    = "http://222.73.201.152:9100"
	MOCK_CLIENT_ID          = "abcd0c7edcdf914228ed8aa7c6cee2f2bc6155e2"
	MOCK_CLIENT_SECRET      = "fc9ef8b171a74e197b17f85ba23799860ddf3b9c"
	MOCK_INIT_USER_ID       = "init@qbox.net"
	MOCK_INIT_USER_PASSWORD = "qboxtest123"
)

func create(host, username, password string, l rpc.Logger) (err error) {
	client := rpc.Client{
		http.DefaultClient,
	}
	err = client.CallWithForm(l, nil, host+"/user/create", map[string][]string{"email": {username}, "password": {password}})
	return
}

func createUser(t *testing.T, l *xlog.Logger) (email, password string) {
	email = "a" + bson.NewObjectId().Hex() + "@example.com"
	password = email
	err := create(HOST, email, password, l)
	assert.NoError(t, err)
	return
}

func TestSDK_ChangePassword(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	l := xlog.NewDummy()
	email, password := createUser(t, l)
	new_password := "new_password"

	a, err := NewService(HOST, MOCK_CLIENT_ID, MOCK_CLIENT_SECRET, email, password)
	assert.NoError(t, err)

	err = a.ChangePassword(password, new_password, l)
	assert.NoError(t, err)

	_, err = NewService(HOST, MOCK_CLIENT_ID, MOCK_CLIENT_SECRET, email, password)
	assert.Equal(t, err.Error(), "failed_authentication")

	_, err = NewServiceByRefreshToken(HOST, MOCK_CLIENT_ID, MOCK_CLIENT_SECRET, a.OAuthToken().RefreshToken)
	assert.Equal(t, err.Error(), "expired_token")

	_, err = NewService(HOST, MOCK_CLIENT_ID, MOCK_CLIENT_SECRET, email, new_password)
	assert.NoError(t, err)
}

func TestSDK_Logout(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	l := xlog.NewDummy()
	email, password := createUser(t, l)

	a, err := NewService(HOST, MOCK_CLIENT_ID, MOCK_CLIENT_SECRET, email, password)
	assert.NoError(t, err)

	refreshToken := a.OAuthToken().RefreshToken
	_, err = NewServiceByRefreshToken(HOST, MOCK_CLIENT_ID, MOCK_CLIENT_SECRET, refreshToken)
	assert.NoError(t, err)

	a.Logout(refreshToken, l)

	_, err = NewServiceByRefreshToken(HOST, MOCK_CLIENT_ID, MOCK_CLIENT_SECRET, refreshToken)
	assert.Equal(t, err.Error(), "expired_token")
}

func TestSDK_UserInfo(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	l := xlog.NewDummy()
	email, password := createUser(t, l)

	a, err := NewService(HOST, MOCK_CLIENT_ID, MOCK_CLIENT_SECRET, email, password)
	assert.NoError(t, err)

	res, err := a.UserInfo(l)
	assert.NoError(t, err)
	assert.Equal(t, res.Email, email)
	assert.Equal(t, res.UserId, email)
}

func TestSDK_Child(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	l := xlog.NewDummy()
	email, password := createUser(t, l)

	a, err := NewService(HOST, MOCK_CLIENT_ID, MOCK_CLIENT_SECRET, email, password)
	assert.NoError(t, err)

	res, err := a.UserInfo(l)
	assert.NoError(t, err)

	param := url.Values{}
	param.Set("utype", "36")
	param.Set("child_email_domain", "example.com")
	_, err = getAdminService(t).UserUpdate(res.Uid, param, l)
	assert.NoError(t, err)

	a, err = NewService(HOST, MOCK_CLIENT_ID, MOCK_CLIENT_SECRET, email, password)

	password = bson.NewObjectId().Hex()
	email = password + "@example.com"
	info, err := a.UserCreateChild(email, password, l)
	assert.NoError(t, err)
	assert.Equal(t, info.UserId, email)
	assert.Equal(t, info.Email, email)
	assert.Equal(t, info.ParentUid, res.Uid)
	assert.Equal(t, info.UserType, uint32(4))
	assertUserActive(t, email, password)

	// 禁用用户
	info, err = a.UserDisableChild(info.Uid, "test", l)
	assert.NoError(t, err)

	// 再次禁用
	_, err = a.UserDisableChild(info.Uid, "test", l)
	assert.Equal(t, err.Error(), "user_already_disabled")

	// 启用
	info, err = a.UserEnableChild(info.Uid, l)
	assert.NoError(t, err)

	// 再次启用
	_, err = a.UserEnableChild(info.Uid, l)
	assert.Equal(t, err.Error(), "user_already_enabled")

	infos, err := a.UserChildren(0, 10, l)
	assert.Equal(t, len(infos), 1)
	assert.Equal(t, infos[0].Uid, info.Uid)
}

func assertUserActive(t *testing.T, email, password string) {
	a, err := NewService(HOST, MOCK_CLIENT_ID, MOCK_CLIENT_SECRET, email, password)
	assert.NoError(t, err)
	_, err = a.UserInfo(nil)
	assert.NoError(t, err)
}

func getAdminService(t *testing.T) *account.Service {
	a, err := account.NewService(
		HOST,
		MOCK_CLIENT_ID,
		MOCK_CLIENT_SECRET,
		MOCK_INIT_USER_ID,
		MOCK_INIT_USER_PASSWORD)
	assert.NoError(t, err)
	return a
}
