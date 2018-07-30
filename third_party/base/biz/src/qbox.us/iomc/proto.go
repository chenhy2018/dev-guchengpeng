package iomc

type Message struct {
	Uid        uint32 `json:"uid"`
	Bucket     string `json:"bucket"`
	Key        string `json:"key"`
	Itbl       uint32 `json:"itbl"`
	Op         int    `json:"op"`
	Fh         []byte `json:"fh"`
	PutTime    int64  `json:"put_time"`
	ChangeTime int64  `json:"change_time"` //time.Now().Unix()
	IoIgnored  bool   `json:"io_ignored"`
	Version    string `json:"version"`
}

const (
	OpDelete = iota
	OpIns
	OpPut
	OpFailed
	OpUpdateFh
	OpDeleteAfterDays
	OpChgm
	OpChtype
	OpCopy
	OpChstatus
	OpSetMd5
)
