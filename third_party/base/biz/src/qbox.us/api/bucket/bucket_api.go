package bucket

import "qbox.us/api/one/domain"

type Entry struct {
	Tbl               string `json:"tbl" bson:"tbl"`
	Uid               uint32 `json:"uid" bson:"uid"`
	Itbl              uint32 `json:"itbl" bson:"itbl"`
	PhyTbl            string `json:"phy" bson:"phy"`
	Ctime             int64  `json:"ctime" bson:"ctime"`
	DropTime          int64  `json:"drop" bson:"drop"` // !=0时，表示该条目被删除（包括uc和domain）
	Region            string `json:"region" bson:"region"`
	Zone              string `json:"zone" bson:"zone"`
	Global            bool   `json:"global" bson:"global"`
	Line              bool   `json:"line" bson:"line"`
	VersioningEnabled bool   `bson:"versioning_enabled" json:"versioning_enabled"`

	Ouid  uint32 `json:"ouid" bson:"ouid,omitempty"`
	Oitbl uint32 `json:"oitbl" bson:"oitbl,omitempty"`
	Otbl  string `json:"otbl" bson:"otbl,omitempty"`
	Perm  uint32 `json:"perm" bson:"perm,omitempty"`

	Val        string       `json:"val" bson:"val,omitempty"`
	DomainInfo []DomainInfo `json:"domain_info" bson:"domain_info,omitempty"`
}

type DomainInfo struct {
	Domain    string           `json:"domain" bson:"domain"`
	Refresh   bool             `json:"refresh" bson:"refresh"`
	Global    bool             `json:"global" bson:"global"`
	AntiLeech domain.AntiLeech `json:"antileech,omitempty" bson:"antileech,omitempty"`
}
