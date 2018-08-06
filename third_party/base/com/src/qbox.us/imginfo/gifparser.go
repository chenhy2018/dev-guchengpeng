// GIF specification: http://www.w3.org/Graphics/GIF/spec-gif89a.txt
// Qiniu BPL of GIF:  https://github.com/qbox/bpl/blob/develop/formats/gif.bpl
// Golang Reference:  https://golang.org/src/image/gif/reader.go
package imginfo

import (
	"bufio"
	"fmt"
	"image/color"
	"io"
)

type GifRichInfo struct {
	Width  int
	Height int

	FrameNumber     int
	BackgroundIndex byte
	Version         string
	LoopCount       int
	DelayTime       int
	ColorModel      color.Palette
}

func GetGifRichInfo(r io.Reader) (GifRichInfo, error) {

	var d decoder
	if err := d.decode(r); err != nil {
		return GifRichInfo{}, err
	}

	return GifRichInfo{
		Width:           d.width,
		Height:          d.height,
		FrameNumber:     d.imageNumber,
		ColorModel:      d.globalColorTable,
		BackgroundIndex: d.backgroundIndex,
		Version:         d.vers,
		LoopCount:       d.loopCount,
		DelayTime:       d.delayTime,
	}, nil
}

const (
	// Fields.
	fColorTable         = 1 << 7
	fInterlace          = 1 << 6
	fColorTableBitsMask = 7

	// Graphic control flags.
	gcTransparentColorSet = 1 << 0
	gcDisposalMethodMask  = 7 << 2
)

// Section indicators.
const (
	sExtension       = 0x21
	sImageDescriptor = 0x2C
	sTrailer         = 0x3B
)

// Extensions.
const (
	eText           = 0x01 // Plain Text
	eGraphicControl = 0xF9 // Graphic Control
	eComment        = 0xFE // Comment
	eApplication    = 0xFF // Application
)

type reader interface {
	io.Reader
	io.ByteReader
}

type decoder struct {
	r reader

	// From header.
	vers            string
	width           int
	height          int
	loopCount       int
	delayTime       int
	backgroundIndex byte
	disposalMethod  byte

	// Computed.
	globalColorTable color.Palette

	// From image data.
	imageNumber int

	// From image descriptor.
	imageFields byte

	// Used when decoding.
	tmp [1024]byte // must be at least 768 so we can read color table
}

func (d *decoder) decode(r io.Reader) error {

	if rr, ok := r.(reader); ok {
		d.r = rr
	} else {
		d.r = bufio.NewReader(r)
	}

	err := d.readHeaderAndScreenDescriptor()
	if err != nil {
		return err
	}

	for {
		c, err := d.r.ReadByte()
		if err != nil {
			return err
		}
		switch c {
		case sExtension:
			if err := d.readExtension(); err != nil {
				return err
			}
		case sImageDescriptor:
			if err := d.readImageDescriptor(); err != nil {
				return err
			}
			useLocalColorTable := d.imageFields&fColorTable != 0
			if useLocalColorTable {
				if _, err = d.readColorTable(d.imageFields); err != nil {
					return err
				}
			}

			litWidth, err := d.r.ReadByte()
			if err != nil {
				return err
			}
			if litWidth < 2 || litWidth > 8 {
				return fmt.Errorf("gif: pixel size in decode out of range: %d", litWidth)
			}
			for {
				n, err := d.readBlock()
				if err != nil {
					return err
				}
				if n == 0 {
					break
				}
			}
			d.imageNumber++

		case sTrailer:
			if d.imageNumber == 0 {
				return io.ErrUnexpectedEOF
			}
			return nil
		default:
			return fmt.Errorf("gif: unknown block type: 0x%.2x", c)
		}
	}
}

