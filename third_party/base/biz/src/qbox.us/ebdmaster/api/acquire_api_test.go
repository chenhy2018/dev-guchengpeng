package api

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"qbox.us/pfd/api/types"
	pfdstgapi "qbox.us/pfdstg/api"
)

func TestStripeBuildInfo(t *testing.T) {
	m := pfdstgapi.MigrateInfo{
		Flag:      1,
		GidCount:  3,
		FidCounts: []uint32{1, 0, 2},
		Gids:      []types.Gid{types.NewGid(1), types.NewGid(2), types.NewGid(3)},
		FidInfos: []pfdstgapi.FileInfo{
			{Fid: 11, Off: 0, Fsize: 1024},
			{Fid: 31, Off: 0, Fsize: 1024},
			{Fid: 32, Off: 1321, Fsize: 101},
		},
		FirstFidDataBegin: 123,
		LastFidDataEnd:    321,
	}
	sbi := &StripeBuildInfo{
		Sid:      0xa23e,
		Dgid:     0xb21c,
		Psectors: [N + M]uint64{},
		Migrate:  m,
	}
	for i := range sbi.Psectors {
		sbi.Psectors[i] = uint64(rand.Int63())
	}

	buf := bytes.NewBuffer(nil)
	err := EncodeStripeBuildInfo(buf, sbi)
	assert.NoError(t, err)
	b := buf.Bytes()

	sbi2, err := ReadStripeBuildInfo(bytes.NewReader(b))
	assert.NoError(t, err)
	sbi.Crc64 = sbi2.Crc64
	sbi.Version = CurrentVersion()
	assert.Equal(t, sbi, sbi2)

	buf = bytes.NewBuffer(nil)
	err = EncodeStripeBuildInfo(buf, sbi2)
	assert.NoError(t, err)
	assert.Equal(t, b, buf.Bytes())

	err = EncodeStripeBuildInfo(buf, sbi)
	assert.NoError(t, err)
	b = buf.Bytes()

	sbi2, err = ReadStripeBuildInfo(bytes.NewReader(b))
	assert.NoError(t, err)

	versions = []int64{
		int64(CurrentVersion()),
	}
	err = EncodeStripeBuildInfo(buf, sbi)
	assert.NoError(t, err)
	b = buf.Bytes()

	sbi2, err = ReadStripeBuildInfo(bytes.NewReader(b))
	assert.NoError(t, err)

	versions = []int64{
		int64(CurrentVersion()-1),
	}
	sbi2, err = ReadStripeBuildInfo(bytes.NewReader(b))
	assert.Error(t, ErrVersion, err.Error())
}

func TestStripeRepairInfo(t *testing.T) {
	sri := &StripeRepairInfo{
		Sid:      0xa23e,
		Bads:     [M]int8{},
		Psectors: [N + M]uint64{},
		Crc32s:   [N + M]uint32{},
	}
	for i := range sri.Bads {
		sri.Bads[i] = int8(rand.Int())
	}
	for i := range sri.Psectors {
		sri.Psectors[i] = uint64(rand.Int63())
	}
	for i := range sri.Crc32s {
		sri.Crc32s[i] = uint32(rand.Int63())
	}

	buf := bytes.NewBuffer(nil)
	err := EncodeStripeRepairInfo(buf, sri)
	assert.NoError(t, err)
	b := buf.Bytes()

	sri2, err := ReadStripeRepairInfo(bytes.NewReader(b))
	assert.NoError(t, err)
	sri.Crc64 = sri2.Crc64
	sri.Version = CurrentVersion()
	assert.Equal(t, sri, sri2)

	buf = bytes.NewBuffer(nil)
	err = EncodeStripeRepairInfo(buf, sri2)
	assert.NoError(t, err)
	assert.Equal(t, b, buf.Bytes())
}

func TestStripeRecycleInfo(t *testing.T) {
	sri := &StripeRecycleBuildInfo{
		Sid:           0xa23e,
		Psectors:      [N + M]uint64{},
		FirstFidBegin: 0,
		LastFidEnd:    1,
		Fids:          []uint64{1, 2, 3, 4},
		FileInfo: []FileInfo{
			FileInfo{
				Soff:     12345,
				Fsize:    54321,
				Suids:    make([]uint64, 5),
				Psectors: make([]uint64, 5),
			},
			FileInfo{
				Soff:     56789,
				Fsize:    98765,
				Suids:    make([]uint64, 10),
				Psectors: make([]uint64, 10),
			},
			FileInfo{
				Soff:     67890,
				Fsize:    9876,
				Suids:    make([]uint64, 8),
				Psectors: make([]uint64, 8),
			},
			FileInfo{
				Soff:     78901,
				Fsize:    10987,
				Suids:    make([]uint64, 1),
				Psectors: make([]uint64, 1),
			},
		},
	}

	buf := bytes.NewBuffer(nil)
	err := EncodeStripeRecycleInfo(buf, sri)
	assert.NoError(t, err)
	b := buf.Bytes()

	sri2, err := ReadStripeRecycleInfo(bytes.NewReader(b))
	assert.NoError(t, err)
	sri.Crc64 = sri2.Crc64
	sri.Version = CurrentVersion()
	assert.Equal(t, sri, sri2)

	buf = bytes.NewBuffer(nil)
	err = EncodeStripeRecycleInfo(buf, sri2)
	assert.NoError(t, err)
	assert.Equal(t, b, buf.Bytes())
}
