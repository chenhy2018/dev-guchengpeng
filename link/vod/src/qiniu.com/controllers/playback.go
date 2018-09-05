package controllers

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"google.golang.org/grpc"
	"qiniu.com/m3u8"
	"qiniu.com/models"
	pb "qiniu.com/proto"
)

var (
	SegMod           *models.SegmentKodoModel
	fastForwardClint pb.FastForwardClient
)

func init() {
	SegMod = &models.SegmentKodoModel{}
	SegMod.Init()
	gRPCinit()

}
func gRPCinit() {
	//	conn, err := grpc.Dial("47.105.118.51:50051", grpc.WithInsecure())
	conn, err := grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())

	if err != nil {
		fmt.Println("Init gprc failedgrpcgrpc")

	}

	fastForwardClint = pb.NewFastForwardClient(conn)
}

// sample requset url = /playback/12345.m3u8?from=1532499345&to=1532499345&e=1532499345&token=xxxxxx
func GetPlayBackm3u8(c *gin.Context) {

	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	if params.to <= params.from {
		xl.Errorf("bad from/to time, from = %v, to = %v", params.from, params.to)
		c.JSON(400, gin.H{
			"error": "bad from/to time, from great or equal to",
		})
		return
	}

	fullUrl := "http://" + c.Request.Host + c.Request.URL.String()
	if !VerifyToken(xl, params.expire, params.token, fullUrl, params.uid) {
		xl.Errorf("verify token falied")
		c.JSON(401, gin.H{
			"error": "bad token",
		})
		return
	}
	xl.Infof("uid= %v, uaid = %v, from = %v, to = %v, namespace = %v", params.uid, params.uaid, params.from, params.to, params.namespace)

	if params.speed != 1 {
		getFastForwardStream(params, c)
		return
	}
	dayInMilliSec := int64((24 * time.Hour).Seconds() * 1000)
	if (params.to - params.from) > dayInMilliSec {
		xl.Errorf("bad from/to time, from = %v, to = %v", params.from, params.to)
		c.JSON(400, gin.H{
			"error": "bad from/to time, currently we only support playback in 24 hours",
		})
		return
	}
	segs, _, err := SegMod.GetSegmentTsInfo(xl, params.from, params.to, params.namespace, params.uaid, 0, "")
	if err != nil {
		xl.Errorf("getTsInfo error, error =  %#v", err)
		c.JSON(500, nil)
		return
	}

	if len(segs) == 0 {
		c.JSON(200, nil)
		return
	}
	var playlist []map[string]interface{}

	var total int64
	for _, v := range segs {
		start, ok := v[models.SEGMENT_ITEM_START_TIME].(int64)
		if !ok {
			xl.Errorf("start time format error %#v", v)
			c.JSON(500, nil)
			return
		}
		end, ok := v[models.SEGMENT_ITEM_END_TIME].(int64)
		if !ok {
			xl.Errorf("end time format error %#v", v)
			c.JSON(500, nil)
			return
		}
		duration := float64(end-start) / 1000
		total += int64(duration)
		filename, ok := v[models.SEGMENT_ITEM_FILE_NAME].(string)

		if !ok {
			xl.Errorf("filename format error %#v", v)
			c.JSON(500, nil)
			return
		}
		realUrl := GetUrlWithDownLoadToken(xl, "http://pdwjeyj6v.bkt.clouddn.com/", filename, total)

		m := map[string]interface{}{
			"duration": duration,
			"url":      realUrl,
		}
		playlist = append(playlist, m)

	}
	c.Header("Content-Type", "application/x-mpegURL")
	c.Header("Access-Control-Allow-Origin", "*")
	c.String(200, m3u8.Mkm3u8(playlist, xl))
}

func (s *testStream) Read(b []byte) (n int, err error) {
	n = len(s.Stream)
	copy(b, s.Stream)
	return n, nil
}

func (f *testStream) Seek(offset int64, whence int) (ret int64, err error) {
	if whence == io.SeekEnd {
		ret = int64(len(f.Stream))
	}
	return ret, nil
}

type testStream struct {
	Stream []byte
}

func getFastForwardStream(params *requestParams, c *gin.Context) bool {
	req := new(pb.FastForwardInfo)
	req.Baseurl = params.url
	req.From = params.from
	req.To = params.to
	req.Expire = params.expire
	req.Token = params.token
	req.Speed = params.speed

	r, err := fastForwardClint.GetTsSteam(context.Background(), req)
	if err != nil {
		fmt.Println(err)
	}
	if r == nil {
		fmt.Println("get ts file error")
	}

	c.Header("Content-Type", "video/mp2t")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Stream(func(w io.Writer) bool {
		if ret, err := r.Recv(); err == nil {
			c.Header("Accept-Ranges", "bytes")
			w.Write(ret.Stream)
			return true
		}
		return false
	})
	/*
		if ret, err := r.Recv(); err == nil || err != io.EOF {
			//retN := &testStream{}
			//copy(retN.Stream, ret.Stream)
			r := bytes.NewReader(ret.Stream)
			fmt.Println(len(ret.Stream))
			http.ServeContent(c.Writer, c.Request, "test2.ts", time.Now(), r)
		}
	*/
	/*
		c.Stream(func(w io.Writer) bool {
			if ret, err := r.Recv(); err == nil || err != io.EOF {
				fmt.Println(len(ret.Stream))
				//http.ServeContent(w, c.Request, "test2.ts", time.Now(), retN)
				w.Write(ret.Stream)
				return true
			}
			return false
		})
	*/
	return false
	/*
		c.Header("Content-Type", "video/MP2T")
			c.Header("Access-Control-Allow-Origin", "*")
				c.Status(206)
				c.Stream(func(w io.Writer) bool {
					if ret, err := r.Recv(); err == nil || err != io.EOF {
						w.Write(ret.Stream)
						return true
					}
					return false
				})
				return false
	*/
}
