package auth

import (
	"errors"
	"fmt"
	"testing"

	"github.com/bouk/monkey"
	xlog "github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify.v1/suite"
	redis "gopkg.in/redis.v5"
)

type AuthTestSuite struct {
	suite.Suite
	xl *xlog.Logger
}

func (suite *AuthTestSuite) SetupSuite() {
	RedisClint = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0})
	pong, err := RedisClint.Ping().Result()
	fmt.Println(pong, err)
	RedisClint.Set("4294967295", "adoNgymaNDST:CaRtAtemORgI", 0)
	RedisClint.Set("4294967294", "aNtrAlbeRSdu:siOvErAGerAM", 0)
	RedisClint.Set("4294967293", "ViArytORMOLo:eRYMiGRacTat", 0)
	RedisClint.Set("4294967292", "oCKUleFeScIp:SElERsEplAWF", 0)
	suite.xl = xlog.NewDummy()
}
func (suite *AuthTestSuite) TearDownSuite() {
	RedisClint = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0})
	pong, err := RedisClint.Ping().Result()
	fmt.Println(pong, err)
	RedisClint.FlushAll()
}
func (suite *AuthTestSuite) TestGetAkSkFromRedis() {
	defer monkey.UnpatchAll()
	//monkey.Patch(VerifyAuth, func(xl *xlog.Logger, req *http.Request) (bool, error) { return true, nil })
	ak, sk, err := getAKSKByUid(suite.xl, 4294967295, "adoNgymaNDST")
	suite.Equal(err, nil)
	suite.Equal(ak, "adoNgymaNDST", "")
	suite.Equal(sk, "CaRtAtemORgI", "")
}
func (suite *AuthTestSuite) TestGetAkSkFromQconf() {
	defer monkey.UnpatchAll()
	monkey.Patch(getSKByAK, func(ak string) (string, error) { return "atERglEafeRV", nil })
	ak, sk, err := getAKSKByUid(suite.xl, 4294967291, "ARARYLOwagon")
	suite.Equal(err, nil)
	suite.Equal(ak, "ARARYLOwagon", "")
	suite.Equal(sk, "atERglEafeRV", "")

	// now we shoud get it from redis
	defer monkey.UnpatchAll()
	ak, sk, err = getAKSKByUid(suite.xl, 4294967291, "ARARYLOwagon")
	suite.Equal(err, nil)
	suite.Equal(ak, "ARARYLOwagon", "")
	suite.Equal(sk, "atERglEafeRV", "")

}

func (suite *AuthTestSuite) TestRedisIsOKANDQconfIsFailed() {
	monkey.Patch(getSKByAK, func(ak string) (string, error) { return "", errors.New("get Accessinfo failed") })
	ak, sk, err := getAKSKByUid(suite.xl, 4294967289, "ARARYLOwagon")
	suite.Equal(err, errors.New("get sk failed"))
	suite.Equal(ak, "", "")
	suite.Equal(sk, "", "")
}

func (suite *AuthTestSuite) TestUpdateDataInRedis() {
	monkey.Patch(getSKByAK, func(ak string) (string, error) { return "MarLEDHicEIg", nil })
	ak, sk, err := getAKSKByUid(suite.xl, 4294967291, "iNDICKyJacHi")
	suite.Equal(err, nil)
	suite.Equal(ak, "iNDICKyJacHi", "")
	suite.Equal(sk, "MarLEDHicEIg", "")

	//  aksk in redis should be update
	ak, sk, err = getAKSKByUid(suite.xl, 4294967291, "iNDICKyJacHi")
	suite.Equal(err, nil)
	suite.Equal(ak, "iNDICKyJacHi", "")
	suite.Equal(sk, "MarLEDHicEIg", "")
}
func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
