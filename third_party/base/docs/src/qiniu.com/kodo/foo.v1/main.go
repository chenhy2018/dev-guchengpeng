package foo

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
)

type cmdArgs struct {
	CmdArgs []string
}

type fooInfo struct {
	A   int    `json:"a"`
	Bar string `json:"bar"`
}

// ---------------------------------------------------------------------------

type Config struct {
}

/*
	一般来说，服务器比较少会自己维护状态，通常数据存储在数据库中（MongoDB或者MySQL之类）。
	这个样例为了简单，维护了内存中的状态（map和mutex都是状态）。这会导致服务器无法 Scale（不能启动多个实例）。
	所以这只是一个服务器编写框架的样例。
*/
type Service struct {
	foos  map[string]*fooInfo
	mutex sync.RWMutex
}

func New(cfg *Config) (p *Service, err error) {

	p = &Service{
		foos: make(map[string]*fooInfo),
	}
	return
}

func genId() (id string, err error) {

	b := make([]byte, 12)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ---------------------------------------------------------------------------

