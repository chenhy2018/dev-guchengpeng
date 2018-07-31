package api

import (
	"encoding/binary"
	"hash"
	"hash/crc64"
	"io"
)

type Crc64StreamReader struct {
	reader io.Reader
	crc    uint64
	table  *crc64.Table
}

func NewCrc64StreamReader(r io.Reader) io.Reader {
	return &Crc64StreamReader{
		reader: r,
		table:  crc64.MakeTable(crc64.ECMA),
	}
}

func (self *Crc64StreamReader) Read(p []byte) (n int, err error) {
	n, err = self.reader.Read(p)
	if n > 0 {
		self.crc = crc64.Update(self.crc, self.table, p[:n])
	}
	return
}

func (self *Crc64StreamReader) Sum64() uint64 {
	return self.crc
}

type Crc64StreamWriter struct {
	writer io.Writer
	h      hash.Hash64
}

func NewCrc64StreamWriter(w io.Writer) io.Writer {
	h1 := crc64.New(crc64.MakeTable(crc64.ECMA))
	h1.Reset()
	return &Crc64StreamWriter{
		writer: w,
		h:      h1,
	}
}

func (self *Crc64StreamWriter) Write(p []byte) (n int, err error) {
	if n, err = self.writer.Write(p); err != nil {
		return
	}
	return self.h.Write(p)
}

// WriteSum64直接将当前writer的内容计算出Crc64，并写入到writer的末尾。
func (self *Crc64StreamWriter) WriteSum64() (crc64Sum uint64, err error) {
	crc64Sum = self.h.Sum64()
	sumBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(sumBuf, crc64Sum)
	_, err = self.writer.Write(sumBuf)
	return
}
