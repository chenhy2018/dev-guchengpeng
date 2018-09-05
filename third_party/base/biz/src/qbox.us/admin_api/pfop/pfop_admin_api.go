package pfop

import (
	"net/http"

	"github.com/qiniu/rpc.v1"
)

type Pipeline struct {
	Id    string `json:"id"`
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

type IdRet struct {
	Id string `json:"id"`
}

type StatRet struct {
	Todo  int `json:"todo"`
	Doing int `json:"doing"`
}

type Client struct {
	Host string
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) Client {
	client := &http.Client{Transport: t}
	return Client{Host: host, Conn: rpc.Client{client}}
}

func (c *Client) CreatePipeline(l rpc.Logger, uid, name string) (pipelineId string, err error) {
	var ret IdRet
	params := map[string][]string{"uid": {uid}, "name": {name}}
	err = c.Conn.CallWithForm(l, &ret, c.Host+"/pipeline/new", params)
	return ret.Id, err
}

func (c *Client) RemovePipeline(l rpc.Logger, pipelineId string) (err error) {
	params := map[string][]string{"id": {pipelineId}}
	return c.Conn.CallWithForm(l, nil, c.Host+"/pipeline/rm", params)
}

func (c *Client) ListPipelines(l rpc.Logger, uid string) (ret []*Pipeline, err error) {
	params := map[string][]string{"uid": {uid}}
	err = c.Conn.CallWithForm(l, &ret, c.Host+"/pipeline/ls", params)
	return
}

func (c *Client) StatPipeline(l rpc.Logger, pipelineId string) (ret *StatRet, err error) {
	param := map[string][]string{"id": {pipelineId}}
	err = c.Conn.CallWithForm(l, &ret, c.Host+"/pipeline/stat", param)

	return
}
