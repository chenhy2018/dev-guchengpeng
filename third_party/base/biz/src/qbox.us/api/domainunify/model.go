package domainunify

import (
	"labix.org/v2/mgo/bson"
)

type M bson.M

type Domain struct {
	Domain string `json:"domain" bson:"domain"`
	UID    uint32 `json:"uid" bson:"uid"`
	Prod   string `json:"prod" bson:"prod"`
	Status int    `json:"status" bson:"status"`
}

type Log struct {
	Domain string     `bson:"domain" json:"domain"`
	Logs   []LogEntry `bson:"logs" json:"logs"`
}

type LogEntry struct {
	StartTime int64  `bson:"starttime" json:"starttime"`
	UID       uint32 `json:"uid" bson:"uid"`
	Prod      string `json:"prod" bson:"prod"`
	Status    int    `json:"status" bson:"status"`
}

const (
	WV = 0 // waiting verify
	VC = 1 // verify complete
	VS = 2 // verify success
	VF = 3 // verify fail
)

type VerifyState struct {
	Domain string    `json:"domain" bson:"domain"`
	UID    uint32    `json:"uid" bson:"uid"`
	State  int       `json:"state" bson:"state"`
	Data   StateData `json:"data" bson:"data"`
}

type StateData struct {
	FileName   string `json:"filename" bson:"filename"`
	Content    string `json:"content" bson:"content"`
	NewCname   bool   `json:"newcname" bson:"newcname"`
	TargetProd string `json:"targetprod" bson:"targetprod"`
	OldProd    string `json:"oldprod"`
	OldUid     uint32 `json:"olduid"`
	Time       int64  `json:"time" bson:"time"`
}

type VerifyLog struct {
	Domain     string `json:"domain" bson:"domain"`
	OldUID     uint32 `json:"olduid" bson:"olduid"`
	UID        uint32 `json:"uid" bson:"uid"`
	Time       int64  `bson:"time" json:"time"`
	State      int    `json:"state" bson:"state"`
	OldProd    string `json:"oldprod" bson:"oldprod"`
	TargetProd string `json:"targetprod" bson:"targetprod"`
	NewCname   bool   `json:"newcname" bson:"newcname"`
	Msg        string `json:"msg" bson:"msg"`
}
