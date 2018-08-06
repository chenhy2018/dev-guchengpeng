// +build go1.7

package localaddrpool

import (
	"net"
	"testing"

	"github.com/stretchr/testify.v2/assert"
)

const IpDataFile = "test/17monipdb.dat"

func TestLocalAddrPool(t *testing.T) {
	p, err := NewLocalAddrPool(IpDataFile, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	{
		addr := p.Pick(net.ParseIP("192.168.0.1"))
		assert.Nil(t, addr)
	}

	{
		fakeAddr := &net.TCPAddr{IP: net.IPv4(1, 2, 3, 4)}
		p.ispMap["电信"] = []net.Addr{fakeAddr}
		addr := p.Pick(net.ParseIP("115.239.210.27"))
		assert.Equal(t, fakeAddr, addr)
	}

	{
		fakeAddr0 := &net.TCPAddr{IP: net.IPv4(1, 1, 1, 1)}
		fakeAddr1 := &net.TCPAddr{IP: net.IPv4(2, 2, 2, 2)}
		p.ispMap["移动"] = []net.Addr{fakeAddr0, fakeAddr1}

		count0, count1 := 0, 0
		for i := 0; i < 1000; i++ {
			addr := p.Pick(net.ParseIP("111.13.100.92"))
			assert.True(t, fakeAddr0 == addr || fakeAddr1 == addr)
			if fakeAddr0 == addr {
				count0++
			} else {
				count1++
			}
		}
		assert.True(t, count0 > 0)
		assert.True(t, count1 > 0)
	}

	{
		addr := p.Pick(net.ParseIP("[2001:DF0:1003::F]"))
		assert.Nil(t, addr)
	}

	p, err = NewLocalAddrPool(IpDataFile, "联通", []string{"电信"})
	if err != nil {
		t.Fatal(err)
	}
}
