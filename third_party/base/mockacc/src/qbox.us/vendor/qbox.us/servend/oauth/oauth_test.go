package oauth_test

import (
	"fmt"
	"qbox.us/mockacc"
	"qbox.us/oauth"
	"testing"
	"time"
	//	"qbox.us/account"
)

func TestOAuthWithMockAccount(t *testing.T) {

	sa := &mockacc.SaInstance
	go mockacc.Run(":7899", sa)
	time.Sleep(1e9)

	// Set up a configuration
	config := &oauth.Config{
		ClientId:     "<ClientId>",
		ClientSecret: "<ClientSecret>",
		Scope:        "<Scope>",
		AuthURL:      "<AuthURL>",
		TokenURL:     "http://localhost:7899/oauth2/token",
		RedirectURL:  "<RedirectURL>",
	}

	transport := &oauth.Transport{Config: config}

	token, _, err := transport.ExchangeByPassword("qboxtest", "qboxtest123")
	if err != nil {
		t.Fatal("ExchangeByPassword:", err)
	}

	fmt.Println(token)
}

/*
func TestOAuthWithRealAccount(t *testing.T) {
	go account.Run(":7891")
	time.Sleep(1e9)

	// Set up a configuration
	config := &Config{
		ClientId:     account.MOCK_CLIENT_ID,
		ClientSecret: account.MOCK_CLIENT_SECRET,
		Scope:        "<Scope>",
		AuthURL:      "<AuthURL>",
		TokenURL:     "http://localhost:7891/oauth2/token",
		RedirectURL:  "<RedirectURL>",
	}

	transport := &Transport{Config: config}

	token, _, err := transport.ExchangeByPassword(account.MOCK_USERID, account.MOCK_PASSWORD)
	if err != nil {
		t.Fatal("ExchangeByPassword:", err)
	}

	fmt.Println(token)
}
*/
