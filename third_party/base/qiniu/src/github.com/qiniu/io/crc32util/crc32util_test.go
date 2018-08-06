package crc32util

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/qiniu/errors"
	"github.com/stretchr/testify.v1/assert"
	"github.com/stretchr/testify.v1/require"
)

func TestSize(t *testing.T) {
	fsizes := []int64{
		0,
		1,
		5,
		BufSize - 5,
		BufSize - 4,
		BufSize - 3,
		BufSize - 2,
		BufSize - 1,
		BufSize,
		BufSize + 1,
		BufSize + 2,
		2 * (BufSize - 5),
		2 * (BufSize - 4),
		2 * (BufSize - 3),
		2 * (BufSize - 2),
		2 * (BufSize - 1),
		2 * (BufSize),
		2 * (BufSize + 1),
		2 * (BufSize + 2),
	}
	for _, fsize := range fsizes {
		total := EncodeSize(fsize)
		fsize2 := DecodeSize(total)
		if fsize2 != fsize {
			t.Errorf("fsize: %v, totalSize: %v, fsize2: %v\n", fsize, total, fsize2)
		}
	}
}

// 1. 64K
// 2. 64K -4
// 3. 128K
// 4. 1M
// 5. 0
func TestSimpleEncoderDecoder(t *testing.T) {
	size1 := int64(64*1024 - 4)
	size2 := int64(64 * 1024)
	size3 := int64(64*1024 + 4)
	size4 := int64(1024 * 1024)
	size5 := int64(0)
	data := []struct {
		bodysize int64
	}{
		{bodysize: size1},
		{bodysize: size2},
		{bodysize: size3},
		{bodysize: size4},
		{bodysize: size5},
	}

	for index, d := range data {
		b1 := randData(d.bodysize)
		body := bytes.NewReader(b1)
		enc := SimpleEncoder(body, nil)

		dec := SimpleDecoder(enc, nil)
		b2 := make([]byte, d.bodysize)
		_, err := io.ReadFull(dec, b2)
		assert.NoError(t, err, "index is %v", index)
		assert.Equal(t, crc32.ChecksumIEEE(b1), crc32.ChecksumIEEE(b2), "index is %v", index)
	}
}

// 测试64K * n < length <= 64K * n +4 的数据流
func TestSimpleDecoder_InvalidLength(t *testing.T) {
	tests := []struct {
		bodysize int64
		ok       bool
	}{
		{bodysize: 0, ok: true},
		{bodysize: 1, ok: false},
		{bodysize: 4, ok: false},
		{bodysize: 5, ok: true},
		{bodysize: 64 * 1024, ok: true},
		{bodysize: 64*1024 + 1, ok: false},
		{bodysize: 64*1024 + 4, ok: false},
	}

	for i, test := range tests {
		if test.ok {
			body := getCrcEncodedReader(test.bodysize)
			dec := SimpleDecoder(body, nil)
			n, err := io.Copy(ioutil.Discard, dec)
			assert.NoError(t, err, strconv.Itoa(i))
			assert.Equal(t, DecodeSize(test.bodysize), n, strconv.Itoa(i))
		}

		if test.bodysize != 0 {
			body := getRandomReader(test.bodysize)
			dec := SimpleDecoder(body, nil)
			_, err := io.Copy(ioutil.Discard, dec)
			assert.Error(t, err, strconv.Itoa(i))
		}
	}
}

// esize = 0 ||(4, 64K] + 64K * n
func getCrcEncodedReader(esize int64) io.Reader {

	fsize := DecodeSize(esize)
	data := randData(fsize)
	body := bytes.NewReader(data)
	enc := SimpleEncoder(body, nil)
	return enc
}

// 前面是正确的数据，最后一片是随机的。
// esize >= 0
func getRandomReader(esize int64) io.Reader {
	chunkSize := int64(64 * 1024)
	wrap := (esize - 1) / chunkSize // 64K * n 时，最后一片需要是错误的
	left := esize - wrap*chunkSize
	var b1 io.Reader = noBody{}
	if wrap != 0 {
		b1 = getCrcEncodedReader(wrap * chunkSize)
	}
	var b2 io.Reader = noBody{}
	if left != 0 {
		data := randData(left)
		b2 = bytes.NewReader(data)
	}
	r := io.MultiReader(b1, b2)
	return r
}

