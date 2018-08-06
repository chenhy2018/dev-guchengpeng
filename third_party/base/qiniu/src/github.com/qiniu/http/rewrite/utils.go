package rewrite

import (
	"strconv"
	"strings"
	"time"

	"qiniupkg.com/api.v7/auth/qbox"
)

func makePrivateUrl(baseUrl string, expires int64, accessKey, secretKey string) (privateUrl string) {
	if expires <= 0 {
		expires = 3600
	}
	deadline := time.Now().Unix() + expires
	if strings.Contains(baseUrl, "?") {
		baseUrl += "&e="
	} else {
		baseUrl += "?e="
	}
	baseUrl += strconv.FormatInt(deadline, 10)
	token := qbox.Sign(qbox.NewMac(accessKey, secretKey), []byte(baseUrl))
	return baseUrl + "&token=" + token
}
