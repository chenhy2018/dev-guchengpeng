package controllers
  
import (
        //"time"
        "errors"
        "strconv"
        "io/ioutil"
        "github.com/gin-gonic/gin"
        xlog "github.com/qiniu/xlog.v1"
        "qiniu.com/models"
        "encoding/json"
)

var (
        UaMod *models.UaModel
)

func init() {
        UaMod = &models.UaModel{}
        UaMod.Init()
}

type uabody struct {
        Uid        string `json:"uid"`
        Uaid       string `json:"uaid"`
        Namespace  string `json:"namespace"`
        CreatedAt  int64  `json:"createdAt"`
        UpdatedAt  int64  `json:"updatedAt"`
        Password   string `json:"password"`
}

type params struct {
        uid       string
        namespace string
        uaid      string
        token     string
        expire    int64
        limit     int
        marker    string
}

func parseRequest(c *gin.Context, xl *xlog.Logger) (*params, error) {
        uaid := c.Param("uaid")
        namespace := c.Param("namespace")
        limit := c.DefaultQuery("limit", "1000")
        marker := c.DefaultQuery("marker", "")

        limitT, err := strconv.ParseInt(limit, 10, 32)
        if err != nil {
                return nil, errors.New("Parse limit time failed")
        }
        if limitT > 1000 {
                limitT = 1000
        }

        param := &params{
                uaid        : uaid,
                namespace : namespace,
                limit     : int(limitT),
                marker    : marker,
        }
        return param, nil
}

// sample requset url = /v1/namespaces/<Namespace>/uas
func RegisterUa(c *gin.Context) {
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
        var uaData uabody
        err = json.Unmarshal(body, &uaData)
        xl.Infof("%#v", uaData)

        if err != nil {
                xl.Errorf("parse request body failed, body = %#v", body)
                c.JSON(500, gin.H{
                        "error": "read callback body failed",
                })
                return
        }
        xl.Infof("%s %s", uaData.Uaid, params.namespace)
        
        ua := models.UaInfo{
                Uid  : params.uid,
                UaId : uaData.Uaid,
                Namespace : params.namespace,
                Password  : uaData.Password,
        }
        err = UaMod.Register(xl, ua)
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

// sample requset url = /v1/namespaces/<Namespace>/uas/<Encodedua>
func DeleteUa(c *gin.Context) {
        xl := xlog.New(c.Writer, c.Request)
        params, err := ParseRequest(c, xl)
        if err != nil {
                xl.Errorf("parse request falied error = %#v", err.Error())
                c.JSON(500, gin.H{
                        "error": err.Error(),
                })
                return
        }
        err = UaMod.Delete(xl, params.namespace, params.uaid)
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

// sample requset url = /v1/namespaces/<Namespace>/uas/<Encodedua>
func UpdateUa(c *gin.Context) {
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
        var uaData uabody
        err = json.Unmarshal(body, &uaData)
        if err != nil || uaData.Uid=="" || uaData.Uaid=="" || uaData.Namespace=="" {
                xl.Errorf("parse request body failed, body = %#v", body)
                c.JSON(500, gin.H{
                        "error": "read callback body failed",
                })
                return
        }

        xl.Infof("%#v", uaData)
        ua := models.UaInfo{
                Uid  : uaData.Uid,
                UaId : uaData.Uaid,
                Namespace : uaData.Namespace,
                Password  : uaData.Password,
        }
        err = UaMod.UpdateUa(xl, params.namespace, params.uaid, ua)
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

// sample requset url = /v1/namespaces/<Namespace>/uas&regular=<Regular>&limit=<Limit>&marker=<Marker>
func GetUaInfos(c *gin.Context) {
        xl := xlog.New(c.Writer, c.Request)
        params, err := ParseRequest(c, xl)
        if err != nil {
                xl.Errorf("parse request falied error = %#v", err.Error())
                c.JSON(500, gin.H{
                        "error": err.Error(),
                })
                return
        }
        xl.Infof("limit %d, marker %s, regular %s namespace %s", params.limit, params.marker, params.regular, params.namespace)
        r, nextMark, err1 := UaMod.GetUaInfos(xl, params.limit, params.marker,params.namespace, models.UA_ITEM_UAID, params.regular)
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

// sample requset url = /v1/namespaces/<Namespace>/uas/<EncodedUa>
func GetUaInfo(c *gin.Context) {
        xl := xlog.New(c.Writer, c.Request)
        params, err := ParseRequest(c, xl)
        if err != nil {
                xl.Errorf("parse request falied error = %#v", err.Error())
                c.JSON(500, gin.H{
                        "error": err.Error(),
                })
                return
        }
        r, err1 := UaMod.GetUaInfo(xl, params.namespace, params.uaid)
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
