package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	xlog "github.com/qiniu/xlog.v1"
	"io/ioutil"
	"net/http"
	"qiniu.com/models"
)

var (
	namespaceMod *models.NamespaceModel
)

func init() {
	namespaceMod = &models.NamespaceModel{}
	namespaceMod.Init()
}

type namespacebody struct {
	Uid       string `json:"uid"`
	Bucket    string `json:"bucket"`
	Domain    string `json:"domain"`
	Namespace string `json:"namespace"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

func getDomain(xl *xlog.Logger, bucket string) (string, error) {
	client := http.Client{}
	url := fmt.Sprintf("http://api.qiniu.com/v6/domain/list?tbl=%s", bucket)
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		xl.Errorf("%#v", err)
	}
	mac := qbox.NewMac("JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ", "G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS")
	token, _ := mac.SignRequest(request)

	request.Header.Set("Authorization", "QBox "+token)
	xl.Infof("%#v", request)
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	xl.Infof("%#v", resp)
	body, err := ioutil.ReadAll(resp.Body)
	var domain []string
	err = json.Unmarshal(body, &domain)
	if err != nil || len(domain) == 0 {
		return "", err
	}
	return domain[0], err
}

func checkbucket(xl *xlog.Logger, bucket string) error {
	xl.Infof("bucket %s", bucket)
	info, err := namespaceMod.GetNamespaceByBucket(xl, bucket)
	if err != nil {
		xl.Infof("%s", err.Error())
		if err.Error() != "not found" {
			return err
		}
	}
	if len(info) != 0 {
		return fmt.Errorf("bucket is already register")
	}
	return nil
}

// sample requset url = /v1/namespaces/<Namespace>
func RegisterNamespace(c *gin.Context) {
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
	if err != nil {
		xl.Errorf("parse request body failed, body = %#v", body)
		c.JSON(400, gin.H{
			"error": "read callback body failed",
		})
		return
	}
	xl.Infof("%s", body)
	var namespaceData namespacebody
	err = json.Unmarshal(body, &namespaceData)
	xl.Infof("%#v", namespaceData)

	if err != nil {
		xl.Errorf("parse request body failed, body = %#v", body)
		c.JSON(400, gin.H{
			"error": "read callback body failed",
		})
		return
	}
	xl.Infof("%s %s", namespaceData.Namespace, params.namespace)
	err = checkbucket(xl, namespaceData.Bucket)
	if err != nil {
		xl.Errorf("bucket is already register, err = %#v", err)
		c.JSON(403, gin.H{
			"error": "bucket is already register",
		})
		return
	}

	domain, err := getDomain(xl, namespaceData.Bucket)
	if err != nil || domain == "" {
		xl.Errorf("bucket is not correct, err = %#v", err)
		c.JSON(403, gin.H{
			"error": "bucket is not correct",
		})
		return
	}
	xl.Infof("domain %s", domain)
	namespace := models.NamespaceInfo{
		Uid:    params.uid,
		Space:  params.namespace,
		Bucket: namespaceData.Bucket,
		Domain: domain,
	}
	err = namespaceMod.Register(xl, namespace)
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

// sample requset url = /v1/namespaces/<Namespace>
func DeleteNamespace(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	err = namespaceMod.Delete(xl, params.uid, params.namespace)
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

//need test.
func updateNamespace(xl *xlog.Logger, uid, space, newSpace string) error {
	if space != newSpace && newSpace != "" {
		err := namespaceMod.UpdateNamespace(xl, uid, space, newSpace)
		if err != nil {
			return err
		}
		model := models.UaModel{}
		mark := ""
		for {
			uas, nextmark, err := model.GetUaInfos(xl, 0, mark, space, models.UA_ITEM_UAID, "")
			if err != nil {
				return err
			}
			for i := 0; i < len(uas); i++ {
				model.UpdateNamespace(xl, uas[i].Namespace, uas[i].UaId, newSpace)
			}
			if nextmark != "" {
				mark = nextmark
			} else {
				break
			}
		}
	}
	return nil
}

func updateBucket(xl *xlog.Logger, uid, space, bucket, newBucket, domain string) error {
	if bucket != newBucket && newBucket != "" {
		err := checkbucket(xl, newBucket)
		if err != nil {
			xl.Errorf("bucket is already register, err = %#v", err)
			return err
		}
		err = namespaceMod.UpdateBucket(xl, uid, space, bucket, domain)
		if err != nil {
			xl.Errorf("Update falied error = %#v", err.Error())
			return err
		}
	}
	return nil
}

// sample requset url = /v1/namespaces/<Namespace>
func UpdateNamespace(c *gin.Context) {
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
	xl.Infof("%s", body)
	var namespaceData namespacebody
	err = json.Unmarshal(body, &namespaceData)
	if err != nil || namespaceData.Uid == "" {
		xl.Errorf("parse request body failed, body = %#v", body)
		c.JSON(400, gin.H{
			"error": "read callback body failed",
		})
		return
	}

	domain, err := getDomain(xl, namespaceData.Bucket)
	if err != nil || domain == "" {
		xl.Errorf("bucket is not correct, err = %#v", err)
		c.JSON(403, gin.H{
			"error": "bucket is not correct",
		})
		return
	}
	xl.Infof("%#v", namespaceData)
	/*
	   namespace := models.NamespaceInfo{
	           Uid  : namespaceData.Uid,
	           Space : namespaceData.Namespace,
	           Bucket  : namespaceData.Bucket,
	           Domain : domain,
	   }
	*/
	xl.Infof("%s %s", params.uid, params.namespace)
	oldinfo, err := namespaceMod.GetNamespaceInfo(xl, params.uid, params.namespace)
	if err != nil || len(oldinfo) == 0 {
		xl.Errorf("Can't find namespace")
		c.JSON(400, gin.H{
			"error": "Can't find namespace info",
		})
		return
	}
	err = updateNamespace(xl, params.uid, params.namespace, namespaceData.Namespace)
	if err != nil {
		xl.Errorf("update namespace failed, err = %#v", err)
		c.JSON(400, gin.H{
			"error": "update namespace failed",
		})
		return
	}
	err = updateBucket(xl, params.uid, params.namespace, oldinfo[0].Bucket, namespaceData.Bucket, domain)
	if err != nil {
		xl.Errorf("update bucket failed, err = %#v", err)
		c.JSON(400, gin.H{
			"error": "update bucket failed",
		})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

// sample requset url = /v1/namespaces?regex=<Regex>&limit=<Limit>&marker=<Marker>&exact=<Exact>
func GetNamespaceInfo(c *gin.Context) {
	xl := xlog.New(c.Writer, c.Request)
	params, err := ParseRequest(c, xl)
	if err != nil {
		xl.Errorf("parse request falied error = %#v", err.Error())
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	nextMark := ""
	var r []models.NamespaceInfo
	xl.Infof("limit %d, marker %s, regex %s uid %s", params.limit, params.marker, params.regex, params.uid)
	if params.exact {
		r, err = namespaceMod.GetNamespaceInfo(xl, params.uid, params.regex)
	} else {
		r, nextMark, err = namespaceMod.GetNamespaceInfos(xl, params.limit, params.marker, params.uid, models.NAMESPACE_ITEM_ID, params.regex)
	}
	if err != nil {
		xl.Errorf("Update falied error = %#v", err.Error())
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
