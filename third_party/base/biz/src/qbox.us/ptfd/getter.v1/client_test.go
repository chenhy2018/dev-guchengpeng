package getter

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"

	"qbox.us/fh/fhver"
	pfdtypes "qbox.us/pfd/api/types"
	pfdapi "qbox.us/pfdstg/api"
	"qbox.us/ptfd/api.v1"
	"qbox.us/ptfd/masterapi.v1"
)

var (
	errEblockNotFound = errors.New("eblock not found")
	errFidNotFound    = errors.New("fid not found")
	errFhNotFound     = httputil.NewError(pfdapi.StatusAllocedEntry, "fh not found")
)

type mockStg struct {
	Eblocks map[string][]byte
}

func (p *mockStg) Get(xl *xlog.Logger, eblocks []string, from, to int64) (io.ReadCloser, error) {

	var data []byte
	for _, eblock := range eblocks {
		b, ok := p.Eblocks[eblock]
		if !ok {
			return nil, errEblockNotFound
		}
		data = append(data, b...)
	}
	return ioutil.NopCloser(bytes.NewReader(data[from:to])), nil
}

type mockPfd struct {
	Datas map[string][]byte
}

func (p *mockPfd) Get(l rpc.Logger, fh []byte, from, to int64) (io.ReadCloser, int64, error) {

	data, ok := p.Datas[string(fh)]
	if !ok {
		return nil, 0, errFhNotFound
	}
	return ioutil.NopCloser(bytes.NewReader(data[from:to])), to - from, nil
}

type mockMaster struct {
	Entrys map[uint64]*masterapi.Entry
}

func (p *mockMaster) Query(l rpc.Logger, fh []byte) (*masterapi.Entry, error) {

	fhi, err := pfdtypes.DecodeFh(fh)
	if err != nil {
		return nil, err
	}
	fid := fhi.Fid
	entry, ok := p.Entrys[fid]
	if !ok {
		return nil, errFidNotFound
	}
	return entry, nil
}

func TestClient(t *testing.T) {

	data := []byte("helloworld")
	stg := &mockStg{
		Eblocks: map[string][]byte{
			"eblock1": data,
		},
	}
	pfd := &mockPfd{
		Datas: map[string][]byte{},
	}
	master := &mockMaster{
		Entrys: make(map[uint64]*masterapi.Entry),
	}
	l := xlog.NewDummy()
	p := NewWith(stg, master, pfd)
	// Read 0 size file
	fh := pfdtypes.EncodeFh(&pfdtypes.FileHandle{
		Ver: fhver.FhPfdV2,
		Tag: pfdtypes.FileHandle_ChunkBits,
		Fid: 1,
	})
	var from, to int64
	rc, length, err := p.Get(l, fh, from, to)
	assert.NoError(t, err)
	n, err := io.Copy(ioutil.Discard, rc)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, 0, length)

	// Query fid not found
	to = int64(len(data))
	fh = pfdtypes.EncodeFh(&pfdtypes.FileHandle{
		Ver:   fhver.FhPfdV2,
		Tag:   pfdtypes.FileHandle_ChunkBits,
		Fid:   1,
		Fsize: to,
	})
	_, _, err = p.Get(l, fh, from, to)
	assert.Equal(t, errFidNotFound, errors.Err(err))

	// Query fsize not matched
	master.Entrys[1] = &masterapi.Entry{Fsize: to - 1}
	_, _, err = p.Get(l, fh, from, to)
	assert.Equal(t, api.ErrInvalidArgs, errors.Err(err))

	// Exceed fsize
	master.Entrys[1] = &masterapi.Entry{Fsize: to}
	_, _, err = p.Get(l, fh, from, to+1)
	assert.Equal(t, api.ErrExceedFsize, errors.Err(err))

	// Eblock not found
	master.Entrys[1] = &masterapi.Entry{Fsize: to, Eblocks: []string{"eblock"}}
	_, _, err = p.Get(l, fh, from, to)
	assert.Equal(t, errEblockNotFound, errors.Err(err))

	// Success get
	master.Entrys[1] = &masterapi.Entry{Fsize: to, Eblocks: []string{"eblock1"}}
	froms := []int64{0, 1, 1, 0}
	tos := []int64{to, to, to - 1, to - 1}
	for i := 0; i < len(froms); i++ {
		from, to := froms[i], tos[i]
		rc, length, err = p.Get(l, fh, from, to)
		assert.NoError(t, err)
		assert.Equal(t, to-from, length)
		buf := bytes.NewBuffer(nil)
		n, err = io.Copy(buf, rc)
		assert.NoError(t, err)
		assert.Equal(t, to-from, n)
		assert.Equal(t, to-from, n)
		assert.Equal(t, data[from:to], buf.Bytes())
	}
}
