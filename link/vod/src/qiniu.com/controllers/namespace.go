package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/api.v7/auth/qbox"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/models"
)

const (
	DOMAIN_URL     = "http://api.qiniu.com/v6/domain/list?tbl="
	DEFAULT_EXPIRE = 7
)

type namespacebody struct {
	Bucket       string `json:"bucket"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
	AutoCreateUa bool   `json:"auto"`
	Expire       int    `json:"expire"`
}

func getDomain(xl *xlog.Logger, bucket string, info *userInfo) (string, error) {
	client := http.Client{}
	url := fmt.Sprintf("%s%s", DOMAIN_URL, bucket)
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		xl.Errorf("%#v", err)
	}
	mac := qbox.NewMac(info.ak, info.sk)

	token, _ := mac.SignRequest(request)

	request.Header.Set("Authorization", "QBox "+token)
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var domain []string
	for {
		if err := dec.Decode(&domain); err == io.EOF {
			break

		} else if err != nil {
			return "", err
		}
	}
	if len(domain) == 0 {
		return "", nil
	}
	return domain[0], nil
}

func checkbucket(xl *xlog.Logger, bucket string) error {
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

	var namespaceData namespacebody
	dec := json.NewDecoder(c.Request.Body)
	for {
		if err := dec.Decode(&namespaceData); err == io.EOF {
			break
		} else if err != nil {
			xl.Errorf("json decode failed %#v", err)
			c.JSON(400, gin.H{
				"error": "json decode failed",
			})
			return
		}
	}
	err = checkbucket(xl, namespaceData.Bucket)
	if err != nil {
		xl.Errorf("bucket is already register, err = %#v", err)
		c.JSON(403, gin.H{
			"error": "bucket is already register",
		})
		return
	}

	info, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	oldinfo, err := namespaceMod.GetNamespaceInfo(xl, getUid(info.uid), params.namespace)
	if err != nil {
		xl.Errorf("get Namesapce Info error %#v", err)
		c.JSON(500, gin.H{
			"error": "Service Internal Error",
		})
		return
	}
	if len(oldinfo) != 0 {
		xl.Errorf("namespace is exist")
		c.JSON(400, gin.H{
			"error": "namespace is exist",
		})
		return
	}
	expire := namespaceData.Expire
	if expire <= 0 {
		expire = DEFAULT_EXPIRE
	}
	namespace := models.NamespaceInfo{
		Uid:          getUid(info.uid),
		Space:        params.namespace,
		Bucket:       namespaceData.Bucket,
		AutoCreateUa: namespaceData.AutoCreateUa,
		Expire:       expire,
	}

	err = namespaceMod.Register(xl, namespace)
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
	info, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	oldinfo, err := namespaceMod.GetNamespaceInfo(xl, getUid(info.uid), params.namespace)
	if err != nil {
		xl.Errorf("get Namesapce Info error %#v", err)
		c.JSON(500, gin.H{
			"error": "Service Internal Error",
		})
		return
	}
	if len(oldinfo) == 0 {
		xl.Errorf("namespace doesn't exist")
		c.JSON(400, gin.H{
			"error": "namespace doesn't exist",
		})
		return
	}
	err = namespaceMod.Delete(xl, getUid(info.uid), params.namespace)
	if err != nil {
		xl.Errorf("Delete falied error = %#v", err.Error())
		c.JSON(500, gin.H{
			"error": "Service Internal Error",
		})
		return
	}
	model := models.UaModel{}
	cond := map[string]interface{}{models.UA_ITEM_NAMESPACE: params.namespace}
	model.Delete(xl, cond)
	c.JSON(200, gin.H{"success": true})
}

//need test.
func updateNamespace(xl *xlog.Logger, uid, space, newSpace string) error {
	if space != newSpace && newSpace != "" {
		xl.Errorf("Can't support modify namespace")
		return fmt.Errorf("Can't support modify namespace")
	}
	return nil
}

func updateBucket(xl *xlog.Logger, uid, space, bucket, newBucket string, info *userInfo) error {
	if bucket != newBucket && newBucket != "" {
		err := checkbucket(xl, newBucket)
		if err != nil {
			xl.Errorf("bucket is already register, err = %#v", err)
			return err
		}
		domain, err := getDomain(xl, newBucket, info)
		if err != nil || domain == "" {
			xl.Errorf("bucket is not correct")
			return fmt.Errorf("bucket is not correct")

		}
		err = namespaceMod.UpdateBucket(xl, uid, space, newBucket, domain)
		if err != nil {
			xl.Errorf("Update falied error = %#v", err.Error())
			return err

		}

	}
	return nil
}

func updateAutoCreateUa(xl *xlog.Logger, uid, space string, auto, newauto bool) error {
	if auto != newauto {
		namespaceMod := models.NamespaceModel{}
		err := namespaceMod.UpdateAutoCreateUa(xl, uid, space, newauto)
		if err != nil {
			xl.Errorf("Update falied error = %#v", err.Error())
			return err
		}
	}
	return nil
}

func updateExpire(xl *xlog.Logger, uid, space string, expire, newExpire int) error {
	if expire != newExpire && newExpire != 0 {
		namespaceMod := models.NamespaceModel{}
		err := namespaceMod.UpdateExpire(xl, uid, space, newExpire)
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

	var namespaceData namespacebody
	dec := json.NewDecoder(c.Request.Body)
	for {
		if err := dec.Decode(&namespaceData); err == io.EOF {
			break
		} else if err != nil {
			xl.Errorf("json decode failed")
			c.JSON(400, gin.H{
				"error": "json decode failed",
			})
			return
		}
	}

	info, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return

	}

	/*
	   namespace := models.NamespaceInfo{
	           Uid  : info.Uid,
	           Space : namespaceData.Namespace,
	           Bucket  : namespaceData.Bucket,
	           Domain : domain,
	   }
	*/
	oldinfo, err := namespaceMod.GetNamespaceInfo(xl, getUid(info.uid), params.namespace)
	if err != nil || len(oldinfo) == 0 {
		xl.Errorf("Can't find namespace")
		c.JSON(400, gin.H{
			"error": "Can't find namespace info",
		})
		return
	}
	err = updateBucket(xl, getUid(info.uid), params.namespace, oldinfo[0].Bucket, namespaceData.Bucket, info)
	if err != nil {
		xl.Errorf("update bucket failed, err = %#v", err)
		c.JSON(400, gin.H{
			"error": "update bucket failed",
		})
		return
	}
	err = updateAutoCreateUa(xl, getUid(info.uid), params.namespace, oldinfo[0].AutoCreateUa, namespaceData.AutoCreateUa)
	if err != nil {
		xl.Errorf("update auto create ua failed")
		c.JSON(500, gin.H{
			"error": "Service Internal Error",
		})
		return
	}
	err = updateExpire(xl, getUid(info.uid), params.namespace, oldinfo[0].Expire, namespaceData.Expire)
	if err != nil {
		xl.Errorf("update expire failed")
		c.JSON(500, gin.H{
			"error": "update expire failed",
		})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

// sample requset url = /v1/namespaces?prefix=<Prefix>&limit=<Limit>&marker=<Marker>&exact=<Exact>
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

	info, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	xl.Infof("limit %d, marker %s, regex %s uid %s", params.limit, params.marker, params.prefix, getUid(info.uid))
	if params.exact {
		r, err = namespaceMod.GetNamespaceInfo(xl, getUid(info.uid), params.prefix)
	} else {
		r, nextMark, err = namespaceMod.GetNamespaceInfos(xl, params.limit, params.marker, getUid(info.uid), params.prefix)
	}
	if err != nil {
		xl.Errorf("get namesapce failed, error = %#v", err.Error())
		c.JSON(500, gin.H{
			"error": "Service Internal Error",
		})
		return
	}

	c.Header("Content-Type", "application/json")
	c.JSON(200, gin.H{"item": r,
		"marker": nextMark})
}
