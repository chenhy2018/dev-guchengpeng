package api

import (
	"testing"

	"github.com/stretchr/testify.v2/assert"
	"qbox.us/pfdcfg/api"
)

func TestDownSelector_SelDiskNode(t *testing.T) {

	{
		dg := NewDGNode("jjh", 101, api.DEFAULT)

		dg.add(&DiskNode{Dgid: 101, HostUrl: "a.com:9030", Idc: "xs"})
		dg.add(&DiskNode{Dgid: 101, HostUrl: "b.com:9030", Idc: "xs"})
		dg.add(&DiskNode{Dgid: 101, HostUrl: "c.com:9030", Idc: "nb"})

		sel := NewDownSelector("jjh", []string{"xs", "nb"})
		var badUrl []string
		for try := 0; try < 3; try++ {
			node, err := sel.SelDiskNode(101, false, dg, badUrl)
			assert.NoError(t, err)
			assert.Equal(t, node.Idc, "xs")
		}
		node, err := sel.SelDiskNode(101, false, dg, badUrl)
		assert.NoError(t, err)
		assert.Equal(t, node.Idc, "xs")
		badUrl = append(badUrl, node.HostUrl)
		for try := 0; try < 3; try++ {
			node, err = sel.SelDiskNode(101, false, dg, badUrl)
			assert.NoError(t, err)
			assert.Equal(t, node.Idc, "xs")
		}
		badUrl = append(badUrl, node.HostUrl)
		for try := 0; try < 3; try++ {
			node, err = sel.SelDiskNode(101, false, dg, badUrl)
			assert.NoError(t, err)
			assert.Equal(t, node.Idc, "nb")
		}

	}
	{
		dg := NewDGNode("jjh", 101, api.DEFAULT)
		dg.add(&DiskNode{Dgid: 101, HostUrl: "a.com:9030", Idc: "xs"})
		dg.add(&DiskNode{Dgid: 101, HostUrl: "b.com:9030", Idc: "nb"})
		dg.add(&DiskNode{Dgid: 101, HostUrl: "c.com:9030", Idc: "nb"})

		sel := NewDownSelector("jjh", []string{"xs", "nb"})
		var badUrl []string
		for try := 0; try < 3; try++ {
			node, err := sel.SelDiskNode(101, false, dg, badUrl)
			assert.NoError(t, err)
			assert.Equal(t, node.Idc, "xs")
		}
		node, err := sel.SelDiskNode(101, false, dg, badUrl)
		assert.NoError(t, err)
		assert.Equal(t, node.Idc, "xs")
		badUrl = append(badUrl, node.HostUrl)
		for try := 0; try < 3; try++ {
			node, err = sel.SelDiskNode(101, false, dg, badUrl)
			assert.NoError(t, err)
			assert.Equal(t, node.Idc, "nb")
		}
		badUrl = append(badUrl, node.HostUrl)
		for try := 0; try < 3; try++ {
			node, err = sel.SelDiskNode(101, false, dg, badUrl)
			assert.NoError(t, err)
			assert.Equal(t, node.Idc, "nb")
		}

	}

}
