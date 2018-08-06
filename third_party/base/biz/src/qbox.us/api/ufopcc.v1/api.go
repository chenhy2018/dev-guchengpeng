// Please keep track of "qbox.us/qcckeeper/keeperapi/external" package.
package ufopcc

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/qiniu/rpc.v2"
	"github.com/qiniu/rpc.v2/failover"
	"github.com/qiniu/xlog.v1"
)

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

//------------------------------------------------------
// Create a new uapp.

type NewUappArgs struct {
	Name string
}

func (p *Client) NewUapp(l rpc.Logger, args *NewUappArgs) (err error) {

	xl := xlog.NewWith(l)
	params := map[string][]string{
		"name": {args.Name},
	}
	return p.Conn.CallWithForm(xl, nil, "POST", "/uapps", params)
}

//------------------------------------------------------
// Change uapp properties.

type UappModifyArgs struct {
	Uapp       string
	Version    int
	FlavorName string
}

func (p *Client) UappModify(l rpc.Logger, args *UappModifyArgs) (err error) {

	xl := xlog.NewWith(l)
	params := map[string][]string{
		"flavor_name": {args.FlavorName},
		"version":     {strconv.Itoa(args.Version)},
	}
	return p.Conn.CallWithForm(xl, nil, "POST", "/uapps/"+args.Uapp, params)
}

//------------------------------------------------------
// Delete the whole uapp.

type DeleteUappArgs struct {
	Uapp string
}

func (p *Client) DeleteUapp(l rpc.Logger, args *DeleteUappArgs) (err error) {

	xl := xlog.NewWith(l)
	return p.Conn.Call(xl, nil, "DELETE", "/uapps/"+args.Uapp)
}

//------------------------------------------------------
// Get uapp info.

type UappInfoArgs struct {
	Uapp string
}

type UappInfoRet struct {
	Name       string `json:"name" bson:"name"`
	Owner      uint32 `json:"owner" bson:"owner"`
	CreatedAt  int64  `json:"created_at" bson:"created_at"`   // ns
	LastUpdate int64  `json:"last_update" bson:"last_update"` // ns
	Version    uint32 `json:"version" bson:"version"`
	Quota      struct {
		MaxDuplication uint32 `json:"max_duplication" bson:"max_duplication"`
	} `json:"quota" bson:"quota"`
	Flavor struct {
		Name     string `json:"name" bson:"name"`
		Resource struct {
			Cpu  uint64 `json:"cpu" bson:"cpu"`
			Mem  uint64 `json:"mem" bson:"mem"`   // in MB
			Net  uint64 `json:"net" bson:"net"`   // in Kbps
			Disk uint64 `json:"disk" bson:"disk"` // in GB
			Iops uint64 `json:"iops" bson:"iops"`
		} `json:"resource" bson:"resource"`
	} `json:"flavor" bson:"flavor"`
	Duplication uint32            `json:"duplication" bson:"duplication"`
	Envs        map[string]string `json:"envs" bson:"envs"`
}

func (p *Client) UappInfo(l rpc.Logger, args *UappInfoArgs) (ui UappInfoRet, err error) {

	xl := xlog.NewWith(l)
	err = p.Conn.Call(xl, &ui, "GET", "/uapps/"+args.Uapp)
	return
}

//------------------------------------------------------

type rmVersionRet struct {
	Version uint32 `json:"version"`
	Message string `json:"message"`
}

type RmVerionArgs struct {
	Uapp     string
	Versions []uint32
}

// Remove version of some uapp
func (p *Client) RmVersion(l rpc.Logger, args *RmVerionArgs) (ret []rmVersionRet, err error) {
	xl := xlog.NewWith(l)
	verstrs := make([]string, len(args.Versions))
	for i, ver := range args.Versions {
		verstrs[i] = strconv.FormatUint(uint64(ver), 10)
	}
	params := map[string][]string{
		"versions": verstrs,
	}
	err = p.Conn.CallWithForm(xl, &ret, "DELETE", "/uapps/"+args.Uapp+"/version", params)
	return
}

//------------------------------------------------------
// [Admin] Set uapp quota.

type UappQuotaArgs struct {
	Uapp           string
	MaxDuplication uint32
}

