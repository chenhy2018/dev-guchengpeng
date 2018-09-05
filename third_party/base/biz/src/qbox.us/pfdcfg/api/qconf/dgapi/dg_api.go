package dgapi

import (
	"strconv"
	"strings"
	"syscall"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	qconf "qbox.us/qconf/qconfapi"
)

// ---------------------------------------------------------

var ErrNoSuchEntry = httputil.NewError(612, "no such entry")

type Client struct {
	Conn *qconf.Client
}

// ---------------------------------------------------------

type Host struct {
	Url      string
	IsBackup bool
}

func (h Host) String() string {
	if h.IsBackup {
		return h.Url + "(backup)"
	}
	return h.Url
}

type hostsRet struct {
	Hosts    [][2]string `bson:"hosts"`
	IsBackup []bool      `bson:"is_backup"`
	Idc      []string    `bson:"idc"`
	Repair   []bool      `bson:"repair"`
}

func (r Client) Hosts(l rpc.Logger, guid string, dgid uint32) (hosts []Host, err error) {

	var ret hostsRet
	err = r.Conn.Get(l, &ret, makeDgId(guid, dgid), 0)
	if err != nil {
		return
	}

	if len(ret.Hosts) == 0 {
		err = ErrNoSuchEntry
		return
	}

	hosts = make([]Host, len(ret.Hosts))
	for i := range ret.Hosts {
		hosts[i] = Host{
			Url: ret.Hosts[i][0],
		}
		if len(ret.IsBackup) > i {
			hosts[i].IsBackup = ret.IsBackup[i]
		}
	}

	return hosts, nil
}

func (r Client) EcHosts(l rpc.Logger, guid string, dgid uint32) (echosts []string, err error) {

	echosts, _, err = r.EcHostsAndIdc(l, guid, dgid)
	return
}

func (r Client) EcHostsAndIdc(l rpc.Logger, guid string, dgid uint32) (echosts []string, idc []string, err error) {

	var ret hostsRet
	err = r.Conn.Get(l, &ret, makeDgId(guid, dgid), 0)
	if err != nil {
		return
	}

	if len(ret.Hosts) == 0 {
		err = ErrNoSuchEntry
		return
	}

	echosts = make([]string, len(ret.Hosts))
	for i := range ret.Hosts {
		echosts[i] = ret.Hosts[i][1]
	}

	return echosts, ret.Idc, nil
}

func (r Client) EcCheckHosts(l rpc.Logger, guid string, dgid uint32) (echosts []string, err error) {
	var ret hostsRet
	err = r.Conn.Get(l, &ret, makeDgId(guid, dgid), 0)
	if err != nil {
		return
	}

	for i := range ret.Hosts {
		if ret.Repair[i] {
			xl := xlog.NewWith(l)
			xl.Info("skip broken disk", ret.Hosts[i][1])
			continue
		}
		echosts = append(echosts, ret.Hosts[i][1])
	}
	if len(ret.Hosts) == 0 {
		err = ErrNoSuchEntry
		return
	}
	return
}

// ---------------------------------------------------------

const (
	groupPrefix    = "dg:"
	groupPrefixLen = len(groupPrefix)
)

func makeDgId(guid string, dgid uint32) string {
	return "dg:" + guid + ":" + strconv.FormatUint(uint64(dgid), 36)
}

func ParseDgId(id string) (guid string, dgid uint32, err error) {
	if !strings.HasPrefix(id, groupPrefix) {
		return "", 0, syscall.EINVAL
	}
	ss := strings.Split(id[groupPrefixLen:], ":")
	if len(ss) != 2 {
		return "", 0, syscall.EINVAL
	}
	dgid64, err := strconv.ParseUint(ss[1], 36, 32)
	if err != nil {
		return
	}
	guid, dgid = ss[0], uint32(dgid64)
	return
}

// ---------------------------------------------------------
