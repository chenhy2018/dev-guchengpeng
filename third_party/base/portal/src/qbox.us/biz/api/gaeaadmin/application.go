package gaeaadmin

import (
	"fmt"
	"time"

	"labix.org/v2/mgo/bson"
	"qbox.us/zone"
)

func (s *gaeaAdminService) ApplicationGet(id string) (res Application, err error) {
	var (
		resp struct {
			apiResultBase
			Data Application `json:"data"`
		}
		api = fmt.Sprintf("%s/api/application/%s", s.host, id)
	)

	err = s.client.GetCallWithForm(s.reqLogger, &resp, api, nil)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	res = resp.Data
	return
}

type Application struct {
	Id          bson.ObjectId
	Version     int
	CreatedBy   bson.ObjectId
	Uid         uint32
	Zone        zone.Zone
	Type        AppType
	Description string
	Status      AppStatus
	AuditedBy   *bson.ObjectId
	AuditedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DetailId    bson.ObjectId
}

const (
	_AppTypeMin         AppType = 1
	AppTypeWriteoff     AppType = 1 // 销账
	AppTypeRecharge     AppType = 2 // 银行转账
	AppTypeFreeReward   AppType = 3 // 充值赠送
	AppTypeSetUserPrice AppType = 4 // 设置用户价格
	AppTypeManualBlock  AppType = 5 // 手动冻结用户
	_AppTypeMax         AppType = 5
)

type AppType int

type AppStatus int

const (
	_AppStatusMin       AppStatus = 0
	AppStatusNew        AppStatus = 0 // 新建
	AppStatusProcessing AppStatus = 1 // 处理中
	AppStatusFinished   AppStatus = 2 // 处理完毕
	AppStatusRejected   AppStatus = 3 // 拒绝
	AppStatusUncertain  AppStatus = 4 // 不确定
	_AppStatusMax       AppStatus = 4
)
