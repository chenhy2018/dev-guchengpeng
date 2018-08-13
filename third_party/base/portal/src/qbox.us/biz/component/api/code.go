package api

type Code interface {
	Code() int
	Humanize() string
}

// 通用错误
const (
	OK              code = 200 // ok
	NeedRedirect    code = 301 // 需要重定向
	InvalidArgs     code = 400 // 请求参数错误，或者数据未通过验证
	Unauthorized    code = 401 // 提供的授权数据未通过（登录已过期，或者不正确）
	Forbidden       code = 403 // 不允许使用此接口
	NotFound        code = 404 // 资源不存在
	Conflict        code = 409 // 资源冲突/重复
	TooManyRequests code = 429 // 访问频率超过限制
	ResultError     code = 500 // 请求结果发生错误
	DatabaseError   code = 598 // 后端数据库查询错误
	CSRFDetected    code = 599 // 检查到 CSRF
)

// 特殊错误
const (
	// just a example
	ErrorcodeExample code = 5000 // 特殊错误代码以 5000 起始

	// access error
	SigninWrongInfo  code = 5100 // 账户或密码错误
	SigninFailed     code = 5101 // 登录失败，可能服务器错误
	SigninBlocked    code = 5102 // 超过5次，被Block，等待5分钟
	InvalidToken     code = 5103 // token, refresh_token 过期或错误
	OverQuota        code = 5104 // 超过配额
	OpIsNotConfirmed code = 5105 // 需要密码确认的操作没有确认密码

)

var codeHumanize = map[code]string{
	OK:              "ok",
	NeedRedirect:    "need redirect",
	InvalidArgs:     "invalid args",
	Unauthorized:    "unauthorized",
	Forbidden:       "forbidden",
	NotFound:        "not found",
	Conflict:        "entry exist",
	TooManyRequests: "too many requests",
	ResultError:     "response result error",
	DatabaseError:   "database err",
	CSRFDetected:    "csrf attack detected",

	// access error
	SigninWrongInfo:  "signin wrong info, username or password is wrong",
	SigninFailed:     "signin failed, server return error",
	SigninBlocked:    "user is blocked 5 minutes",
	InvalidToken:     "signout failed, expired or invalid token",
	OverQuota:        "over quota",
	OpIsNotConfirmed: "op is not confirmed",
}

type code int

func (c code) Code() int {
	return int(c)
}

func (c code) Humanize() string {
	return codeHumanize[c]
}