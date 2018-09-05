package biz

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"qbox.us/oauth"
)

const (
	HOST                    = "http://222.73.201.152:8093"
	MOCK_INIT_USER_ID       = "init@qbox.net"
	MOCK_INIT_USER_PASSWORD = "qboxtest123"
	MOCK_CLIENT_ID          = "abcd0c7edcdf914228ed8aa7c6cee2f2bc6155e2"
	MOCK_CLIENT_SECRET      = "fc9ef8b171a74e197b17f85ba23799860ddf3b9c"
)

func TestSDKByDigestAuth(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" || os.Getenv("JENKINS_URL") != "" {
		t.SkipNow()
		return
	}
	if os.Getenv("ADMIN_ACCESS_KEY") == "" || os.Getenv("ADMIN_SECRET_KEY") == "" {
		t.Log("ADMIN_ACCESS_KEY or ADMIN_SECRET_KEY not set")
		t.SkipNow()
		return
	}
	client := digest.NewClient(
		&digest.Mac{
			AccessKey: os.Getenv("ADMIN_ACCESS_KEY"),
			SecretKey: []byte(os.Getenv("ADMIN_SECRET_KEY")),
		}, nil)
	testSDKByClient(t, client)
}

func TestSDKByOAuth(t *testing.T) {
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		t.SkipNow()
		return
	}
	transport := &oauth.Transport{
		Config: &oauth.Config{
			ClientId:     MOCK_CLIENT_ID,
			ClientSecret: MOCK_CLIENT_SECRET,
			Scope:        "Scope",
			AuthURL:      "",
			TokenURL:     "http://222.73.201.152:9100/oauth2/token",
			RedirectURL:  "",
		},
		Transport: http.DefaultTransport, // it is default
	}
	_, _, err := transport.ExchangeByPassword(MOCK_INIT_USER_ID, MOCK_INIT_USER_PASSWORD)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	client := &http.Client{
		Transport: transport,
	}
	testSDKByClient(t, client)
}

func testSDKByClient(t *testing.T, client *http.Client) {
	a := NewBizService(
		HOST,
		&rpc.Client{client},
	)
	assert.NotNil(t, a)

	l := xlog.NewDummy()

	c, err := a.CountDeveloper(l)
	assert.NoError(t, err)
	assert.True(t, c > 0)

	c, err = a.CountDeveloperByTime(l, time.Now())
	assert.NoError(t, err)
	assert.True(t, c > 0)
}
