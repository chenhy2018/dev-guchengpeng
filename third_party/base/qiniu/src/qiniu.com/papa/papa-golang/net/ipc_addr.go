package net

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"strings"
	"syscall"
)

// tcp:<Host>:<Port> -> tcp <Host>:<Port>
// unix:<Path> -> unix <Path>
//
func ParseIPCAddr(fullAddr, dftProto string) (proto, addr string) {
	idx := strings.Index(fullAddr, "/")
	if idx >= 0 {
		dftProto = "unix"
	}
	idx = strings.Index(fullAddr, ":")
	if idx < 0 {
		return dftProto, fullAddr
	}

	switch fullAddr[:idx] {
	case "tcp", "tcp4", "tcp6":
		fallthrough
	case "udp", "udp4", "udp6":
		fallthrough
	case "ip", "ip4", "ip6":
		fallthrough
	case "unix", "unixgram", "unixpacket":
		return fullAddr[:idx], fullAddr[idx+1:]
	}
	return dftProto, fullAddr
}

func IsNetworkIPC(fullAddr string) bool {

	netw, _ := ParseIPCAddr(fullAddr, "tcp")
	switch netw {
	case "tcp", "tcp4", "tcp6":
		fallthrough
	case "udp", "udp4", "udp6":
		fallthrough
	case "ip", "ip4", "ip6":
		return true
	}
	return false
}

func IsLocalIPC(fullAddr string) bool {

	netw, _ := ParseIPCAddr(fullAddr, "tcp")
	switch netw {
	case "unix", "unixgram", "unixpacket":
		return true
	}
	return false
}

// tcp:<Host>:<Port> -> <Host>
//
func SplitNetworkHost(fullAddr string) (host uint64, err error) {

	if !IsNetworkIPC(fullAddr) {
		return 0, errors.New("not network address")
	}
	_, hostport := ParseIPCAddr(fullAddr, "tcp")
	idx := strings.Index(hostport, ":")
	if idx < 0 {
		return Atoip(hostport)
	}
	return Atoip(hostport[:idx])
}

// ---------------------------------------------------------------------------

func Iptoa(ip uint64) string {

	v := make([]byte, 4)
	binary.BigEndian.PutUint32(v, uint32(ip))
	return net.IP(v).String()
}

func Atoip(v string) (ip uint64, err error) {

	if ret := net.ParseIP(v); ret != nil {
		if ipv4 := ret.To4(); ipv4 != nil {
			return uint64(binary.BigEndian.Uint32(ipv4)), nil
		}
	}
	return 0, syscall.EINVAL
}

func Ipton(ip uint64) net.IP {

	v := uint32(ip)
	return net.IPv4(
		byte(v>>24&0xff),
		byte(v>>16&0xff),
		byte(v>>8&0xff),
		byte(v&0xff),
	)
}

func NtoIp(ip net.IP) uint64 {

	ip4 := ip.To4()
	return uint64(binary.BigEndian.Uint32(ip4))
}

func ChangeNetwork(from net.IP, to *net.IPNet) net.IP {

	maskn := binary.BigEndian.Uint32(to.Mask)
	fromn := binary.BigEndian.Uint32(from.To4())
	ton := binary.BigEndian.Uint32(to.IP.To4())

	ipn := ton&maskn + fromn & ^maskn
	return Ipton(uint64(ipn))
}

func SplitHostPort(addr net.Addr) (host uint32, port uint16, err error) {

	h, p, err := net.SplitHostPort(addr.String())
	if err != nil {
		return
	}
	host64, err := Atoip(h)
	if err != nil {
		return
	}
	port64, err := strconv.ParseUint(p, 10, 16)
	if err != nil {
		return
	}
	return uint32(host64), uint16(port64), nil
}

// ---------------------------------------------------------------------------
