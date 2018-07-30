package webutil

import (
	"http"
	"net"
	"qbox.us/log"
	"runtime/debug"
	"strconv"
	"strings"
)

// --------------------------------------------------------------

func SafeHandler(f func(w http.ResponseWriter, req *http.Request)) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			p := recover()
			if p != nil {
				w.WriteHeader(597)
				log.Warnf("panic fired in %v.panic:%v", f, p)
				log.Warn(string(debug.Stack()))
			}
		}()
		f(w, req)
	}
}

// --------------------------------------------------------------

var (
	LocalIPTable *IPTables = NewIPTable([]string{"127.255.255.255"})
	LanIPTable   *IPTables = NewIPTable([]string{"192.255.255.255", "10.255.255.255"})
	DebugIPTable *IPTables = NewIPTable([]string{"255.255.255.255"})
)

type IPTables struct {
	masks []*net.IPMask
}

func NewIPTable(masks []string) *IPTables {
	ms := make([]*net.IPMask, len(masks))
	for i, m := range masks {
		m2 := parseMask(m)
		if m2 == nil {
			return nil
		}
		ms[i] = m2
	}
	return &IPTables{ms}
}

func (p *IPTables) Check(req *http.Request) bool {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return false
	}
	return p.CheckIP(ip)
}

func (p *IPTables) CheckIP(ip string) bool {
	ip2 := net.ParseIP(ip)
	for _, m := range p.masks {
		if ip2.Equal(ip2.Mask(*m)) {
			return true
		}
	}
	return false
}

// Parse ipv4 mask (d.d.d.d).
func parseMask(s string) *net.IPMask {
	ds := strings.Split(s, ".", -1)
	if len(ds) != 4 {
		return nil
	}
	bs := make([]byte, 4)
	for i, d := range ds {
		n, err := strconv.Atoi(d)
		if err != nil {
			return nil
		}
		bs[i] = byte(n)
	}
	mask := net.IPv4Mask(bs[0], bs[1], bs[2], bs[3])
	return &mask
}
