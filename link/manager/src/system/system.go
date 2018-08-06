package system

import (
	"encoding/json"
	"os"
)

type DBConfiguration struct {
	Host     string `json:"host"`
	Db       string `json:"db"`
	Mode     string `json:"mode"`
	TimeOut  int    `json:"timeout"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type Configuration struct {
	Bind   string          `json:"bind"`
	DbConf DBConfiguration `json:"db_conf"`
}

func LoadConf(file string) (conf *Configuration, err error) {
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		return
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&conf)
	return conf, nil
}
