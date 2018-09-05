package configs

import "github.com/teapots/teapot"

var Env Config
var Repo repo

type Config struct {
	Teapot *teapot.Teapot `conf:"-"`

	RepoPath string `conf:"repo_path"`
}

type RepoConfig struct {
	Auth map[string]struct {
		Service string `json:"service"`
		Secret  string `json:"secret"`
	} `json:"auth"`
}

type AuthInfo struct {
	Service   string
	AccessKey string
	SecretKey string
}
