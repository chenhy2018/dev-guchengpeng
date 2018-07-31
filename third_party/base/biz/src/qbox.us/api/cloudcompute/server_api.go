package cc

import (
	"net/http"

	"github.com/qiniu/rpc.v2"
)

// ------------------------------------------------------------------------------------------

type Service struct {
	Host      string
	ApiPrefix string
	Conn      rpc.Client
}

func New(host, prefix string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{host, prefix, rpc.Client{client}}
}

// ------------------------------------------------------------------------------------------

type CreateServerArgs struct {
	Name      string
	FlavorId  string
	ImageId   string
	NetworkId string
}

type CreateServerRet struct {
	Id        string `json:"id"`
	AdminPass string `json:"admin_pass"`
}

func (r *Service) CreateServer(l rpc.Logger, args *CreateServerArgs) (ret CreateServerRet, err error) {

	params := map[string][]string{
		"name":    {args.Name},
		"flavor":  {args.FlavorId},
		"image":   {args.ImageId},
		"network": {args.NetworkId},
	}

	err = r.Conn.CallWithForm(l, &ret, "POST", r.Host+r.ApiPrefix+"/servers", params)
	return
}

// ------------------------------------------------------------------------------------------

func (r *Service) ChangeServerName(l rpc.Logger, serverId, name string) (ret ServerDetail, err error) {
	params := map[string][]string{
		"name": {name},
	}
	err = r.Conn.CallWithForm(l, &ret, "PUT", r.Host+r.ApiPrefix+"/servers/"+serverId, params)
	return
}

// ------------------------------------------------------------------------------------------

func (r *Service) DeleteServer(l rpc.Logger, serverId string) (err error) {
	return r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/servers/"+serverId)
}

// ------------------------------------------------------------------------------------------

type ServerBrief struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ListServersRet []ServerBrief

func (r *Service) ListServers(l rpc.Logger) (ret ListServersRet, err error) {

	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/servers")
	return
}

// ------------------------------------------------------------------------------------------

type ServerFlavor struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type ServerImage struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type ServerAddr struct {
	Addr string `json:"addr,omitempty"`
	Type string `json:"type,omitempty"`
}

