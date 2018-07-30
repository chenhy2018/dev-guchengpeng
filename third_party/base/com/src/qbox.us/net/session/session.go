package session

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type SessionManager struct {
	CookieName    string
	mutex         *sync.Mutex
	sessions      map[string][2]map[string]interface{}
	expires       int
	timerDuration time.Duration
}

func New(cookieName string, expires int, timerDuration time.Duration) *SessionManager {
	if cookieName == "" {
		cookieName = "GoLangerSession"
	}

	if expires <= 0 {
		expires = 3600
	}

	if timerDuration <= 0 {
		timerDuration, _ = time.ParseDuration("24h")
	}

	s := &SessionManager{
		CookieName:    cookieName,
		mutex:         &sync.Mutex{},
		sessions:      map[string][2]map[string]interface{}{},
		expires:       expires,
		timerDuration: timerDuration,
	}

	time.AfterFunc(s.timerDuration, func() { s.GC() })

	return s
}

func (s *SessionManager) Get(rw http.ResponseWriter, req *http.Request) (http.ResponseWriter, map[string]interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var sessionSign string
	if c, err := req.Cookie(s.CookieName); err == nil {
		sessionSign = c.Value
		if _, ok := s.sessions[sessionSign]; !ok {
			sessionSign = s.new(rw)
		}
	} else {
		sessionSign = s.new(rw)
	}

	return rw, s.sessions[sessionSign][1]
}

func (s *SessionManager) Len() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return int64(len(s.sessions))
}

func (s *SessionManager) new(rw http.ResponseWriter) string {
	timeNano := time.Now().UnixNano()
	sessionSign := s.sessionSign()
	s.sessions[sessionSign] = [2]map[string]interface{}{
		map[string]interface{}{
			"create": timeNano,
		},
		map[string]interface{}{},
	}

	s.setCookie(rw, s.CookieName, sessionSign, 0, "/")

	return sessionSign
}

func (s *SessionManager) Clear(sessionSign string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.sessions, sessionSign)
}

func (s *SessionManager) GC() {
	s.mutex.Lock()
	for sessionSign, _ := range s.sessions {
		if (s.sessions[sessionSign][0]["create"].(int64) + int64(s.expires)) <= time.Now().Unix() {
			delete(s.sessions, sessionSign)
		}
	}

	s.mutex.Unlock()
	time.AfterFunc(s.timerDuration, func() { s.GC() })
}

func (s *SessionManager) sessionSign() string {
	var n int = 24
	b := make([]byte, n)
	io.ReadFull(rand.Reader, b)

	//return length:32
	return base64.URLEncoding.EncodeToString(b)
}

/*
cookie[0] => name string
cookie[1] => value string
cookie[2] => expires string
cookie[3] => path string
cookie[4] => domain string
*/
func (s *SessionManager) setCookie(rw http.ResponseWriter, args ...interface{}) {
	if len(args) < 2 {
		return
	}

	length := 5
	cookie := make([]interface{}, length)

	for k, v := range args {
		if k >= length {
			break
		}

		cookie[k] = v
	}

	var (
		name    string
		value   string
		expires int
		path    string
		domain  string
	)

	if v, ok := cookie[0].(string); ok {
		name = v
	} else {
		return
	}

	if v, ok := cookie[1].(string); ok {
		value = v
	} else {
		return
	}

	if v, ok := cookie[2].(int); ok {
		expires = v
	}

	if v, ok := cookie[3].(string); ok {
		path = v
	}

	if v, ok := cookie[4].(string); ok {
		domain = v
	}

	bCookie := &http.Cookie{
		Name:   name,
		Value:  url.QueryEscape(value),
		Path:   path,
		Domain: domain,
	}

	if expires > 0 {
		d, _ := time.ParseDuration(strconv.Itoa(expires) + "s")
		bCookie.Expires = time.Now().Add(d)
	}

	http.SetCookie(rw, bCookie)
}
