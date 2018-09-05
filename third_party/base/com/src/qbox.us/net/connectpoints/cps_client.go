package connectpoints

import (
	"bufio"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"qbox.us/cc"
	"github.com/qiniu/log.v1"
	"strconv"
	"strings"
)

// -------------------------------------------------------

type requestImpl struct {
	Method int64
	Body   []byte
}

type Request struct {
	*requestImpl
}

func (r Request) Decode(data interface{}) (err error) {

	body := cc.NewBytesReader(r.Body)
	return gob.NewDecoder(body).Decode(data)
}

// -------------------------------------------------------

type Handler interface {
	Process(req Request)
}

// -------------------------------------------------------

type ServeMux struct {
	handlers map[int64]func(req Request)
	defaulth func(req Request)
}

func NewServeMux(defaulth func(req Request)) *ServeMux {
	return &ServeMux{
		make(map[int64]func(req Request)),
		defaulth,
	}
}

func (mux *ServeMux) HandleFunc(method int64, handler func(req Request)) {
	mux.handlers[method] = handler
}

func (mux *ServeMux) Process(req Request) {
	if handle, ok := mux.handlers[req.Method]; ok {
		handle(req)
	} else if mux.defaulth != nil {
		mux.defaulth(req)
	} else {
		log.Println("Unknown event:", req.Method)
	}
}

// -------------------------------------------------------

type Client struct {
	source io.ReadCloser
}

func Connect(url1 string, eventTypes int64) (c Client, err error) {

	url2, err := url.ParseRequestURI(url1)
	if err != nil {
		log.Println("ParseRequestURI failed:", err)
		return
	}

	if strings.Index(url2.Host, ":") == -1 { // default port: 80
		url2.Host += ":80"
	}

	conn, err := net.Dial("tcp", url2.Host)
	if err != nil {
		log.Println("Dial failed:", err)
		return
	}
	io.WriteString(conn, "CONNECT "+url2.Path+strconv.FormatInt(eventTypes, 16)+" HTTP/1.0\n\n")

	// Require successful HTTP response
	// before switching to RPC protocol.
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == connected {
		return Client{conn}, nil
	}
	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	log.Println("Connect failed:", err)
	conn.Close()
	return
}

func (c Client) Close() (err error) {

	return c.source.Close()
}

func (c Client) RunLoop(handler Handler) (err error) {

	var method int64
	var l uint16
	var hdr [10]byte
	var body []byte
	var req Request
	for {
		_, err = io.ReadFull(c.source, hdr[:])
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("RunLoop: read header failed -", err)
			return
		}
		method = int64(binary.LittleEndian.Uint64(hdr[:]))
		l = binary.LittleEndian.Uint16(hdr[8:])
		body = make([]byte, l)
		_, err = io.ReadFull(c.source, body)
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			log.Println("RunLoop: read body failed -", err)
			return
		}
		req = Request{&requestImpl{method, body}}
		//		log.Println("Request:", method, body)
		handler.Process(req)
	}

	log.Println("Connection closed!")
	return nil
}

// -------------------------------------------------------