type noBody struct{}

func (noBody) Read([]byte) (int, error) { return 0, io.EOF }

func TestEncodeDecode(t *testing.T) {

	data := randData(128*1024 + 80)
	fsize := int64(len(data)) // 128k+80
	r := bytes.NewReader(data)
	w := bytes.NewBuffer(nil)
	err := Encode(w, r, fsize, nil)
	require.NoError(t, err)

	_64k := int64(BufSize)

	tcs := []struct {
		base  int64
		from  int64
		to    int64
		fsize int64
		tail  int64
	}{
		{from: 0, to: 0, fsize: 0},
		{from: fsize, to: fsize, fsize: fsize},
		{from: 0, to: fsize, fsize: fsize},

		{from: _64k - 4, to: fsize, fsize: fsize},
		{from: _64k + 0, to: fsize, fsize: fsize},
		{from: _64k + 4, to: fsize, fsize: fsize},
		{from: _64k + 5, to: fsize, fsize: fsize},

		{from: _64k - 4, to: fsize - 1, fsize: fsize},
		{from: _64k + 0, to: fsize - 1, fsize: fsize},
		{from: _64k + 4, to: fsize - 1, fsize: fsize},
		{from: _64k + 5, to: fsize - 1, fsize: fsize},

		{from: _64k + 4, to: fsize - _64k, fsize: fsize},
		{from: _64k + 4, to: fsize - _64k - 4, fsize: fsize},
		{from: _64k + 4, to: fsize - _64k - 5, fsize: fsize},
		{from: 0, to: fsize - _64k - 4 - _64k - 4, fsize: fsize},

		{base: 1, from: _64k + 4, to: fsize, fsize: fsize},
		{base: 0, from: _64k + 4, to: fsize, fsize: fsize, tail: 1000},
		{base: _64k + 4, from: _64k + 4, to: fsize, fsize: fsize, tail: _64k + 4},
		{base: _64k + 5, from: _64k + 4, to: fsize, fsize: fsize, tail: _64k + 5},
	}
	chunk := make([]byte, BufSize)
	for _, tc := range tcs {
		// log.Printf("%+v\n", tc)
		disk := append(make([]byte, tc.base), w.Bytes()...)
		disk = append(disk, make([]byte, tc.tail)...)

		dr := RangeDecoder(bytes.NewReader(disk), tc.base, chunk, tc.from, tc.to, tc.fsize)
		all, err := ioutil.ReadAll(dr)
		assert.NoError(t, err, "%+v", tc)
		assert.Equal(t, crc32.ChecksumIEEE(data[tc.from:tc.to]), crc32.ChecksumIEEE(all), "%+v", tc)
	}
}

func TestDecoderFail(t *testing.T) {
	data := bytes.Repeat([]byte("hellowor"), 2*1024*8+10)
	fsize := int64(len(data))
	r := bytes.NewReader(data)
	w := bytes.NewBuffer(nil)
	err := Encode(w, r, fsize, nil)
	assert.NoError(t, err)

	r2 := bytes.NewReader(w.Bytes())
	er := &errorReaderAt{r2, 3}
	dr := RangeDecoder(er, 0, nil, 0, 10, 10)
	all := make([]byte, fsize)
	_, err = io.ReadFull(dr, all)
	assert.Error(t, err)
}

