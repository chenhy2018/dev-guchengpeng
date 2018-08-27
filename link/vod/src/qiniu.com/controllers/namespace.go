package controllers
  
import (
        //"time"
        //"errors"
        //"strconv"
        "io/ioutil"
        "github.com/gin-gonic/gin"
        xlog "github.com/qiniu/xlog.v1"
        "qiniu.com/models"
        "encoding/json"
)

var (
        NamespaceMod *models.NamespaceModel
)

func init() {
        NamespaceMod = &models.NamespaceModel{}
        NamespaceMod.Init()
}

type namespacebody struct {
        Uid        string `json:"uid"`
        Bucket     string `json:"bucket"`
        Namespace  string `json:"namespace"`
        CreatedAt  int64  `json:"createdAt"`
        UpdatedAt  int64  `json:"updatedAt"`
}

// sample requset url = /v1/uids/<Uid>/namespaces
func RegisterNamespace(c *gin.Context) {
        xl := xlog.New(c.Writer, c.Request)
        params, err := ParseRequest(c, xl)
        if err != nil {
                xl.Errorf("parse request falied error = %#v", err.Error())
                c.JSON(500, gin.H{
                        "error": err.Error(),
                })
                return
        }

        body, err := ioutil.ReadAll(c.Request.Body)
        xl.Infof("%s", body)
        var namespaceData namespacebody
        err = json.Unmarshal(body, &namespaceData)
        xl.Infof("%#v", namespaceData)

        if err != nil {
                xl.Errorf("parse request body failed, body = %#v", body)
                c.JSON(500, gin.H{
                        "error": "read callback body failed",
                })
                return
        }
        xl.Infof("%s %s", namespaceData.Namespace, params.namespace)

        namespace := models.NamespaceInfo{
                Uid  : params.uid,
                Space : namespaceData.Namespace,
                Bucket  : namespaceData.Bucket,
        }
        err = NamespaceMod.Register(xl, namespace)
        if err != nil {
                xl.Errorf("Register falied error = %#v", err.Error())
                c.JSON(500, gin.H{
                        "error": err.Error(),
                })
                return
        } else {
               c.JSON(200, gin.H{ "success": true })
        }
}

// sample requset url = /v1/uids/<Uid>/namespaces/<Encodednamespace>
func DeleteNamespace(c *gin.Context) {
        xl := xlog.New(c.Writer, c.Request)
        params, err := ParseRequest(c, xl)
        if err != nil {
                xl.Errorf("parse request falied error = %#v", err.Error())
                c.JSON(500, gin.H{
                        "error": err.Error(),
                })
                return
        }
        err = NamespaceMod.Delete(xl, params.uid, params.namespace)
        if err != nil {
                xl.Errorf("Register falied error = %#v", err.Error())
                c.JSON(500, gin.H{
                        "error": err.Error(),
                })
                return
        } else {
               c.JSON(200, gin.H{ "success": true })
        }
}

// sample requset url = /v1/uids/<Uid>/namespaces/<Encodednamespace>
func UpdateNamespace(c *gin.Context) {
        xl := xlog.New(c.Writer, c.Request)
        params, err := ParseRequest(c, xl)
        if err != nil {
                xl.Errorf("parse request falied error = %#v", err.Error())
                c.JSON(500, gin.H{
                        "error": err.Error(),
                })
                return
        }

        body, err := ioutil.ReadAll(c.Request.Body)
        xl.Infof("%s", body)
        var namespaceData namespacebody
        err = json.Unmarshal(body, &namespaceData)
        if err != nil || namespaceData.Uid=="" || namespaceData.Namespace=="" {
                xl.Errorf("parse request body failed, body = %#v", body)
                c.JSON(500, gin.H{
                        "error": "read callback body failed",
                })
                return
        }

        xl.Infof("%#v", namespaceData)
        namespace := models.NamespaceInfo{
                Uid  : namespaceData.Uid,
                Space : namespaceData.Namespace,
                Bucket  : namespaceData.Bucket,
        }
        err = NamespaceMod.UpdateNamespace(xl, params.uid, params.namespace, namespace)
        if err != nil {
                xl.Errorf("Update falied error = %#v", err.Error())
                c.JSON(500, gin.H{
                        "error": err.Error(),
                })
                return
        } else {
               c.JSON(200, gin.H{ "success": true })
        }
}

// sample requset url = /v1/uids/<Uid>/namespaces&regular=<Regular>&limit=<Limit>&marker=<Marker>
func GetNamespaceInfos(c *gin.Context) {
        xl := xlog.New(c.Writer, c.Request)
        params, err := ParseRequest(c, xl)
        if err != nil {
                xl.Errorf("parse request falied error = %#v", err.Error())
                c.JSON(500, gin.H{
                        "error": err.Error(),
                })
                return
        }
        xl.Infof("limit %d, marker %s, regular %s uid %s", params.limit, params.marker, params.regular, params.uid)
        r, nextMark, err1 := NamespaceMod.GetNamespaceInfos(xl, params.limit, params.marker,params.uid, models.NAMESPACE_ITEM_ID, params.regular)
        if err1 != nil {
                xl.Errorf("Update falied error = %#v", err1.Error())
                c.JSON(500, gin.H{
                        "error": err.Error(),
                })
                return
        } else {
                c.Header("Content-Type", "application/json")
                c.Header("Access-Control-Allow-Origin", "*")
                c.JSON(200, gin.H{ "item" : r,
                                   "marker" : nextMark, })
        }
}

// sample requset url = /v1/uids/<Uid>/namespaces/<EncodedNamespace>
func GetNamespaceInfo(c *gin.Context) {
        xl := xlog.New(c.Writer, c.Request)
        params, err := ParseRequest(c, xl)
        if err != nil {
                xl.Errorf("parse request falied error = %#v", err.Error())
                c.JSON(500, gin.H{
                        "error": err.Error(),
                })
                return
        }
        r, err1 := NamespaceMod.GetNamespaceInfo(xl, params.uid, params.namespace)
        if err1 != nil {
                xl.Errorf("Update falied error = %#v", err1.Error())
                c.JSON(500, gin.H{
                        "error": err1.Error(),
                })
                return
        } else {
                str, err2 := json.Marshal(r)
                if err2 != nil {
                        xl.Errorf("Update falied error = %#v", err2.Error())
                        c.JSON(500, gin.H{
                                "error": err2.Error(),
                        })
                }
                c.Header("Content-Type", "application/json")
                c.Header("Access-Control-Allow-Origin", "*")
                c.String(200, string(str))
        }
}
