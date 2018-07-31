package volume

import (
	"github.com/qiniu/rpc.v2"
	"ustack.com/admin_api.v1/ustack"
)

type Client struct {
	ProjectId string
	Conn      ustack.Conn
}

func New(services ustack.Services, project string) Client {
	conn, ok := services.Find("cinder")
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

// ------------------------------------------------------------------
// 获取 quota 设置

type Quota struct {
	QuotaSet struct {
		Gigabytes     uint32 `json:"gigabytes,omitempty"`
		GigabytesSSD  uint32 `json:"gigabytes_ssd,omitempty"`
		GigabytesSATA uint32 `json:"gigabytes_sata,omitempty"`
		Id            string `json:"id,omitempty"`
		Snapshots     int    `json:"snapshots,omitempty"`
		SnapshotsSSD  int    `json:"snapshots_ssd,omitempty"`
		SnapshotsSATA int    `json:"snapshots_sata,omitempty"`
		Volumes       int    `json:"volumes,omitempty"`
		VolumesSSD    int    `json:"volumes_ssd,omitempty"`
		VolumesSATA   int    `json:"volumes_sata,omitempty"`
	} `json:"quota_set"`
}

func (p *Client) GetQuota(l rpc.Logger, projectId string) (ret Quota, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2/"+p.ProjectId+"/os-quota-sets/"+projectId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// ------------------------------------------------------------------
// 修改 quota 设置

func (p *Client) PutQuota(l rpc.Logger, projectId string, args Quota) (ret Quota, err error) {
	err = p.Conn.CallWithJson(l, &ret, "PUT", "/v2/"+p.ProjectId+"/os-quota-sets/"+projectId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}
