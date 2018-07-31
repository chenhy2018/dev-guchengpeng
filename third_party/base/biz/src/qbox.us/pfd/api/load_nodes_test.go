package api

import (
	"fmt"
	"testing"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	cfgapi "qbox.us/pfdcfg/api"
)

func init() {
	log.SetOutputLevel(1)
}

const GUID_LOAD = "1234567890"

func makeDgiLoad(dgid uint32, stgHost string, dt cfgapi.DiskType, idc string) *cfgapi.DiskGroupInfo {

	needStgHosts := make([][2]string, 0)
	var oneGroupStgHosts [2]string
	oneGroupStgHosts[0] = stgHost
	oneGroupStgHosts[1] = stgHost
	needStgHosts = append(needStgHosts, oneGroupStgHosts)

	return &cfgapi.DiskGroupInfo{
		Guid:     GUID_LOAD,
		Dgid:     dgid,
		Hosts:    needStgHosts,
		DiskType: dt,
		Idc:      []string{idc},
	}
}

func TestMatrix(t *testing.T) {
	xl := xlog.NewWith("TestMatrix")

	var dgis []*cfgapi.DiskGroupInfo

	totalDgidNum := 36 + 9*12
	totalDgidWeightNum := totalDgidNum + 1 + 3 + 2
	weights := map[uint32]uint32{0: 1, 2: 3, 5: 2}

	dgid := 0
	//stgs:
	stgNodeId := 0
	stgHost := fmt.Sprintf("http://%v.pfd.info", stgNodeId)

	for i := 0; i < 36; i++ {
		dgi := makeDgiLoad(uint32(dgid), stgHost, cfgapi.DEFAULT, "")
		if weight, ok := weights[uint32(dgid)]; ok {
			dgi.Weight = weight
		}
		dgis = append(dgis, dgi)
		dgid += 1
	}

	for stgNodeId = 1; stgNodeId < 10; stgNodeId++ {
		stgHost := fmt.Sprintf("http://%v.pfd.info", stgNodeId)

		for j := 0; j < 12; j++ {
			dgi := makeDgiLoad(uint32(dgid), stgHost, cfgapi.DEFAULT, "")
			if weight, ok := weights[uint32(dgid)]; ok {
				dgi.Weight = weight
			}
			dgis = append(dgis, dgi)
			dgid += 1
		}
	}

	pfdNodeMgr := NewPfdNodeMgr("", nil)
	pfdNodeMgr.LoadPfdNodes(xl, dgis)
	badHostUrls := make([]string, 0)

	for i := 0; i < totalDgidWeightNum; i++ {
		diskNode, err := pfdNodeMgr.SelectUpDisk(xl, cfgapi.DEFAULT, badHostUrls, []uint32{})
		if err != nil {
			xl.Errorf("select node for up error. err:%v", err)
			return
		}
		xl.Infof("select time:%v selected diskNode:%v", i, diskNode)
	}

	for dgid = 0; dgid < totalDgidNum; dgid++ {
		load := pfdNodeMgr.upSel.GetLoad(loadId(uint32(dgid), 0))
		assert.Equal(t, load, 1)
	}
	for dgid, weight := range weights {
		for i := uint32(1); i <= weight; i++ {
			load := pfdNodeMgr.upSel.GetLoad(loadId(dgid, i))
			assert.Equal(t, load, 1)
		}
	}
}
