package connectpoints

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"net/http"
	"github.com/qiniu/log.v1"
	"strconv"
	"strings"
	"sync"
)

var connected = "200 Connected"

// -------------------------------------------------------------------------

var evbufFull = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00}

type connection struct {
	io.Writer
	eventTypes int64
	events     chan []byte
	efull      bool
}

func (c *connection) fire(ev []byte) bool {

	select {
	case c.events <- ev:
		return true
	default:
		log.Println("connection.fire event failed: MQ FULL")
		c.efull = true
	}
	return false
}

func (c *connection) session() {

	var err error
	for {
		if c.efull {
			_, err = c.Write(evbufFull)
			c.efull = false
		} else {
			ev := <-c.events
			//			log.Println("Fire:", ev)
			_, err = c.Write(ev)
		}
		if err != nil {
			break
		}
	}

	log.Println("connection session closed")
}

// -------------------------------------------------------------------------

var MailBoxSize = 128

type Server struct {
	conns       map[int]*connection
	mutex       sync.Mutex
	mailBoxSize int
	baseFd      int
}

func NewServer() *Server {

	return &Server{conns: make(map[int]*connection), mailBoxSize: MailBoxSize}
}

func (r *Server) sink(c *connection) (fd int) {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.baseFd++
	fd = r.baseFd
	r.conns[fd] = c
	return
}

func (r *Server) unsink(fd int) {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	defer delete(r.conns, fd)
}

func (r *Server) Fire(eventType int64, data interface{}) {

	var evtbuf [10]byte
	binary.LittleEndian.PutUint64(evtbuf[:], uint64(eventType))

	b := bytes.NewBuffer(nil)
	b.Write(evtbuf[:])
	if data != nil {
		gob.NewEncoder(b).Encode(data)
	}

	b2 := b.Bytes()

	binary.LittleEndian.PutUint16(b2[8:], uint16(len(b2)-10))

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, conn := range r.conns {
		if (eventType & conn.eventTypes) != 0 {
			conn.fire(b2)
		}
	}
}

// ServeHTTP implements an http.Handler that answers requests.
func (r *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must CONNECT\n")
		return
	}

	pos := strings.LastIndex(req.URL.Path, "/")
	evts1 := req.URL.Path[pos+1:]
	evts, err := strconv.ParseInt(evts1, 16, 64)
	if err != nil || evts == 0 {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		io.WriteString(w, "400 Invalid arguments\n")
		return
	}

	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking ", req.RemoteAddr, ": ", err.Error())
		return
	}
	io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	r.ServeConn(conn, evts)
}

func (r *Server) ServeConn(conn io.WriteCloser, evts int64) {

	defer conn.Close()

	events := make(chan []byte, r.mailBoxSize)
	c := &connection{Writer: conn, eventTypes: evts, events: events}

	fd := r.sink(c)
	defer r.unsink(fd)

	c.session()
}

// -------------------------------------------------------------------------
