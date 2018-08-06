package job

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type Service struct {
	Host string
	Conn *http.Client
}

func New(host string, transport http.RoundTripper) *Service {
	return &Service{host, &http.Client{Transport: transport}}
}

// ==================================================================

type JobInfo struct { // http POST body
	EncodedEntryURI string `json:"encoded_entry_uri"`
	Name            []byte `json:"name"`
	FopParams       string `json:"fop_params"`
}

type Output struct {
	JobInfos []JobInfo `json:"job_infos"`
	FileName []byte    `json:"file_name"`
	Expires  int64     `json:"expires"`
	Uid      uint32    `json:"uid"`
}

func (s *Service) GetJob(jobId string) (output Output, err error) {
	resp, err := s.Conn.Get(s.Host + "/getJob/" + jobId)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode/100 != 2 {
		err = errors.New(string(data))
		return
	}
	err = json.Unmarshal(data, &output)
	return
}
