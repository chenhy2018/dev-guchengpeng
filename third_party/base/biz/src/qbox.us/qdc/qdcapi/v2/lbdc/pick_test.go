package lbdc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"qbox.us/dht"
)

var pickCases = [][2][]string{
	[2][]string{
		[]string{"http://a:1"},
		[]string{"http://a:1", "http://a:1"},
	},
	[2][]string{
		[]string{"http://a:1", "http://b:1"},
		[]string{"http://a:1", "http://b:1", "http://a:1", "http://b:1"},
	},
	[2][]string{
		[]string{"http://a:1", "http://a:2"},
		[]string{"http://a:1", "http://a:2", "http://a:1", "http://a:2"},
	},
	[2][]string{
		[]string{"http://a:1", "http://a:2", "http://b:1"},
		[]string{"http://a:1", "http://b:1", "http://a:2", "http://a:1", "http://b:1", "http://a:2"},
	},
	[2][]string{
		[]string{"a:1", "a:2", "a:3", "b:1"},
		[]string{"a:1", "b:1", "a:3", "a:2", "a:1", "b:1", "a:3", "a:2"},
	},
}

func TestPicker(t *testing.T) {

	for i, cs := range pickCases {
		p := picker{hosts: cs[0]}
		for j, host := range cs[1] {
			assert.Equal(t, p.One(), host, "case %v %v", i, j)
		}
	}
}

func TestPickerProxy(t *testing.T) {

	_, err := newPickerProxy(dht.RouterInfos{})
	assert.Equal(t, err, EServerNotAvailable)

	for i, cs := range pickCases {
		routers := make(dht.RouterInfos, len(cs[0]))
		for i, host := range cs[0] {
			routers[i] = dht.RouterInfo{Host: host}
		}
		p, _ := newPickerProxy(routers)
		for j, host := range cs[1] {
			assert.Equal(t, p.One(), host, "case %v %v", i, j)
		}
	}
}
