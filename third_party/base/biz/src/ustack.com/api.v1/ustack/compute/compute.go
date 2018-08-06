// nova
package compute

import (
	"github.com/qiniu/rpc.v2"
	"ustack.com/api.v1/ustack"
)

// --------------------------------------------------

type Client struct {
	ProjectId string
	Conn      ustack.Conn
}

func New(services ustack.Services, project string) Client {

	conn, ok := services.Find("compute")
	if !ok {
		panic("compute api not found")
	}
	return Client{
		ProjectId: project,
		Conn:      conn,
	}
}

func fakeError(err error) bool {

	if rpc.HttpCodeOf(err)/100 == 2 {
		return true
	}
	return false
}

// --------------------------------------------------
// 创建云主机

type Network struct {
	Uuid string `json:"uuid"`
}

type CreateServerArgs struct {
	Server struct {
		Name      string    `json:"name"`
		FlavorRef string    `json:"flavorRef"`
		ImageRef  string    `json:"imageRef"`
		MaxCount  int       `json:"max_count"`
		MinCount  int       `json:"min_count"`
		Networks  []Network `json:"networks"`
	} `json:"server"`
}

type CreateServerRet struct {
	Server struct {
		Id        string `json:"id"`
		AdminPass string `json:"adminPass"`
	} `json:"server"`
}

