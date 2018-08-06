package config

import (
	"fmt"
	qconf "qbox.us/qconf/qconfapi"
	"testing"
)

func TestPubConfig(t *testing.T) {
	//pubConf, err := NewPubConfig()
	pubConf, err := LoadPubConfig("./pub_server.conf")
	if err == nil {
		fmt.Println("test section ...... ", pubConf.GetSection("account").GetSection("test").String("test"))

		fmt.Println("master_hosts ", pubConf.GetSection("access").Strings("master_hosts"))
		fmt.Println("mc_hosts ", pubConf.GetSection("access").Strings("mc_hosts"))
		fmt.Println("access ", pubConf.GetSection("access").String("access_key"))
		fmt.Println("secret ", pubConf.GetSection("access").String("secret_key"))

		fmt.Println("client id", pubConf.GetSection("account").String("client_id"))
		fmt.Println("secret ", pubConf.GetSection("access").String("secret_key"))
		fmt.Println("client host", pubConf.GetSection("account").String("host"))
	} else {
		fmt.Println("config parse failed: ", err)
	}
}

func TestConfig(t *testing.T) {

	pubConf, err := LoadPubConfig("./pub_server.conf")
	if err != nil {
		fmt.Println(err)
	}
	cfg := &qconf.Config{
		MasterHosts:       pubConf.GetSection(CONF_ACCESS).Strings(MASTER_HOSTS),
		McHosts:           pubConf.GetSection(CONF_ACCESS).Strings(MC_HOSTS),
		AccessKey:         pubConf.GetSection(CONF_ACCESS).String(ACCESS_KEY),
		SecretKey:         pubConf.GetSection(CONF_ACCESS).String(SECRET_KEY),
		LcacheExpires:     300000,
		LcacheDuration:    60000,
		LcacheChanBufSize: 16000,
		McRWTimeout:       100,
	}
	fmt.Println("config: ", cfg)
}