func (p *Client) SetUappQuota(l rpc.Logger, args *UappQuotaArgs) (err error) {

	xl := xlog.NewWith(l)
	params := map[string][]string{
		"max_duplication": {strconv.Itoa(int(args.MaxDuplication))},
	}
	return p.Conn.CallWithForm(xl, nil, "POST", "/uapps/"+args.Uapp+"/quota", params)
}

//------------------------------------------------------
// [Admin] Set provider's uapp quota
type ProviderQuotaArgs struct {
	Uid       string
	UappQuota uint32
}

func (p *Client) SetProviderQuota(l rpc.Logger, args *ProviderQuotaArgs) (err error) {

	xl := xlog.NewWith(l)
	params := map[string][]string{
		"quota": {strconv.Itoa(int(args.UappQuota))},
	}
	return p.Conn.CallWithForm(xl, nil, "POST", "/provider/"+args.Uid+"/quota", params)
}

// -----------------------
// build a image

func (p *Client) GetNewImageName(l rpc.Logger, uid uint32, uapp string) (tarName string, err error) {

	params := url.Values{}
	params.Set("uid", strconv.FormatUint(uint64(uid), 10))
	params.Set("uapp", uapp)

	var ret struct {
		TarName string `json:"tar_name"`
	}

	err = p.Conn.CallWithForm(l, &ret, "GET", "/image/name", params)
	return ret.TarName, err
}

func (p *Client) GetNewDockerimage(l rpc.Logger, uapp, desc string) (image, registry, token string, err error) {
	params := url.Values{}
	params.Set("uapp", uapp)
	params.Set("desc", desc)

	var ret struct {
		ImageName       string `json:"image_name"`
		RegistryAddress string `registry_address`
		Token           string `json:"token"`
	}

	err = p.Conn.CallWithForm(l, &ret, "POST", "/dockerimage", params)
	return ret.ImageName, ret.RegistryAddress, ret.Token, err
}

func (p *Client) SetDockerimageState(l rpc.Logger, imageName string, pushSuccess bool, pushDuration int64) (err error) {
	params := url.Values{}
	params.Set("image_name", imageName)
	params.Set("push_success", strconv.FormatBool(pushSuccess))
	params.Set("push_duration", strconv.FormatInt(pushDuration, 10))

	err = p.Conn.CallWithForm(l, nil, "POST", "/dockerimage/state", params)
	return
}

func (p *Client) GetImageInfo(l rpc.Logger, uid uint32, uapp string, ver int) (images []UserImage, err error) {

	params := url.Values{}
	params.Set("uid", strconv.FormatUint(uint64(uid), 10))
	params.Set("uapp", uapp)
	params.Set("version", strconv.Itoa(ver))

	var ret struct {
		Images []UserImage `json:"images"`
	}

	err = p.Conn.CallWithForm(l, &ret, "GET", "/image/info", params)
	return ret.Images, err
}

func (p *Client) BuildImage(l rpc.Logger, tarName string, desc string) (err error) {

	params := url.Values{}
	params.Set("tar_name", tarName)
	params.Set("desc", desc)

	err = p.Conn.CallWithForm(l, nil, "POST", "/image/new", params)
	return
}

//-----------------------------------------
type ModifyEnvArgs struct {
	Uapp      string
	Operation string
	Key       string
	Value     string
}

func (p *Client) ChangeEnvironment(l rpc.Logger, args *ModifyEnvArgs) (err error) {

	params := url.Values{}
	params.Set("op", args.Operation)
	params.Set("key", args.Key)
	params.Set("value", args.Value)

	err = p.Conn.CallWithForm(l, nil, "POST", "/uapps/"+args.Uapp+"/env", params)
	return
}

//------------------------------------------
type ModifyDescArgs struct {
	Uapp    string
	Version int
	Desc    string
}

func (p *Client) ChangeVersionDesc(l rpc.Logger, args *ModifyDescArgs) (err error) {

	params := url.Values{}
	params.Set("desc", args.Desc)

	version := strconv.Itoa(args.Version)

	err = p.Conn.CallWithForm(l, nil, "POST", "/uapps/"+args.Uapp+"/version/"+version+"/desc", params)
	return
}
