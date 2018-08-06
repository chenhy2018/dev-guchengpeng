package qauth

import (
	"fmt"
	. "qiniu.com/qauth/config"
	"strings"

	qconf "qbox.us/qconf/qconfapi"
	proto "qiniu.com/auth/proto.v1"
)

//------------------------------------------------------
// verify qiniu access key
//------------------------------------------------------

func GetUserInfoByAccessKey(accessKey string) (proto.AccessInfo, error) {
	resp := proto.AccessInfo{}
	if QConfClient == nil {
		return resp, fmt.Errorf("qconf client has not been initialized")
	}
	err := QConfClient.Get(nil, &resp, MakeId(accessKey), qconf.Cache_NoSuchEntry)
	if err != nil {
		return resp, fmt.Errorf("get account info failed: %v", err)
	}
	return resp, nil
}

func MakeId(key string) string {
	return GROUP_PREFIX + key
}

func ParseId(id string) (key string, err error) {
	if !strings.HasPrefix(id, GROUP_PREFIX) {
		return "", fmt.Errorf("invalid group prefix")
	}
	return id[len(GROUP_PREFIX):], nil
}
