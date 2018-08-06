package localaddrpool

import (
	"encoding/binary"
	"math/rand"
	"net"
	"time"

	"github.com/wangtuanjie/ip17mon"
	log "qiniupkg.com/x/log.v7"
)

type LocalAddrPool struct {
	ispMap     map[string][]net.Addr
	defaultIps []net.Addr
	ipLoc      *ip17mon.Locator
}

func NewLocalAddrPool(IpDataFile string, defaultIsp string, excludeIsps []string) (*LocalAddrPool, error) {
	rand.Seed(time.Now().Unix())
	ipLoc, err := ip17mon.NewLocator(IpDataFile)
	if err != nil {
		log.Warn("newLocalAddrPool: ip17mon.NewLocator failed", err)
		return nil, err
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	ips := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ip := ipnet.IP.To4(); ip != nil {
				ips = append(ips, ip)
			}
		}
	}
	ispMap := make(map[string][]net.Addr)
	for _, ip := range ips {
		isp := ipLoc.FindByUint(binary.BigEndian.Uint32([]byte(ip))).Isp
		if isp == ip17mon.Null {
			continue
		}
		log.Println("newLocalAddrPool:", isp, ip)
		ispMap[isp] = append(ispMap[isp], &net.TCPAddr{IP: ip})
	}
	log.Println("newLocalAddrPool: excludeIsps:", excludeIsps)
	for _, excludeIsp := range excludeIsps {
		delete(ispMap, excludeIsp)
	}
	var defaultIps []net.Addr
	if defaultIsp != "" {
		log.Println("newLocalAddrPool: defaultIsp:", defaultIsp)
		defaultIps = ispMap[defaultIsp]
	}
	return &LocalAddrPool{
		ispMap:     ispMap,
		defaultIps: defaultIps,
		ipLoc:      ipLoc,
	}, nil
}

func (self *LocalAddrPool) Pick(ip net.IP) net.Addr {
	ip = ip.To4()
	if ip == nil {
		return nil
	}
	isp := self.ipLoc.FindByUint(binary.BigEndian.Uint32([]byte(ip))).Isp
	ips, ok := self.ispMap[isp]
	if !ok {
		ips = self.defaultIps
	}
	if len(ips) == 0 {
		return nil
	}
	if n := len(ips); n > 1 {
		i := rand.Intn(n)
		return ips[i]
	}
	return ips[0]
}
