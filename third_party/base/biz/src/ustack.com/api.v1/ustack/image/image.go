// glance
package image

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

	conn, ok := services.Find("image")
	if !ok {
		panic("image api not found")
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
// 列出所有可用镜像

type Image struct {
	Id         string   `json:"id"`
	Name       string   `json:"name"`
	Status     string   `json:"status"`
	Visibility string   `json:"visibility"`
	Size       int      `json:"size"`
	CheckSum   string   `json:"checksum"`
	Tags       []string `json:"tags"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
	Self       string   `json:"self"`
	File       string   `json:"file"`
	Schema     string   `json:"schema"`
	ImageLabel string   `json:"image_label"`
}

type ListImagesRet struct {
	Images []Image `json:"images"`
	Next   string  `json:"next"`
}

func (p Client) ListImages(l rpc.Logger) (ret []Image, err error) {
	var uret *ListImagesRet
	next := "/v2/images"

	for next != "" {
		uret = &ListImagesRet{}
		err = p.Conn.Call(l, uret, "GET", next)
		if err != nil {
			if fakeError(err) {
				err = nil
			} else {
				return
			}
		}

		ret = append(ret, uret.Images...)
		next = uret.Next
	}

	return
}

// --------------------------------------------------
// 查看单个镜像

func (p Client) ImageInfo(l rpc.Logger, id string) (ret *Image, err error) {

	ret = &Image{}
	err = p.Conn.Call(l, ret, "GET", "/v2/images/"+id)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 删除镜像

func (p Client) DeleteImage(l rpc.Logger, imageId string) (err error) {
	err = p.Conn.Call(l, nil, "DELETE", "/v2/images/"+imageId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}
