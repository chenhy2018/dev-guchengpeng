package link

import (
	"time"

	"github.com/qiniu/xlog.v1"
)

func nowTimestamp() int64 { return time.Now().Unix() }

type App struct {
	App string `json:"app" bson:"app"`

	Bucket               string `json:"bucket" bson:"bucket"`
	SegmentExpireDays    int    `json:"segmentExpireDays" bson:"segmentExpireDays"`
	BucketDownloadDomain string `json:"bucketDownloadDomain" bson:"bucketDownloadDomain"`

	CreatedAt int64 `json:"createdAt" bson:"createdAt"`
	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt"`

	State int `json:"state" bson:"state"`
}

type appStg struct {
}

func (s *appStg) Create(xl *xlog.Logger, app App) (*App, error) {

	now := nowTimestamp()

	app.CreatedAt = now
	app.UpdatedAt = now

	return nil, nil
}

func (s *appStg) Update(xl *xlog.Logger, uid int, app App) (*App, error) {

	app.UpdatedAt = nowTimestamp()
	return nil, nil
}
