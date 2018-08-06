package cc

import (
	"strconv"

	"github.com/qiniu/rpc.v2"
)

type VolumeInfo struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"created_at"`
	Size        int    `json:"size"`
	VolumeType  string `json:"volume_type"`
	Attachments []struct {
		ServerId string `json:"server_id"`
	} `attachments`
}

type CreateVolumeArgs struct {
	Name       string
	Size       int
	SnapshotId string
	VolumeType string // "ssd" or "sata"
}

func (r *Service) CreateVolume(l rpc.Logger, args *CreateVolumeArgs) (ret VolumeInfo, err error) {

	params := map[string][]string{
		"name":        {args.Name},
		"size":        {strconv.Itoa(args.Size)},
		"snapshot_id": {args.SnapshotId},
		"volume_type": {args.VolumeType},
	}

	err = r.Conn.CallWithForm(l, &ret, "POST", r.Host+r.ApiPrefix+"/volumes", params)
	return
}

// ------------------------------------------------------------

func (r *Service) UpdateVolume(l rpc.Logger, volumeId, name, desc string) (ret VolumeInfo, err error) {
	params := map[string][]string{
		"name":        {name},
		"description": {desc},
	}

	err = r.Conn.CallWithForm(l, &ret, "PUT", r.Host+r.ApiPrefix+"/volumes/"+volumeId, params)
	return
}

// ------------------------------------------------------------

type VolumeMountInfo struct {
	Id       string `json:"id"`
	Device   string `json:"device"`
	ServerId string `json:"server_id"`
	VolumeId string `json:"volume_id"`
}

func (r *Service) MountVolume(
	l rpc.Logger, serverId, volumeId string) (ret VolumeMountInfo, err error) {

	params := map[string][]string{
		"volume_id": {volumeId},
	}

	err = r.Conn.CallWithForm(l, &ret, "POST",
		r.Host+r.ApiPrefix+"/servers/"+serverId+"/mount", params)
	return
}

func (r *Service) UnmountVolume(l rpc.Logger, serverId, volumeId string) (err error) {

	params := map[string][]string{
		"volume_id": {volumeId},
	}

	return r.Conn.CallWithForm(l, nil, "POST",
		r.Host+r.ApiPrefix+"/servers/"+serverId+"/unmount", params)
}

// ------------------------------------------------------------

func (r *Service) ResizeVolume(l rpc.Logger, volumeId string, newSize int) (err error) {

	params := map[string][]string{
		"new_size": {strconv.Itoa(newSize)},
	}

	return r.Conn.CallWithForm(l, nil, "POST",
		r.Host+r.ApiPrefix+"/volumes/"+volumeId+"/resize", params)
}

// ------------------------------------------------------------

func (r *Service) DeleteVolume(l rpc.Logger, volumeId string) (err error) {

	return r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/volumes/"+volumeId)
}

// ------------------------------------------------------------

type ListVolumesRet []VolumeInfo

func (r *Service) ListVolumes(l rpc.Logger) (ret ListVolumesRet, err error) {

	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/volumes")
	return
}

// ------------------------------------------------------------

func (r *Service) GetVolumeInfo(l rpc.Logger, volumeId string) (ret VolumeInfo, err error) {

	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/volumes/"+volumeId)
	return
}

// ------------------------------------------------------------

type VolSnapshot struct {
	Status    string `json:"status"`
	Name      string `json:"name"`
	Desc      string `json:"description"`
	CreatedAt string `json:"created_at"`
	VolumeId  string `json:"volume_id"`
	Size      int    `json:"size"`
	Id        string `json:"id"`
}

type VolSnapshotRet struct {
	Snapshot VolSnapshot `json:"snapshot"`
}

type MakeVolSnapshotArgs struct {
	VolumeId string
	Name     string
	Desc     string // optional
	Force    string // optional, default true  是否强制创建硬盘快照(当云硬盘被挂载时)
}

func (r *Service) MakeVolSnapshot(l rpc.Logger, args MakeVolSnapshotArgs) (ret VolSnapshotRet, err error) {
	params := map[string][]string{
		"name":  {args.Name},
		"desc":  {args.Desc},
		"force": {args.Force},
	}
	err = r.Conn.CallWithForm(l, &ret, "POST", r.Host+r.ApiPrefix+"/volumes/"+args.VolumeId+"/snapshots", params)
	return
}

// ------------------------------------------------------------

type VolSnapshotList struct {
	Snapshots []VolSnapshot `json:"snapshots"`
}

func (r *Service) GetVolSnapshots(l rpc.Logger) (ret VolSnapshotList, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/snapshots")
	return
}

// ------------------------------------------------------------

func (r *Service) DeleteVolSnapshot(l rpc.Logger, volSnapshotId string) (err error) {
	err = r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/snapshots/"+volSnapshotId)
	return
}

// ------------------------------------------------------------
