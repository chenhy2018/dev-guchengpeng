package api

import (
	"fmt"
	"testing"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetOutputLevel(0)
}

func TestSelDiskNode(t *testing.T) {
	xl := xlog.NewWith("TestSelDiskNode")

	var upDiskNodes []*DiskNode
	upSel := NewUpSelector()

	for dgid := 1; dgid <= 100; dgid++ {
		diskNode := DiskNode{
			Dgid:     uint32(dgid),
			DiskType: 0,
			ReadOnly: 0,
			IsBackup: false,
			HostUrl:  fmt.Sprintf("http://%v.pfd.info", dgid),
			LoadId:   loadId(uint32(dgid), 0),
		}

		upDiskNodes = append(upDiskNodes, &diskNode)

		//设置diskNode的负载, 倒序的顺序：第一个节点的负载为100， 最后一个节点的负载为1
		upSel.GetLoad(loadId(uint32(dgid), 0)) //必须先get，才能在map中创建laod节点。在非单元测试的正规流程中，肯定会先Get，然后Incre/Decre
		for i := 0; i <= 100-dgid; i++ {
			upSel.IncreLoad(diskNode.LoadId)
		}
	}

	badHostUrls := []string{"http://1.pfd.info", "http://2.pfd.info"}
	diskNode, err := upSel.SelDiskNode(xl, upDiskNodes, badHostUrls, []uint32{})
	xl.Infof("selDiskNode:", diskNode)
	assert.NoError(t, err)
	assert.Equal(t, MAX_DISK_NODE_CHOOSE_NUM+len(badHostUrls), diskNode.Dgid)
	upSel.DecreLoad(diskNode.LoadId)

	upSel.diskIdx = 0
	badHostUrls = []string{"http://12.pfd.info", "http://13.pfd.info"}
	diskNode, err = upSel.SelDiskNode(xl, upDiskNodes, badHostUrls, []uint32{})
	xl.Info("selDiskNode:", diskNode)
	assert.NoError(t, err)
	assert.Equal(t, MAX_DISK_NODE_CHOOSE_NUM+len(badHostUrls), diskNode.Dgid)
	upSel.DecreLoad(diskNode.LoadId)

	upSel.diskIdx = 0
	badHostUrls = []string{"http://37.pfd.info", "http://38.pfd.info"}
	diskNode, err = upSel.SelDiskNode(xl, upDiskNodes, badHostUrls, []uint32{})
	xl.Info("selDiskNode:", diskNode)
	assert.NoError(t, err)
	assert.Equal(t, MAX_DISK_NODE_CHOOSE_NUM, diskNode.Dgid)
	upSel.DecreLoad(diskNode.LoadId)
}
