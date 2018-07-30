package lbdc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAcl(t *testing.T) {

	cfg := &AclConfig{
		MaxBdCount:  1,
		MaxLbdCount: 2,
		MaxIpCount:  3,
		MaxIdcCount: 4,
		IdcBds:      map[string][]uint16{"bj": []uint16{2, 3, 4, 5, 6, 7}},
	}
	p := NewAcl(cfg)

	// bd++, lbd++, ip++
	releaseFunc, err := p.AcquireWithBd("http://host1:123", 1)
	assert.NoError(t, err)

	// bd err
	_, err = p.AcquireWithBd("http://host1:123", 1)
	assert.Equal(t, err, ErrAccessBd)

	releaseFunc()
	releaseFunc, err = p.AcquireWithBd("http://host1:123", 1)
	assert.NoError(t, err)

	_, err = p.AcquireWithBd("http://host1:123", 1)
	assert.Equal(t, err, ErrAccessBd)

	// lbd++, ip++, idc++
	releaseFunc, err = p.AcquireWithBd("http://host1:123", 2)
	assert.NoError(t, err)

	// lbd err
	_, err = p.AcquireWithBd("http://host1:123", 3)
	assert.Equal(t, err, ErrAccessLbd)

	releaseFunc()
	releaseFunc, err = p.AcquireWithBd("http://host1:123", 3)
	assert.NoError(t, err)

	releaseFunc()
	releaseFunc, err = p.AcquireWithBd("http://host1:123", 2)
	assert.NoError(t, err)

	_, err = p.AcquireWithBd("http://host1:123", 3)
	assert.Equal(t, err, ErrAccessLbd)

	// ip++, idc++
	releaseFunc, err = p.AcquireWithBd("http://host1:234", 3)
	assert.NoError(t, err)

	// ip err
	_, err = p.AcquireWithBd("http://host1:234", 4)
	assert.Equal(t, err, ErrAccessIp)

	releaseFunc()
	releaseFunc, err = p.AcquireWithBd("http://host1:234", 4)
	assert.NoError(t, err)

	releaseFunc()
	releaseFunc, err = p.AcquireWithBd("http://host1:234", 3)
	assert.NoError(t, err)

	_, err = p.AcquireWithBd("http://host1:234", 4)
	assert.Equal(t, err, ErrAccessIp)

	// idc++
	releaseFunc, err = p.AcquireWithBd("http://host2:123", 5)
	assert.NoError(t, err)

	// idc++
	releaseFunc, err = p.AcquireWithBd("http://host2:123", 6)
	assert.NoError(t, err)

	// idc err
	_, err = p.AcquireWithBd("http://host3:123", 7)
	assert.Equal(t, err, ErrAccessIdc)

	releaseFunc()
	releaseFunc, err = p.AcquireWithBd("http://host4:123", 7)
	assert.NoError(t, err)

	releaseFunc()
	releaseFunc, err = p.AcquireWithBd("http://host3:123", 6)
	assert.NoError(t, err)

	_, err = p.AcquireWithBd("http://host3:123", 7)
	assert.Equal(t, err, ErrAccessIdc)

	// other idc
	_, err = p.AcquireWithBd("http://host3:123", 8)
	assert.NoError(t, err)

	// mem stat
	assert.Equal(t, 1, p.bds[1])
	assert.Equal(t, 1, p.bds[2])
	assert.Equal(t, 1, p.bds[3])
	assert.Equal(t, 1, p.bds[5])
	assert.Equal(t, 1, p.bds[6])
	assert.Equal(t, 1, p.bds[8])
	assert.Equal(t, 2, p.lbds["http://host1:123"])
	assert.Equal(t, 1, p.lbds["http://host1:234"])
	assert.Equal(t, 1, p.lbds["http://host2:123"])
	assert.Equal(t, 2, p.lbds["http://host3:123"])
	assert.Equal(t, 3, p.ips["http://host1"])
	assert.Equal(t, 1, p.ips["http://host2"])
	assert.Equal(t, 2, p.ips["http://host3"])
	assert.Equal(t, 2, p.idcs[""])
	assert.Equal(t, 4, p.idcs["bj"])

	p.MaxLbdCount += 1
	p.MaxIpCount += 2

	// lbd++, ip++
	releaseFunc, err = p.Acquire("http://host1:123")
	assert.NoError(t, err)

	// lbd err
	_, err = p.Acquire("http://host1:123")
	assert.Equal(t, err, ErrAccessLbd)

	releaseFunc()
	releaseFunc, err = p.Acquire("http://host1:123")
	assert.NoError(t, err)

	_, err = p.Acquire("http://host1:123")
	assert.Equal(t, err, ErrAccessLbd)

	// ip++
	releaseFunc, err = p.Acquire("http://host1:234")
	assert.NoError(t, err)

	// ip err
	_, err = p.Acquire("http://host1:567")
	assert.Equal(t, err, ErrAccessIp)

	releaseFunc()
	releaseFunc, err = p.Acquire("http://host1:567")
	assert.NoError(t, err)

	releaseFunc()
	releaseFunc, err = p.Acquire("http://host1:234")
	assert.NoError(t, err)

	_, err = p.Acquire("http://host1:567")
	assert.Equal(t, err, ErrAccessIp)

	// other ip
	_, err = p.Acquire("http://host2:123")
	assert.NoError(t, err)

	// mem stat
	assert.Equal(t, 1, p.bds[1])
	assert.Equal(t, 1, p.bds[2])
	assert.Equal(t, 1, p.bds[3])
	assert.Equal(t, 1, p.bds[5])
	assert.Equal(t, 1, p.bds[6])
	assert.Equal(t, 1, p.bds[8])
	assert.Equal(t, 3, p.lbds["http://host1:123"])
	assert.Equal(t, 2, p.lbds["http://host1:234"])
	assert.Equal(t, 2, p.lbds["http://host2:123"])
	assert.Equal(t, 2, p.lbds["http://host3:123"])
	assert.Equal(t, 5, p.ips["http://host1"])
	assert.Equal(t, 2, p.ips["http://host2"])
	assert.Equal(t, 2, p.ips["http://host3"])
	assert.Equal(t, 2, p.idcs[""])
	assert.Equal(t, 4, p.idcs["bj"])
}
