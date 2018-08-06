package cc

import (
	"bytes"
	"crypto/sha1"
	"io"
	"os"
	"qbox.us/api"
)

type MockDc struct {
	Stg map[string][]byte
}

func NewMockDc() *MockDc {
	dc := &MockDc{
		Stg: make(map[string][]byte),
	}
	return dc
}

type ClosingBuffer struct {
	*bytes.Buffer
}

func (cb *ClosingBuffer) Close() (err error) {
	return
}

func (s *MockDc) Get(key []byte) (r io.ReadCloser, length int64, err error) {
	b, ok := s.Stg[string(key)]
	if !ok {
		err = os.ErrNotExist
		return
	}
	reader := bytes.NewBuffer(b)
	return &ClosingBuffer{reader}, int64(len(b)), nil
}

func (s *MockDc) RangeGet(key []byte, from, to int64) (r io.ReadCloser, length int64, err error) {
	b, ok := s.Stg[string(key)]
	if !ok {
		err = os.ErrNotExist
		return
	}
	if from < 0 {
		from = 0
	}
	if to > int64(len(b)) {
		to = int64(len(b))
	}
	b = b[from:to]
	reader := bytes.NewBuffer(b)
	return &ClosingBuffer{reader}, int64(len(b)), nil
}

func (s *MockDc) Set(key []byte, r io.Reader, length int64) (err error) {
	s.Stg[string(key)] = make([]byte, length)
	_, err = io.ReadFull(r, s.Stg[string(key)])
	return
}

func (s *MockDc) SetEx(key []byte, r io.Reader, length int64, checksum []byte) (err error) {

	s.Stg[string(key)] = make([]byte, length)
	_, err = io.ReadFull(r, s.Stg[string(key)])

	h := sha1.New()
	h.Write(s.Stg[string(key)])
	if !bytes.Equal(h.Sum(nil), checksum) {
		err = api.EDataVerificationFail
		s.Stg[string(key)] = nil
		return
	}
	return
}
