package collector

import (
	"bytes"
	"os"
	"path/filepath"
	"time"

	"qiniu.com/probe/common/bufio"
	"qiniu.com/probe/common/protocol/line"
)

var EXEC string = ""

func init() {
	EXEC = filepath.Base(os.Args[0])
}

func Mark(
	measurement string,
	tim time.Time,
	tagK []string, tagV []string,
	fieldK []string, fieldV []interface{}) {

	MarkDirect(measurement+"."+EXEC, tim, tagK, tagV, fieldK, fieldV)
}

func MarkDirect(
	measurement string,
	tim time.Time,
	tagK []string, tagV []string,
	fieldK []string, fieldV []interface{}) {

	if DefaultCollector != nil {
		DefaultCollector.Mark(measurement, tim, tagK, tagV, fieldK, fieldV)
	}
}

var DefaultCollector *Collector

type Config struct {
	Urls []string
	Max  int64
}

type Collector struct {
	*bufio.Client
	Config
}

func NewCollector(config Config) *Collector {
	return &Collector{
		Client: bufio.NewClient(config.Max, NewClient(config.Urls).Post, false),
		Config: config,
	}
}

func Init(config Config) {
	DefaultCollector = NewCollector(config)
}

func (r *Collector) Mark(
	measurement string,
	tim time.Time,
	tagK, tagV []string,
	fieldK []string, fieldV []interface{}) {

	buf := bytes.NewBuffer(nil)
	line.MarshalArrayWrite(buf, measurement, tim, tagK, tagV, fieldK, fieldV)
	buf.WriteByte('\n')
	r.Client.Write(buf.Bytes())
}
