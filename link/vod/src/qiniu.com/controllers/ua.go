package controllers

import (
	//"time"
	"encoding/json"
	//"errors"
	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"io/ioutil"
	"qiniu.com/models"
	//"strconv"
)

var (
	UaMod *models.UaModel
)

func init() {
	UaMod = &models.UaModel{}
	UaMod.Init()
}

type uabody struct {
	Uaid      string `json:"uaid"`
	Namespace string `json:"namespace"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
	Password  string `json:"password"`
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

	body, err := ioutil.ReadAll(c.Request.Body)
	var uaData uabody
	err = json.Unmarshal(body, &uaData)

	if err != nil {
		xl.Errorf("parse request body failed, body = %#v", body)
		c.JSON(400, gin.H{
			"error": "read callback body failed",
		})
		return
	}
	user, err := getUserInfo(xl, c.Request)
	if err != nil {
		c.JSON(403, gin.H{
			"error": "get user info falied",
		})
	}

	ua := models.UaInfo{
		Uid:       getUid(user.uid),
		UaId:      params.uaid,
		Namespace: params.namespace,
		Password:  uaData.Password,
	}
	model := models.NamespaceModel{}
	r, err := model.GetNamespaceInfo(xl, getUid(user.uid), params.namespace)
	if err != nil || len(r) == 0 {
		xl.Errorf("namespace is not correct")
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}

	info, err := UaMod.GetUaInfo(xl, uaData.Namespace, params.uaid)
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
		c.JSON(400, gin.H{
			"error": err.Error(),
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
	err = UaMod.Delete(xl, params.namespace, params.uaid)
	if err != nil {
		xl.Errorf("Delete falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
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

	body, err := ioutil.ReadAll(c.Request.Body)
	var uaData uabody
	err = json.Unmarshal(body, &uaData)
	if err != nil {
		xl.Errorf("parse request body failed, body = %#v", body)
		c.JSON(400, gin.H{
			"error": "read callback body failed",
		})
		return
	}

	user, err := getUserInfo(xl, c.Request)
	if err != nil {
		c.JSON(403, gin.H{
			"error": "get user info falied",
		})
	}

	ua := models.UaInfo{
		Uid:       getUid(user.uid),
		UaId:      uaData.Uaid,
		Namespace: uaData.Namespace,
		Password:  uaData.Password,
	}

	model := models.NamespaceModel{}
	r, err := model.GetNamespaceInfo(xl, getUid(user.uid), params.namespace)
	if err != nil || len(r) == 0 {
		xl.Errorf("namespace is not correct")
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}

	if ua.UaId == "" {
		ua.UaId = params.uaid
	}

	if ua.Namespace == "" {
		ua.Namespace = params.namespace
	}

	info, err := UaMod.GetUaInfo(xl, params.namespace, params.uaid)
	if len(info) == 0 || err != nil {
		xl.Errorf("ua is not correct")
		c.JSON(400, gin.H{
			"error": "ua is not correct",
		})
		return
	}

	if ua.Password == "" {
		ua.Password = info[0].Password
	}

	if ua.Namespace != params.namespace {
		info, _ = UaMod.GetUaInfo(xl, ua.Namespace, ua.UaId)
		if len(info) != 0 {
			xl.Errorf("ua is exist")
			c.JSON(400, gin.H{
				"error": "ua is exist",
			})
			return
		}
	}
	err = UaMod.UpdateUa(xl, params.namespace, params.uaid, ua)
	if err != nil {
		xl.Errorf("Update falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
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
	var nextMark = ""
	var r []models.UaInfo
	xl.Infof("limit %d, marker %s, regex %s namespace %s", params.limit, params.marker, params.regex, params.namespace)
	if params.exact {
		r, err = UaMod.GetUaInfo(xl, params.namespace, params.regex)
	} else {
		r, nextMark, err = UaMod.GetUaInfos(xl, params.limit, params.marker, params.namespace, models.UA_ITEM_UAID, params.regex)
	}
	if err != nil {
		xl.Errorf("Get falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	} else {
		c.Header("Content-Type", "application/json")
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(200, gin.H{"item": r,
			"marker": nextMark})
	}
}
