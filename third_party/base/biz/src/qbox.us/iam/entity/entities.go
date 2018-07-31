package entity

import (
	"time"

	"gopkg.in/mgo.v2/bson"
	"qbox.us/iam/enums"
)

type (
	User struct {
		ID            string    `json:"id"`
		RootUID       uint32    `json:"root_uid"`        // 根用户ID
		IUID          uint32    `json:"iuid"`            // Iam用户UID
		Alias         string    `json:"alias"`           // 用户别名
		CreatedAt     time.Time `json:"created_at"`      // 创建时间
		UpdatedAt     time.Time `json:"updated_at"`      // 更新时间
		LastLoginTime time.Time `json:"last_login_time"` // 最后一次登录时间
		Enabled       bool      `json:"enabled"`         // 是否启用
	}

	UserGroup struct {
		ID          string    `json:"id"`
		RootUID     uint32    `json:"root_uid"`    // 根用户ID
		Alias       string    `json:"alias"`       // 组别名
		Description string    `json:"description"` // 描述
		CreatedAt   time.Time `json:"created_at"`  // 创建时间
		UpdatedAt   time.Time `json:"updated_at"`  // 更新时间
		Enabled     bool      `json:"enabled"`     // 是否启用
	}

	RootUser struct {
		ID        string    `json:"id"`
		UID       uint32    `json:"uid"`        // 用户ID
		Alias     string    `json:"alias"`      // 别名
		CreatedAt time.Time `json:"created_at"` // 创建时间
		UpdatedAt time.Time `json:"updated_at"` // 更新时间
		Enabled   bool      `json:"enabled"`    // 是否启用
	}
)

type (
	Statement struct {
		Action   []string     `json:"action"`   // 动作
		Resource []string     `json:"resource"` // 资源
		Effect   enums.Effect `json:"effect"`   // 效果
	}

	Policy struct {
		ID          string               `json:"id"`
		RootUID     uint32               `json:"root_uid"`    // 根用户ID （自定义策略有）
		Alias       string               `json:"alias"`       // 策略别名
		Description string               `json:"description"` // 描述
		EditType    enums.PolicyEditType `json:"edit_type"`   // 策略编辑类型
		CreatedAt   time.Time            `json:"created_at"`  // 创建时间
		UpdatedAt   time.Time            `json:"updated_at"`  // 更新时间
		Enabled     bool                 `json:"enabled"`     // 是否启用
		Statement   []Statement          `json:"statement"`   // 声明
	}

	AuditLog struct {
		ID            string                 `json:"id"`
		RootUID       uint32                 `json:"root_uid"`       // 根用户ID
		IUID          uint32                 `json:"iuid"`           // Iam用户UID
		Service       string                 `json:"service"`        // 服务（产品）
		Action        string                 `json:"action"`         // 动作
		Resources     []string               `json:"resources"`      // 资源
		CreatedAt     time.Time              `json:"created_at"`     // 创建时间
		SourceIP      string                 `json:"source_ip"`      // 发送API请求的源IP地址
		UserAgent     string                 `json:"user_agent"`     // 用户代理
		RequestParams map[string]interface{} `json:"request_params"` // 请求参数
		RequestID     string                 `json:"request_id"`     // 请求ID
		EventTime     time.Time              `json:"event_time"`     // 发出请求的时间
		EventVersion  string                 `json:"event_version"`  // 日志事件格式的版本
		EventSource   string                 `json:"event_source"`   // 处理请求的server
		ErrorCode     int                    `json:"error_code"`     // 错误码
		ErrorMessage  string                 `json:"error_message"`  // 错误信息
	}

	Keypair struct {
		Id        bson.ObjectId `json:"id"`
		CreatedAt time.Time     `json:"created_at"`
		AccessKey string        `json:"access_key"`
		SecretKey string        `json:"secret_key"`
		UserId    bson.ObjectId `json:"user_id"`
		Enabled   bool          `json:"enabled" bson:"enabled"`
	}

	SecretInfo struct {
		ID        string `json:"id"`
		SecretKey string `json:"secret_key"`
		AccessKey string `json:"access_key"`
		RootUID   uint32 `json:"root_uid"`
		IUID      uint32 `json:"iuid"`
		Alias     string `json:"alias"`
	}
	Action struct {
		Id        bson.ObjectId     `json:"id"`
		Name      string            `json:"name"`       // 动作名
		Alias     string            `json:"alias"`      // 动作别名
		Service   string            `json:"service"`    // 服务（产品）
		Scope     enums.ActionScope `json:"scope"`      // 动作类型（区分是否可以绑定资源）
		CreatedAt time.Time         `json:"created_at"` // 创建时间
		UpdatedAt time.Time         `json:"updated_at"` // 更新时间
		Enabled   bool              `json:"enabled"`    // 是否启用
	}

	QConfInfo struct {
		AccessKey string      `json:"access_key"`
		SecretKey string      `json:"secret_key"`
		RootUID   uint32      `json:"root_uid"`
		IUID      uint32      `json:"iuid"`
		UserAlias string      `json:"user_alias"`
		Enabled   bool        `json:"enabled"`
		Version   string      `json:"version"`
		Statement []Statement `json:"statement"`
	}
)

func (i *QConfInfo) GetSecret() []byte {
	return []byte(i.SecretKey)
}

func (i *QConfInfo) GetIUID() uint32 {
	return i.IUID
}

func (i *QConfInfo) GetVersion() string {
	return i.Version
}

func (i *QConfInfo) GetStatments() []Statement {
	return i.Statement
}

func (i *QConfInfo) IsEnabled() bool {
	return i.Enabled
}
