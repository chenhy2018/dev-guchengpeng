package config

import (
	"fmt"
	"os"
	"path/filepath"
)

//------------------------------------------------------
// config parser for json, ini, xml ...
// just implements json config now.
//------------------------------------------------------

type Configer interface {
	Set(key string, val interface{}) error
	Get(key string) interface{}
	DupData() interface{}
}

type ConfigParser interface {
	Parse(key string) (Configer, error)
	ParseData(data []byte) (Configer, error)
}

var providers = make(map[string]ConfigParser)

func RegisterProvider(name string, provider ConfigParser) error {
	if provider == nil {
		return fmt.Errorf("config parser: provider is nil")
	}
	if _, ok := providers[name]; ok {
		return fmt.Errorf("config parser: %s called twice ", name)
	}
	providers[name] = provider
	return nil
}

func NewConfig(providerName, filename string) (Configer, error) {
	provider, ok := providers[providerName]
	if !ok {
		return nil, fmt.Errorf("config: unknown config provider %s", providerName)
	}
	return provider.Parse(filename)
}

func NewConfigData(providerName string, data []byte) (Configer, error) {
	provider, ok := providers[providerName]
	if !ok {
		return nil, fmt.Errorf("config: unknown config provider %s", providerName)
	}
	return provider.ParseData(data)
}

//------------------------------------------------------
// check config file path
//------------------------------------------------------

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func LoadConfigPath(path string) (string, error) {
	absConfigPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if !FileExists(absConfigPath) {
		return "", fmt.Errorf("pub_server.conf doesn't exist\n")
	}

	return absConfigPath, nil
}
