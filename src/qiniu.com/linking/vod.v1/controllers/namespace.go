package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	xlog "github.com/qiniu/xlog.v1"
	"qiniu.com/linking/vod.v1/models"
)

const (
	DEFAULT_EXPIRE = 7
)

type namespacebody struct {
	Bucket       string `json:"bucket"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
	AutoCreateUa bool   `json:"auto"`
	Expire       int    `json:"expire"`
	Domain       string `json:"domain"`
	Category     string `json:"category"`
	Remark       string `json:"remark"`
}

func checkBucketInKodo(bucket string, user *userInfo) error {
	service, err := newRsService(user, bucket)
	if err != nil {
		return err
	}
	_, code, err := service.Bucket(bucket)
	if code != 200 {
		return errors.New("get Bucket Info falied")
	}
	return err
}
func checkbucket(xl *xlog.Logger, bucket string, user *userInfo) error {

	if err := checkBucketInKodo(bucket, user); err != nil {
		return fmt.Errorf("check bucket availability failed=%#v", err.Error())
	}
	return nil
}

func checkdomain(xl *xlog.Logger, domain, bucket string, user *userInfo) error {
	domains, err := getDomain(xl, bucket, user)
	if err != nil || domain == "" {
		xl.Errorf("bucket is not correct")
		return fmt.Errorf("bucket is not correct")
	}
	for _, s := range domains {
		if s == domain {
			return nil
		}
	}
	return fmt.Errorf("domain is not correct")
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
	info, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed%v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	err = checkbucket(xl, namespaceData.Bucket, info)
	if err != nil {
		xl.Errorf("bucket is not correct, err = %#v", err)
		c.JSON(403, gin.H{
			"error": "bucket is not correct",
		})
		return
	}

	err = checkdomain(xl, namespaceData.Domain, namespaceData.Bucket, info)
	if err != nil {
		xl.Errorf("domain is not correct, err = %#v", err)
		c.JSON(403, gin.H{
			"error": "domain is not correct",
		})
		return
	}

	oldinfo, err := namespaceMod.GetNamespaceInfo(xl, info.uid, params.namespace)
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
		Uid:          info.uid,
		Space:        params.namespace,
		Bucket:       namespaceData.Bucket,
		AutoCreateUa: namespaceData.AutoCreateUa,
		Expire:       expire,
		Domain:       namespaceData.Domain,
		Category:     namespaceData.Category,
		Remark:       namespaceData.Remark,
	}

	err = namespaceMod.Register(xl, namespace)
	if err != nil {
		xl.Errorf("Register falied error = %#v", err.Error())
		c.JSON(500, gin.H{
			"error": "Service Internal Error",
		})
		return
	} else {
		fmt.Printf("name %#v %#v", info.uid, params.namespace)
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
	oldinfo, err := namespaceMod.GetNamespaceInfo(xl, info.uid, params.namespace)
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
	err = namespaceMod.Delete(xl, info.uid, params.namespace)
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

func updateDomain(xl *xlog.Logger, uid, space, bucket, domain, newDomain string, info *userInfo) error {
	if domain != newDomain && newDomain != "" {
		err := checkdomain(xl, newDomain, bucket, info)
		if err != nil {
			xl.Errorf("err = %#v", err)
			return err
		}
		cond := map[string]interface{}{
			models.NAMESPACE_ITEM_DOMAIN: newDomain,
		}
		err = namespaceMod.UpdateFunction(xl, uid, space, models.NAMESPACE_ITEM_DOMAIN, cond)
		if err != nil {
			xl.Errorf("Update falied error = %#v", err.Error())
			return err
		}
	}
	return nil
}

func updateBucket(xl *xlog.Logger, uid, space, bucket, newBucket string, info *userInfo) error {
	if bucket != newBucket && newBucket != "" {
		err := checkbucket(xl, newBucket, info)
		if err != nil {
			xl.Errorf("bucket is not correct, err = %#v", err)
			return err
		}
		domain, err := getDomain(xl, newBucket, info)
		if err != nil || len(domain) == 0 {
			xl.Errorf("domain is not correct")
			return fmt.Errorf("domain is not correct")
		}
		cond := map[string]interface{}{
			models.NAMESPACE_ITEM_BUCKET: newBucket,
		}
		err = namespaceMod.UpdateFunction(xl, uid, space, models.NAMESPACE_ITEM_BUCKET, cond)
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
		cond := map[string]interface{}{
			models.NAMESPACE_ITEM_AUTO_CREATE_UA: newauto,
		}
		err := namespaceMod.UpdateFunction(xl, uid, space, models.NAMESPACE_ITEM_AUTO_CREATE_UA, cond)
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
		cond := map[string]interface{}{
			models.NAMESPACE_ITEM_EXPIRE: newExpire,
		}
		err := namespaceMod.UpdateFunction(xl, uid, space, models.NAMESPACE_ITEM_EXPIRE, cond)
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

	user, err := getUserInfo(xl, c.Request)
	if err != nil {
		xl.Errorf("get user Info failed %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	xl.Errorf("get namespace info ,  %#v %#v", user.uid, params.namespace)
	namespaceInfo, err := namespaceMod.GetNamespaceInfo(xl, user.uid, params.namespace)
	if err != nil {
		xl.Errorf("get namespace info error, error =  %#v", err)
		c.JSON(400, gin.H{"error =  %#v": "err"})
		return
	}
	if len(namespaceInfo) == 0 {
		xl.Errorf("namespace is not correct")
		c.JSON(400, gin.H{
			"error": "namespace is not correct",
		})
		return
	}

	namespaceData := namespacebody{
		Bucket:       namespaceInfo[0].Bucket,
		AutoCreateUa: namespaceInfo[0].AutoCreateUa,
		Expire:       namespaceInfo[0].Expire,
		Domain:       namespaceInfo[0].Domain,
		Category:     namespaceInfo[0].Category,
		Remark:       namespaceInfo[0].Remark,
	}

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

	err = checkbucket(xl, namespaceData.Bucket, user)
	if err != nil {
		xl.Errorf("bucket is not correct, err = %#v", err)
		c.JSON(403, gin.H{
			"error": "bucket is not correct",
		})
		return
	}

	err = checkdomain(xl, namespaceData.Domain, namespaceData.Bucket, user)
	if err != nil {
		xl.Errorf("domain is not correct, err = %#v", err)
		c.JSON(403, gin.H{
			"error": "domain is not correct",
		})
		return
	}

	namespace := models.NamespaceInfo{
		Uid:          user.uid,
		Space:        params.namespace,
		Bucket:       namespaceData.Bucket,
		AutoCreateUa: namespaceData.AutoCreateUa,
		Expire:       namespaceData.Expire,
		Domain:       namespaceData.Domain,
		Category:     namespaceData.Category,
		Remark:       namespaceData.Remark,
	}
	err = namespaceMod.Update(xl, namespace)
	if err != nil {
		xl.Errorf("Register falied error = %#v", err.Error())
		c.JSON(500, gin.H{
			"error": "Service Internal Error",
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

	xl.Infof("limit %d, marker %s, regex %s uid %s", params.limit, params.marker, params.prefix, info.uid)
	if params.exact {
		r, err = namespaceMod.GetNamespaceInfo(xl, info.uid, params.prefix)
	} else {
		r, nextMark, err = namespaceMod.GetNamespaceInfos(xl, params.limit, params.marker, info.uid, params.prefix)
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
