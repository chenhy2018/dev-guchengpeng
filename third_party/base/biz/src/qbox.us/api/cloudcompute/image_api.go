package cc

import (
	"github.com/qiniu/rpc.v2"
)

type ImageInfo struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Visibility string `json:"visibility"`
	Size       int    `json:"size"`
	ImageLabel string `json:"image_label"`
	CreatedAt  string `json:"created_at"`
}

type ListImagesRet []ImageInfo

func (r *Service) ListImages(l rpc.Logger) (ret ListImagesRet, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/images")
	return
}

func (r *Service) DeleteImage(l rpc.Logger, imageId string) (err error) {
	err = r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/images/"+imageId)
	return
}
