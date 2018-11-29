package db

import (
	"fmt"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"testing"
)

func TestDbInit(t *testing.T) {
	url := "mongodb://127.0.0.1:27017"
	dbName := "vod"
	config := MgoConfig{
		Host:     url,
		DB:       dbName,
		Mode:     "strong",
		Username: "root",
		Password: "public",
		AuthDB:   "admin",
		Proxies:  nil,
	}
	xl := xlog.NewDummy()
	xl.Infof("TestDbInit")
	err := InitDb(&config)
	assert.Equal(t, err, nil, "they should be equal")
	// already init also return nil err.
	err = InitDb(&config)
	assert.Equal(t, err, nil, "they should be equal")
	// wrong port or url. return err.
	DinitDb()
	config.Host = "mongodb://127.0.0.1:27018"
	err = InitDb(&config)
	assert.Equal(t, err.Error(), "db open failed: no reachable servers", "they should be equal")
	config.Host = "mongodb://127.0.0.3:27017"
	assert.Equal(t, err.Error(), "db open failed: no reachable servers", "they should be equal")
}

func TestDbConnect(t *testing.T) {
	url := "mongodb://127.0.0.1:27017"
	dbName := "vod"
	config := MgoConfig{
		Host:     url,
		DB:       dbName,
		Mode:     "strong",
		Username: "root",
		Password: "public",
		AuthDB:   "admin",
		Proxies:  nil,
	}
	xl := xlog.NewDummy()
	xl.Infof("TestDbConnect")
	err := InitDb(&config)
	assert.Equal(t, err, nil, "they should be equal")
	err = WithCollection(
		"test",
		func(c *mgo.Collection) error {
			return nil
		},
	)
	assert.Equal(t, err, nil, "they should be equal")

	err = WithCollection(
		"test",
		func(c *mgo.Collection) error {
			return fmt.Errorf("function err")
		},
	)
	assert.Equal(t, err.Error(), "function err", "they should be equal")
	// if not init. return err
	DinitDb()
	err = WithCollection(
		"test",
		func(c *mgo.Collection) error {
			return nil
		},
	)
	assert.Equal(t, err.Error(), "Data base is not init", "they should be equal")
}
