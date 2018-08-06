package teapotlog

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/audit/jsonlog"
	. "github.com/qiniu/http/audit/proto"
	"github.com/qiniu/largefile/log"
	//	"github.com/qiniu/http/supervisor"
	"qbox.us/biz/services.v2/account"
)

type decoder struct {
	jsonlog.BaseDecoder
}

func (r *decoder) DecodeRequest(userInfo *account.UserInfo, req *http.Request) (url string, header, params M) {

	url, header, params = r.BaseDecoder.DecodeRequest(req)

	if userInfo != nil {
		token := M{
			"uid":   userInfo.Uid,
			"utype": userInfo.UserType,
			"email": userInfo.Email,
		}
		header["Token"] = token
	}
	return
}

func (r *decoder) DecodeResponse(header http.Header, bodyThumb []byte, h, params M) (resph M, body []byte) {
	return r.BaseDecoder.DecodeResponse(header, bodyThumb, h, params)
}

type Logger struct {
	w     jsonlog.LogWriter
	dec   decoder
	Mod   string
	Limit int
	Xlog  bool
}

type Config struct {
	//Supervisor supervisor.Config `json:"supervisor"`
	LogDir        string `json:"logdir"`
	LogNamePrefix string `json:"lognameprefix"`
	TimeMode      int64  `json:"timemode"` // log按时间切分模式, 单位:秒, 最小1s, 最大1天(24*3600), 且能被24*3600整除, 默认10分钟(600)
	ChunkBits     uint   `json:"chunkbits"`
	//NoXlog        uint              `json:"noxlog"`
	BodyLimit int `json:"bodylimit"`
}

func Open(module string, cfg *Config) (al *Logger, err error) {

	if cfg.BodyLimit == 0 {
		cfg.BodyLimit = 1024
	}

	logf, err := log.Open(cfg.LogDir, cfg.ChunkBits)
	if err != nil {
		err = errors.Info(err, "jsonlog.Open: largefile/log.Open failed").Detail(err)
		return
	}

	var dec decoder
	al = New(module, logf, dec, 100, false)
	return
}

func New(mod string, w jsonlog.LogWriter, dec decoder, limit int, xlog bool) *Logger {

	return &Logger{w, dec, mod, limit, xlog}
}

func (this *Logger) processBody(body []byte) (code int, bodyStr string) {
	bodyStr = string(body)
	if len(bodyStr) < 6 {
		code = 200
		return
	}

	bodyStr = strings.Replace(bodyStr, "\n", "", -1)
	bodyStr = strings.Replace(bodyStr, " ", "", -1)

	indexStart := strings.Index(bodyStr, "\"code\":")

	var codeBytes []byte
	for i := indexStart + 7; bodyStr[i] >= '0' && bodyStr[i] <= '9'; i++ {
		codeBytes = append(codeBytes, bodyStr[i])
	}

	code, err := strconv.Atoi(string(codeBytes))
	if err != nil {
		return -1, bodyStr
	}
	return code, bodyStr
}

func (this *Logger) Log(userInfo *account.UserInfo, rw *ResponseWriter, req *http.Request) (err error) {

	url_, headerM, paramsM := this.dec.DecodeRequest(userInfo, req)
	if url_ == "" { // skip
		errors.Info(err, "jsonlog.Handler: Log failed").Detail(err).Warn()
		return
	}

	var header, params, resph []byte
	if len(headerM) != 0 {
		header, _ = json.Marshal(headerM)
	}
	if len(paramsM) != 0 {
		params, _ = json.Marshal(paramsM)
		if len(params) > 4096 {
			params, _ = json.Marshal(M{"discarded": len(params)})
		}
	}

	b := bytes.NewBuffer(nil)
	startTime := rw.StartT

	startTime /= 100
	endTime := time.Now().UnixNano() / 100

	b.WriteString("REQ\t")
	b.WriteString(this.Mod)
	b.WriteByte('\t')

	b.WriteString(strconv.FormatInt(startTime, 10))
	b.WriteByte('\t')
	b.WriteString(req.Method)
	b.WriteByte('\t')
	b.WriteString(url_)
	b.WriteByte('\t')
	b.Write(header)
	b.WriteByte('\t')
	b.Write(params)
	b.WriteByte('\t')

	resphM, _ := this.dec.DecodeResponse(rw.Header(), rw.Body.Bytes(), rw.Extra, paramsM)
	if len(resphM) != 0 {
		resph, _ = json.Marshal(resphM)
	}

	code, bodyStr := this.processBody(rw.Body.Bytes())

	b.WriteString(strconv.Itoa(code))
	b.WriteByte('\t')
	b.Write(resph)
	b.WriteByte('\t')
	b.Write([]byte(bodyStr))
	b.WriteByte('\t')
	b.WriteString(strconv.FormatInt(rw.WrittenBytes, 10))
	b.WriteByte('\t')
	b.WriteString(strconv.FormatInt(endTime-startTime, 10))

	err = this.w.Log(b.Bytes())
	if err != nil {
		errors.Info(err, "jsonlog.Handler: Log failed").Detail(err).Warn()
		return
	}

	return
}

// ----------------------------------------------------------
