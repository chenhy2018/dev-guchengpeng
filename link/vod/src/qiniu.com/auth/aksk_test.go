package auth

import (
	"errors"
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	rpc "github.com/qiniu/rpc.v1"
	xlog "github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify.v1/suite"
	"qbox.us/qconf/qconfapi"
)

type AuthTestSuite struct {
	suite.Suite
	xl *xlog.Logger
}

func (suite *AuthTestSuite) TestIfQconfIsNil() {
	QConfClient = nil
	SetSkFromUser(xlog.NewDummy(), []byte("xxaass"))
	info, err := GetUserInfoFromQconf(xlog.NewDummy(), "xssss")
	// if qconf client has not been initialized, also return nil and use load sk
	suite.Equal(info.Secret, []byte("xxaass"))

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*qconfapi.Client)(nil)), "Get", func(ss *qconfapi.Client, logger rpc.Logger, ret interface{}, ak string, entry int) error {
			return errors.New("get info error")
		})
	cfg := &qconfapi.Config{
		McHosts:     []string{"10.200.20.23:11211"},
		MasterHosts: []string{"http://10.200.20.25:8510"},
	}
	QConfClient = qconfapi.New(cfg)
	_, err = GetUserInfoFromQconf(xlog.NewDummy(), "xssss")
	suite.Equal(err, errors.New("get account info failed"), nil)

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*qconfapi.Client)(nil)), "Get", func(ss *qconfapi.Client, logger rpc.Logger, ret interface{}, ak string, entry int) error {
			return nil
		})

	_, err = GetUserInfoFromQconf(xlog.NewDummy(), "xssss")
	suite.Equal(err, nil)
}
func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
