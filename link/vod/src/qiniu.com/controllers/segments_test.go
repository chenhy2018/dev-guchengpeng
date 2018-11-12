package controllers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/qiniu/api.v7/auth/qbox"

	"github.com/bouk/monkey"
	xlog "github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/suite"
	"qiniu.com/models"
)

type SegmentsTestSuite struct {
	suite.Suite
	r http.Handler
}

func (suite *SegmentsTestSuite) SetupTest() {
	suite.r = GetRouter()

}

type segInfos struct {
	Segs   []segInfo `json:"segments"`
	Marker string    `json:"marker"`
}

func (suite *SegmentsTestSuite) TestFromEndInTwoSeg1() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1532499325&to=1532499345", nil)
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetFragmentTsInfo", func(ss *models.SegmentKodoModel, xl *xlog.Logger, count int, starttime, endtime int64, bucketurl, uaid, mark string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			return nil, "", nil
		})
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "link",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			info := []models.NamespaceInfo{}
			item := models.NamespaceInfo{
				Space:        "ipcamera",
				Regtime:      111111,
				UpdateTime:   3333333,
				Bucket:       "ipcamera",
				Uid:          "link",
				AutoCreateUa: true,
			}
			info = append(info, item)
			return info, nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 1)
			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499327000), models.SEGMENT_ITEM_END_TIME: int64(1532499331000), models.SEGMENT_ITEM_FRAGMENT_START_TIME: int64(1532499327000)}
			rets = append(rets, item)
			return rets, "", nil
		})
	w := PerformRequest(suite.r, req)
	body, _ := ioutil.ReadAll(w.Body)
	var segInfos segInfos
	json.Unmarshal(body, &segInfos)
	suite.Equal(200, w.Code, "200 ok for request")
	suite.Equal(0, len(segInfos.Segs), "empty frames")
}

//seg           A1----------A2  B1------B2  C1--------C2
//url        D1-------------------------------------------D2
//result        A1----------A2  B1------B2  C1--------C2
func (suite *SegmentsTestSuite) TestFromEndInTwoSeg2() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1532499325&to=1532499345", nil)
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetFragmentTsInfo", func(ss *models.SegmentKodoModel, xl *xlog.Logger, count int, starttime, endtime int64, bucketurl, uaid, mark string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 3)
			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499327000), models.SEGMENT_ITEM_END_TIME: int64(1532499331000)}
			rets = append(rets, item)

			item = map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499332000), models.SEGMENT_ITEM_END_TIME: int64(1532499337000)}
			rets = append(rets, item)

			item = map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499339000), models.SEGMENT_ITEM_END_TIME: int64(1532499342000)}
			rets = append(rets, item)
			return rets, "", nil
		})
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "link",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})
	monkey.Patch(GetBucket, func(xl *xlog.Logger, uid, namespace string) (string, error) {
		return "ipcamera", nil
	})

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			info := []models.NamespaceInfo{}
			item := models.NamespaceInfo{
				Space:        "ipcamera",
				Regtime:      111111,
				UpdateTime:   3333333,
				Bucket:       "ipcamera",
				Uid:          "link",
				AutoCreateUa: true,
			}
			info = append(info, item)
			return info, nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 1)
			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499327000), models.SEGMENT_ITEM_END_TIME: int64(1532499331000), models.SEGMENT_ITEM_FRAGMENT_START_TIME: int64(1532499327000)}
			rets = append(rets, item)
			return rets, "", nil
		})

	w := PerformRequest(suite.r, req)

	body, _ := ioutil.ReadAll(w.Body)
	var segInfos segInfos
	err := json.Unmarshal(body, &segInfos)
	suite.Equal(nil, err)
	suite.Equal(200, w.Code, "200 ok for request")
	suite.Equal(3, len(segInfos.Segs), "empty frames")
	suite.Equal(200, w.Code, "401 for bad token")
}

