package api

import (
	"fmt"
	"strconv"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
)

type Client struct {
	client rpc.Client
	Host   string
}

var (
	ErrNoSuchKey = httputil.NewError(612, "no such key")
)

type SidReverseItem struct {
	Sid  uint64   `json:"sid"`
	Fids []uint64 `json:"fids"`
}

func NewClient(host string) *Client {
	return &Client{
		Host:   host,
		client: rpc.DefaultClient,
	}
}

func (self *Client) Put(l rpc.Logger, item *SidReverseItem) error {
	sid := item.Sid
	fidsstr := EncodeFids(item.Fids)
	params := map[string][]string{
		"sid":     {strconv.FormatUint(sid, 10)},
		"fidsstr": {fidsstr},
	}
	url := fmt.Sprintf("%v/put", self.Host)
	return self.client.CallWithForm(l, nil, url, params)
}

func (self *Client) Get(l rpc.Logger, sid uint64) (*SidReverseItem, error) {
	url := fmt.Sprintf("%v/get/%v", self.Host, sid)
	item := &SidReverseItem{}
	err := self.client.Call(l, item, url)
	if err != nil {
		code := httputil.DetectCode(err)
		if code == 612 {
			return nil, ErrNoSuchKey
		}
		return nil, err
	}
	return item, nil
}

func (self *Client) Delete(l rpc.Logger, sid uint64) error {
	url := fmt.Sprintf("%v/delete/%v", self.Host, sid)
	err := self.client.Call(l, nil, url)
	return err
}

func (self *Client) Update(l rpc.Logger, sid, firstFid, lastFid uint64) error {
	params := map[string][]string{
		"sid":      {strconv.FormatUint(sid, 10)},
		"firstfid": {strconv.FormatUint(firstFid, 10)},
		"lastfid":  {strconv.FormatUint(lastFid, 10)},
	}
	url := fmt.Sprintf("%v/update", self.Host)
	return self.client.CallWithForm(l, nil, url, params)
}
