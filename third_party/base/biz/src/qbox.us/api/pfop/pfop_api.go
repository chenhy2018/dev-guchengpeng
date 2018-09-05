package pfop

import (
	"errors"
	"github.com/qiniu/rpc.v1"
	"net/http"
	"strconv"
	"strings"
)

type StateCode int

const (
	StateNormal StateCode = iota
	StateProcessing
)

var descs = [...]string{"waiting", "processing"}

func (s StateCode) String() string {
	if 0 <= int(s) && int(s) < len(descs) {
		return descs[s]
	}
	return ""
}

type Job struct {
	Id         string `json:"id"`
	PipelineId string `json:"pipelineId"`
	State      int32  `json:"state"`

	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	NotifyURL string `json:"notifyURL"`
	Fops      string `json:"fops"`
}

type JobsListRet struct {
	Marker string `json:"marker,omitempty"`
	Items  []*Job `json:"items"`
}

type Client struct {
	Host string
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) Client {
	client := &http.Client{Transport: t}
	return Client{Host: host, Conn: rpc.Client{client}}
}

func (c Client) PersistentFop(l rpc.Logger, pipeline, bucket, key, fops, notifyURL string, force int) (persistentId string, err error) {
	var ret struct {
		PersistentId string `json:"persistentId"`
	}
	params := map[string][]string{
		"pipeline":  {pipeline},
		"bucket":    {bucket},
		"key":       {key},
		"fops":      {fops},
		"notifyURL": {notifyURL},
		"force":     {strconv.Itoa(force)},
	}
	err = c.Conn.CallWithForm(l, &ret, c.Host+"/pfop", params)
	return ret.PersistentId, err
}

func (c Client) PersistentFopWithCallBackDelaySecs(l rpc.Logger, pipeline, bucket, key, fops, notifyURL string, force int, callbackDelaySecs int) (persistentId string, err error) {
	var ret struct {
		PersistentId string `json:"persistentId"`
	}
	params := map[string][]string{
		"pipeline":          {pipeline},
		"bucket":            {bucket},
		"key":               {key},
		"fops":              {fops},
		"notifyURL":         {notifyURL},
		"force":             {strconv.Itoa(force)},
		"callbackDelaySecs": {strconv.Itoa(callbackDelaySecs)},
	}
	err = c.Conn.CallWithForm(l, &ret, c.Host+"/pfop", params)
	return ret.PersistentId, err
}

func (c Client) ListJobs(l rpc.Logger, pipelineId, marker string, limit int) (ret JobsListRet, err error) {
	params := map[string][]string{
		"pipelineId": {pipelineId},
		"marker":     {marker},
		"limit":      {strconv.Itoa(limit)},
	}
	err = c.Conn.CallWithForm(l, &ret, c.Host+"/pipeline/lsjobs", params)
	return
}

// --------------------------------------------------------
func ParsePipelineId(id string) (zone, uid, name string, err error) {
	zone = "z0"
	if strings.HasPrefix(id, "z") {
		parts := strings.SplitN(id, ".", 3)
		if len(parts) != 3 {
			err = errors.New("invalid id")
			return
		}
		zone, uid, name = parts[0], parts[1], parts[2]
	} else {
		parts := strings.SplitN(id, ".", 2)
		if len(parts) != 2 {
			err = errors.New("invalid id")
			return
		}
		uid, name = parts[0], parts[1]
	}
	return
}

func ParseMessageId(id string) (zone, hexId string) {
	zone = "z0"
	idx := strings.Index(id, ".")
	if idx != -1 {
		zone = id[:idx]
		id = id[idx+1:]
	}
	hexId = id
	return
}

func GenPipelineId(uid, name string) (id string) {
	id = uid + "." + name
	return id
}
