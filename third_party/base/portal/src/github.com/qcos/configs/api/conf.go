package api

import (
	"net/http"
	"os"
	"strings"
)

type Config struct {
	AccessKey string
	SecretKey string

	Hosts     []string
	Transport http.RoundTripper
}

func (c *Config) LoadFromEnv() {
	c.AccessKey = os.Getenv("CONFIGS_ACCESS_KEY")
	c.SecretKey = os.Getenv("CONFIGS_SECRET_KEY")
	for _, host := range strings.Split(os.Getenv("CONFIGS_HOSTS"), ",") {
		host = strings.TrimSpace(host)
		if host != "" {
			c.Hosts = append(c.Hosts, host)
		}
	}
}

func (c *Config) Clean() {
	for i, host := range c.Hosts {
		for strings.HasSuffix(host, "/") {
			host = strings.TrimSuffix(host, "/")
		}
		c.Hosts[i] = host
	}
}
