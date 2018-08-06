package mockbd

import (
	"crypto/sha1"
	"encoding/base64"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"qbox.us/auditlog2"
	"qbox.us/cc/time"
	"qbox.us/rpc"
	"qbox.us/servestk"
	"strconv"
	"sync/atomic"
)

var tmpFileIndex uint64

func getPath(dirs []string, key string) (string, string, string) {

	dir := dirs[(int(key[0])+int(key[1])*256)%len(dirs)]
	path := filepath.Join(dir, key[2:4], key[4:7])
	return dir, path, filepath.Join(path, key)
}

func put(w http.ResponseWriter, req *http.Request, dirs []string) {

	xl := xlog.New(w, req)
	totalTm := time.NewMeter()
	defer func() {
		xl.Info("put: total(100ns):", totalTm.Elapsed()/100)
	}()

	v := req.URL.Query()
	l := v.Get("len")
	k := v.Get("key")
	xl.Info("put: ...", k, l)

	if l == "" {
		w.WriteHeader(400)
		return
	}

	length, err := strconv.Atoi(l)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	readTm := time.NewMeter()
	buff := make([]byte, length)
	n, err := io.ReadFull(req.Body, buff)
	xl.Info("put: io.ReadFull(100ns):", readTm.Elapsed()/100)
	if n != length || err != nil {
		w.WriteHeader(400)
		return
	}

	hash := sha1.New()
	hash.Write(buff)
	h := hash.Sum(nil)
	key := base64.URLEncoding.EncodeToString(h)
	if k != "" && k != key {
		xl.Error("put: k :", k, key)
		w.WriteHeader(412)
		return
	}

	root, dir, path := getPath(dirs, key)
	xl.Debug("put: getPath", root, dir, path)

	if err = os.MkdirAll(dir, 0777); err != nil {
		xl.Error("put: os.MkdirAll", dir, err)
		w.WriteHeader(500)
		return
	}

	index := atomic.AddUint64(&tmpFileIndex, 1)
	tmpPath := filepath.Join(root, "tmp", strconv.FormatUint(index, 10))

	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		xl.Error("put: os.Create", tmpPath, err)
		w.WriteHeader(500)
		return
	}

	writeTm := time.NewMeter()
	n, err = tmpFile.Write(buff)
	if n != length || err != nil {
		xl.Error("put: tmpFile.Write:", tmpFile.Name(), err, n, length)
		tmpFile.Close()
		os.Remove(tmpPath)
		w.WriteHeader(500)
		return
	}
	xl.Info("put: tmpFile.Write(100ns):", writeTm.Elapsed()/100)
	tmpFile.Close()

	if err = os.Rename(tmpPath, path); err != nil {
		xl.Error("put: os.Rename", tmpPath, path, err)
		os.Remove(tmpPath)
		w.WriteHeader(500)
		return
	}

	w1 := rpc.ResponseWriter{w}
	w1.ReplyWithParam(200, "application/octet-stream", h)
}

func get(w http.ResponseWriter, req *http.Request, dirs []string) {

	xl := xlog.New(w, req)
	totalTm := time.NewMeter()
	defer func() {
		xl.Info("get: total(100ns):", totalTm.Elapsed()/100)
	}()

	v := req.URL.Query()
	k := v.Get("key")
	f := v.Get("from")
	t := v.Get("to")
	xl.Info("get: ", k, f, t)

	if k == "" {
		w.WriteHeader(400)
		return
	}

	_, _, path := getPath(dirs, k)
	xl.Debug("get: getPath", path)
	//	file, err := os.Open(dir + "/" + k)
	file, err := os.Open(path)
	if err != nil {
		xl.Error("get: not exist", k, err)
		w.WriteHeader(404)
		return
	}
	defer file.Close()

	//	fileinfo, err := os.Stat(dir + "/" + k)
	fileinfo, err := os.Stat(path)
	if err != nil {
		xl.Error("get: not exist", k, err)
		w.WriteHeader(404)
		return
	}

	//	buff := make([]byte, int(fileinfo.Size))
	//	n, err := io.ReadFull(file, buff)
	//	if n != int(fileinfo.Size) || err != nil {
	//		log.Error("read file err", k, err)
	//		w.WriteHeader(500)
	//		return
	//	}
	from := 0
	to := int(fileinfo.Size())
	if f != "" {
		from, err = strconv.Atoi(f)
		if err != nil {
			w.WriteHeader(500)
			return
		}
	}
	if t != "" {
		to, err = strconv.Atoi(t)
		if err != nil {
			w.WriteHeader(500)
			return
		}
	}
	if to > int(fileinfo.Size()) {
		to = int(fileinfo.Size())
	}

	if from > to {
		w.WriteHeader(400)
		return
	}

	xl.Info("get: key :", k)
	w1 := rpc.ResponseWriter{w}

	file.Seek(int64(from), 0)

	copyTm := time.NewMeter()
	n, err := io.CopyN(w1, file, int64(to-from))
	xl.Info("get: io.CopyN(100ns):", copyTm.Elapsed()/100)
	if err != nil && n == 0 {
		w.WriteHeader(500)
		return
	}
	//	w1.ReplyWithParam(200, "application/octet-stream", buff[from:to])
}

// -----------------------------------------------------------

func initTmpDir(dirs []string) error {

	for _, dir := range dirs {
		tmpDir := filepath.Join(dir, "tmp")
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Error("initTmpDir: os.RemoveAll", err, dir)
			// not fatal error
		}

		if err := os.MkdirAll(tmpDir, 0777); err != nil {
			return err
		}
	}
	return nil
}

// -----------------------------------------------------------

type Config struct {
	Addr   string
	Dirs   []string
	LogCfg *auditlog2.Config
}

func RegisterHandlers(mux1 *http.ServeMux, cfg *Config) (lh auditlog2.Instance, err error) {

	lh, err = auditlog2.Open("SBD", cfg.LogCfg, nil)
	if err != nil {
		return
	}
	mux := servestk.New(mux1, lh.Handler())

	mux.HandleFunc("/put", func(w http.ResponseWriter, req *http.Request) { put(w, req, cfg.Dirs) })
	mux.HandleFunc("/get", func(w http.ResponseWriter, req *http.Request) { get(w, req, cfg.Dirs) })
	return
}

func Run(cfg *Config) error {

	mux, err := Init(cfg)
	if err != nil {
		return err
	}
	return http.ListenAndServe(cfg.Addr, mux)
}

func Init(cfg *Config) (mux *http.ServeMux, err error) {

	if err = initTmpDir(cfg.Dirs); err != nil {
		return
	}

	mux = http.NewServeMux()
	_, err = RegisterHandlers(mux, cfg)
	return
}
