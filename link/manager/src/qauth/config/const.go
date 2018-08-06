package config

const (
	CONF_TYPE_JSON = "json"

	CONF_ACCESS  = "access"
	MASTER_HOSTS = "master_hosts"
	MC_HOSTS     = "mc_hosts"
	SECRET_KEY   = "secret_key"
	ACCESS_KEY   = "access_key"

	CONF_ACCOUNT         = "account"
	GROUP_PREFIX         = "ak:"
	CLIENT_ID            = "client_id"
	CLIENT_SECRET        = "client_secret"
	CLIENT_INIT_USER     = "user"
	CLIENT_INIT_PASSWORD = "passwd"

	MC_RW_TIMEOUT             = 100
	LOCAL_CACHE_EXPIRES       = 300000
	LOCAL_CACHE_DURATION      = 60000
	LOCAL_CACHE_CHAN_BUF_SIZE = 16000

	HOST         = "host"
	SCOPE        = "Scope"
	TOKEN_URL    = "/oauth2/token"
	USR_INFO_URL = "/user/info"
)
