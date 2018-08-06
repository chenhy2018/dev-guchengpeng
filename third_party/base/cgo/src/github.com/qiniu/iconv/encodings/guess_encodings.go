package encodings

import (
	"github.com/qiniu/iconv"
	"github.com/qiniu/log.v1"
	"syscall"
)

// ---------------------------------------------------

const statDeltaMax = 65536

func StatEncodingDelta(cd iconv.Iconv, text []byte, outbuf []byte) (delta int, tlen int, err error) {

	out, inleft, err := cd.Conv(text, outbuf)
	if err != nil {
		return
	}
	if inleft != 0 || (len(out)&1) != 0 {
		err = syscall.EINVAL
		return
	}

	var statDelta int64 = statDeltaMax
	for i := 0; i < len(out); i += 2 {
		wch := int(out[i]) + (int(out[i+1]) << 8)
		statDelta += int64(utf16_delta_table[wch])
	}

	tlen = len(out) >> 1
	delta = int(statDelta / (int64(tlen) + 1))
	return
}

// ---------------------------------------------------

func Encodings(toEncoding string, encodings []string) (cds []iconv.Iconv, err error) {

	cds = make([]iconv.Iconv, 0, len(encodings))
	for _, encoding := range encodings {
		cd, err2 := iconv.Open(toEncoding, encoding)
		if err2 != nil {
			for _, cd := range cds {
				cd.Close()
			}
			return nil, err2
		}
		cds = append(cds, cd)
	}
	return
}

func GuessEncoding(cds []iconv.Iconv, text []byte, outbuf []byte) (encoding int, delta int) {

	encoding, delta = -1, statDeltaMax
	for i := 0; i < len(cds); i++ {
		tdelta, tlen, err := StatEncodingDelta(cds[i], text, outbuf)
		if err != nil {
			log.Debug("GuessEncoding: StatEncodingDelta failed", encoding, err)
			continue
		}
		log.Debug("GuessEncoding: StatEncodingDelta", i, tdelta, tlen)
		if tdelta < delta {
			encoding, delta = i, tdelta
		}
	}
	return
}

// ---------------------------------------------------

type Iconvs struct {
	toUtf8  []iconv.Iconv
	toUtf16 []iconv.Iconv
}

func Open(encodings []string) (cds *Iconvs, err error) {

	toUtf16, err := Encodings("UCS-2LE", encodings)
	if err != nil {
		return
	}
	cds = &Iconvs{nil, toUtf16}

	toUtf8, err := Encodings("UTF8", encodings)
	if err != nil {
		cds.Close()
		return
	}
	cds.toUtf8 = toUtf8
	return
}

func (cds *Iconvs) Close() (err error) {

	for _, cd := range cds.toUtf16 {
		cd.Close()
	}
	cds.toUtf16 = nil

	if cds.toUtf8 != nil {
		for _, cd := range cds.toUtf8 {
			cd.Close()
		}
		cds.toUtf8 = nil
	}
	return nil
}

func (cds *Iconvs) Item(encoding int) iconv.Iconv {

	return cds.toUtf8[encoding]
}

func (cds *Iconvs) Guess(text []byte, outbuf []byte) (encoding int, delta int) {

	return GuessEncoding(cds.toUtf16, text, outbuf)
}

func (cds *Iconvs) Conv(text []byte, outbuf []byte, encodingUtf8 int) (out []byte, converted bool) {

	encoding, _ := GuessEncoding(cds.toUtf16, text, outbuf)
	if encoding < 0 || encoding == encodingUtf8 {
		return
	}
	out, _, _ = cds.toUtf8[encoding].Conv(text, outbuf)
	converted = true
	return
}

// ---------------------------------------------------
