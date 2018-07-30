package billing

import (
	"net/url"
	"strconv"

	"github.com/qiniu/rpc.v2"
	"ustack.com/admin_api.v1/ustack"
)

type Client struct {
	Conn ustack.Conn
}

func New(services ustack.Services) Client {
	conn, ok := services.Find("gringotts")
	if !ok {
		panic("billing api not found")
	}

	return Client{conn}
}

func isFakeError(err error) bool {
	if rpc.HttpCodeOf(err)/100 == 2 {
		return true
	}
	return false
}

// ----------------------------------------------------------

type ExpenseArgs struct {
	ProjectId string
	Offset    int
	Limit     int
	StartTime string
	EndTime   string
}

type Order struct {
	UserId       string `json:"user_id"`
	ProjectId    string `json:"project_id"`
	OrderId      string `json:"order_id"`
	ResourceId   string `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UnitPrice    string `json:"unit_price"`
	TotalPrice   string `json:"total_price"`
}

type ExpenseRet struct {
	Orders []Order `json:"orders"`
}

func (p Client) GetExpense(l rpc.Logger, args ExpenseArgs) (ret []Order, err error) {
	path := "/v1/orders?"
	query := make(url.Values)
	query.Set("project_id", args.ProjectId)
	query.Set("offset", strconv.Itoa(args.Offset))
	query.Set("limit", strconv.Itoa(args.Limit))
	query.Set("start_time", args.StartTime)
	query.Set("end_time", args.EndTime)

	url := path + query.Encode()
	uret := &ExpenseRet{}
	err = p.Conn.Call(l, uret, "GET", url)
	if err != nil && isFakeError(err) {
		err = nil
	}
	ret = uret.Orders
	return
}

// ----------------------------------------------------------
// WARNING: 该接口会删掉用户资源，调用前请确保已知道影响！
//          该接口建议仅用于注销用户，先使用其它接口逐个删除计费相关资源（虚机、磁盘、IP、快照等），插入计费日志，再调用此接口删除其它非计费相关资源。

func (p Client) DeleteProjectResources(l rpc.Logger, projectId string) (err error) {
	err = p.Conn.Call(l, nil, "DELETE", "/v2/resources?project_id="+projectId)
	if err != nil && isFakeError(err) {
		err = nil
	}
	return
}