func TestDecoderShort(t *testing.T) {
	data := make([]byte, chunkLen+1)
	r := bytes.NewReader(data)
	w := bytes.NewBuffer(nil)
	err := Encode(w, r, int64(len(data)), nil)
	assert.NoError(t, err)

	for off := 0; off <= 4; off++ {
		r2 := bytes.NewReader(w.Bytes()[:BufSize+off])
		dr := RangeDecoder(r2, 0, nil, 0, int64(len(data)), int64(len(data)))
		b, err := ioutil.ReadAll(dr)
		assert.Equal(t, io.ErrUnexpectedEOF, err)
		assert.Equal(t, data[:chunkLen], b)
	}

	r2 := bytes.NewReader(w.Bytes())
	dr := RangeDecoder(r2, 0, nil, 0, int64(len(data)), int64(len(data)))
	b, err := ioutil.ReadAll(dr)
	assert.Equal(t, nil, err)
	assert.Equal(t, data, b)
}

func randData(size int64) []byte {

	b := make([]byte, size)
	rand.Read(b)
	return b
}

func TestAppend(t *testing.T) {

	dir := "./testappend/"
	os.MkdirAll(dir, 0777)
	defer os.RemoveAll(dir)

	f, _ := os.Create(dir + "1")

	size := int64(3*1024*1024 - 112123)
	data := randData(size)

	sizes := []int64{0, 50*1024 + 123, 64*1024 - 4, 64 * 1024, 134*1024 + 121, 320 * 1024, 1024*1024 + 123, size}
	for i := 1; i < len(sizes); i++ {
		osize := sizes[i-1]
		nsize := sizes[i]
		ndata := data[osize:nsize]
		r := bytes.NewReader(ndata)
		err := AppendEncode(f, 4112, osize, r, nsize-osize, nil)
		assert.NoError(t, err, "%v", i)

		all := make([]byte, nsize)
		dr := RangeDecoder(f, 4112, nil, 0, nsize, nsize)
		_, err = io.ReadFull(dr, all)
		assert.NoError(t, err, "%v", i)

		crc := crc32.ChecksumIEEE(data[:nsize])
		assert.Equal(t, crc, crc32.ChecksumIEEE(all), "%v", i)
	}

	f.Close()

	f, _ = os.Create(dir + "2")

	size0 := int64(3*1024 + 1)
	data0 := randData(size0)

	var r io.Reader
	r = bytes.NewReader(data0)
	err := AppendEncode(f, 123, 0, r, size0, nil)
	assert.NoError(t, err)

	crc := crc32.ChecksumIEEE(data0)
	all := make([]byte, size0)
	dr := RangeDecoder(f, 123, nil, 0, size0, size0)
	_, err = io.ReadFull(dr, all)
	assert.NoError(t, err)
	assert.Equal(t, crc, crc32.ChecksumIEEE(all))

	size1 := int64(23*1024 + 13)
	data1 := randData(size1)

	r = &errorReader{bytes.NewReader(data1), 20*1024 + 12}
	err = AppendEncode(f, 123, size0, r, size1, nil)
	assert.Error(t, err)

	all = make([]byte, size0)
	dr = RangeDecoder(f, 123, nil, 0, size0, size0)
	_, err = io.ReadFull(dr, all)
	assert.NoError(t, err)
	assert.Equal(t, crc, crc32.ChecksumIEEE(all))

	r = bytes.NewReader(data1)
	err = AppendEncode(f, 123, size0, r, size1, nil)
	assert.NoError(t, err)
	crc = crc32.Update(crc, crc32.IEEETable, data1)

	all = make([]byte, size0+size1)
	dr = RangeDecoder(f, 123, nil, 0, size0+size1, size0+size1)
	_, err = io.ReadFull(dr, all)
	assert.NoError(t, err)
	assert.Equal(t, crc, crc32.ChecksumIEEE(all))

	size2 := int64(123*1024 + 313)
	data2 := randData(size2)

	r = &errorReader{bytes.NewReader(data2), 100*1024 + 436}
	err = AppendEncode(f, 123, size0+size1, r, size2, nil)
	assert.Error(t, err)

	all = make([]byte, size0+size1)
	dr = RangeDecoder(f, 123, nil, 0, size0+size1, size0+size1)
	_, err = io.ReadFull(dr, all)
	assert.NoError(t, err)
	assert.Equal(t, crc, crc32.ChecksumIEEE(all))

	r = bytes.NewReader(data2)
	err = AppendEncode(f, 123, size0+size1, r, size2, nil)
	assert.NoError(t, err)
	crc = crc32.Update(crc, crc32.IEEETable, data2)

	all = make([]byte, size0+size1+size2)
	dr = RangeDecoder(f, 123, nil, 0, size0+size1+size2, size0+size1+size2)
	_, err = io.ReadFull(dr, all)
	assert.NoError(t, err)
	assert.Equal(t, crc, crc32.ChecksumIEEE(all))

	f.Close()
}

