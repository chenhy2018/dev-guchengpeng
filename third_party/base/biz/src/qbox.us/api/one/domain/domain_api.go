package domain

import (
	"net/http"
	"strconv"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
)

type Client struct {
	Host string //功能废除,兼容保留
	Conn *lb.Client
}

func New(host string, t http.RoundTripper) Client {
	cfg := &lb.Config{
		Hosts:    []string{host},
		TryTimes: 1,
	}
	client := lb.New(cfg, t)
	return Client{
		Host: host,
		Conn: client,
	}
}

func NewWithMultiHosts(hosts []string, t http.RoundTripper) Client {
	cfg := &lb.Config{
		Hosts:    hosts,
		TryTimes: uint32(len(hosts)),
	}
	client := lb.New(cfg, t)
	return Client{
		Conn: client,
	}
}

func (p Client) Publish(l rpc.Logger, domain string, owner uint32, tbl string) (err error) {

	err = p.Conn.CallWithForm(l, nil, "/publish", map[string][]string{
		"domain": {domain},
		"owner":  {strconv.FormatUint(uint64(owner), 10)},
		"tbl":    {tbl},
	})
	return
}

func (p Client) PublishWithGlobal(l rpc.Logger, domain string, owner uint32, tbl string, global bool) (err error) {

	err = p.Conn.CallWithForm(l, nil, "/publish", map[string][]string{
		"domain": {domain},
		"owner":  {strconv.FormatUint(uint64(owner), 10)},
		"tbl":    {tbl},
		"global": {strconv.FormatBool(global)},
	})
	return
}

// 需要proxy_auth鉴权
func (p Client) PublishCheck(l rpc.Logger, domain string, tbl string) (err error) {

	err = p.Conn.CallWithForm(l, nil, "/publish/check", map[string][]string{
		"domain": {domain},
		"tbl":    {tbl},
	})
	return
}

func (p Client) PublishCheckWithGlobal(l rpc.Logger, domain string, tbl string, global bool) (err error) {

	err = p.Conn.CallWithForm(l, nil, "/publish/check", map[string][]string{
		"domain": {domain},
		"tbl":    {tbl},
		"global": {strconv.FormatBool(global)},
	})
	return
}

func (p Client) Unpublish(l rpc.Logger, domain string, owner uint32) (err error) {

	err = p.Conn.CallWithForm(l, nil, "/unpublish", map[string][]string{
		"domain": {domain},
		"owner":  {strconv.FormatUint(uint64(owner), 10)},
	})
	return
}

func (p Client) Clearpublish(l rpc.Logger, owner uint32, tbl string, itbl uint32) (err error) {

	err = p.Conn.CallWithForm(l, nil, "/clearpublish", map[string][]string{
		"owner": {strconv.FormatUint(uint64(owner), 10)},
		"tbl":   {tbl},
		"itbl":  {strconv.FormatUint(uint64(itbl), 10)},
	})
	return
}

type GetByDomainRet struct {
	PhyTbl    string `json:"phy"`
	Tbl       string `json:"tbl"`
	Uid       uint32 `json:"owner"`
	Itbl      uint32 `json:"itbl"`
	Refresh   bool   `json:"refresh"`
	Global    bool   `json:"global"`
	Domain    string `json:"domain"`
	AntiLeech `json:"antileech,omitempty" bson:"antileech,omitempty"`
}

func (p Client) GetByDomain(l rpc.Logger, domain string) (ret GetByDomainRet, err error) {

	err = p.Conn.CallWithForm(l, &ret, "/getbydomain", map[string][]string{
		"domain": {domain},
	})
	return
}

type IdomainRet struct {
	Domain string `json:"domain"`
	Cname  string `json:"cname"`
}

func (p Client) Idomain(l rpc.Logger, owner uint32, tbl, channel string) (ret []IdomainRet, err error) {

	err = p.Conn.CallWithForm(l, &ret, "/idomain", map[string][]string{
		"owner":   {strconv.FormatUint(uint64(owner), 10)},
		"tbl":     {tbl},
		"channel": {channel},
	})
	return
}

