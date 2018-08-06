package account

// ---------------------------------------------------------------------------------------

type UserInfo struct {
	Uid     uint32 `json:"uid"`
	Sudoer  uint32 `json:"suid,omitempty"`
	Utype   uint32 `json:"ut"`
	UtypeSu uint32 `json:"sut,omitempty"`
	Devid   uint32 `json:"dev,omitempty"`
	Appid   uint32 `json:"app,omitempty"`
	Expires uint32 `json:"e,omitempty"`
}

// ---------------------------------------------------------------------------------------
