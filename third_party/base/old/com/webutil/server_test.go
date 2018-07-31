package webutil

import (
	"testing"
)

func TestIPTables(t *testing.T) {

	cases := []struct {
		masks  []string
		ips    []string
		result bool
	}{
		{[]string{"255.255.255.255"}, []string{"192.168.1.100"}, true},
		{[]string{"127.255.255.255"}, []string{"127.0.0.1"}, true},
		{[]string{"127.255.255.255"}, []string{"192.0.0.1"}, false},
		{[]string{"0.0.0.0"}, []string{"127.0.0.1"}, false},
		{[]string{"10.0.0.101"}, []string{"10.0.0.101"}, true},
		{[]string{"10.0.0.101"}, []string{"10.0.0.102"}, false},
	}

	for _, c := range cases {
		tb := NewIPTable(c.masks)
		for _, ip := range c.ips {
			b := tb.CheckIP(ip)
			if b != c.result {
				t.Fatal(c)
			}
		}
	}
}
