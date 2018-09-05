// This package is deprecated, and will be removed soon.
// Please use "qbox.us/api/uc.v2"
package pub

import (
	"net/http"
	"strconv"

	"qbox.us/api/uc"
	"qbox.us/rpc"
)

const FILE_NOT_FOUND_KEY = "errno-404"

type Service struct {
	Host string
	Conn rpc.Client
}

func NewService(host string, t http.RoundTripper) *Service {
	return &Service{host, rpc.Client{&http.Client{Transport: t}}}
}

func (pub Service) Image(bucketName string, srcSiteUrls []string, srcHost string, expires int) (code int, err error) {
	url := pub.Host + "/image/" + bucketName
	for _, srcSiteUrl := range srcSiteUrls {
		url += "/from/" + rpc.EncodeURI(srcSiteUrl)
	}
	if expires != 0 {
		url += "/expires/" + strconv.Itoa(expires)
	}
	if srcHost != "" {
		url += "/host/" + rpc.EncodeURI(srcHost)
	}
	return pub.Conn.Call(nil, url)
}

func (pub Service) Unimage(bucketName string) (code int, err error) {
	return pub.Conn.Call(nil, pub.Host+"/unimage/"+bucketName)
}

func (pub Service) Info(bucketName string) (info uc.BucketInfo, code int, err error) {
	code, err = pub.Conn.Call(&info, pub.Host+"/info/"+bucketName)
	return
}

func (pub Service) AccessMode(bucketName string, mode int) (code int, err error) {
	return pub.Conn.Call(nil, pub.Host+"/accessMode/"+bucketName+"/mode/"+strconv.Itoa(mode))
}

func (pub Service) Separator(bucketName string, sep string) (code int, err error) {
	return pub.Conn.Call(nil, pub.Host+"/separator/"+bucketName+"/sep/"+rpc.EncodeURI(sep))
}

func (pub Service) Style(bucketName string, name string, style string) (code int, err error) {
	return pub.Conn.Call(nil, pub.Host+"/style/"+bucketName+"/name/"+rpc.EncodeURI(name)+"/style/"+rpc.EncodeURI(style))
}

func (pub Service) Unstyle(bucketName string, name string) (code int, err error) {
	return pub.Conn.Call(nil, pub.Host+"/unstyle/"+bucketName+"/name/"+rpc.EncodeURI(name))
}

// ----------------------------------------------------------
