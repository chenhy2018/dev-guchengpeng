package stater

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fmt"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
	"qbox.us/pfd/api/types"
	"qbox.us/qconf/qconfapi"
)

func TestGidStater(t *testing.T) {

	e := Entry{
		Group:  "xs",
		Dgid:   101,
		EC:     true,
		ECTime: bson.Now(),
	}
	tracker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		b, _ := bson.Marshal(e)
		w.Write(b)
	}))
	defer tracker.Close()

	s := NewGidStater(&qconfapi.Config{
		McHosts:           []string{"127.0.0.1:11211"},
		MasterHosts:       []string{tracker.URL},
		LcacheExpires:     2000,
		LcacheDuration:    5000,
		LcacheChanBufSize: 1000,
	})
	xl := xlog.NewDummy()
	egid := types.EncodeGid(types.Gid{})

	for i := 0; i < 5; i++ {
		dgid, isECed, err := s.State(xl, egid)
		assert.NoError(t, err)
		assert.Equal(t, e.Dgid, dgid)
		assert.Equal(t, e.EC, isECed)
		dgid, isECed, ecTime, err := s.StateWithECtime(xl, egid)
		assert.NoError(t, err)
		assert.Equal(t, e.Dgid, dgid)
		assert.Equal(t, e.EC, isECed)
		assert.Equal(t, e.ECTime, ecTime)
		fmt.Println(e.ECTime, ecTime)
		group, dgid, isECed, err := s.StateWithGroup(xl, egid)
		assert.NoError(t, err)
		assert.Equal(t, e.Dgid, dgid)
		assert.Equal(t, e.EC, isECed)
		assert.Equal(t, e.Group, group)
		e2, err := s.StateEntry(xl, egid)
		assert.NoError(t, err)
		assert.Equal(t, e, e2)
	}
}