type VolumeAttachment struct {
	Id         string `json:"id,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
}

type MetaData struct {
	AdminPass string `json:"admin_pass,omitempty"`
}

type ServerDetail struct {
	Id        string             `json:"id,omitempty"`
	Name      string             `json:"name,omitempty"`
	Status    string             `json:"status,omitempty"`
	Addresses []ServerAddr       `json:"addresses,omitempty"`
	Image     ServerImage        `json:"image,omitempty"`
	Flavor    ServerFlavor       `json:"flavor,omitempty"`
	Keys      string             `json:"keys,omitempty"` // "key1,key2,key3..."
	Volumes   []VolumeAttachment `json:"volumes,omitempty"`
	Created   int64              `json:"created,omitempty"` // 100ns
	AdminPass string             `json:"admin_pass,omitempty"`
	Metadata  MetaData           `json:"metadata,omitempty"`
}

// ------------------------------------------------------------------------------------------

func (r *Service) ListServersDetail(l rpc.Logger) (ret []ServerDetail, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/servers/detail")
	return
}

// ------------------------------------------------------------------------------------------

func (r *Service) ServerInfo(l rpc.Logger, serverId string) (ret ServerDetail, err error) {

	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/servers/"+serverId)
	return
}

// ------------------------------------------------------------------------------------------
func (r *Service) ServerSnapshot(l rpc.Logger, serverId, name string) (err error) {

	params := map[string][]string{
		"name": {name},
	}
	return r.Conn.CallWithForm(l, nil, "POST", r.Host+r.ApiPrefix+"/servers/"+serverId+"/image", params)
}

// ------------------------------------------------------------------------------------------

func (r *Service) StartServer(l rpc.Logger, serverId string) (err error) {

	params := map[string][]string{
		"command": {"start"},
	}
	return r.Conn.CallWithForm(l, nil, "POST", r.Host+r.ApiPrefix+"/servers/"+serverId+"/action", params)
}

func (r *Service) StopServer(l rpc.Logger, serverId string) (err error) {

	params := map[string][]string{
		"command": {"stop"},
	}
	return r.Conn.CallWithForm(l, nil, "POST", r.Host+r.ApiPrefix+"/servers/"+serverId+"/action", params)
}

func (r *Service) ShutdownServer(l rpc.Logger, serverId string) (err error) {

	params := map[string][]string{
		"command": {"shutdown"},
	}
	return r.Conn.CallWithForm(l, nil, "POST", r.Host+r.ApiPrefix+"/servers/"+serverId+"/action", params)
}

func (r *Service) RebootServer(l rpc.Logger, serverId string) (err error) {

	params := map[string][]string{
		"command": {"reboot"},
	}
	return r.Conn.CallWithForm(l, nil, "POST", r.Host+r.ApiPrefix+"/servers/"+serverId+"/action", params)
}

func (r *Service) RebuildServer(l rpc.Logger, serverId, imageId string) (ret ServerDetail, err error) {
	params := map[string][]string{
		"command": {"rebuild/" + imageId},
	}
	err = r.Conn.CallWithForm(l, &ret, "POST", r.Host+r.ApiPrefix+"/servers/"+serverId+"/action", params)
	return
}

func (r *Service) ResizeServer(l rpc.Logger, serverId, flavorId string) (err error) {
	params := map[string][]string{
		"command": {"resize/" + flavorId},
	}
	err = r.Conn.CallWithForm(l, nil, "POST", r.Host+r.ApiPrefix+"/servers/"+serverId+"/action", params)
	return
}

// ------------------------------------------------------------------------------------------

func (r *Service) ChangeServerPasswd(l rpc.Logger, serverId, newPasswd string) (err error) {

	params := map[string][]string{
		"new_passwd": {newPasswd},
	}
	return r.Conn.CallWithForm(l, nil, "POST", r.Host+r.ApiPrefix+"/servers/"+serverId+"/passwd", params)
}

// ------------------------------------------------------------------------------------------

func (r *Service) AttachInterface(l rpc.Logger, serverId, portId string) (err error) {
	params := map[string][]string{
		"port_id": {portId},
	}
	return r.Conn.CallWithForm(l, nil, "POST", r.Host+r.ApiPrefix+"/servers/"+serverId+"/interfaces", params)
}

// ------------------------------------------------------------------------------------------

func (r *Service) DetachInterface(l rpc.Logger, serverId, portId string) (err error) {
	return r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/servers/"+serverId+"/interfaces/"+portId)
}

// ------------------------------------------------------------------------------------------

func (r *Service) ImportKeypair(l rpc.Logger, name, publickey string) (err error) {

	params := map[string][]string{
		"name":       {name},
		"public_key": {publickey},
	}

	return r.Conn.CallWithForm(l, nil, "POST", r.Host+r.ApiPrefix+"/keypairs", params)
}

// ------------------------------------------------------------------------------------------

func (r *Service) DeleteKeypair(l rpc.Logger, name string) (err error) {

	return r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/keypairs/"+name)
}

// ------------------------------------------------------------------------------------------

func (r *Service) AttachKeypair(l rpc.Logger, serverId string, keypairs []string) (err error) {

	params := map[string][]string{
		"key_name": keypairs,
	}
	return r.Conn.CallWithForm(l, nil, "POST",
		r.Host+r.ApiPrefix+"/servers/"+serverId+"/attach", params)
}

func (r *Service) DetachKeypair(l rpc.Logger, serverId string, keypairs []string) (err error) {

	params := map[string][]string{
		"key_name": keypairs,
	}
	return r.Conn.CallWithForm(l, nil, "POST",
		r.Host+r.ApiPrefix+"/servers/"+serverId+"/detach", params)
}

// ------------------------------------------------------------------------------------------

type KeypairBrief struct {
	Name        string `json:"name"`
	PublicKey   string `json:"public_key"`
	Fingerprint string `json:"fingerprint"`
}

type ListKeypairsRet []KeypairBrief

func (r *Service) ListKeypairs(l rpc.Logger) (ret ListKeypairsRet, err error) {

	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/keypairs")
	return
}

// ------------------------------------------------------------------------------------------

type FixedIp struct {
	IpAddress string `json:"ip_address"`
}

type PortInfo struct {
	PortId    string    `json:"port_id"`
	PortState string    `json:"port_state"`
	MacAddr   string    `json:"mac_addr"`
	FixedIps  []FixedIp `json:"fixed_ips"`
}

type ListServerPortsRet []PortInfo

func (r *Service) ListServerPorts(l rpc.Logger, serverId string) (ret ListServerPortsRet, err error) {

	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/servers/"+serverId+"/ports")
	return
}
