package ufopdistro

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/qiniu/rpc.v2"
	"github.com/qiniu/rpc.v2/failover"
	"github.com/qiniu/xlog.v1"
)

// Common type.
type InstanceState int

const (
	STATE_NOEXIST InstanceState = iota
	STATE_RUNNING InstanceState = iota
	STATE_STOPPED InstanceState = iota
	STATE_UNKNOWN InstanceState = iota
)

func (is InstanceState) String() string {
	switch is {
	case STATE_NOEXIST:
		return "NoExist"
	case STATE_RUNNING:
		return "Running"
	case STATE_STOPPED:
		return "Stopped"
	case STATE_UNKNOWN:
		return "Unknown"
	}
	return "bad state"
}

type Client struct {
	Conn *failover.Client
}

func New(hosts []string, tr http.RoundTripper) *Client {

	return &Client{
		Conn: failover.New(hosts, &failover.Config{
			Http: &http.Client{
				Transport: tr,
			},
		}),
	}
}

//-----------------
type UappStatsArgs struct {
	Uapp   string
	Index  int
	Follow bool
}

func (p *Client) GetStats(l rpc.Logger, args *UappStatsArgs) (rc io.ReadCloser, err error) {

	xl := xlog.NewWith(l)
	params := map[string][]string{
		"index": {strconv.Itoa(args.Index)},
		//"follow": {strconv.FormatBool(args.Follow)},
	}
	resp, err := p.Conn.DoRequestWithForm(
		xl, "GET", "/uapps/"+args.Uapp+"/stats", params)
	if err != nil {
		return
	}

	if resp.StatusCode/100 != 2 {
		err = rpc.CallRet(l, nil, resp)
		return
	}
	return resp.Body, nil
}

//------------------
type UappLogsArgs struct {
	Uapp   string
	Index  int
	Type   string
	Tail   int
	Stream bool
}

func (p *Client) GetLogs(l rpc.Logger, args *UappLogsArgs) (rc io.ReadCloser, err error) {

	xl := xlog.NewWith(l)
	params := map[string][]string{
		"index":  {strconv.Itoa(args.Index)},
		"type":   {args.Type},
		"tail":   {strconv.Itoa(args.Tail)},
		"stream": {strconv.FormatBool(args.Stream)},
	}
	resp, err := p.Conn.DoRequestWithForm(
		xl, "GET", "/uapps/"+args.Uapp+"/logs", params)
	if err != nil {
		return
	}

	if resp.StatusCode/100 != 2 {
		err = rpc.CallRet(l, nil, resp)
		return
	}
	return resp.Body, nil
}

//-----------------------
// [Admin] Query operations history

type QueryArg struct {
	Uid    uint32
	From   int64
	To     int64
	Marker string
	Limit  int
}

type InstanceInfo struct {
	// app setting
	Uapp         string        `json:"uapp"`
	Uid          uint32        `json:"uid"`
	Index        int           `json:"index"`
	Version      uint32        `json:"version"`
	HouseName    string        `json:"house_name"`
	DesiredState InstanceState `json:"desired_state"`
	Flavor       string        `json:"flavor"`

	// auto
	Active    bool  `json:"active"`
	CreatedAt int64 `json:"created_at"`
	StoppedAt int64 `json:"stopped_at"`
}

type QueryInstanceRst struct {
	Marker  string         `json:"marker"`
	Results []InstanceInfo `json:"results"`
}

func (p *Client) QueryInstance(l rpc.Logger, args *QueryArg) (result QueryInstanceRst, err error) {

	if args.From == 0 || args.To == 0 {
		err = errors.New("arg from or to should not be nil")
		return
	}
	params := url.Values{}
	params.Set("from", strconv.FormatInt(args.From, 10))
	params.Set("to", strconv.FormatInt(args.To, 10))
	if args.Uid != 0 {
		params.Set("uid", strconv.FormatInt(int64(args.Uid), 10))
	}
	if args.Marker != "" {
		params.Set("marker", args.Marker)
	}
	if args.Limit != 0 {
		params.Set("limit", strconv.Itoa(args.Limit))
	}

	xl := xlog.NewWith(l)
	err = p.Conn.CallWithForm(xl, &result, "POST", "/query/instance", params)
	return
}

//-------------------------------------------------

func (p *Client) HostMigrate(l rpc.Logger, host string) (err error) {
	xl := xlog.NewWith(l)
	return p.Conn.Call(xl, nil, "POST", "/hosts/"+host+"/migrate")
}

//-------------------------------------------------
type MigrateArgs struct {
	InstanceID  string
	Source      string
	Destination string
}

func (p *Client) Migrate(l rpc.Logger, args *MigrateArgs) (err error) {

	xl := xlog.NewWith(l)
	params := map[string][]string{
		"source":      {args.Source},
		"destination": {args.Destination},
	}
	return p.Conn.CallWithForm(xl, nil, "POST", "/instances/"+args.InstanceID+"/migrate", params)
}

//------------------------------------------------------
// [Admin] List all houses.

type HouseInfo struct {
	HouseName     string `json:"house_name" bson:"house_name"`
	HouseHost     string `json:"house_host" bson:"house_host"`
	LastHeartBeat int64  `json:"last_heartbeat" bson:"last_heartbeat"`
	ReadOnly      bool   `json:"read_only" bson:"read_only"`
}

