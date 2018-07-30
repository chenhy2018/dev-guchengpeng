package stgapi

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"io"
	"testing"

	"qbox.us/ptfd/stgapi.v1/api"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {

	cfg := newMockCfg(g_hosts)
	stg := newMockStg(cfg, 3)

	p := &Client{
		stg:        stg,
		cfg:        cfg,
		maxPutTrys: 3,
		idc:        "nb",
	}

	xl := xlog.NewDummy()
	fsize := int64(2*api.MaxDataSize + 1212123)
	data := randData(uint32(fsize))
	var eblocks []string
	for i := uint32(0); i < 3; i++ {
		max := uint32(api.MaxDataSize)
		if i == 2 {
			max = 1212123
		}
		data := data[i*api.MaxDataSize : i*api.MaxDataSize+max]
		ret, err := p.Create(xl, max, bytes.NewReader(data), max)
		assert.NoError(t, err)
		ctx, _ := api.DecodePositionCtx(ret)
		eblocks = append(eblocks, ctx.Eblock)
	}

	froms := []int64{0, 0, 1, 1, 0, 0, 0, 0, api.MaxDataSize, api.MaxDataSize + 1}
	tos := []int64{1, fsize, fsize, fsize - 1, fsize - 1, api.MaxDataSize - 1, api.MaxDataSize, api.MaxDataSize + 1, fsize, fsize - 12123}
	for i := range froms {
		xl = xlog.NewWith(fmt.Sprintf("reader-succes%v", i))
		from, to := froms[i], tos[i]
		rc, err := p.Get(xl, eblocks, from, to)
		assert.NoError(t, err, "%v", i)
		buf := bytes.NewBuffer(nil)
		n, err := io.Copy(buf, rc)
		assert.NoError(t, err, "%v", i)
		assert.Equal(t, to-from, n, "%v", i)
		assert.Equal(t, n, buf.Len(), "%v", i)
		rc.Close()
		crc := crc32.ChecksumIEEE(data[from:to])
		assert.Equal(t, crc, crc32.ChecksumIEEE(buf.Bytes()), "%v", i)
	}

	stg.failedGet = true
	for i := range froms {
		xl = xlog.NewWith(fmt.Sprintf("reader-failed%v", i))
		from, to := froms[i], tos[i]
		rc, err := p.Get(xl, eblocks, from, to)
		assert.NoError(t, err, "%v", i)
		buf := bytes.NewBuffer(nil)
		n, err := io.Copy(buf, rc)
		assert.NoError(t, err, "%v", i)
		assert.Equal(t, to-from, n, "%v", i)
		assert.Equal(t, n, buf.Len(), "%v", i)
		rc.Close()
		crc := crc32.ChecksumIEEE(data[from:to])
		assert.Equal(t, crc, crc32.ChecksumIEEE(buf.Bytes()), "%v", i)
	}
}