func (p Client) CreateServer(l rpc.Logger, args *CreateServerArgs) (ret *CreateServerRet, err error) {

	ret = &CreateServerRet{}
	err = p.Conn.CallWithJson(l, ret, "POST", "/v2/"+p.ProjectId+"/servers", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 修改主机名字

func (p Client) ChangeServerName(l rpc.Logger, serverId, name string) (ret ServerDetailRet, err error) {
	type changeSrvNameArgs struct {
		Server struct {
			Name string `json:"name"`
		} `json:"server"`
	}

	args := changeSrvNameArgs{}
	args.Server.Name = name
	err = p.Conn.CallWithJson(l, &ret, "PUT", "/v2/"+p.ProjectId+"/servers/"+serverId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 删除主机

func (p Client) DeleteServer(l rpc.Logger, serverId string) (err error) {

	path := "/v2/" + p.ProjectId + "/servers/" + serverId
	err = p.Conn.Call(l, nil, "DELETE", path)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 列出所有主机基本信息

type sharedAddress struct {
	Addr string `json:"addr"`
	Type string `json:"OS-EXT-IPS:type"`
	// "UOS-EXT-IPS:subnet_name": "shared_subnet",
	// "UOS-EXT-IPS:id": "e37e5768-7dd9-46cb-a644-6f0a57c514a4",
	// "UOS-EXT-IPS:subnet_shared": true,
	// "version": 4,
	// "UOS-EXT-IPS:subnet_id": "0383c257-d730-4bbc-bd20-c7482bc925e8"
}

type addresses struct {
	Shared []sharedAddress `json:"shared"`
}

type ServerBrief struct {
	Addresses addresses `json:"addresses"`
	Id        string    `json:"id"`
	Name      string    `json:"name"`
}

type ListServerBriefRet struct {
	Servers []ServerBrief `json:"servers"`
}

func (p Client) ListServers(l rpc.Logger) (ret *ListServerBriefRet, err error) {
	ret = &ListServerBriefRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2/"+p.ProjectId+"/servers")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 列出所有主机的详细信息

type image struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type flavor struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type volumeAttachment struct {
	Id         string `json:"id"`
	DeviceName string `json:"device_name"`
}

type Metadata struct {
	AdminPass string `json:"admin_pass"`
}

type ServerDetail struct {
	Id      string `json:"id"`
	Status  string `json:"status"`
	Updated string `json:"updated"`

	Addresses addresses `json:"addresses"`
	Image     image     `json:"image"`
	Flavor    flavor    `json:"flavor"`

	UserId   string `json:"user_id"`
	Name     string `json:"name"`
	Created  string `json:"created"`
	TenantId string `json:"tenant_id"`

	KeyName         string             `json:"key_name"`
	VolumesAttached []volumeAttachment `json:"os-extended-volumes:volumes_attached"`

	AdminPass string   `json:"adminPass"`
	Metadata  Metadata `json:"metadata"`
	//...
}

type ServerDetailRet struct {
	Server ServerDetail `json:"server"`
}

type ListServersDetailRet struct {
	Servers []ServerDetail `json:"servers"`
}

func (p Client) ListServersDetail(l rpc.Logger) (ret *ListServersDetailRet, err error) {
	ret = &ListServersDetailRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2/"+p.ProjectId+"/servers/detail")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看单个主机详情

func (p Client) ServerInfo(l rpc.Logger, serverId string) (ret *ServerDetailRet, err error) {
	ret = &ServerDetailRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2/"+p.ProjectId+"/servers/"+serverId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 开启主机

func (p Client) StartServer(l rpc.Logger, serverId string) (err error) {

	type startArgs struct {
		OSStart interface{} `json:"os-start"`
	}
	path := "/v2/" + p.ProjectId + "/servers/" + serverId + "/action"
	err = p.Conn.CallWithJson(l, nil, "POST", path, startArgs{nil})
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 软关机

func (p Client) ShutdownServer(l rpc.Logger, serverId string) (err error) {

	type shutdownArgs struct {
		OSShutdown interface{} `json:"os-shutdown"`
	}
	path := "/v2/" + p.ProjectId + "/servers/" + serverId + "/action"
	err = p.Conn.CallWithJson(l, nil, "POST", path, shutdownArgs{nil})
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 强制关机

func (p Client) StopServer(l rpc.Logger, serverId string) (err error) {

	type stopArgs struct {
		OSStop interface{} `json:"os-stop"`
	}
	path := "/v2/" + p.ProjectId + "/servers/" + serverId + "/action"
	err = p.Conn.CallWithJson(l, nil, "POST", path, stopArgs{nil})
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// Rebuild Server

func (p Client) RebuildServer(l rpc.Logger, serverId, imageId string) (ret *ServerDetailRet, err error) {
	type rebuildArgs struct {
		Rebuild struct {
			ImgRef string `json:"imageRef"`
		} `json:"rebuild"`
	}

	var args rebuildArgs
	args.Rebuild.ImgRef = imageId

	ret = &ServerDetailRet{}
	err = p.Conn.CallWithJson(l, ret, "POST", "/v2/"+p.ProjectId+"/servers/"+serverId+"/action", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// Resize Server

func (p Client) ResizeServer(l rpc.Logger, serverId, flavorId string) (err error) {
	type reseizeArgs struct {
		Resize struct {
			FlavorRef string `json:"flavorRef"`
		} `json:"localResize"`
	}

	var args reseizeArgs
	args.Resize.FlavorRef = flavorId

	err = p.Conn.CallWithJson(l, nil, "POST", "/v2/"+p.ProjectId+"/servers/"+serverId+"/action", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 重启主机

func (p Client) RebootServer(l rpc.Logger, serverId string) (err error) {

	type rebootArgs struct {
		Reboot struct {
			Type string `json:"type"`
		} `json:"reboot"`
	}

	var args rebootArgs
	args.Reboot.Type = "SOFT"

	path := "/v2/" + p.ProjectId + "/servers/" + serverId + "/action"
	err = p.Conn.CallWithJson(l, nil, "POST", path, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 挂载网卡到主机

func (p Client) ServerAttachPort(l rpc.Logger, serverId, portId string) (err error) {
	type attachPortArgs struct {
		Attach struct {
			PortId string `json:"port_id"`
		} `json:"interfaceAttachment"`
	}

	var args attachPortArgs
	args.Attach.PortId = portId
	err = p.Conn.CallWithJson(l, nil, "POST", "/v2/"+p.ProjectId+"/servers/"+serverId+"/os-interface", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 将网卡从主机卸载

func (p Client) ServerDetachPort(l rpc.Logger, serverId, portId string) (err error) {
	err = p.Conn.Call(l, nil, "DELETE", "/v2/"+p.ProjectId+"/servers/"+serverId+"/os-interface/"+portId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 获取 vnc

func (p Client) GetServerVNC(l rpc.Logger, serverId string) (vncUrl string, err error) {
	type vncRet struct {
		Console struct {
			URL  string `json:"url"`
			Type string `json:"type"`
		} `json:"console"`
	}

	type vncArgs struct {
		GetConsole struct {
			Type string `json:"type"`
		} `json:"os-getVNCConsole"`
	}

	var args vncArgs
	args.GetConsole.Type = "novnc"
	var uret vncRet
	path := "/v2/" + p.ProjectId + "/servers/" + serverId + "/action"

	err = p.Conn.CallWithJson(l, &uret, "POST", path, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	vncUrl = uret.Console.URL
	return
}

// --------------------------------------------------
// 修改主机密码

func (p Client) ChangePassword(l rpc.Logger, serverId, newpass string) (err error) {

	type chgpwArgs struct {
		ChangePassword struct {
			AdminPass string `json:"adminPass"`
		} `json:"changePassword"`
	}

	var args chgpwArgs
	args.ChangePassword.AdminPass = newpass

	path := "/v2/" + p.ProjectId + "/servers/" + serverId + "/action"
	err = p.Conn.CallWithJson(l, nil, "POST", path, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 创建密钥

type Keypair struct {
	Name        string `json:"name"`
	PublicKey   string `json:"public_key"`
	PrivateKey  string `json:"private_key"`
	UserId      string `json:"user_id"`
	Fingerprint string `json:"fingerprint"`
}

type CreateKeypairRet struct {
	Keypair Keypair `json:"keypair"`
}

func (p Client) CreateKeypair(l rpc.Logger, keypair string) (ret *CreateKeypairRet, err error) {

	type createArgs struct {
		Keypair struct {
			Name string `json:"name"`
		} `json:"keypair"`
	}

	var args createArgs
	args.Keypair.Name = keypair

	err = p.Conn.CallWithJson(l, nil, "POST", "/v2/"+p.ProjectId+"os-keypairs", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 导入密钥

type ImportKeypairRet struct {
	Keypair Keypair `json:"keypair"`
}

func (p Client) ImportKeypair(l rpc.Logger, keypair, publicKey string) (ret *ImportKeypairRet, err error) {

	type importArgs struct {
		Keypair struct {
			Name      string `json:"name"`
			PublicKey string `json:"public_key"`
		} `json:"keypair"`
	}

	var args importArgs
	args.Keypair.Name = keypair
	args.Keypair.PublicKey = publicKey

	ret = &ImportKeypairRet{}
	err = p.Conn.CallWithJson(l, ret, "POST", "/v2/"+p.ProjectId+"/os-keypairs", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 删除密钥

func (p Client) DeleteKeypair(l rpc.Logger, keypair string) (err error) {

	err = p.Conn.Call(l, nil, "DELETE", "/v2/"+p.ProjectId+"/os-keypairs/"+keypair)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 列出密钥

type KeypairInfo struct {
	Keypair struct {
		Name        string `json:"name"`
		PublicKey   string `json:"public_key"`
		Fingerprint string `json:"fingerprint"`
	} `json:"keypair"`
}

type ListKeypairsRet struct {
	Keypairs []KeypairInfo `json:"keypairs"`
}

func (p Client) ListKeypairs(l rpc.Logger) (ret *ListKeypairsRet, err error) {

	ret = &ListKeypairsRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2/"+p.ProjectId+"/os-keypairs")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 挂载密钥

func (p Client) AttachKeypair(l rpc.Logger, serverId string, keypairs []string) (err error) {

	type attachArgs struct {
		AttachKeypairs struct {
			KeyNames []string `json:"key_names"`
		} `json:"attachKeypairs"`
	}

	var args attachArgs
	args.AttachKeypairs.KeyNames = keypairs

	err = p.Conn.CallWithJson(l, nil, "POST", "/v2/"+p.ProjectId+"/servers/"+serverId+"/action", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 卸载密钥

func (p Client) DetachKeypair(l rpc.Logger, serverId string, keypairs []string) (err error) {

	type detachArgs struct {
		DetachedKeypairs struct {
			KeyNames []string `json:"key_names"`
		} `json:"detachKeypairs"`
	}

	var args detachArgs
	args.DetachedKeypairs.KeyNames = keypairs

	err = p.Conn.CallWithJson(l, nil, "POST", "/v2/"+p.ProjectId+"/servers/"+serverId+"/action", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 列出 flavor 配置(基本信息)

type ListFlavorsRet struct {
	Flavors []flavor `json:"flavors"`
}

func (p Client) ListFlavors(l rpc.Logger) (ret *ListFlavorsRet, err error) {

	ret = &ListFlavorsRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2/"+p.ProjectId+"/flavors")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 列出 flavor 配置(详细信息)

type ListFlavorsDetailRet struct {
	Flavors []Flavor `json:"flavors"`
}

func (p Client) ListFlavorsDetail(l rpc.Logger) (ret *ListFlavorsDetailRet, err error) {

	ret = &ListFlavorsDetailRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2/"+p.ProjectId+"/flavors/detail")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看单个 flavor 配置

type Flavor struct {
	Name  string `json:"name"`
	Ram   int    `json:"ram"` // MB
	Vcpus int    `json:"vcpus"`
	// Swap       string    `json:"swap"` // GB
	// RxtxFactor int    `json:"rxtx_factor"`
	Disk int    `json:"disk"` // GB
	Id   string `json:"id"`
}

type FlavorInfoRet struct {
	Flavor Flavor `json:"flavor"`
}

func (p Client) FlavorInfo(l rpc.Logger, id string) (ret *FlavorInfoRet, err error) {

	ret = &FlavorInfoRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2/"+p.ProjectId+"/flavors/"+id)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 列出主机的虚拟网卡

type fixedIp struct {
	SubnetId  string `json:"subnet_id"`
	IpAddress string `json:"ip_address"`
}

type portInfo struct {
	PortState string    `json:"port_state"`
	FixedIps  []fixedIp `json:"fixed_ips"`
	PortId    string    `json:"port_id"`
	NetId     string    `json:"net_id"`
	MacAddr   string    `json:"mac_addr"`
}

type ListServerPortsRet struct {
	InterfaceAttachments []portInfo `json:"interfaceAttachments"`
}

func (p Client) ListServerPorts(
	l rpc.Logger, id string) (ret *ListServerPortsRet, err error) {

	ret = &ListServerPortsRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2/"+p.ProjectId+"/servers/"+id+"/os-interface")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 挂载硬盘

type MountVolumeArgs struct {
	VolumeAttachment struct {
		VolumeId string `json:"volumeId"`
	} `json:"volumeAttachment"`
}

type MountVolumeRet struct {
	VolumeAttachment struct {
		Id       string `json:"id"`
		Device   string `json:"device"`
		ServerId string `json:"serverId"`
		VolumeId string `json:"volumeId"`
	} `json:"volumeAttachment"`
}

func (p Client) MountVolume(
	l rpc.Logger, serverId, volumeId string) (ret *MountVolumeRet, err error) {

	args := MountVolumeArgs{}
	args.VolumeAttachment.VolumeId = volumeId

	ret = &MountVolumeRet{}
	err = p.Conn.CallWithJson(
		l, ret, "POST", "/v2/"+p.ProjectId+"/servers/"+serverId+"/os-volume_attachments", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 卸载硬盘

func (p Client) UnmountVolume(l rpc.Logger, serverId, volumeId string) (err error) {

	err = p.Conn.Call(l, nil, "DELETE",
		"/v2/"+p.ProjectId+"/servers/"+serverId+"/os-volume_attachments/"+volumeId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 创建虚机快照

func (p Client) CreateServerImage(l rpc.Logger, serverId, name string) (err error) {
	type createSvrImgArgs struct {
		CreateImage struct {
			Name string `json:"name"`
		} `json:"createImage"`
	}

	args := createSvrImgArgs{}
	args.CreateImage.Name = name

	err = p.Conn.CallWithJson(l, nil, "POST", "/v2/"+p.ProjectId+"/servers/"+serverId+"/action", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