type errorReader struct {
	io.Reader
	remain int
}

func (p *errorReader) Read(b []byte) (int, error) {

	if p.remain == 0 {
		return 0, errors.New("errorReader: hit")
	}
	if len(b) > p.remain {
		b = b[:p.remain]
	}
	n, err := p.Reader.Read(b)
	p.remain -= n
	return n, err
}

type errorReaderAt struct {
	io.ReaderAt
	remain int
}

func (p *errorReaderAt) ReadAt(b []byte, off int64) (int, error) {

	if p.remain == 0 {
		return 0, errors.New("errorReader: hit")
	}
	if len(b) > p.remain {
		b = b[:p.remain]
	}
	n, err := p.ReaderAt.ReadAt(b, off)
	p.remain -= n
	return n, err
}

// length = 0
// length < 64K - 4
// length = 64K - 4
// length > 64K - 4
func TestEncodeWriteCloser(t *testing.T) {
	tests := []struct {
		length int64
	}{
		{0},
		{5},
		{64*1024 - 4},
		{64*1024 + 5},
	}

	for i, test := range tests {
		data, edata := getDataAndEdata(test.length)
		body := bytes.NewReader(data)
		ewc := NewEncodeWriteCloser(newCompareWriteCloser(edata))
		n, err := io.Copy(ewc, body)
		assert.NoError(t, err, strconv.Itoa(i))
		assert.Equal(t, test.length, n, strconv.Itoa(i))
		err = ewc.Close()
		assert.NoError(t, err, strconv.Itoa(i))
	}

}

func TestNewEncodeWriteCloser_Concurrency(t *testing.T) {

	ewc := NewEncodeWriteCloser(&nopWriteCloser{os.Stdout})
	go func() {
		b := make([]byte, 1024)
		io.ReadFull(rand.Reader, b)
		for i := 0; i < 1000; i++ {
			ewc.Write(b)
		}
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			ewc.Close()
		}
	}()
}

type nopWriteCloser struct {
	io.Writer
}

func (w *nopWriteCloser) Close() error {
	return nil
}

func getDataAndEdata(size int64) (data []byte, data1 []byte) {
	if size < 0 {
		panic("should not < 0")
	}
	data = randData(size)
	enc := SimpleEncoder(bytes.NewReader(data), nil)
	data1, err := ioutil.ReadAll(enc)
	if err != nil {
		panic("err should be nil")
	}
	return
}

// 对比写入的数据和标准数据是否相同
type compareWriteCloser struct {
	std []byte // 正确的 byte
	off int
}

func newCompareWriteCloser(b []byte) (w *compareWriteCloser) {
	return &compareWriteCloser{b, 0}
}

func (c *compareWriteCloser) Write(p []byte) (n int, err error) {
	if c.off+len(p) > len(c.std) {
		msg := fmt.Sprintf("write more than std data: c.off:%v, len p:%v, len std: %v", c.off, len(p), len(c.std))
		return 0, errors.New(msg)
	}
	for i := 0; i < len(p); i++ {
		if c.std[c.off+i] != p[i] {
			msg := fmt.Sprintf("check writer: not equal slice off:%v, index:%v, std:%v, p:%v", c.off, i, c.std[c.off+i], p[i])
			return 0, errors.New(msg)
		}
	}
	c.off += len(p)
	return len(p), err
}

func (c *compareWriteCloser) Close() (err error) {
	if c.off != len(c.std) {
		msg := fmt.Sprintf("write less than std data len p:%v, len std: %v", c.off, len(c.std))
		return errors.New(msg)
	}
	return nil
}
