package api

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"qbox.us/fh/fhver"
	"qbox.us/pfd/api/types"
	pfdcfg "qbox.us/pfdcfg/api"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	s := &mockStater{cache: false, actual: false}
	pfd := &mockPfd{err612: false}
	ebd := &mockEbd{}
	c := New(s, pfd, ebd)
	xl := xlog.NewDummy()
	fh := types.EncodeFh(&types.FileHandle{
		Ver:   fhver.FhPfdV2,
		Tag:   types.FileHandle_ChunkBits,
		Fsize: 3,
	})

	{
		s.cache, s.actual = false, false
		pfd.err612 = false

		rc, _, err := c.Get(xl, fh, 0, 3)
		assert.NoError(t, err)
		b, _ := ioutil.ReadAll(rc)
		rc.Close()
		assert.Equal(t, "pfd", string(b))
	}
	{
		s.cache, s.actual = false, true
		pfd.err612 = false

		rc, _, err := c.Get(xl, fh, 0, 3)
		assert.NoError(t, err)
		b, _ := ioutil.ReadAll(rc)
		rc.Close()
		assert.Equal(t, "pfd", string(b))
	}
	{
		s.cache, s.actual = false, true
		pfd.err612 = true

		rc, _, err := c.Get(xl, fh, 0, 3)
		assert.NoError(t, err)
		b, _ := ioutil.ReadAll(rc)
		rc.Close()
		assert.Equal(t, "ebd", string(b))
	}
	{
		s.cache, s.actual = true, true
		pfd.err612 = false

		rc, _, err := c.Get(xl, fh, 0, 3)
		assert.NoError(t, err)
		b, _ := ioutil.ReadAll(rc)
		rc.Close()
		assert.Equal(t, "ebd", string(b))
	}
	{
		s.cache, s.actual = true, true
		pfd.err612 = true

		rc, _, err := c.Get(xl, fh, 0, 3)
		assert.NoError(t, err)
		b, _ := ioutil.ReadAll(rc)
		rc.Close()
		assert.Equal(t, "ebd", string(b))
	}
	{
		s.cache, s.actual = true, true
		pfd.err612 = false
		ebd.err599 = true

		rc, _, err := c.Get(xl, fh, 0, 3)
		assert.NoError(t, err)
		b, _ := ioutil.ReadAll(rc)
		rc.Close()
		assert.Equal(t, "pfd", string(b))
	}
}

type mockStater struct {
	cache  bool
	actual bool
}

func (self *mockStater) State(l rpc.Logger, egid string) (dgid uint32, isECed bool, err error) {
	return 0, self.cache, nil
}

func (self *mockStater) StateWithGroup(l rpc.Logger, egid string) (guid string, dgid uint32, isECed bool, err error) {
	return "", 0, self.cache, nil
}

func (self *mockStater) ForceUpdate(l rpc.Logger, egid string) (dgid uint32, isECed bool, err error) {
	return 0, self.actual, nil
}

type mockPfd struct {
	err612 bool
}

func (self *mockPfd) Get(l rpc.Logger, fh []byte, from, to int64) (rc io.ReadCloser, fsize int64, err error) {
	if self.err612 {
		return nil, 0, httputil.NewError(612, "error for fun")
	}
	return ioutil.NopCloser(strings.NewReader("pfd")), 3, nil
}

func (self *mockPfd) GetType(l rpc.Logger, fh []byte) (typ pfdcfg.DiskType, err error) {
	return pfdcfg.SSD, nil
}

type mockEbd struct {
	err599 bool
}

func (self *mockEbd) Get(l rpc.Logger, fh []byte, from, to int64) (rc io.ReadCloser, fsize int64, err error) {
	if self.err599 {
		return nil, 0, httputil.NewError(599, "error for fun")
	}
	return ioutil.NopCloser(strings.NewReader("ebd")), 3, nil
}

func (self *mockEbd) GetType(l rpc.Logger, fh []byte) (typ pfdcfg.DiskType, err error) {
	return pfdcfg.DEFAULT, nil
}
