package pilihub

// 该组建用来建立中间表
import (
	"errors"
	"time"

	"github.com/qiniu/db/mgoutil.v3"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type HubUniq interface {
	Create(hubName string, uid uint32) (HubInfo, error)
	Delete(hubName string, uid uint32) (err error)
	Get(hubName string) (HubInfo, error)
}

type Version string

const (
	V1 Version = "v1"
	V2 Version = "v2"
)

var (
	ErrHubExist    = errors.New("hub already exists")
	ErrHubNotFound = errors.New("hub not found")
)

type HubInfo struct {
	HubName   string    `json:"hub" bson:"_id"`
	Uid       uint32    `json:"uid" bson:"uid"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	Version   Version   `json:"version" bson:"version"`
}

//---------------------------------
type hubUniq struct {
	coll mgoutil.Collection
	ver  Version //pili version
}

func New(coll mgoutil.Collection, ver Version) HubUniq {
	return &hubUniq{
		coll: coll,
		ver:  ver,
	}
}

func (h *hubUniq) Create(hubName string, uid uint32) (HubInfo, error) {
	c := h.coll.CopySession()
	defer c.CloseSession()

	hub := HubInfo{
		HubName:   hubName,
		Uid:       uid,
		Version:   h.ver,
		CreatedAt: time.Now().Round(time.Millisecond),
	}

	err := c.Insert(hub)
	if mgo.IsDup(err) {
		err = ErrHubExist
	}
	return hub, err
}

func (h *hubUniq) Delete(hubName string, uid uint32) (err error) {
	c := h.coll.CopySession()
	defer c.CloseSession()

	err = c.Remove(bson.M{
		"_id":     hubName,
		"uid":     uid,
		"version": h.ver,
	})
	if err == mgo.ErrNotFound {
		err = ErrHubNotFound
	}
	return
}

func (h *hubUniq) Get(hubName string) (HubInfo, error) {
	c := h.coll.CopySession()
	defer c.CloseSession()

	var hub HubInfo
	err := c.FindId(hubName).One(&hub)
	if err == mgo.ErrNotFound {
		err = ErrHubNotFound
	}
	return hub, err
}
