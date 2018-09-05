package sstore

import (
	"errors"
	"strconv"

	"qbox.us/fh"
	"qbox.us/fh/fhver"
)

var ErrUnknownBdInfo = errors.New("unknown bd info")

type IdcParser struct {
	bdIdcs map[uint16]uint16
}

func NewIdcParser(strBdIdcs map[string]uint16) (*IdcParser, error) {

	bdIdcs := make(map[uint16]uint16, len(strBdIdcs))
	for k, v := range strBdIdcs {
		bd, err := strconv.ParseUint(k, 10, 16)
		if err != nil {
			return nil, err
		}
		bdIdcs[uint16(bd)] = v
	}

	return &IdcParser{bdIdcs}, nil
}

func (p *IdcParser) Parse(fhb []byte) (uint16, error) {

	if fhver.FhVer(fhb) == fhver.FhUnknown {
		return 0, ErrUnknownBdInfo
	}
	ibd := fh.Ibd(fhb)
	idc, ok := p.bdIdcs[ibd]
	if !ok {
		return 0, ErrUnknownBdInfo
	}
	return idc, nil
}
