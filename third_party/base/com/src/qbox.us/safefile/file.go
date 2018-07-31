package safefile

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

var (
	ErrInvalidPath = errors.New("path is an invalid safe file path")
	indexFileName  = "index"
)

type SafeFile struct {
	dir  string // 实际上有两个文件：dir + /0|1
	curr int    // 0|1
	file *os.File
}

func readIndex(safePath string) (curr int, err error) {
	data, err := ioutil.ReadFile(path.Join(safePath, indexFileName))
	if err != nil {
		return
	}
	curr, err = strconv.Atoi(string(data))
	if err != nil {
		return
	}
	if curr > 1 || curr < 0 {
		err = ErrInvalidPath
	}
	return
}

func writeIndex(safePath string, curr int) (err error) {
	return ioutil.WriteFile(path.Join(safePath, indexFileName), []byte(strconv.Itoa(curr)), os.ModePerm)
}

func Open(safePath string) (f *os.File, err error) {
	curr, err := readIndex(safePath)
	if err != nil {
		return
	}
	f, err = os.Open(path.Join(safePath, strconv.Itoa(curr)))
	if err != nil {
		return
	}
	return
}

func createDataFile(safePath string, curr int) (f *os.File, err error) {
	return os.Create(path.Join(safePath, strconv.Itoa(curr)))
}

func ensureDir(safePath string, e error) (err error) {
	if e == nil {
		return
	}
	if !os.IsNotExist(e) {
		return
	}
	return os.MkdirAll(safePath, os.ModePerm)
}

//
// 这个组件已经过时。请使用 github.com/qiniu/reliable.ReadFile/WriteFile
//
func Create(safePath string) (s *SafeFile, err error) {
	curr, err := readIndex(safePath)

	err = ensureDir(safePath, err)
	if err != nil {
		return nil, err
	}

	dataFile, err := createDataFile(safePath, (1 ^ curr))
	if err != nil {
		return nil, err
	}
	return &SafeFile{safePath, curr, dataFile}, nil
}

func (s *SafeFile) Write(data []byte) (n int, err error) {
	return s.file.Write(data)
}

func (s *SafeFile) Close() (err error) {
	err = s.file.Close()
	if err != nil {
		return
	}
	return writeIndex(s.dir, (1 ^ s.curr))
}