func (p *Client) ListHouses(l rpc.Logger) (ret []HouseInfo, err error) {

	xl := xlog.NewWith(l)
	err = p.Conn.Call(xl, &ret, "GET", "/houses")
	return
}

//------------------------------------------------------
// [Admin] Mark one house to read_only.

type MarkHouseArgs struct {
	HouseName string
	ReadOnly  bool
}

func (p *Client) MarkHouse(l rpc.Logger, args *MarkHouseArgs) (err error) {

	xl := xlog.NewWith(l)
	params := map[string][]string{
		"read_only": {strconv.FormatBool(args.ReadOnly)},
	}
	return p.Conn.CallWithForm(xl, nil, "POST", "/houses/"+args.HouseName, params)
}

//------------------------------------------------------
// Delete service

type PreDeleteUappArgs struct {
	Uapp string
}

func (p *Client) PreDeleteUapp(l rpc.Logger, args *PreDeleteUappArgs) (err error) {

	xl := xlog.NewWith(l)
	return p.Conn.Call(xl, nil, "DELETE", "/uapps/"+args.Uapp)
}

//------------------------------------------------------
// Get duplication of active instances

type GetUappDuplicArgs struct {
	Uapp string
}

func (p *Client) GetUappDuplication(l rpc.Logger, args *GetUappDuplicArgs) (ret int, err error) {
	xl := xlog.NewWith(l)
	err = p.Conn.Call(xl, &ret, "GET", "/uapps/"+args.Uapp+"/Duplication")
	return
}

//------------------------------------------------------
// Start/State instances of uapp.

type UappActionArgs struct {
	Uapp       string
	IndexStart int
	IndexEnd   int
}

func (p *Client) StartUapp(
	l rpc.Logger, args *UappActionArgs) (ret []ActionResult, err error) {

	xl := xlog.NewWith(l)
	params := map[string][]string{
		"command":     {"start"},
		"index_start": {strconv.Itoa(args.IndexStart)},
		"index_end":   {strconv.Itoa(args.IndexEnd)},
	}
	err = p.Conn.CallWithForm(xl, &ret, "POST", "/uapps/"+args.Uapp+"/action", params)
	return
}

func (p *Client) UpgradeUapp(
	l rpc.Logger, args *UappActionArgs, upgradeVersion int) (ret []ActionResult, err error) {

	xl := xlog.NewWith(l)
	params := map[string][]string{
		"command":         {"upgrade"},
		"index_start":     {strconv.Itoa(args.IndexStart)},
		"index_end":       {strconv.Itoa(args.IndexEnd)},
		"upgrade_version": {strconv.Itoa(upgradeVersion)},
	}
	err = p.Conn.CallWithForm(xl, &ret, "POST", "/uapps/"+args.Uapp+"/action", params)
	return
}

func (p *Client) StopUapp(
	l rpc.Logger, args *UappActionArgs) (ret []ActionResult, err error) {

	xl := xlog.NewWith(l)
	params := map[string][]string{
		"command":     {"stop"},
		"index_start": {strconv.Itoa(args.IndexStart)},
		"index_end":   {strconv.Itoa(args.IndexEnd)},
	}
	err = p.Conn.CallWithForm(xl, &ret, "POST", "/uapps/"+args.Uapp+"/action", params)
	return
}

func (p *Client) StateUapp(
	l rpc.Logger, args *UappActionArgs) (ret []ActionResult, err error) {

	xl := xlog.NewWith(l)
	params := map[string][]string{
		"command":     {"state"},
		"index_start": {strconv.Itoa(args.IndexStart)},
		"index_end":   {strconv.Itoa(args.IndexEnd)},
	}
	err = p.Conn.CallWithForm(xl, &ret, "POST", "/uapps/"+args.Uapp+"/action", params)
	return
}

//------------------------------------------------------
// Resize the instances of uapp.

type ResizeUappArgs struct {
	Uapp        string
	Duplication uint32
}

type ActionResult struct {
	Index   int    `json:"index"`
	State   string `json:"state"`
	Error   string `json:"error"`
	Flavor  string `json:"flavor"`
	Version uint32 `json:"version"`

	StartedAt int64 `json:"started_at"`
}

// Return messages include the info of delta part of instances' operations.
func (p *Client) ResizeUapp(
	l rpc.Logger, args *ResizeUappArgs) (ret []ActionResult, err error) {

	xl := xlog.NewWith(l)
	params := map[string][]string{
		"duplication": {strconv.Itoa(int(args.Duplication))},
	}
	err = p.Conn.CallWithForm(xl, &ret, "POST", "/uapps/"+args.Uapp+"/resize", params)
	return
}

//----------------------------------------------------
type FindInstancesArgs struct {
	Uapp  string
	Index int
}

func (p *Client) FindInstances(l rpc.Logger, args *FindInstancesArgs) (ret []InstanceInfo, err error) {
	xl := xlog.NewWith(l)
	params := map[string][]string{
		"index": {strconv.Itoa(args.Index)},
	}
	err = p.Conn.CallWithForm(xl, &ret, "GET", "/uapps/"+args.Uapp+"/instances", params)
	return
}
