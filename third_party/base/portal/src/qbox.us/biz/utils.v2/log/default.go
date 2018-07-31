package log

import (
	"encoding/base64"
	"encoding/binary"
	"log"
	"os"
	"time"

	"github.com/teapots/teapot"
)

var X teapot.LoggerAdv

func init() {
	logOut := log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)

	// 每机器每进程的 RequestId
	hostname, _ := os.Hostname()

	logger := teapot.NewWithId(logOut, hostname+"]["+NewReqId())
	logger.SetLineInfo(true)
	logger.SetShortLine(true)
	logger.SetColorMode(false)
	logger.EnableLogStack(teapot.LevelCritical)

	X = logger
}

var pid = uint32(os.Getpid())

func NewReqId() string {
	var b [12]byte
	binary.LittleEndian.PutUint32(b[:], pid)
	binary.LittleEndian.PutUint64(b[4:], uint64(time.Now().UnixNano()))
	return base64.URLEncoding.EncodeToString(b[:])
}
