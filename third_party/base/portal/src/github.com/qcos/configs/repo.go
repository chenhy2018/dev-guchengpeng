package configs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/howeyc/fsnotify"

	"qiniupkg.com/x/log.v7"

	"github.com/teapots/teapot"
)

type repoFile struct {
	Name    string `json:"name,omitempty"`
	Version int    `json:"version,omitempty"`
	Content string `json:"content,omitempty"`
}

type repo struct {
	RepoConfig
	mux        sync.RWMutex
	log        teapot.Logger
	configFile string
}

func (r *repo) Init(log teapot.Logger) (err error) {
	r.log = log
	r.configFile = filepath.Join(Env.RepoPath, ".config.json")
	// TODO cache config file

	err = r.loadConfig()
	if err == nil {
		go r.notifyAndReload()
	}
	return
}

func (r *repo) LoadFile(app, name string, version int) (file repoFile, err error) {

	appPath := filepath.Join(Env.RepoPath, app)
	name = filepath.Join("/", name) // clean path

	f, err := os.Open(appPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = ErrNotExists
			return
		}
		r.log.Error("RepoPath load err", err)
		return
	}

	var dirs []os.FileInfo
	func() {
		defer f.Close()
		dirs, err = f.Readdir(-1)
	}()
	if err != nil {
		r.log.Error("RepoPath, Readdir err", err)
		return
	}

	versions := make([]int, 0, len(dirs))
	for _, dir := range dirs {
		i, err := strconv.ParseInt(dir.Name(), 10, 64)
		if err != nil {
			continue
		}
		ver := int(i)

		if ver > version {
			continue
		}

		versions = append(versions, ver)
	}

	if len(versions) == 0 {
		err = ErrNotExists
		return
	}

	sort.Sort(sort.Reverse(sort.IntSlice(versions)))

	for _, ver := range versions {
		filename := filepath.Join(appPath, strconv.FormatInt(int64(ver), 10), name)
		b, er := ioutil.ReadFile(filename)
		if er != nil {
			if !os.IsNotExist(er) {
				log.Error("Readfile", filename, er)
			}
			continue
		}
		file.Name = strings.TrimPrefix(name, "/")
		file.Version = ver
		file.Content = string(b)
		return
	}

	err = ErrNotExists
	return
}

func (r *repo) AccessRepo(key string) (info *AuthInfo, err error) {
	r.mux.RLock()
	defer r.mux.RUnlock()
	v, ok := r.Auth[key]
	if !ok {
		err = fmt.Errorf("not found AccessKey: %v", key)
		return
	}

	info = &AuthInfo{}
	info.Service = v.Service
	info.AccessKey = key
	info.SecretKey = v.Secret
	return
}

func (r *repo) loadConfig() (err error) {
	b, err := ioutil.ReadFile(r.configFile)
	if err != nil {
		r.log.Error("RepoConfig read:", r.configFile, err)
		return
	}

	r.mux.Lock()
	defer r.mux.Unlock()
	err = json.Unmarshal(b, &r.RepoConfig)
	if err != nil {
		r.log.Error("RepoConfig unmarshal:", r.configFile, err)
		return
	}

	r.log.Info("RepoConfig loaded:", r.configFile)
	return
}

func (r *repo) notifyAndReload() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		r.log.Error("notifyAndReload fsnotify.NewWatcher:", err)
		return
	}

	err = watcher.Watch(r.configFile)
	if err != nil {
		r.log.Error("notifyAndReload watcher.Watch:", err)
		return
	}

	for {
		select {
		case evt := <-watcher.Event:
			r.log.Info("notifyAndReload", evt)
			r.loadConfig()
		}
	}
}
