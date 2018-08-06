package config

import (
	"fmt"
	"net/http"

	admin "qbox.us/admin_api/account.v2"
	account "qbox.us/api/account.v2"
	"qbox.us/oauth"
	qconf "qbox.us/qconf/qconfapi"
)

//------------------------------------------------------
// load pub_server.conf
//------------------------------------------------------

var inputArgs = []string{
	"../pub_server.conf",
	"./pub_server.conf",
}

type PubConfig struct {
	content JsonConfigContainer
}

var (
	Pubconf      *PubConfig
	QConfClient  *qconf.Client
	QConfAdmin   *admin.Service
	QConfAccount *account.Service
	QTransport   *oauth.Transport
)

func InitConf() {

	Pubconf = &PubConfig{}
	for i, arg := range inputArgs {
		jc, err := Pubconf.LoadConfig(CONF_TYPE_JSON, arg)
		if err == nil {
			Pubconf.content.data = jc.DupData().(map[string]interface{})
			break
		}
		if err != nil && i == len(inputArgs)-1 {
			Pubconf = nil
			fmt.Printf("load pubserver config failed: %v\n", err)
		}
	}

	QConfClient, err := InitQConfClient()
	if QConfClient == nil {
		fmt.Printf("create qconf client failed: %v\n", err)
	}
	err = InitQConfService()
	if err != nil {
		fmt.Printf("create qconf service failed: %v\n", err)
	}
}

func LoadPubConfig(path string) (*PubConfig, error) {
	pubConf := &PubConfig{}
	jc, err := pubConf.LoadConfig(CONF_TYPE_JSON, path)
	if err == nil {
		pubConf.content.data = jc.DupData().(map[string]interface{})
		return pubConf, nil
	}
	return nil, err
}

func (this *PubConfig) GetSection(sectionKey string) *PubConfig {
	data, ok := this.content.Get(sectionKey).(map[string]interface{})
	if ok {
		s := &PubConfig{
			content: JsonConfigContainer{
				data: data,
			},
		}
		return s
	}
	return this
}

func (this *PubConfig) Get(key string) interface{} {
	return this.content.Get(key)
}

func (this *PubConfig) String(key string) string {
	return this.content.String(key)
}

func (this *PubConfig) Strings(key string) []string {
	return this.content.Strings(key)
}

func (this *PubConfig) Int64(key string) (int64, error) {
	val, err := this.content.Int64(key)
	if err == nil {
		return val, nil
	}
	return 0, err
}

func (this *PubConfig) Int(key string) (int, error) {
	val, err := this.content.Int(key)
	if err == nil {
		return val, nil
	}
	return 0, err
}

func (this *PubConfig) LoadConfig(providerName, path string) (Configer, error) {

	RegisterProvider(providerName, &JsonConfigParser{})

	absPath, err := LoadConfigPath(path)
	if err != nil {
		return nil, err
	}
	return NewConfig(providerName, absPath)
}

// init qconf client
func InitQClientConfig() (cfg *qconf.Config, err error) {
	if Pubconf == nil {
		return nil, fmt.Errorf("pub_server.conf is not initialized")
	}
	cfg = &qconf.Config{
		MasterHosts:       Pubconf.GetSection(CONF_ACCESS).Strings(MASTER_HOSTS),
		McHosts:           Pubconf.GetSection(CONF_ACCESS).Strings(MC_HOSTS),
		AccessKey:         Pubconf.GetSection(CONF_ACCESS).String(ACCESS_KEY),
		SecretKey:         Pubconf.GetSection(CONF_ACCESS).String(SECRET_KEY),
		LcacheExpires:     LOCAL_CACHE_EXPIRES,
		LcacheDuration:    LOCAL_CACHE_DURATION,
		LcacheChanBufSize: LOCAL_CACHE_CHAN_BUF_SIZE,
		McRWTimeout:       MC_RW_TIMEOUT,
	}
	return cfg, nil
}

func InitQTransport() *oauth.Transport {
	if Pubconf == nil {
		return nil
	}
	QTransport := &oauth.Transport{
		Config: &oauth.Config{
			ClientId:     Pubconf.GetSection(CONF_ACCOUNT).String(CLIENT_ID),
			ClientSecret: Pubconf.GetSection(CONF_ACCOUNT).String(CLIENT_SECRET),
			Scope:        SCOPE,
			TokenURL:     Pubconf.GetSection(CONF_ACCOUNT).String(HOST) + TOKEN_URL,
			AuthURL:      "",
			RedirectURL:  "",
		},
		Transport: http.DefaultTransport,
	}
	return QTransport
}

func InitQConfClient() (*qconf.Client, error) {
	cfg, err := InitQClientConfig()
	if err != nil {
		return nil, fmt.Errorf("config parse failed: %v", err)
	}
	QConfClient = qconf.New(cfg)
	if QConfClient == nil {
		return nil, fmt.Errorf("init qconf client failed")
	}
	return QConfClient, nil
}

func InitQConfAdmin() (*admin.Service, error) {
	if QTransport == nil {
		return nil, fmt.Errorf("init qconf admin failed: qtransport has not been initialized\n")
	}

	QConfAdmin = admin.New(Pubconf.GetSection(CONF_ACCOUNT).String(HOST), QTransport)
	if QConfAdmin == nil {
		return nil, fmt.Errorf("get qconf admin failed\n")
	}

	return QConfAdmin, nil
}

func InitQConfAccount() (*account.Service, error) {
	if QTransport == nil {
		return nil, fmt.Errorf("init qconf account failed: qtransport has not been initialized\n")
	}

	QConfAccount = account.New(Pubconf.GetSection(CONF_ACCOUNT).String(HOST), QTransport)
	if QConfAccount == nil {
		return nil, fmt.Errorf("get qconf admin failed\n")
	}
	return QConfAccount, nil
}

func InitQConfService() error {
	QTransport = InitQTransport()
	if QTransport == nil {
		return fmt.Errorf("init qconf transport failed\n")
	}

	_, err := InitQConfAdmin()
	if err != nil {
		return fmt.Errorf("init qconf admin failed: %v\n")
	}

	_, err = InitQConfAccount()
	if err != nil {
		return fmt.Errorf("init qconf account failed : %v\n")
	}
	return nil
}
