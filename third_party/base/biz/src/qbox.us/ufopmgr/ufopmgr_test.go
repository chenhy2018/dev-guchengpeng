package ufopmgr

import (
	"testing"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/stretchr/testify/assert"
	"qbox.us/api/qconf/ufopg"
)

type mockUfopLister struct{}

func (mul mockUfopLister) List(rpc.Logger) (ret ufopg.ListRet, err error) {

	return ufopg.ListRet{
		Entries: []ufopg.AclEntry{
			ufopg.AclEntry{
				Ufop:    "ufop1",
				AclMode: 0,
				AclList: []uint32{1, 2, 3},
			},
			ufopg.AclEntry{
				Ufop:    "ufop2",
				AclMode: 1,
				AclList: []uint32{4, 5, 6},
			},
			ufopg.AclEntry{
				Ufop:    "ufop3",
				AclMode: 2,
				AclList: []uint32{7, 8, 9},
			},
		},
	}, nil
}

func TestUfopMgr(t *testing.T) {

	lister := mockUfopLister{}
	um := NewUfopMgr(lister, 3)

	time.Sleep(1 * time.Second)

	ok := um.IsValidUfop("ufop", 1)
	assert.Equal(t, ok, false, "test fail")

	ok = um.IsValidUfop("ufop1", 1)
	assert.Equal(t, ok, true, "test fail")

	ok = um.IsValidUfop("ufop1", 4)
	assert.Equal(t, ok, true, "test fail")

	ok = um.IsValidUfop("ufop2", 1)
	assert.Equal(t, ok, false, "test fail")

	ok = um.IsValidUfop("ufop2", 4)
	assert.Equal(t, ok, true, "test fail")

	ok = um.IsValidUfop("ufop3", 0)
	assert.Equal(t, ok, false, "test fail")

	ok = um.IsValidUfop("ufop3", 7)
	assert.Equal(t, ok, true, "test fail")

}