//seg           A1----------A2  B1------B2  C1--------C2
//url               D1---------------------------D2
//result            D1------A2  B1------B2  C1---D2
func (suite *SegmentsTestSuite) TestFromEndInTwoSeg3() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1532499325&to=1532499341", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, nil
	})
	// hack models.SegMentKodomodel.GetFragmentTsInfo
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetFragmentTsInfo", func(ss *models.SegmentKodoModel, xl *xlog.Logger, count int, starttime, endtime int64, bucketurl, uaid, mark string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 3)
			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499322000), models.SEGMENT_ITEM_END_TIME: int64(1532499331000)}
			rets = append(rets, item)

			item = map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499332000), models.SEGMENT_ITEM_END_TIME: int64(1532499337000)}
			rets = append(rets, item)

			item = map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499339000), models.SEGMENT_ITEM_END_TIME: int64(1532499342000)}
			rets = append(rets, item)
			return rets, "", nil
		})
	monkey.Patch(GetBucket, func(xl *xlog.Logger, uid, namespace string) (string, error) {
		return "ipcamera", nil
	})

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			info := []models.NamespaceInfo{}
			item := models.NamespaceInfo{
				Space:        "ipcamera",
				Regtime:      111111,
				UpdateTime:   3333333,
				Bucket:       "ipcamera",
				Uid:          "link",
				AutoCreateUa: true,
			}
			info = append(info, item)
			return info, nil
		})

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "link",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 1)
			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499327000), models.SEGMENT_ITEM_END_TIME: int64(1532499331000), models.SEGMENT_ITEM_FRAGMENT_START_TIME: int64(1532499327000)}
			rets = append(rets, item)
			return rets, "", nil
		})

	w := PerformRequest(suite.r, req)

	body, _ := ioutil.ReadAll(w.Body)
	var segInfos segInfos
	json.Unmarshal(body, &segInfos)
	suite.Equal(200, w.Code, "200 ok for request")
	suite.Equal(3, len(segInfos.Segs), "empty frames")

	suite.Equal(int64(1532499325), segInfos.Segs[0].StartTime, "should be 1532499325")
	suite.Equal(int64(1532499331), segInfos.Segs[0].EndTime, "should be 1532499331")

	suite.Equal(int64(1532499332), segInfos.Segs[1].StartTime, "should be 1532499332")
	suite.Equal(int64(1532499337), segInfos.Segs[1].EndTime, "should be 1532499337")

	suite.Equal(int64(1532499339), segInfos.Segs[2].StartTime, "should be 1532499339")
	suite.Equal(int64(1532499341), segInfos.Segs[2].EndTime, "should be 1532499341")
}

//seg           A1----------A2  B1------B2
//url               D1----------------D2
//result            D1------A2  B1----D2
func (suite *SegmentsTestSuite) TestFromEndInTwoSeg4() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1532499329&to=1532499335", nil)
	defer monkey.UnpatchAll()
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, nil
	})
	// hack models.SegMentKodomodel.GetFragmentTsInfo
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetFragmentTsInfo", func(ss *models.SegmentKodoModel, xl *xlog.Logger, count int, starttime, endtime int64, bucketurl, uaid, mark string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 3)
			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499327000), models.SEGMENT_ITEM_END_TIME: int64(1532499331000)}
			rets = append(rets, item)

			item = map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499332000), models.SEGMENT_ITEM_END_TIME: int64(1532499337000)}
			rets = append(rets, item)

			return rets, "", nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "link",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 1)
			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499327000), models.SEGMENT_ITEM_END_TIME: int64(1532499331000), models.SEGMENT_ITEM_FRAGMENT_START_TIME: int64(1532499327000)}
			rets = append(rets, item)
			return rets, "", nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			info := []models.NamespaceInfo{}
			item := models.NamespaceInfo{
				Space:        "ipcamera",
				Regtime:      111111,
				UpdateTime:   3333333,
				Bucket:       "ipcamera",
				Uid:          "link",
				AutoCreateUa: true,
			}
			info = append(info, item)
			return info, nil
		})
	monkey.Patch(GetBucket, func(xl *xlog.Logger, uid, namespace string) (string, error) {
		return "ipcamera", nil
	})

	w := PerformRequest(suite.r, req)

	body, _ := ioutil.ReadAll(w.Body)
	var segInfos segInfos
	json.Unmarshal(body, &segInfos)
	suite.Equal(200, w.Code, "200 ok for request")
	suite.Equal(2, len(segInfos.Segs), "empty frames")

	suite.Equal(int64(1532499329), segInfos.Segs[0].StartTime, "should be 1532499329")
	suite.Equal(int64(1532499331), segInfos.Segs[0].EndTime, "should be 1532499331")

	suite.Equal(int64(1532499332), segInfos.Segs[1].StartTime, "should be 1532499332")
	suite.Equal(int64(1532499335), segInfos.Segs[1].EndTime, "should be 1532499335")
}