func (p Client) Domains(l rpc.Logger, owner uint32, tbl string) (domains []string, err error) {

	err = p.Conn.CallWithForm(l, &domains, "/domains", map[string][]string{
		"owner": {strconv.FormatUint(uint64(owner), 10)},
		"tbl":   {tbl},
	})
	return
}

func (p Client) AllDomains(l rpc.Logger) (entries []Entry, err error) {

	err = p.Conn.Call(l, &entries, "/alldomains")
	return
}

func (p Client) SetDomain(l rpc.Logger, domain string, owner uint32, refresh bool) (err error) {
	err = p.Conn.CallWithForm(l, nil, "/setdomain", map[string][]string{
		"domain":  {domain},
		"owner":   {strconv.FormatUint(uint64(owner), 10)},
		"refresh": {strconv.FormatBool(refresh)},
	})
	return
}

func (p Client) SetGlobal(l rpc.Logger, domain string, owner uint32) (err error) {
	err = p.Conn.CallWithForm(l, nil, "/setdomain", map[string][]string{
		"domain": {domain},
		"owner":  {strconv.FormatUint(uint64(owner), 10)},
		"global": {"true"},
	})
	return
}

func (p Client) Republish(l rpc.Logger, domain string, srcOwner uint32, dstOwner uint32, dstTbl string) (err error) {
	err = p.Conn.CallWithForm(l, nil, "/republish", map[string][]string{
		"domain":   {domain},
		"srcOwner": {strconv.FormatUint(uint64(srcOwner), 10)},
		"dstOwner": {strconv.FormatUint(uint64(dstOwner), 10)},
		"dstTbl":   {dstTbl},
	})
	return
}

func (p Client) RepublishAll(l rpc.Logger, srcOwner uint32, srcTbl string, dstOwner uint32, dstTbl string) (err error) {
	err = p.Conn.CallWithForm(l, nil, "/republishall", map[string][]string{
		"srcOwner": {strconv.FormatUint(uint64(srcOwner), 10)},
		"srcTbl":   {srcTbl},
		"dstOwner": {strconv.FormatUint(uint64(dstOwner), 10)},
		"dstTbl":   {dstTbl},
	})
	return
}

type AntiLeech struct {
	ReferWhiteList []string `json:"refer_wl,omitempty" bson:"refer_wl"`
	ReferBlackList []string `json:"refer_bl,omitempty" bson:"refer_bl"`
	ReferNoRefer   bool     `json:"no_refer" bson:"no_refer"`
	AntiLeechMode  int      `json:"anti_leech_mode" bson:"anti_leech_mode"` // 0:off,1:wl,2:bl
	AntiLeechUsed  bool     `json:"anti_leech_used" bson:"anti_leech_used"` // 表示是否设置过,只要设置过了就应该一直为true
	SourceEnabled  bool     `json:"source_enabled" bson:"source_enabled"`
}

func (p Client) SetAntiLeech(l rpc.Logger, domain string, antiLeech AntiLeech) (err error) {

	err = p.Conn.CallWithForm(l, nil, "/setantileech", map[string][]string{
		"domain":          {domain},
		"refer_wl":        antiLeech.ReferWhiteList,
		"refer_bl":        antiLeech.ReferBlackList,
		"no_refer":        {strconv.FormatBool(antiLeech.ReferNoRefer)},
		"anti_leech_mode": {strconv.Itoa(antiLeech.AntiLeechMode)},
		"anti_leech_used": {"true"},
	})
	return
}

type Entry struct {
	Domain    string `json:"domain" bson:"domain"`
	Tbl       string `json:"tbl" bson:"tbl"`
	Owner     uint32 `json:"owner" bson:"owner"`
	Refresh   bool   `json:"refresh,omitempty" bson:"refresh"`
	Global    bool   `json:"global,omitempty" bson:"global"`
	AntiLeech `json:"antileech,omitempty" bson:"antileech,omitempty"`
}
