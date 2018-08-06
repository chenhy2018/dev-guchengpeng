package model

import "time"

const (
	StatusWaiting = "waiting"
	StatusDoing   = "doing"
	StatusSuccess = "success"
	StatusFailure = "failure"
)

// 分发锁，多个实例具有分发能力，锁用于保证同时只有一个实例在分发。
type DispatchLock struct {
	Key        string    `bson:"key"        json:"-"`
	Lock       int64     `bson:"lock"       json:"lock"`
	Owner      string    `bson:"owner"      json:"owner"`
	UpdateTime time.Time `bson:"updateTime" json:"updateTime"`
}

// 分发进度。
type DispatchProgress struct {
	Key        string                  `bson:"key"        json:"-"`
	StartTime  time.Time               `bson:"startTime"  json:"startTime"`
	UpdateTime time.Time               `bson:"updateTime" json:"updateTime"`
	FinishTime time.Time               `bson:"finishTime" json:"finishTime"`
	Progress   []*NodeDispatchProgress `bson:"progress"   json:"progress"`
}

// 单节点分发进度。
type NodeDispatchProgress struct {
	Name       string    `bson:"name"       json:"name"`
	StartTime  time.Time `bson:"startTime"  json:"startTime"`
	UpdateTime time.Time `bson:"updateTime" json:"updateTime"`
	Status     string    `bson:"status"     json:"status"`
	Desc       string    `bson:"desc"       json:"desc"`
}