//seg           B1------B2  C1--------C2
//url              D1-----------------------D2
//result           D1---B2  C1--------C2
func (suite *SegmentsTestSuite) TestFromEndInTwoSeg5() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1532499335&to=1532499345", nil)
	defer monkey.UnpatchAll()

	// hack models.SegMentKodomodel.GetFragmentTsInfo
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetFragmentTsInfo", func(ss *models.SegmentKodoModel, xl *xlog.Logger, count int, starttime, endtime int64, bucketurl, uaid, mark string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 3)

			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499332000), models.SEGMENT_ITEM_END_TIME: int64(1532499337000)}
			rets = append(rets, item)

			item = map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499339000), models.SEGMENT_ITEM_END_TIME: int64(1532499342000)}
			rets = append(rets, item)
			return rets, "", nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "link",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 1)
			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499327000), models.SEGMENT_ITEM_END_TIME: int64(1532499331000), models.SEGMENT_ITEM_FRAGMENT_START_TIME: int64(1532499327000)}
			rets = append(rets, item)
			return rets, "", nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			info := []models.NamespaceInfo{}
			item := models.NamespaceInfo{
				Space:        "ipcamera",
				Regtime:      111111,
				UpdateTime:   3333333,
				Bucket:       "ipcamera",
				Uid:          "link",
				AutoCreateUa: true,
			}
			info = append(info, item)
			return info, nil
		})
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, nil
	})
	monkey.Patch(GetBucket, func(xl *xlog.Logger, uid, namespace string) (string, error) {
		return "ipcamera", nil
	})

	w := PerformRequest(suite.r, req)

	body, _ := ioutil.ReadAll(w.Body)
	var segInfos segInfos
	json.Unmarshal(body, &segInfos)
	suite.Equal(200, w.Code, "200 ok for request")
	suite.Equal(2, len(segInfos.Segs))
	suite.Equal(int64(1532499335), segInfos.Segs[0].StartTime, "should be 1532499335")
	suite.Equal(int64(1532499337), segInfos.Segs[0].EndTime, "should be 1532499337")

	suite.Equal(int64(1532499339), segInfos.Segs[1].StartTime, "should be 1532499339")
	suite.Equal(int64(1532499342), segInfos.Segs[1].EndTime, "should be 1532499342")
}

//seg           A1----------A2  B1------B2
//url        D1-----------------------D2
//result        A1----------A2  B1----D2
func (suite *SegmentsTestSuite) TestFromEndInTwoSeg6() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1532499325&to=1532499335", nil)
	defer monkey.UnpatchAll()

	// hack models.SegMentKodomodel.GetFragmentTsInfo
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetFragmentTsInfo", func(ss *models.SegmentKodoModel, xl *xlog.Logger, count int, starttime, endtime int64, bucketurl, uaid, mark string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 3)
			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499327000), models.SEGMENT_ITEM_END_TIME: int64(1532499331000)}
			rets = append(rets, item)

			item = map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499332000), models.SEGMENT_ITEM_END_TIME: int64(1532499337000)}
			rets = append(rets, item)
			return rets, "", nil
		})
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "link",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 1)
			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499327000), models.SEGMENT_ITEM_END_TIME: int64(1532499331000), models.SEGMENT_ITEM_FRAGMENT_START_TIME: int64(1532499327000)}
			rets = append(rets, item)
			return rets, "", nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			info := []models.NamespaceInfo{}
			item := models.NamespaceInfo{
				Space:        "ipcamera",
				Regtime:      111111,
				UpdateTime:   3333333,
				Bucket:       "ipcamera",
				Uid:          "link",
				AutoCreateUa: true,
			}
			info = append(info, item)
			return info, nil
		})
	monkey.Patch(GetBucket, func(xl *xlog.Logger, uid, namespace string) (string, error) {
		return "ipcamera", nil
	})

	w := PerformRequest(suite.r, req)

	body, _ := ioutil.ReadAll(w.Body)
	var segInfos segInfos
	json.Unmarshal(body, &segInfos)
	suite.Equal(200, w.Code, "200 ok for request")
	suite.Equal(2, len(segInfos.Segs), "empty frames")

	suite.Equal(int64(1532499327), segInfos.Segs[0].StartTime, "should be 1532499327")
	suite.Equal(int64(1532499331), segInfos.Segs[0].EndTime, "should be 1532499331")

	suite.Equal(int64(1532499332), segInfos.Segs[1].StartTime, "should be 1532499332")
	suite.Equal(int64(1532499335), segInfos.Segs[1].EndTime, "should be 1532499335")
}

