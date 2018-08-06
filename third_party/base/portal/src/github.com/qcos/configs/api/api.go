package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"code.google.com/p/go.net/context"
)

type BatchFile struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

type BatchFileRes struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    FileContent `json:"data"`
}

func (p *Client) BatchConfig(files []*BatchFile) (res []*BatchFileRes, err error) {
	buf, err := json.Marshal(&struct {
		Files []*BatchFile `json:"files"`
	}{files})
	if err != nil {
		return
	}

	body := bytes.NewReader(buf)
	resp, err := p.Client.Do(context.Background(), "POST", "/v1/config/batch", "application/json", body, len(buf))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("resp err: %d", resp.StatusCode)
		return
	}

	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	err = dec.Decode(&struct {
		Files *[]*BatchFileRes `json:"files"`
	}{&res})
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		return
	}

	if len(files) != len(res) {
		err = fmt.Errorf("config files res not matched req:%d != res:%d\n", len(files), len(res))
	}
	return
}
