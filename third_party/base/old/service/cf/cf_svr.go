package cf

import (
	"http"
	"io"
	"launchpad.net/gobson/bson"
	"launchpad.net/mgo"
	"os"
	"qbox.us/log"
	"qbox.us/webutil"
	"strconv"
	"strings"
)

type M map[string]interface{}

type Config struct {
	Mgo mgo.Collection
}

type ColumnFamily struct {
	Config
}

func New(cfg *Config) (p *ColumnFamily, err os.Error) {
	p = &ColumnFamily{*cfg}
	err = p.Mgo.EnsureIndex(mgo.Index{Key: []string{"key"}})
	if err != nil {
		log.Error("fs.EnsureIndex kye:", err)
		return
	}
	return
}

func (p *ColumnFamily) get(w http.ResponseWriter, req *http.Request) {
	query := strings.Split(req.RawURL[1:], "/", -1)
	log.Debug("get query:", query)
	if len(query) != 5 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var key, col string
	for i := 1; i+1 < len(query); i += 2 {
		switch query[i] {
		case "key":
			key = query[i+1]
		case "col":
			col = query[i+1]
		}
	}
	if key == "" || col == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var ret bson.M
	col = "c_" + col
	err := p.Mgo.Find(M{"key": key}).Select(M{col: 1}).One(&ret)
	log.Debug("ret:", err, ret, len(ret))
	if ret == nil || len(ret) <= 1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Write(ret[col].([]byte))
}

func (p *ColumnFamily) set(w http.ResponseWriter, req *http.Request) {
	query := strings.Split(req.RawURL[1:], "/", -1)
	log.Debug("set query:", query)
	if len(query) != 7 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var key, col string
	var length int
	for i := 1; i+1 < len(query); i += 2 {
		switch query[i] {
		case "key":
			key = query[i+1]
		case "col":
			col = query[i+1]
		case "len":
			var err os.Error
			length, err = strconv.Atoi(query[i+1])
			if err != nil {
				log.Info("len error:", query[i+1], length)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
	}

	log.Debug("request:", key, col, length)
	if key == "" || col == "" || length <= 0 || length > 1024*1024 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	buff := make([]byte, length)
	n, err := io.ReadFull(req.Body, buff)
	if n != len(buff) || err != nil {
		log.Warn("bad data:", n, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	col = "c_" + col
	_, err = p.Mgo.Upsert(M{"key": key}, M{"$set": M{"key": key, col: buff}})
	if err != nil {
		log.Warn("upsert err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (p *ColumnFamily) del(w http.ResponseWriter, req *http.Request) {
	query := strings.Split(req.RawURL[1:], "/", -1)
	log.Debug("del query:", query)
	if len(query) != 5 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var key, col string
	for i := 1; i+1 < len(query); i += 2 {
		switch query[i] {
		case "key":
			key = query[i+1]
		case "col":
			col = query[i+1]
		}
	}
	if key == "" || col == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	col = "c_" + col
	err := p.Mgo.Update(M{"key": key}, M{"$unset": M{col: 1}})
	log.Debug(err)
}

func RegisterHandlers(mux *http.ServeMux, cfg *Config) os.Error {
	log.Debug("registerHandlers")
	p, err := New(cfg)
	if err != nil {
		log.Warn("new cfg error:", err)
		return err
	}

	mux.HandleFunc("/get/", webutil.SafeHandler(func(w http.ResponseWriter, req *http.Request) { p.get(w, req) }))
	mux.HandleFunc("/set/", webutil.SafeHandler(func(w http.ResponseWriter, req *http.Request) { p.set(w, req) }))
	mux.HandleFunc("/del/", webutil.SafeHandler(func(w http.ResponseWriter, req *http.Request) { p.del(w, req) }))
	return nil
}

func Run(addr string, cfg *Config) os.Error {

	mux := http.DefaultServeMux
	err := RegisterHandlers(mux, cfg)
	if err != nil {
		return err
	}
	return http.ListenAndServe(addr, mux)
}