//seg           A1----------A2
//url              D1-----D2
//result           D1-----D2
func (suite *SegmentsTestSuite) TestFromEndInTwoSeg7() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1532499326&to=1532499330", nil)
	defer monkey.UnpatchAll()

	// hack models.SegMentKodomodel.GetFragmentTsInfo
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetFragmentTsInfo", func(ss *models.SegmentKodoModel, xl *xlog.Logger, count int, starttime, endtime int64, bucketurl, uaid, mark string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 3)
			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499325000), models.SEGMENT_ITEM_END_TIME: int64(1532499331000)}
			rets = append(rets, item)
			return rets, "", nil
		})
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "link",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 1)
			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499327000), models.SEGMENT_ITEM_END_TIME: int64(1532499331000), models.SEGMENT_ITEM_FRAGMENT_START_TIME: int64(1532499327000)}
			rets = append(rets, item)
			return rets, "", nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			info := []models.NamespaceInfo{}
			item := models.NamespaceInfo{
				Space:        "ipcamera",
				Regtime:      111111,
				UpdateTime:   3333333,
				Bucket:       "ipcamera",
				Uid:          "link",
				AutoCreateUa: true,
			}
			info = append(info, item)
			return info, nil
		})
	monkey.Patch(GetBucket, func(xl *xlog.Logger, uid, namespace string) (string, error) {
		return "ipcamera", nil
	})

	w := PerformRequest(suite.r, req)

	body, _ := ioutil.ReadAll(w.Body)
	var segInfos segInfos
	json.Unmarshal(body, &segInfos)
	suite.Equal(200, w.Code, "200 ok for request")
	suite.Equal(1, len(segInfos.Segs), "empty frames")

	suite.Equal(int64(1532499326), segInfos.Segs[0].StartTime, "should be 1532499332")
	suite.Equal(int64(1532499330), segInfos.Segs[0].EndTime, "should be 1532499335")
}

//seg           A1----------A2  B1------B2  C1--------C2
//url        D1---------------------------------D2
//result        A1----------A2  B1------B2  C1--D2
func (suite *SegmentsTestSuite) TestFromEndInTwoSeg8() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1532499325&to=1532499340", nil)
	defer monkey.UnpatchAll()

	// hack models.SegMentKodomodel.GetFragmentTsInfo
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetFragmentTsInfo", func(ss *models.SegmentKodoModel, xl *xlog.Logger, count int, starttime, endtime int64, bucketurl, uaid, mark string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 3)
			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499327000), models.SEGMENT_ITEM_END_TIME: int64(1532499331000)}
			rets = append(rets, item)

			item = map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499332000), models.SEGMENT_ITEM_END_TIME: int64(1532499337000)}
			rets = append(rets, item)

			item = map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499339000), models.SEGMENT_ITEM_END_TIME: int64(1532499342000)}
			rets = append(rets, item)
			return rets, "", nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetSegmentTsInfo", func(ss *models.SegmentKodoModel,
			xl *xlog.Logger, starttime, endtime int64, bucketurl, uaid string, limit int, marker string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			rets := make([](map[string]interface{}), 0, 1)
			item := map[string]interface{}{models.SEGMENT_ITEM_START_TIME: int64(1532499327000), models.SEGMENT_ITEM_END_TIME: int64(1532499331000), models.SEGMENT_ITEM_FRAGMENT_START_TIME: int64(1532499327000)}
			rets = append(rets, item)
			return rets, "", nil
		})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.NamespaceModel)(nil)), "GetNamespaceInfo", func(ss *models.NamespaceModel, xl *xlog.Logger, uid, namespace string) ([]models.NamespaceInfo, error) {
			info := []models.NamespaceInfo{}
			item := models.NamespaceInfo{
				Space:        "ipcamera",
				Regtime:      111111,
				UpdateTime:   3333333,
				Bucket:       "ipcamera",
				Uid:          "link",
				AutoCreateUa: true,
			}
			info = append(info, item)
			return info, nil
		})
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, nil
	})

	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "link",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})
	monkey.Patch(GetBucket, func(xl *xlog.Logger, uid, namespace string) (string, error) {
		return "ipcamera", nil
	})

	w := PerformRequest(suite.r, req)
	body, _ := ioutil.ReadAll(w.Body)
	var segInfos segInfos
	json.Unmarshal(body, &segInfos)
	suite.Equal(200, w.Code, "200 ok for request")
	suite.Equal(3, len(segInfos.Segs), "empty frames")

	suite.Equal(int64(1532499327), segInfos.Segs[0].StartTime, "should be 1532499327")
	suite.Equal(int64(1532499331), segInfos.Segs[0].EndTime, "should be 1532499331")

	suite.Equal(int64(1532499332), segInfos.Segs[1].StartTime, "should be 1532499332")
	suite.Equal(int64(1532499337), segInfos.Segs[1].EndTime, "should be 1532499337")

	suite.Equal(int64(1532499339), segInfos.Segs[2].StartTime, "should be 1532499339")
	suite.Equal(int64(1532499340), segInfos.Segs[2].EndTime, "should be 1532499340")
}

