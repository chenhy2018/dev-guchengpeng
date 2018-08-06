package jpegext

import (
	"bufio"
	"image"
	"io"
)

// --------------------------------------------------

type reader interface {
	io.Reader
	Peek(int) ([]byte, error)
}

type decoderExtender struct {
	reader
	quality *int
}

func (p *decoderExtender) DecoderExtend() (r io.Reader, quality *int) {
	return p.reader, p.quality
}

// --------------------------------------------------

type Config struct {
	image.Config
	Quality int
}

func DecodeConfig(r io.Reader) (cfg Config, format string, err error) {

	rr, ok := r.(reader)
	if !ok {
		rr = bufio.NewReader(r)
	}
	ext := &decoderExtender{
		rr, &cfg.Quality,
	}
	cfg.Config, format, err = image.DecodeConfig(ext)
	return
}

// --------------------------------------------------
