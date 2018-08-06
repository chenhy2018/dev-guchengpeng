// cinder
package volume

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

	conn, ok := services.Find("volume")
	if !ok {
		panic("volume api not found")
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
// 创建一个空白硬盘

type CreateVolumeArgs struct {
	Volume struct {
		Size               int    `json:"size"`
		DisplayName        string `json:"display_name"`
		SnapshotId         string `json:"snapshot_id,omitempty"`
		DisplayDescription string `json:"display_description,omitempty"`
		VolumeType         string `json:"volume_type"` // "ssd" or "sata"
	} `json:"volume"`
}

type CreateVolumeRet struct {
	Volume Volume `json:"volume"`
}

func (p Client) CreateVolume(l rpc.Logger, args *CreateVolumeArgs) (ret *CreateVolumeRet, err error) {

	ret = &CreateVolumeRet{}
	err = p.Conn.CallWithJson(l, ret, "POST", "/v2/"+p.ProjectId+"/volumes", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 列出所有可用硬盘

type attachInfo struct {
	ServerId string `json:"server_id"`
}

type Volume struct {
	Id          string       `json:"id"`
	Name        string       `json:"name"`
	Size        int          `json:"size"`
	Status      string       `json:"status"`
	Description string       `json:"description"`
	CreatedAt   string       `json:"created_at"`
	VolumeType  string       `json:"volume_type"`
	SnapshotId  string       `json:"snapshot_id"`
	Bootable    string       `json:"bootable"`
	Attachments []attachInfo `json:"attachments"`
}

type ListVolumesRet struct {
	Volumes []Volume `json:"volumes"`
}

func (p Client) ListVolumes(l rpc.Logger) (ret *ListVolumesRet, err error) {

	ret = &ListVolumesRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2/"+p.ProjectId+"/volumes/detail")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看单个硬盘

type VolumeInfoRet struct {
	Volume Volume `json:"volume"`
}

func (p Client) VolumeInfo(l rpc.Logger, id string) (ret *VolumeInfoRet, err error) {

	ret = &VolumeInfoRet{}
	err = p.Conn.Call(l, ret, "GET", "/v2/"+p.ProjectId+"/volumes/"+id)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 删除硬盘

func (p Client) DeleteVolume(l rpc.Logger, volumeId string) (err error) {

	path := "/v2/" + p.ProjectId + "/volumes/" + volumeId
	err = p.Conn.Call(l, nil, "DELETE", path)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 硬盘扩容

func (p Client) ResizeVolume(l rpc.Logger, volumeId string, newSize int) (err error) {

	type resizeVolumeArgs struct {
		OSExtend struct {
			NewSize int `json:"new_size"`
		} `json:"os-extend"`
	}

	args := resizeVolumeArgs{}
	args.OSExtend.NewSize = newSize

	err = p.Conn.CallWithJson(l, nil,
		"POST", "/v2/"+p.ProjectId+"/volumes/"+volumeId+"/action", args)

	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 创建硬盘快照

type Snapshot struct {
	Status    string `json:"status"`
	Name      string `json:"name"`
	Desc      string `json:"description"`
	CreatedAt string `json:"created_at"`
	VolumeId  string `json:"volume_id"`
	Size      int    `json:"size"`
	Id        string `json:"id"`
}

type SnapshotRet struct {
	Snapshot Snapshot `json:"snapshot"`
}

func (p Client) SnapshotsVolume(l rpc.Logger, volumeId, name, desc, force string) (ret SnapshotRet, err error) {
	type snapshotsVolumeArgs struct {
		Snapshot struct {
			Name     string `json:"name"`
			VolumeId string `json:"volume_id"`
			Desc     string `json:"description"`
			Force    string `json:"force"`
		} `json:"snapshot"`
	}

	args := snapshotsVolumeArgs{}
	args.Snapshot.Name = name
	args.Snapshot.VolumeId = volumeId
	args.Snapshot.Desc = desc
	args.Snapshot.Force = force

	err = p.Conn.CallWithJson(l, &ret, "POST", "/v2/"+p.ProjectId+"/snapshots", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看硬盘快照简单列表

type SnapshotList struct {
	Snapshots []Snapshot `json:"snapshots"`
}

func (p Client) SnapshotList(l rpc.Logger) (ret SnapshotList, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2/"+p.ProjectId+"/snapshots")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 删除硬盘快照

func (p Client) DelVolumeSnapshot(l rpc.Logger, snapshotId string) (err error) {
	err = p.Conn.Call(l, nil, "DELETE", "/v2/"+p.ProjectId+"/snapshots/"+snapshotId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 修改云硬盘的名字或描述

func (p Client) UpdateVolume(l rpc.Logger, volumeId, name, desc string) (ret VolumeInfoRet, err error) {
	type updateVolumeArgs struct {
		Volume struct {
			Name string `json:"display_name,omitempty"`
			Desc string `json:"display_description,omitempty"`
		} `json:"volume"`
	}

	args := updateVolumeArgs{}
	args.Volume.Name = name
	args.Volume.Desc = desc

	err = p.Conn.CallWithJson(l, &ret, "PUT", "/v2/"+p.ProjectId+"/volumes/"+volumeId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}
