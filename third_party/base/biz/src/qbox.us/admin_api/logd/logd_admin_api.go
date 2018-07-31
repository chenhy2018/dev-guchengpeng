package logd

import (
	"net/http"
	"strconv"

	"qbox.us/rpc"
)

// ------------------------------------------------------------------------------------------

type Service struct {
	Host string
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{host, rpc.Client{client}}
}

// ------------------------------------------------------------------------------------------

type Entry struct {
	IsSave     bool   `json:"issave"`
	Format     string `json:"format"`
	SaveBucket string `json:"savebucket"`
	Compress   bool   `json:"compress"`
	BlockSize  int    `json:"blocksize"`
}

type EntryEx struct {
	Uid        uint32 `json:"uid"`
	Bucket     string `json:"bucket"`
	Entry      Entry  `json:"entry"`
	UpdateTime string `json:"update_time"`
}

func (r Service) Group() (entrys []EntryEx, err error) {

	_, err = r.Conn.Call(&entrys, r.Host+"/admin/group")
	return
}

func (r Service) Set(uid uint32, bucket string, formatID int, saveBucket string) (err error) {

	id := 0
	if formatID > 0 {
		id = formatID
	}
	params := map[string][]string{
		"uid":        {strconv.FormatUint(uint64(uid), 10)},
		"bucket":     {bucket},
		"formatid":   {strconv.Itoa(id)},
		"savebucket": {saveBucket},
	}
	_, err = r.Conn.CallWithForm(nil, r.Host+"/admin/set", params)
	return
}

func (r Service) SetEx(uid uint32, bucket string, formatID int, saveBucket string, compress bool, blockSize int) (err error) {

	id := 0
	if formatID > 0 {
		id = formatID
	}
	params := map[string][]string{
		"uid":        {strconv.FormatUint(uint64(uid), 10)},
		"bucket":     {bucket},
		"formatid":   {strconv.Itoa(id)},
		"savebucket": {saveBucket},
		"compress":   {strconv.FormatBool(compress)},
		"blocksize":  {strconv.Itoa(blockSize)},
	}
	_, err = r.Conn.CallWithForm(nil, r.Host+"/admin/set", params)
	return
}
