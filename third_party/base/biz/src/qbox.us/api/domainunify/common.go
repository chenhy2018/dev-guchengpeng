package domainunify

type RetRes struct {
	Rtn int    `json:"rtn"`
	Msg string `json:"msg"`
}

type CreateArgs struct {
	UID     uint32 `json:"uid"`
	NoCheck bool   `json:"nocheck"`
}

type QueryArgs struct {
	UID       uint32 `json:"uid"`
	StartTime int64  `json:"starttime"`
	EndTime   int64  `json:"endtime"`
}

type TimeRes struct {
	RetRes
	Time int64 `json:"time"`
}

type UpdateUIDArgs struct {
	UID       uint32 `json:"uid"`
	TargetUID uint32 `json:"targetuid"`
}

type UpdateProdArgs struct {
	UID        uint32 `json:"uid"`
	TargetProd string `json:"targetprod"`
}

type LogRes struct {
	RetRes
	Logs []LogEntry
}

type VerifyArgs struct {
	Domain     string `json:"domain"`
	NewCname   bool   `json:"newcname"`
	TargetProd string `json:"targetprod"`
}

type DomainstateRes struct {
	RetRes
	IsConflict      bool  `json:"isconflict"`
	IsFinding       bool  `json:"isfinding"`
	LastFindingTime int64 `json:"lastfindingtime"`
}

type VerifystateRes []VerifyLog

type HvRes struct {
	ID     string `json:"_id"`
	Domain string `json:"domain"`
	Time   int64  `json:"time"`
	Prod   string `json:"prod"`
	OldUid uint32 `json:"olduid"`
	NewUid uint32 `json:"newuid"`
}

type HumanVerifyRes []HvRes

type HumanVerifyArgs struct {
	IsAccept bool   `json:"isaccept"`
	Msg      string `json:"msg"`
}

const (
	DOMAIN_INUSE  = 0
	DOMAIN_DELETE = 1
)

const (
	SUCC                = 0
	OK                  = 200
	DUPLICATE           = 400000
	CONFLICT            = 400001
	NOT_AUTHORIZED_AKSK = 400002
	NOT_VALID_DOMAIN    = 400003
	NOT_ADMIN           = 400004
	INVALID_TIME_ARGS   = 400005
	NO_ICP              = 400006
	IN_FIND_STATE       = 400007
	CONTENT_NOT_MATCH   = 400008
	FIND_TOO_OFTEN      = 400009
	NOT_IN_VERIFY_QUEUE = 400010
	INTERNAL_ERR        = 500000
	DB_FAIL             = 500001
)

var CodeToMsg = map[int]string{
	DUPLICATE:           "This domain has been created by you before",
	CONFLICT:            "This domain has been created by others before",
	NOT_AUTHORIZED_AKSK: "not authorized AKSK",
	NOT_VALID_DOMAIN:    "not valid domain",
	NOT_ADMIN:           "you are not admin user",
	INVALID_TIME_ARGS:   "invalid time arguments",
	NO_ICP:              "no icp info",
	IN_FIND_STATE:       "domain in find state",
	CONTENT_NOT_MATCH:   "content not match",
	FIND_TOO_OFTEN:      "find too often",
	NOT_IN_VERIFY_QUEUE: "not in verify queue",
	INTERNAL_ERR:        "server internal error",
}

var (
	ErrDuplicate         = NewRetRes(DUPLICATE, CodeToMsg[DUPLICATE])
	ErrNotAuthAKSK       = NewRetRes(NOT_AUTHORIZED_AKSK, CodeToMsg[NOT_AUTHORIZED_AKSK])
	ErrNotValidDomain    = NewRetRes(NOT_VALID_DOMAIN, CodeToMsg[NOT_VALID_DOMAIN])
	ErrNotAdmin          = NewRetRes(NOT_ADMIN, CodeToMsg[NOT_ADMIN])
	ErrInternalErr       = NewRetRes(INTERNAL_ERR, CodeToMsg[INTERNAL_ERR])
	ErrNoICP             = NewRetRes(NO_ICP, CodeToMsg[NO_ICP])
	ErrDomainInFindState = NewRetRes(IN_FIND_STATE, CodeToMsg[IN_FIND_STATE])
	ErrContentNotMath    = NewRetRes(CONTENT_NOT_MATCH, CodeToMsg[CONTENT_NOT_MATCH])
	ErrFindTooOften      = NewRetRes(FIND_TOO_OFTEN, CodeToMsg[FIND_TOO_OFTEN])
	ErrConflict          = NewRetRes(CONFLICT, CodeToMsg[CONFLICT])
	ErrInvalidTimeArgs   = NewRetRes(INVALID_TIME_ARGS, CodeToMsg[INVALID_TIME_ARGS])
	ErrNotInVerifyQueue  = NewRetRes(NOT_IN_VERIFY_QUEUE, CodeToMsg[NOT_IN_VERIFY_QUEUE])
)

func NewRetRes(rtn int, msg string) (r RetRes) {
	r.Rtn = rtn
	r.Msg = msg
	return
}
