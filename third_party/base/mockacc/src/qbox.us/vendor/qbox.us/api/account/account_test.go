package account

import (
	"os"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
	"labix.org/v2/mgo/bson"
)

const (
	MOCK_CLIENT_ID     = "abcd0c7edcdf914228ed8aa7c6cee2f2bc6155e2"
	MOCK_CLIENT_SECRET = "fc9ef8b171a74e197b17f85ba23799860ddf3b9c"
)

// /user/signup 已经去除
func testSDK(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	a := New("http://222.73.201.152:9100", MOCK_CLIENT_ID, MOCK_CLIENT_SECRET)

	user := "a" + bson.NewObjectId().Hex() + "@example.com"
	token, code, err := a.Signup(user, user, "0")
	assert.NoError(t, err)
	assert.Equal(t, code, 200)
	assert.NotNil(t, token)

	accessToken := token.AccessToken

	info, code, err := a.UserInfo(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, code, 200)
	assert.NotNil(t, info)

	transport, err := a.GetOAuthTransport(user, user)
	assert.NoError(t, err)
	assert.NotNil(t, transport)

	transport, err = a.GetOAuthTransport(user, user+"a")
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "failed_authentication")
	assert.Nil(t, transport)
	transport, err = a.GetOAuthTransport(user, user+"a")
	assert.Equal(t, err.Error(), "failed_authentication")
	transport, err = a.GetOAuthTransport(user, user+"a")
	assert.Equal(t, err.Error(), "failed_authentication")
	transport, err = a.GetOAuthTransport(user, user+"a")
	assert.Equal(t, err.Error(), "failed_authentication")
	transport, err = a.GetOAuthTransport(user, user+"a")
	assert.Equal(t, err.Error(), "failed_authentication")
	transport, err = a.GetOAuthTransport(user, user+"a")
	assert.Equal(t, err.Error(), "user_short_blocked")
	transport, err = a.GetOAuthTransport(user, user)
	assert.Equal(t, err.Error(), "user_short_blocked")
}
