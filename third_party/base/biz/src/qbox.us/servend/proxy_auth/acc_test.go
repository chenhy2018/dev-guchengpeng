package proxy_auth

import (
	"fmt"
	"net/http"
	"qbox.us/servend/account"
	"testing"
)

// ---------------------------------------------------------------------------

func Test(t *testing.T) {

	user := account.UserInfo{
		Uid:   123,
		Utype: 4,
		Devid: 5,
	}
	auth := MakeAuth(user)
	fmt.Println(auth)
	if auth != "QiniuProxy uid=123&ut=4&dev=5" {
		t.Fatal("MakeAuth failed:", user, auth)
	}

	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	req.Header.Set("Authorization", auth)

	user2, err := ParseAuth(req)
	if err != nil {
		t.Fatal("GetAuth failed", err)
	}
	fmt.Println("user:", user2)
	if user2.Uid != user.Uid || user2.Utype != user.Utype || user2.Devid != user.Devid {
		t.Fatal("GetAuth failed:", user, user2)
	}
}

// ---------------------------------------------------------------------------
