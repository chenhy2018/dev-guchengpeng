package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

//------------------------------------------------------
// parse json config
//------------------------------------------------------

type JsonConfigContainer struct {
	data map[string]interface{}
	sync.RWMutex
}

// for processing special characters like comments in json config
var enter = []byte{'\n'}
var comment = byte('#')
var quote = byte('"')

func trimCommentsLine(line []byte) (newLine []byte) {
	quoteCount := 0
	lastIdx := len(line) - 1
	for i := 0; i <= lastIdx; i++ {
		if line[i] == quote {
			quoteCount++
		}
		if line[i] == comment {
			if quoteCount%2 == 0 {
				break
			}
		}
		newLine = append(newLine, line[i])
	}
	return newLine
}

func trimComments(data []byte) []byte {
	confLines := bytes.Split(data, enter)
	for k, line := range confLines {
		confLines[k] = trimCommentsLine(line)
	}
	return bytes.Join(confLines, enter)
}

func (jc *JsonConfigContainer) DupData() interface{} {
	return jc.data
}

func (jc *JsonConfigContainer) Set(key string, val interface{}) error {
	jc.Lock()
	defer jc.Unlock()
	jc.data[key] = val
	return nil
}

func (jc *JsonConfigContainer) Get(key string) interface{} {
	if len(key) == 0 {
		return nil
	}

	jc.RLock()
	defer jc.RUnlock()

	val, ok := jc.data[key].(interface{})
	if !ok {
		return nil
	}
	return val
}

func (jc *JsonConfigContainer) String(key string) string {
	val, ok := jc.Get(key).(string)
	if !ok {
		return ""
	}
	return val
}

func (jc *JsonConfigContainer) Strings(key string) (resp []string) {
	val := jc.Get(key)
	if val == nil {
		return nil
	}

	for _, v := range val.([]interface{}) {
		resp = append(resp, v.(string))
	}

	return resp
}

func (jc *JsonConfigContainer) Int64(key string) (int64, error) {
	val := jc.Get(key)
	if val != nil {
		if v, ok := val.(float64); ok {
			return int64(v), nil
		}
		return 0, fmt.Errorf("not a int value")
	}
	return 0, fmt.Errorf("value[%s] not exist", key)
}

func (jc *JsonConfigContainer) Int(key string) (int, error) {
	val, err := jc.Int64(key)
	if err == nil {
		return int(val), nil
	}
	return 0, err
}

//------------------------------------------------------
// json config parser
//------------------------------------------------------
type JsonConfigParser struct {
}

func (jp *JsonConfigParser) Parse(filename string) (Configer, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	data = trimComments(data)

	return jp.ParseData(data)
}

func (jp *JsonConfigParser) ParseData(data []byte) (Configer, error) {

	jc := &JsonConfigContainer{
		data: make(map[string]interface{}),
	}
	err := json.Unmarshal(data, &jc.data)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json config failed: %v", err)
	}
	return jc, nil
}
