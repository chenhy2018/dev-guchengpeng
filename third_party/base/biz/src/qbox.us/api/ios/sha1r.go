package ios

import (
	"crypto/sha1"
	"io"
	"github.com/qiniu/log.v1"
)

// ----------------------------------------------------------

type Runner interface {
	Run(task func())
}

type Goroutine struct{}

func (r Goroutine) Run(task func()) {
	go task()
}

// ----------------------------------------------------------

type Sha1Calcer struct {
	keys chan []byte
	src  *CancelableReader
}

func NewSha1Calcer(f1 io.Reader, runner Runner, max int, chunkSize int) Sha1Calcer {

	if f2, ok := f1.(*MetricReader); ok {
		f1 = f2.Source
	}

	keys := make(chan []byte, max)
	f := &CancelableReader{Source: f1}
	calcer := Sha1Calcer{keys, f}
	runner.Run(func() { calcer.run(f, chunkSize) })
	return calcer
}

func (r Sha1Calcer) Cancel(err error) {

	log.Debug("Sha1Calcer.Cancel")

	r.src.Cancel()
	for err == nil {
		_, err = r.Get()
	}

	log.Debug("Sha1Calcer.Cancel - done")
}

func (r Sha1Calcer) Get() (keys []byte, err error) {

	key := <-r.keys
	for {
		if key == nil {
			err = io.EOF
			break
		}
		keys = append(keys, key...)
		select {
		case key = <-r.keys:
		default:
			return
		}
	}
	return
}

func (r Sha1Calcer) run(f io.Reader, chunkSize1 int) {

	h := sha1.New()
	chunkSize := int64(chunkSize1)
	for {
		n, err := io.CopyN(h, f, chunkSize)
		if n > 0 {
			r.keys <- h.Sum(nil)
		}
		if err != nil {
			if err != io.EOF {
				log.Warn("Sha1Calc failed:", err)
			}
			break
		}
		h.Reset()
	}
	r.keys <- nil
}

// ----------------------------------------------------------
