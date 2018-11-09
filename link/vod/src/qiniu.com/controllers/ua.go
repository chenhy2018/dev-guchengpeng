package controllers

import (
	//"time"
	"encoding/json"
	//"fmt"
	//"errors"
	"io"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/models"
	//"strconv"
)

type uabody struct {
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
	Password  string `json:"password"`
	Vod       bool   `json:"vod"`
	Live      bool   `json:"live"`
	Online    bool   `json:"online"`
	Expire    int    `json:"expire"`
}

// sample requset url = /v1/namespaces/<Namespace>/uas
func RegisterUa(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	var uaData uabody

	dec := json.NewDecoder(c.Request.Body)
	for {
		if err := dec.Decode(&uaData); err == io.EOF {
			break
		} else if err != nil {
			xl.Errorf("parse request body failed, body = %#v", c.Request.Body)
			c.JSON(400, gin.H{
				"error": "read callback body failed",
			})
			return
		}
	}
	user, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ua := models.UaInfo{
		Uid:       getUid(user.uid),
		UaId:      params.uaid,
		Namespace: params.namespace,
		Password:  uaData.Password,
	}
	model := models.NamespaceModel{}

	r, err := model.GetNamespaceInfo(xl, getUid(user.uid), params.namespace)
	if err != nil {
		xl.Errorf("get namespace error, error =%#v", err)
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}
	if len(r) == 0 {
		xl.Errorf("namespace is not correct")
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}

	info, err := UaMod.GetUaInfo(xl, getUid(user.uid), params.namespace, params.uaid)
	if err != nil {
		xl.Errorf("get ua info failed")
		c.JSON(500, gin.H{
			"error": "Service Internal Error",
		})
		return
	}
	if len(info) != 0 {
		xl.Errorf("ua is exist")
		c.JSON(400, gin.H{
			"error": "ua is exist",
		})
		return
	}

	err = UaMod.Register(xl, ua)
	if err != nil {
		xl.Errorf("Register falied error = %#v", err.Error())
		c.JSON(500, gin.H{
			"error": "Service Internal Error",
		})
		return
	} else {
		c.JSON(200, gin.H{"success": true})
	}
}

// sample requset url = /v1/namespaces/<Namespace>/uas/<Encodedua>
func DeleteUa(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	cond := map[string]interface{}{
		models.ITEM_ID:           getUid(user.uid) + "." + params.namespace + "." + params.uaid,
		models.UA_ITEM_UAID:      params.uaid,
		models.UA_ITEM_NAMESPACE: params.namespace,
	}
	info, err := UaMod.GetUaInfo(xl, getUid(user.uid), params.namespace, params.uaid)
	if err != nil {
		xl.Errorf("get use info error, error =%#v", err)
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	}
	if len(info) == 0 {
		xl.Errorf("ua is not correct")
		c.JSON(400, gin.H{
			"error": "ua is not correct",
		})
		return
	}
	err = UaMod.Delete(xl, cond)
	if err != nil {
		xl.Errorf("Delete falied error = %#v", err.Error())
		c.JSON(500, gin.H{
			"error": "Service Internal Error",
		})
		return
	} else {
		c.JSON(200, gin.H{"success": true})
	}
}

// sample requset url = /v1/namespaces/<Namespace>/uas/<Encodedua>
func UpdateUa(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	uaInfo, err := UaMod.GetUaInfo(xl, getUid(user.uid), params.namespace, params.uaid)
	if err != nil {
		xl.Errorf("get ua info error, error =  %#v", err)
		c.JSON(400, gin.H{"error =  %#v": "err"})
		return
	}
	if len(uaInfo) == 0 {
		xl.Errorf("ua is not correct")
		c.JSON(400, gin.H{
			"error": "ua is not correct",
		})
		return
	}

	uaData := uabody{
		Password: uaInfo[0].Password,
		Vod:      uaInfo[0].Vod,
		Live:     uaInfo[0].Live,
		Online:   uaInfo[0].Online,
		Expire:   uaInfo[0].Expire,
	}

	dec := json.NewDecoder(c.Request.Body)
	for {
		if err := dec.Decode(&uaData); err == io.EOF {
			break
		} else if err != nil {
			xl.Errorf("parse request body failed, body = %#v", c.Request.Body)
			c.JSON(400, gin.H{
				"error": "read callback body failed",
			})
			return
		}
	}

	ua := models.UaInfo{
		Uid:       getUid(user.uid),
		UaId:      params.uaid,
		Namespace: params.namespace,
		Password:  uaData.Password,
		Vod:       uaData.Vod,
		Live:      uaData.Live,
		Online:    uaData.Online,
		Expire:    uaData.Expire,
	}

	model := models.NamespaceModel{}
	r, err := model.GetNamespaceInfo(xl, getUid(user.uid), params.namespace)
	if err != nil {
		xl.Errorf("namespace is not correct")
		c.JSON(500, gin.H{
			"error": "Service Internal Error",
		})
		return
	}
	if len(r) == 0 {
		xl.Errorf("namespace is not correct")
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}

	err = UaMod.UpdateUa(xl, getUid(user.uid), params.namespace, params.uaid, ua)
	if err != nil {
		xl.Errorf("Update falied error = %#v", err.Error())
		c.JSON(500, gin.H{"error": "Service Internal Error"})
		return
	} else {
		c.JSON(200, gin.H{"success": true})
	}
}

// sample requset url = /v1/namespaces/<Namespace>/uas?regex=<Regex>&limit=<Limit>&marker=<Marker>&exact=<Exact>
func GetUaInfo(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var nextMark = ""
	var r []models.UaInfo
	xl.Infof("limit %d, marker %s, prefix %s namespace %s", params.limit, params.marker, params.prefix, params.namespace)
	if params.exact {
		r, err = UaMod.GetUaInfo(xl, getUid(user.uid), params.namespace, params.prefix)
	} else {
		r, nextMark, err = UaMod.GetUaInfos(xl, params.limit, params.marker, getUid(user.uid), params.namespace, params.prefix)
	}
	if err != nil {
		xl.Errorf("Get falied error = %#v", err.Error())
		c.JSON(500, gin.H{
			"error": "Service Internal Error",
		})
		return
	} else {
		c.Header("Content-Type", "application/json")
		c.JSON(200, gin.H{"item": r,
			"marker": nextMark})
	}
}