func (d *decoder) readColorTable(fields byte) (color.Palette, error) {
	n := 1 << (1 + uint(fields&fColorTableBitsMask))
	_, err := io.ReadFull(d.r, d.tmp[:3*n])
	if err != nil {
		return nil, fmt.Errorf("gif: short read on color table: %s", err)
	}
	j, p := 0, make(color.Palette, n)
	for i := range p {
		p[i] = color.RGBA{d.tmp[j+0], d.tmp[j+1], d.tmp[j+2], 0xFF}
		j += 3
	}
	return p, nil
}

func (d *decoder) readImageDescriptor() error {

	if _, err := io.ReadFull(d.r, d.tmp[:9]); err != nil {
		return fmt.Errorf("gif: can't read image descriptor: %s", err)
	}
	//left := int(d.tmp[0]) + int(d.tmp[1])<<8
	//top := int(d.tmp[2]) + int(d.tmp[3])<<8
	//width := int(d.tmp[4]) + int(d.tmp[5])<<8
	//height := int(d.tmp[6]) + int(d.tmp[7])<<8
	d.imageFields = d.tmp[8]
	return nil
}

func (d *decoder) readHeaderAndScreenDescriptor() error {
	_, err := io.ReadFull(d.r, d.tmp[:13])
	if err != nil {
		return err
	}
	d.vers = string(d.tmp[:6])
	if d.vers != "GIF87a" && d.vers != "GIF89a" {
		return fmt.Errorf("gif: can't recognize format %s", d.vers)
	}
	d.width = int(d.tmp[6]) + int(d.tmp[7])<<8
	d.height = int(d.tmp[8]) + int(d.tmp[9])<<8
	if fields := d.tmp[10]; fields&fColorTable != 0 {
		d.backgroundIndex = d.tmp[11]
		// readColorTable overwrites the contents of d.tmp, but that's OK.
		if d.globalColorTable, err = d.readColorTable(fields); err != nil {
			return err
		}
	}
	// d.tmp[12] is the Pixel Aspect Ratio, which is ignored.
	return nil
}

func (d *decoder) readExtension() error {
	extension, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	size := 0
	switch extension {
	case eText:
		size = 13
	case eGraphicControl:
		return d.readGraphicControl()
	case eComment:
		// nothing to do but read the data.
	case eApplication:
		b, err := d.r.ReadByte()
		if err != nil {
			return err
		}
		// The spec requires size be 11, but Adobe sometimes uses 10.
		size = int(b)
	default:
		return fmt.Errorf("gif: unknown extension 0x%.2x", extension)
	}
	if size > 0 {
		if _, err := io.ReadFull(d.r, d.tmp[:size]); err != nil {
			return err
		}
	}

	// Application Extension with "NETSCAPE2.0" as string and 1 in data means
	// this extension defines a loop count.
	if extension == eApplication && string(d.tmp[:size]) == "NETSCAPE2.0" {
		n, err := d.readBlock()
		if n == 0 || err != nil {
			return err
		}
		if n == 3 && d.tmp[0] == 1 {
			d.loopCount = int(d.tmp[1]) | int(d.tmp[2])<<8
		}
	}
	for {
		n, err := d.readBlock()
		if n == 0 || err != nil {
			return err
		}
	}
}

func (d *decoder) readGraphicControl() error {
	if _, err := io.ReadFull(d.r, d.tmp[:6]); err != nil {
		return fmt.Errorf("gif: can't read graphic control: %s", err)
	}
	flags := d.tmp[1]
	d.disposalMethod = (flags & gcDisposalMethodMask) >> 2
	d.delayTime = int(d.tmp[2]) | int(d.tmp[3])<<8
	//if flags&gcTransparentColorSet != 0 {
	//	d.transparentIndex = d.tmp[4]
	//	d.hasTransparentIndex = true
	//}
	return nil
}

func (d *decoder) readBlock() (int, error) {
	n, err := d.r.ReadByte()
	if n == 0 || err != nil {
		return 0, err
	}
	return io.ReadFull(d.r, d.tmp[:n])
}
