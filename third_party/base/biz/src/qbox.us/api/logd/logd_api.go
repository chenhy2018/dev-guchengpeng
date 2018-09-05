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

func (r Service) Get(bucket string) (entry Entry, err error) {

	params := map[string][]string{
		"bucket": {bucket},
	}
	_, err = r.Conn.CallWithForm(&entry, r.Host+"/get", params)
	return
}

func (r Service) Set(bucket string, formatID int, saveBucket string) (err error) {

	id := 0
	if formatID > 0 {
		id = formatID
	}
	params := map[string][]string{
		"bucket":     {bucket},
		"formatid":   {strconv.Itoa(id)},
		"savebucket": {saveBucket},
	}
	_, err = r.Conn.CallWithForm(nil, r.Host+"/set", params)
	return
}

func (r Service) SetEx(bucket string, formatID int, saveBucket string, compress bool, blockSize int) (err error) {

	id := 0
	if formatID > 0 {
		id = formatID
	}
	params := map[string][]string{
		"bucket":     {bucket},
		"formatid":   {strconv.Itoa(id)},
		"savebucket": {saveBucket},
		"compress":   {strconv.FormatBool(compress)},
		"blocksize":  {strconv.Itoa(blockSize)},
	}
	_, err = r.Conn.CallWithForm(nil, r.Host+"/set", params)
	return
}

func (r Service) Enable(bucket string) (err error) {

	params := map[string][]string{
		"bucket": {bucket},
	}
	_, err = r.Conn.CallWithForm(nil, r.Host+"/enable", params)
	return
}

func (r Service) Disable(bucket string) (err error) {

	params := map[string][]string{
		"bucket": {bucket},
	}
	_, err = r.Conn.CallWithForm(nil, r.Host+"/disable", params)
	return
}

func (r Service) Delete(bucket string) (err error) {

	params := map[string][]string{
		"bucket": {bucket},
	}
	_, err = r.Conn.CallWithForm(nil, r.Host+"/delete", params)
	return
}
