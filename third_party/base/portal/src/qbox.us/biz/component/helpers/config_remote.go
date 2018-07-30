package helpers

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/teapots/config"
	"github.com/teapots/teapot"

	"github.com/qcos/configs/api"
)

type RemoteEnv struct {
	tea    *teapot.Teapot
	log    teapot.Logger
	Client *api.Client
}

func NewRemoteFromEnv(tea *teapot.Teapot) *RemoteEnv {
	cfg := api.Config{}
	cfg.LoadFromEnv()

	re := &RemoteEnv{
		tea:    tea,
		log:    tea.Logger(),
		Client: api.New(cfg),
	}
	return re
}

// 加载指定版本的多个 ini 配置文件
// 之后加载的会继承覆盖之前加载的配置项
//
func (p *RemoteEnv) LoadFiles(ver int, env interface{}, files ...string) (err error) {
	batches := make([]*api.BatchFile, 0, len(files))
	for _, file := range files {
		batches = append(batches, &api.BatchFile{
			Name:    file,
			Version: ver,
		})
	}

	res, err := p.Client.BatchConfig(batches)
	if err != nil {
		err = &apiError{errServerErr, err, err.Error()}
		return
	}

	var last config.Configer
	for _, fileRes := range res {
		if fileRes.Code != 200 {
			err = &apiError{errFileLoadErr, nil,
				fmt.Sprintf("RemoteConfig file load failed: %s, code: %d, msg: %s, ver: %d",
					fileRes.Data.Name, fileRes.Code, fileRes.Message, ver),
			}
			return
		}
		data := fileRes.Data
		p.log.Infof("RemoteConfig file loaded name: %s, version: %d", data.Name, data.Version)

		var conf config.Configer
		conf, err = config.LoadIniFromReader(bytes.NewBufferString(data.Content))
		if err != nil {
			err = &apiError{errFileLoadErr, err,
				fmt.Sprintf("RemoteConfig file parsed err: %s %v %v", data.Name, data.Version, err),
			}
			return
		}

		if last != nil {
			conf.SetParent(last)
		}
		last = conf
	}

	if err == nil && last != nil {
		p.tea.ImportConfig(last)
		config.Decode(p.tea.Config, env)
	}
	return
}

type errTyp int

const (
	errServerErr errTyp = iota + 1
	errFileLoadErr
)

type apiError struct {
	typ errTyp
	err error
	msg string
}

func (e *apiError) Error() string {
	if e.msg != "" {
		return e.msg
	}
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}

func (e *apiError) OriginError() error {
	if e.err != nil {
		return e.err
	}
	return errors.New(e.msg)
}

func (e *apiError) IsServerError() bool {
	return e.typ == errServerErr
}

func (e *apiError) IsFileLoadError() bool {
	return e.typ == errFileLoadErr
}

type RemoteEnvApiError interface {
	IsServerError() bool
	IsFileLoadError() bool
	Error() string
	OriginError() error
}

func RemoteEnvError(err error) RemoteEnvApiError {
	e, ok := err.(*apiError)
	if !ok {
		return &apiError{}
	}
	return e
}
