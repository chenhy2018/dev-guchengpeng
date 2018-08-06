package api

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"

	"github.com/qiniu/encoding/binary"
	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"qbox.us/pfd/api/types"
)

const (
	StatusNothingTodo = 202
)

var (
	ErrNothingTodo     = httputil.NewError(StatusNothingTodo, "nothing to do")
	ErrUnmatchChecksum = errors.New("unmatched checksum")
)

// =======================================================================

type MigrateInfo struct {
	// Flag为1表示最后一个 gid 是完整 prepare 的。
	// pfd目前仅在gid无法填满整个条带时，实现了Flag=1。
	Flag              uint32   // 4
	GidCount          uint32   // 8
	FidCounts         []uint32 // GidCounts * 4 + 8
	Gids              []types.Gid
	FidInfos          []FileInfo
	FirstFidDataBegin int64
	LastFidDataEnd    int64
}

func EncodeMigrateInfoWithCrc(m *MigrateInfo) (b []byte) {
	w := encodeMigrateInfo(m)
	h := crc32.NewIEEE()
	h.Write(w)
	bh := make([]byte, 4)
	binary.LittleEndian.PutUint32(bh, h.Sum32())
	return append(w, bh...)
}

func encodeMigrateInfo(m *MigrateInfo) (b []byte) {
	w := new(bytes.Buffer)
	if err := binary.Write(w, binary.LittleEndian, *m); err != nil {
		panic(err) // should not happen
	}
	return w.Bytes()
}

func DecodeMigrateInfoWithCrc(r io.Reader, n int64) (m *MigrateInfo, err error) {
	r, n = appendCrc32Decode(r, n)
	return DecodeMigrateInfo(r)
}

func DecodeMigrateInfo(r io.Reader) (m *MigrateInfo, err error) {

	buf := new(bytes.Buffer)
	tr := io.TeeReader(r, buf)

	m = new(MigrateInfo)
	_, err = io.CopyN(ioutil.Discard, tr, 4) // Flag
	if err != nil {
		return
	}
	err = binary.Read(tr, binary.LittleEndian, &m.GidCount)
	if err != nil {
		return
	}
	m.FidCounts = make([]uint32, m.GidCount)
	m.Gids = make([]types.Gid, m.GidCount)

	err = binary.Read(io.LimitReader(tr, int64(m.GidCount*4)), binary.LittleEndian, m.FidCounts)
	if err != nil {
		return
	}
	fidCount := uint32(0)
	for _, c := range m.FidCounts {
		fidCount += c
	}
	m.FidInfos = make([]FileInfo, fidCount)

	err = binary.Read(io.MultiReader(buf, r), binary.LittleEndian, m)
	return
}

type FileInfo struct {
	Fid   uint64
	Off   int64
	Fsize int64
}

type PreparedInfo struct {
	PreparedGids   []types.Gid `json:"preparedGids"`
	LastGid        types.Gid   `json:"lastGid"`
	LastFid        uint64      `json:"lastFid"` //弱校验
	LastFidOff     int64       `json:"lastFidOff"`
	LastFidFsize   int64       `json:"lastFidFsize"`
	LastFidDataEnd int64       `json:"lastFidDataEnd"`
}

func (c Client) PrepareMigrate(l rpc.Logger, dgid uint32, stripeDataLen uint32, info *PreparedInfo) (ret *MigrateInfo, err error) {
	u := fmt.Sprintf("%v/preparemigrate/%v/len/%v", c.Host, dgid, stripeDataLen)
	resp, err := c.PostClient.PostWithJson(l, u, info)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == StatusNothingTodo {
		err = ErrNothingTodo
		return
	}
	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}

	ret, err = DecodeMigrateInfoWithCrc(resp.Body, resp.ContentLength)
	return
}

func (c Client) SlavePrepareMigrate(l rpc.Logger, dgid uint32, stripeDataLen uint32, info *PreparedInfo, scopeGids []types.Gid) (ret *MigrateInfo, err error) {
	u := fmt.Sprintf("%v/preparemigrate/%v/len/%v?slaveOk=1", c.Host, dgid, stripeDataLen)
	resp, err := c.PostClient.PostWithJson(l, u, struct {
		*PreparedInfo
		ScopeGids []types.Gid `json:"scopeGids"`
	}{info, scopeGids})
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == StatusNothingTodo {
		err = ErrNothingTodo
		return
	}
	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}

	ret, err = DecodeMigrateInfoWithCrc(resp.Body, resp.ContentLength)
	return
}

type appendCrc32Decoder struct {
	appendCrc32Reader io.Reader
	left              int64
	crc               uint32
}

func appendCrc32Decode(r io.Reader, n int64) (io.Reader, int64) {
	return &appendCrc32Decoder{
		appendCrc32Reader: r,
		left:              n - 4,
	}, n - 4
}

func (self *appendCrc32Decoder) Read(p []byte) (n int, err error) {
	if self.left == 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > self.left {
		p = p[:self.left]
	}
	n, err = self.appendCrc32Reader.Read(p)
	self.left -= int64(n)
	self.crc = crc32.Update(self.crc, crc32.IEEETable, p[:n])
	if self.left == 0 {
		crc32Bytes := make([]byte, 4)
		_, err1 := io.ReadFull(self.appendCrc32Reader, crc32Bytes)
		if err1 != nil {
			if err1 == io.EOF {
				err1 = io.ErrUnexpectedEOF
			}
			return 0, err1
		}
		expectedCrc := binary.LittleEndian.Uint32(crc32Bytes)
		if expectedCrc != self.crc {
			err = errors.New("getstripe: unmatched crc32")
			err = errors.Info(err, "expected:", expectedCrc, "actual:", self.crc)
			panic(err)
		}
		if err == nil {
			err = io.EOF
		}
	}
	return n, err
}

func (c Client) GetStripe(l rpc.Logger, dgid uint32, info *MigrateInfo) (rc io.ReadCloser, n int64, err error) {
	u := fmt.Sprintf("%v/getstripe/%v", c.Host, dgid)
	b := EncodeMigrateInfoWithCrc(info)
	resp, err := c.PostClient.PostWith(l, u, "application/octet-stream", bytes.NewReader(b), len(b))
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		err = rpc.ResponseError(resp)
		return
	}
	r, n := appendCrc32Decode(resp.Body, resp.ContentLength)
	return struct {
		io.Reader
		io.Closer
	}{
		Reader: r,
		Closer: io.Closer(resp.Body),
	}, n, nil
}
