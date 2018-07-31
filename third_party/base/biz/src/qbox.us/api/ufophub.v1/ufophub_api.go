package ufophub

import (
	"io"
	"net/http"
	"strconv"

	"github.com/qiniu/rpc.v2"
	"qbox.us/api/ufopcc.v1"
)

type Client struct {
	Host   string
	Prefix string
	Conn   rpc.Client
}

func New(host, prefix string, t http.RoundTripper) Client {
	client := &http.Client{Transport: t}
	return Client{
		Host:   host,
		Prefix: prefix,
		Conn:   rpc.Client{client},
	}
}

func (c Client) Register(l rpc.Logger, ufop string, acl_mode int, desc string) (err error) {
	params := map[string][]string{
		"ufop":     {ufop},
		"acl_mode": {strconv.Itoa(acl_mode)},
		"desc":     {desc},
	}
	return c.Conn.CallWithForm(l, nil, "POST", c.Host+c.Prefix+"/ufops", params)
}

func (c Client) Unregister(l rpc.Logger, ufop string) (err error) {
	return c.Conn.Call(l, nil, "DELETE", c.Host+c.Prefix+"/ufops/"+ufop)
}

func (c Client) Apply(l rpc.Logger, ufop string) (err error) {
	return c.Conn.Call(l, nil, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/apply")
}

func (c Client) Unapply(l rpc.Logger, ufop string) (err error) {
	return c.Conn.Call(l, nil, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/unapply")
}

type PushRet struct {
	Token           string `json:"token"`
	ImageName       string `json:"image_name"`
	UpdateState     bool   `json:"update_state"`
	RegistryAddress string `json:"registry_address"`
}

func (c Client) Push(l rpc.Logger, ufop, desc string) (image, registry string, update bool, token string, err error) {
	ret := PushRet{}
	params := map[string][]string{
		"desc": {desc},
	}

	err = c.Conn.CallWithForm(l, &ret, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/dockerimage", params)
	if err == nil {
		token = ret.Token
		image = ret.ImageName
		update = ret.UpdateState
		registry = ret.RegistryAddress
	}
	return
}

func (c Client) PushDone(l rpc.Logger, imageName string, success bool, pushDuration int64) (err error) {
	params := map[string][]string{
		"image_name":    {imageName},
		"push_duration": {strconv.FormatInt(pushDuration, 10)},
		"push_success":  {strconv.FormatBool(success)},
	}

	return c.Conn.CallWithForm(l, nil, "POST", c.Host+c.Prefix+"/ufops/dockerimage/state", params)
}

type BuildRet struct {
	Token string `json:"token"`
}

func (c Client) Build(l rpc.Logger, ufop string) (ret BuildRet, err error) {
	err = c.Conn.Call(l, &ret, "GET", c.Host+c.Prefix+"/ufops/"+ufop+"/new")
	return
}

type BuildLogRet struct {
	LogUrl string `json:"log_url"`
}

func (c Client) Buildlog(l rpc.Logger, ufop string, ver int) (logUrl string, err error) {
	ret := BuildLogRet{}
	err = c.Conn.Call(l, &ret, "GET", c.Host+c.Prefix+"/ufops/"+ufop+"/buildlog"+"?version="+strconv.Itoa(ver))
	if err == nil {
		logUrl = ret.LogUrl
	}
	return
}

type ImageInfo struct {
	State    ufopcc.ImageState `json:"state"`
	Version  int               `json:"version"`
	CreateAt int64             `json:"create_at"`
	Desc     string            `json:"desc"`
}

type ImgInfoRet struct {
	Images []ImageInfo `json:"images"`
}

func (c Client) ImageInfo(l rpc.Logger, ufop string, ver int) (ret ImgInfoRet, err error) {
	var query string
	if ver != 0 {
		query = "?version=" + strconv.Itoa(ver)
	}
	err = c.Conn.Call(l, &ret, "GET", c.Host+c.Prefix+"/ufops/"+ufop+"/imageinfo"+query)
	return
}

func (c Client) Setbase(l rpc.Logger, ufop, baseImage string) (err error) {
	params := map[string][]string{
		"base_image": {baseImage},
	}
	return c.Conn.CallWithForm(l, nil, "POST", c.Host+c.Prefix+"/ufops/"+ufop, params)
}

func (c Client) Bind(l rpc.Logger, ufop, bucket, key string) (err error) {
	params := map[string][]string{
		"resource_entry": {"qiniu:" + bucket + ":" + key},
	}
	return c.Conn.CallWithForm(l, nil, "POST", c.Host+c.Prefix+"/ufops/"+ufop, params)
}

func (c Client) List(l rpc.Logger) (ufops []string, err error) {
	err = c.Conn.Call(l, &ufops, "GET", c.Host+c.Prefix+"/ufops")
	return
}

func (c Client) ListSelf(l rpc.Logger) (ufops []UfopInfoDigest, err error) {
	err = c.Conn.Call(l, &ufops, "GET", c.Host+c.Prefix+"/ufops/self")
	return
}

func (c Client) Resize(l rpc.Logger, ufop string, duplication int) (states []InstanceState, err error) {
	params := map[string][]string{
		"duplication": {strconv.Itoa(duplication)},
	}
	err = c.Conn.CallWithForm(l, &states, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/resize", params)
	return
}

func (c Client) Quota(l rpc.Logger, ufop string, max_duplication int) (err error) { // need to be admin
	params := map[string][]string{
		"max_duplication": {strconv.Itoa(max_duplication)},
	}
	return c.Conn.CallWithForm(l, nil, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/quota", params)
}

func (c Client) ProviderQuota(l rpc.Logger, uid string, quota string) (err error) {
	params := map[string][]string{
		"quota": {quota},
	}
	return c.Conn.CallWithForm(l, nil, "POST", c.Host+c.Prefix+"/provider/"+uid+"/quota", params)

}

func (c Client) Migrate(l rpc.Logger, instance_id string, src string, dest string) (err error) { // need to be admin
	params := map[string][]string{
		"instance_id": {instance_id},
		"src":         {src},
		"dest":        {dest},
	}
	return c.Conn.CallWithForm(l, nil, "POST", c.Host+c.Prefix+"/ufops/migrate", params)
}

type InstanceState struct {
	Index   int    `json:"index"`
	State   string `json:"state"`
	Error   string `json:"error"`
	Flavor  string `json:"flavor"`
	Version uint32 `json:"version"`

	StartedAt int64 `json:"started_at"`
}

func (c Client) Start(l rpc.Logger, ufop string, indexStart, indexEnd int) (states []InstanceState, err error) {
	params := map[string][]string{
		"index_start": {strconv.Itoa(indexStart)},
		"index_end":   {strconv.Itoa(indexEnd)},
		"command":     {"start"},
	}
	err = c.Conn.CallWithForm(l, &states, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/action", params)
	return
}

func (c Client) Stop(l rpc.Logger, ufop string, indexStart, indexEnd int) (states []InstanceState, err error) {
	params := map[string][]string{
		"index_start": {strconv.Itoa(indexStart)},
		"index_end":   {strconv.Itoa(indexEnd)},
		"command":     {"stop"},
	}
	err = c.Conn.CallWithForm(l, &states, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/action", params)
	return
}

func (c Client) Upgrade(l rpc.Logger, ufop string, indexStart, indexEnd, upgradeVersion int) (states []InstanceState, err error) {
	params := map[string][]string{
		"index_start":     {strconv.Itoa(indexStart)},
		"index_end":       {strconv.Itoa(indexEnd)},
		"command":         {"upgrade"},
		"upgrade_version": {strconv.Itoa(upgradeVersion)},
	}
	err = c.Conn.CallWithForm(l, &states, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/action", params)
	return
}

func (c Client) State(l rpc.Logger, ufop string, indexStart, indexEnd int) (states []InstanceState, err error) {
	params := map[string][]string{
		"index_start": {strconv.Itoa(indexStart)},
		"index_end":   {strconv.Itoa(indexEnd)},
		"command":     {"state"},
	}
	err = c.Conn.CallWithForm(l, &states, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/action", params)
	return
}

type GetLogsOptions struct {
	Type   string
	Tail   int
	Stream bool
}

func (c Client) GetLogs(
	l rpc.Logger, ufop string, index int, opts *GetLogsOptions) (rc io.ReadCloser, err error) {

	params := map[string][]string{
		"index":  {strconv.Itoa(index)},
		"type":   {opts.Type},
		"tail":   {strconv.Itoa(opts.Tail)},
		"stream": {strconv.FormatBool(opts.Stream)},
	}
	resp, err := c.Conn.DoRequestWithForm(l, "GET", c.Host+c.Prefix+"/ufops/"+ufop+"/logs", params)
	if err != nil {
		return
	}

	if resp.StatusCode/100 != 2 {
		err = rpc.CallRet(l, nil, resp)
		return
	}
	return resp.Body, nil
}

func (c Client) Log(l rpc.Logger, ufop string, follow bool, version int, from string, to string) (rc io.ReadCloser, err error) {
	var params map[string][]string
	var url string

	if follow {
		// get real time log
		params = map[string][]string{
			"type":    {"app"},
			"fop":     {ufop},
			"version": {strconv.Itoa(version)},
		}
		url = "/v1/realtimelog"
	} else {
		// get history log
		params = map[string][]string{
			"type": {"app"},
			"fop":  {ufop},
			"from": {from},
			"to":   {to},
		}
		url = "/v1/log"
	}

	resp, err := c.Conn.DoRequestWithForm(l, "GET", c.Host+c.Prefix+url, params)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		err = rpc.CallRet(l, nil, resp)
		return nil, err
	}

	return resp.Body, nil
}

type QuotaInfo struct {
	MaxDuplication uint32 `json:"max_duplication"`
}

type FlavorInfo struct {
	Name     string `json:"name"`
	Resource struct {
		Cpu  uint64 `json:"cpu"`
		Mem  uint64 `json:"mem"`  // in MB
		Net  uint64 `json:"net"`  // in Kbps
		Disk uint64 `json:"disk"` // in GB
		Iops uint64 `json:"iops"`
	} `json:"resource"`
}

type UfopInfo struct {
	Ufop        string            `json:"ufop"`
	Owner       uint32            `json:"owner"`
	AclMode     byte              `json:"acl_mode"`
	AclList     []uint32          `json:"acl_list"`
	CreateTime  int64             `json:"create_time"`
	Method      byte              `json:"method"`
	Url         string            `json:"url"`
	Desc        string            `json:"desc"`
	Duplication uint32            `json:"duplication"`
	Quota       QuotaInfo         `json:"quota"`
	Version     uint32            `json:"version"`
	Flavor      FlavorInfo        `json:"flavor"`
	Envs        map[string]string `json:"envs"`
}

type UfopInfoDigest struct {
	Ufop       string   `json:"ufop"`
	Owner      uint32   `json:"owner"`
	AclMode    byte     `json:"acl_mode"`
	AclList    []uint32 `json:"acl_list"`
	CreateTime int64    `json:"create_time"`
	Desc       string   `json:"desc"`
}

func (c Client) Info(l rpc.Logger, ufop string) (ui UfopInfo, err error) {
	err = c.Conn.Call(l, &ui, "GET", c.Host+c.Prefix+"/ufops/"+ufop)
	return
}

type HouseInfo struct {
	HouseName     string `json:"house_name"`
	HouseHost     string `json:"house_host"`
	LastHeartBeat int64  `json:"last_heartbeat"`
	ReadOnly      bool   `json:"read_only"`
}

func (c Client) Houses(l rpc.Logger) (houses []HouseInfo, err error) {
	err = c.Conn.Call(l, &houses, "GET", c.Host+c.Prefix+"/houses")
	return
}

func (c *Client) ChangeEnv(l rpc.Logger, ufop, operation, key, value string) (err error) {
	params := map[string][]string{
		"op":    {operation},
		"key":   {key},
		"value": {value},
	}
	return c.Conn.CallWithForm(l, nil, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/env", params)
}

func (c *Client) ChangeAclmode(l rpc.Logger, ufop string, aclMode int) (err error) {
	params := map[string][]string{
		"acl_mode": {strconv.Itoa(aclMode)},
	}

	return c.Conn.CallWithForm(l, nil, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/aclmode", params)
}

func (c *Client) ChangeMethod(l rpc.Logger, ufop string, method int) (err error) {
	params := map[string][]string{
		"method": {strconv.Itoa(method)},
	}

	return c.Conn.CallWithForm(l, nil, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/method", params)
}

func (c *Client) ChangeDesc(l rpc.Logger, ufop string, ver int, desc string) (err error) {
	params := map[string][]string{
		"desc": {desc},
	}

	version := strconv.Itoa(ver)

	if ver <= 0 {
		return c.Conn.CallWithForm(l, nil, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/desc", params)
	} else {
		return c.Conn.CallWithForm(l, nil, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/version/"+version+"/desc", params)
	}
}

func (c *Client) ChangeUrl(l rpc.Logger, ufop string, url string) (err error) {
	params := map[string][]string{
		"url": {url},
	}
	return c.Conn.CallWithForm(l, nil, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"/url", params)
}

func (c Client) SwitchUfopVer(l rpc.Logger, ufop string, ver int) error {
	version := strconv.Itoa(ver)
	return c.Conn.Call(l, nil, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"?version="+version)
}

type RmVersionRet struct {
	Version uint32 `json:"version"`
	Message string `json:"message"`
}

func (c Client) RmVersion(l rpc.Logger, ufop string, versions []string) (ret []RmVersionRet, err error) {
	params := map[string][]string{
		"versions": versions,
	}
	err = c.Conn.CallWithForm(l, &ret, "DELETE", c.Host+c.Prefix+"/ufops/"+ufop+"/version", params)
	return
}

func (c Client) CheckUfopExist(l rpc.Logger, ufop string) (ret bool, err error) {

	params := map[string][]string{
		"ufop": {ufop},
	}
	err = c.Conn.CallWithForm(l, &ret, "GET", c.Host+c.Prefix+"/ufop/exist", params)
	return
}

func (c Client) SwitchUfopFlavor(l rpc.Logger, ufop string, flavor string) error {
	return c.Conn.Call(l, nil, "POST", c.Host+c.Prefix+"/ufops/"+ufop+"?flavor="+flavor)
}

type QueryUfopsArgs struct {
	Uid uint32
}

type UfopBrief struct {
	Ufop        string `json:"ufop"`
	Owner       uint32 `json:"owner"`
	AclMode     byte   `json:"acl_mode"`
	AclCount    uint32 `json:"acl_list"`
	CreateTime  int64  `json:"create_time"`
	Duplication int32  `json:"duplication"`
}

func (c Client) QueryUfops(l rpc.Logger, args *QueryUfopsArgs) (ret []UfopBrief, err error) {
	params := map[string][]string{
		"uid": {strconv.Itoa(int(args.Uid))},
	}
	err = c.Conn.CallWithForm(l, &ret, "GET", c.Host+c.Prefix+"/ufops/query", params)
	return
}

type FindInstancesArgs struct {
	Uapp  string `json:"uapp"`
	Index int    `json:"index"`
}

type InstanceInfo struct {
	// app setting
	Uapp         string `json:"uapp"`
	Uid          uint32 `json:"uid"`
	Index        int    `json:"index"`
	Version      uint32 `json:"version"`
	HouseName    string `json:"house_name"`
	DesiredState string `json:"desired_state"`
	Flavor       string `json:"flavor"`
	CreatedAt    int64  `json:"created_at"`
}

func (c Client) FindInstances(l rpc.Logger, args *FindInstancesArgs) (ret []InstanceInfo, err error) {
	params := map[string][]string{
		"index": {strconv.Itoa(args.Index)},
		"uapp":  {args.Uapp},
	}
	err = c.Conn.CallWithForm(l, &ret, "GET", c.Host+c.Prefix+"/ufops/"+args.Uapp+"/instances", params)
	return
}
