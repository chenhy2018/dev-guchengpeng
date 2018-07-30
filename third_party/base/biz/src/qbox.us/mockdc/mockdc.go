package mockdc

import (
	"crypto/sha1"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/facebookgo/httpdown"
	"github.com/qiniu/log.v1"

	"qbox.us/api"
	"qbox.us/errors"
	"qbox.us/net/httputil"
)

type Config struct {
	Key []byte
}

type Service struct {
	*Config
	Dir string
}

func New(cfg *Config) (*Service, error) {
	return &Service{cfg, ""}, nil
}

func (s *Service) getFilePath(key string) string {
	return path.Join(s.Dir, key)
}

func (s *Service) get(w http.ResponseWriter, req *http.Request) {
	key := strings.Split(req.URL.Path, "/")[2]
	log.Info(key)
	fileName := s.getFilePath(key)
	f, err := os.Open(fileName)
	if err != nil {
		httputil.Error(w, api.EFunctionFail)
		return
	}
	defer f.Close()

	info, err := os.Stat(fileName)
	if err != nil {
		httputil.Error(w, api.EFunctionFail)
		return
	}
	fsize := info.Size()
	w.Header().Set("Content-Length", strconv.FormatInt(fsize, 10))
	w.WriteHeader(200)
	io.Copy(w, f)
}

func (s *Service) rangeGet(w http.ResponseWriter, req *http.Request) {
	query := strings.Split(req.URL.Path, "/")
	if len(query) < 5 {
		httputil.ReplyError(w, "url is invalid.", 400)
	}
	key := query[2]
	from, err := strconv.ParseInt(query[3], 10, 64)
	if err != nil {
		httputil.Error(w, err)
	}
	to, err := strconv.ParseInt(query[4], 10, 64)
	if err != nil {
		httputil.Error(w, err)
	}
	log.Info(key, from, to)
	fileName := s.getFilePath(key)
	f, err := os.Open(fileName)
	if err != nil {
		httputil.Error(w, api.EFunctionFail)
		return
	}
	defer f.Close()

	info, err := os.Stat(fileName)
	if err != nil {
		httputil.Error(w, api.EFunctionFail)
		return
	}
	fsize := info.Size()
	if from < 0 {
		from = 0
	}
	if to > fsize {
		to = fsize
	}
	if from >= to {
		httputil.Error(w, api.EInvalidArgs)
		return
	}

	w.Header().Set("Content-Length", strconv.FormatInt(to-from, 10))
	w.WriteHeader(200)
	f.Seek(from, 0)
	io.CopyN(w, f, to-from)

}

//
// Post /set/<Key>/sha1/<checksum>
// Content-Type: application/octet-stream
// Body: <Data>
//
func (s *Service) set(w http.ResponseWriter, req *http.Request) {
	query := strings.Split(req.URL.Path[1:], "/")
	if len(query) < 2 {
		httputil.ReplyWithCode(w, api.InvalidArgs)
		return
	}

	key := query[1]
	log.Info(key)

	checksum := ""
	if len(query) >= 4 && query[2] == "sha1" {
		checksum = query[3]
	}

	if req.ContentLength < 0 {
		log.Println("length is required.")
		w.WriteHeader(http.StatusLengthRequired)
		return
	}
	fileName := s.getFilePath(key)
	f, err := os.Create(fileName)
	if err != nil {
		httputil.Error(w, api.EFunctionFail)
		return
	}
	defer f.Close()
	defer req.Body.Close()

	finish := false
	defer func() {
		if !finish {
			os.Remove(f.Name())
		}
	}()

	var n int64
	if checksum == "" {
		n, err = io.Copy(f, req.Body)
	} else {
		h := sha1.New()
		multiWriter := io.MultiWriter(f, h)
		n, err = io.Copy(multiWriter, req.Body)

		if err != nil {
			err = errors.Info(api.EFunctionFail, "Service.set: io.Copy failed").Detail(err)
		} else if base64.URLEncoding.EncodeToString(h.Sum(nil)) != string(checksum) {
			err = errors.Info(api.EDataVerificationFail, "Service.set: checksum failed").Detail(err)
		}
	}
	if err != nil {
		httputil.Error(w, err)
		return
	}
	if n != req.ContentLength {
		err := errors.Info(api.EFunctionFail, "Service.set: io.Copy n != cb").Detail(err)
		httputil.Error(w, err)
		return
	}
	finish = true

	w.WriteHeader(http.StatusOK)
}

func (s *Service) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/get/", func(w http.ResponseWriter, req *http.Request) {
		s.get(w, req)
	})
	mux.HandleFunc("/rangeGet/", func(w http.ResponseWriter, req *http.Request) {
		s.rangeGet(w, req)
	})
	mux.HandleFunc("/set/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.set(w, req)
	})
}

func (s *Service) Run(addr string, dir string) error {
	s.Dir = dir
	mux := http.NewServeMux()
	s.RegisterHandlers(mux)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	hd := &httpdown.HTTP{}
	err := httpdown.ListenAndServe(server, hd)
	return err
}

type Stopper interface {
	Stop() error
}

func (s *Service) Run2(addr string, dir string) (stopper Stopper, err error) {
	s.Dir = dir
	mux := http.NewServeMux()
	s.RegisterHandlers(mux)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	hd := &httpdown.HTTP{}
	stopper, err = hd.ListenAndServe(server)
	if err != nil {
		return
	}
	return
}
