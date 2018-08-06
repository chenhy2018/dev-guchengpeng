package bufio

import (
	"bytes"
	"io"
	"log"
	"sync"
	"time"
)

type Client struct {
	size, max  int64
	minD, maxD time.Duration
	lastT      time.Time
	buf        *bytes.Buffer
	post       func(io.Reader, int) error
	debug      bool
	*sync.Mutex
}

func NewClient(max int64, post func(io.Reader, int) error, debug bool) *Client {
	if max == 0 {
		max = 1024 * 1024
	}
	return &Client{
		size:  1,
		max:   max,
		minD:  time.Millisecond * 500,
		maxD:  time.Second,
		lastT: time.Now(),
		post:  post,
		debug: debug,
		Mutex: new(sync.Mutex),
	}
}

func (r *Client) Write(bs []byte) error {
	r.Lock()

	buf := r.buf
	if buf == nil {
		buf = bytes.NewBuffer(nil)
		r.buf = buf
	}
	_, _ = buf.Write(bs)
	if int64(buf.Len()) >= r.size {
		r.sync()
	}
	r.Unlock()
	return nil
}

func (r *Client) Copy(reader io.Reader) error {
	r.Lock()

	buf := r.buf
	if buf == nil {
		buf = bytes.NewBuffer(nil)
		r.buf = buf
	}
	io.Copy(buf, reader)
	if int64(buf.Len()) >= r.size {
		r.sync()
	}
	r.Unlock()
	return nil
}

func (r *Client) sync() error {
	size := r.size
	buf := r.buf
	r.buf = nil
	go func(buf *bytes.Buffer, size int64) {
		if err := r.post(buf, buf.Len()); err != nil {
			if r.debug {
				log.Printf("[DEBUG] -probe- bufio.client post failed: %s\n", err.Error())
			}
		} else {
			if r.debug {
				log.Printf("[DEBUG] -probe- bufio.client post done.\n")
			}
		}

	}(buf, size)

	now := time.Now()
	if now.Sub(r.lastT) < r.minD && r.size < r.max {
		r.size = r.size * 2
	} else if now.Sub(r.lastT) > r.maxD && r.size > 1 {
		r.size = r.size / 2
	}
	r.lastT = now
	return nil
}

func (r *Client) Close() error {
	r.Lock()
	defer r.Unlock()
	r.sync()
	return nil
}