func (suite *SegmentsTestSuite) TestFromEndInTwoSeg9() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1532499fdsfs325&to=1532499345", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "400 for request params error")

}
func (suite *SegmentsTestSuite) TestFromEndInTwoSeg10() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1532499355&to=1532499345", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "400 for request from great than to")

}
func (suite *SegmentsTestSuite) TestFromEndInTwoSeg11() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1530499355&to=1532499345", nil)
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "400 for request from great than to")

}

func (suite *SegmentsTestSuite) TestFromEndInTwoSeg12() {
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1537872668&to=1537876443", nil)
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userInfo{}, errors.New("get user info failed")
	})
	w := PerformRequest(suite.r, req)
	suite.Equal(500, w.Code, "500 for get user info  failed")

}
func (suite *SegmentsTestSuite) TestFromEndInTwoSeg13() {
	defer monkey.UnpatchAll()
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1537872668&to=1537876443", nil)
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userinfo, nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "link",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})
	w := PerformRequest(suite.r, req)
	suite.Equal(400, w.Code, "400 for can't find namesapce")
}

func (suite *SegmentsTestSuite) TestFromEndInTwoSeg14() {
	defer monkey.UnpatchAll()
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1537872668&to=1537876443", nil)
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userinfo, nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "link",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})
	monkey.Patch(GetBucket, func(xl *xlog.Logger, uid, namespace string) (string, error) {
		return "ipcamera", nil
	})
	w := PerformRequest(suite.r, req)
	suite.Equal(200, w.Code, "")

}
func (suite *SegmentsTestSuite) TestFromEndInTwoSeg15() {
	defer monkey.UnpatchAll()
	req, _ := http.NewRequest("GET", "/v1/namespaces/ipcamera/uas/testdeviceid8/segments?from=1537872668&to=1537876443", nil)
	monkey.Patch(getUserInfo, func(xl *xlog.Logger, req *http.Request) (*userInfo, error) {
		return &userinfo, nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.UaModel)(nil)), "GetUaInfo", func(ss *models.UaModel, xl *xlog.Logger, uid, namespace, uaid string) ([]models.UaInfo, error) {
			info := []models.UaInfo{}
			item := models.UaInfo{
				Uid:       "link",
				Namespace: "ipcamera",
			}
			info = append(info, item)
			return info, nil
		})
	monkey.Patch(GetBucket, func(xl *xlog.Logger, uid, namespace string) (string, error) {
		return "ipcamera", nil
	})
	monkey.PatchInstanceMethod(
		reflect.TypeOf((*models.SegmentKodoModel)(nil)), "GetFragmentTsInfo", func(ss *models.SegmentKodoModel, xl *xlog.Logger, count int, starttime, endtime int64, bucketurl, uaid, mark string, mac *qbox.Mac) ([]map[string]interface{}, string, error) {
			return nil, "", errors.New("can't  find segments")
		})
	w := PerformRequest(suite.r, req)
	suite.Equal(500, w.Code, "")

}
func TestSegmentsTestSuite(t *testing.T) {
	suite.Run(t, new(SegmentsTestSuite))
}
