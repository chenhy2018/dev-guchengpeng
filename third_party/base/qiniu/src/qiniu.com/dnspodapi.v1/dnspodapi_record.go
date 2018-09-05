package dnspodapi

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
)

/*
记录相关
*/
// 公共参数
// domain_id 域名ID，必选
// offset 记录开始的偏移，第一条记录为 0，依次类推，可选
// length 共要获取的记录的数量，比如获取20条，则为20，可选
// sub_domain 子域名，如果指定则只返回此子域名的记录，可选
// keyword，搜索的关键字，如果指定则只返回符合该关键字的记录，可选
func (c DnspodClient) GetRecordListById(domain_id string, offset, length int64, sub_domain, keyword string) (*RecordListResp, error) {
	xl := xlog.NewDummy()
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.getRecordList(xl, values, offset, length, sub_domain, keyword)
}

// 公共参数
// domain 域名，必选
// offset 记录开始的偏移，第一条记录为 0，依次类推，可选
// length 共要获取的记录的数量，比如获取20条，则为20，可选
// sub_domain 子域名，如果指定则只返回此子域名的记录，可选
// keyword，搜索的关键字，如果指定则只返回符合该关键字的记录，可选
func (c DnspodClient) GetRecordList(xl *xlog.Logger, domain string, offset, length int64, sub_domain, keyword string) (*RecordListResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.getRecordList(xl, values, offset, length, sub_domain, keyword)
}

type RecordCSt struct {
	Id     string
	Name   string
	Status string
}

type RecordCreateResp struct {
	Status Status
	Record RecordCSt
}

func (c DnspodClient) createDomainRecord(xl *xlog.Logger, values url.Values, sub_domain, record_type, record_line, value string,
	mx, ttl int64, status string, weight int) (*RecordCreateResp, error) {

	if sub_domain != "" {
		values.Add("sub_domain", sub_domain)
	}
	values.Add("record_type", record_type)
	values.Add("record_line", record_line)
	if weight != 0 {
		values.Add("weight", strconv.Itoa(weight))
	}

	values.Add("value", value)
	if mx > 0 {
		values.Add("mx", strconv.FormatInt(mx, 10))
	}
	if ttl > 0 {
		values.Add("ttl", strconv.FormatInt(ttl, 10))
	}
	if status != "" {
		values.Add("status", status)
	} else {
		values.Add("status", "enable")
	}
	//log.Info("before post_data values:", values)
	response, err := c.post_dataEx(xl, "/Record.Create", values)
	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret RecordCreateResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 公共参数
// domain_id 域名ID, 必选
// sub_domain 主机记录, 如 www, 默认@，可选
// record_type 记录类型，通过API记录类型获得，大写英文，比如：A, 必选
// record_line 记录线路，通过API记录线路获得，中文，比如：默认, 必选
// value 记录值, 如 IP:200.200.200.200, CNAME: cname.dnspod.com., MX: mail.dnspod.com., 必选
// mx {1-20} MX优先级, 当记录类型是 MX 时有效，范围1-20, MX记录必选
// ttl {1-604800} TTL，范围1-604800，不同等级域名最小值不同, 可选
func (c DnspodClient) CreateDomainRecordById(xl *xlog.Logger, domain_id, sub_domain, record_type, record_line, value string,
	mx, ttl int64, status string) (*RecordCreateResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.createDomainRecord(xl, values, sub_domain, record_type, record_line, value, mx, ttl, status, 0)
}

// 公共参数
// domain 域名, 必选
// sub_domain 主机记录, 如 www, 默认@，可选
// record_type 记录类型，通过API记录类型获得，大写英文，比如：A, 必选
// record_line 记录线路，通过API记录线路获得，中文，比如：默认, 必选
// value 记录值, 如 IP:200.200.200.200, CNAME: cname.dnspod.com., MX: mail.dnspod.com., 必选
// mx {1-20} MX优先级, 当记录类型是 MX 时有效，范围1-20, MX记录必选
// ttl {1-604800} TTL，范围1-604800，不同等级域名最小值不同, 可选
//status [“enable”, “disable”]，记录初始状态，默认为”enable”，如果传入”disable”，解析不会生效，也不会验证负载均衡的限制，可选
func (c DnspodClient) CreateDomainRecord(xl *xlog.Logger, domain, sub_domain, record_type, record_line, value string,
	mx, ttl int64, status string, weight int) (*RecordCreateResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.createDomainRecord(xl, values, sub_domain, record_type, record_line, value, mx, ttl, status, weight)
}

type RecordMSt struct {
	Id     int64
	Name   string
	Value  string
	Status string
}

type RecordModifyResp struct {
	Status Status
	Record RecordMSt
}

func (c DnspodClient) modifyDomainRecord(xl *xlog.Logger, values url.Values, record_id, sub_domain, record_type, record_line, value string,
	mx, ttl int64, weight int) (*RecordModifyResp, error) {

	values.Add("record_id", record_id)
	if sub_domain != "" {
		values.Add("sub_domain", sub_domain)
	}
	values.Add("record_type", record_type)
	values.Add("record_line", record_line)
	values.Add("weight", strconv.Itoa(weight))
	values.Add("value", value)

	if mx > 0 {
		values.Add("mx", strconv.FormatInt(mx, 10))
	}
	if ttl > 0 {
		values.Add("ttl", strconv.FormatInt(ttl, 10))
	}

	response, err := c.post_dataEx(xl, "/Record.Modify", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret RecordModifyResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 公共参数
// domain_id 域名ID，必选
// record_id 记录ID，必选
// sub_domain 主机记录，默认@，如 www，可选
// record_type 记录类型，通过API记录类型获得，大写英文，比如：A，必选
// record_line 记录线路，通过API记录线路获得，中文，比如：默认，必选
// value 记录值, 如 IP:200.200.200.200, CNAME: cname.dnspod.com., MX: mail.dnspod.com.，必选
// mx {1-20} MX优先级, 当记录类型是 MX 时有效，范围1-20, mx记录必选
// ttl {1-604800} TTL，范围1-604800，不同等级域名最小值不同，可选
func (c DnspodClient) ModifyDomainRecordById(xl *xlog.Logger, domain_id, record_id, sub_domain, record_type, record_line, value string,

	mx, ttl int64, weight int) (*RecordModifyResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.modifyDomainRecord(xl, values, record_id, sub_domain, record_type, record_line, value, mx, ttl, weight)
}

// 公共参数
// domain 域名，必选
// record_id 记录ID，必选
// sub_domain 主机记录，默认@，如 www，可选
// record_type 记录类型，通过API记录类型获得，大写英文，比如：A，必选
// record_line 记录线路，通过API记录线路获得，中文，比如：默认，必选
// value 记录值, 如 IP:200.200.200.200, CNAME: cname.dnspod.com., MX: mail.dnspod.com.，必选
// mx {1-20} MX优先级, 当记录类型是 MX 时有效，范围1-20, mx记录必选
// ttl {1-604800} TTL，范围1-604800，不同等级域名最小值不同，可选
func (c DnspodClient) ModifyDomainRecord(xl *xlog.Logger, domain, record_id, sub_domain, record_type, record_line, value string,
	mx, ttl int64, weight int) (*RecordModifyResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.modifyDomainRecord(xl, values, record_id, sub_domain, record_type, record_line, value, mx, ttl, weight)
}

type RecordRemoveResp struct {
	Status Status
}

func (c DnspodClient) removeDomainRecord(xl *xlog.Logger, values url.Values, record_id string) (*RecordRemoveResp, error) {
	values.Add("record_id", record_id)

	response, err := c.post_dataEx(xl, "/Record.Remove", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret RecordRemoveResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 公共参数
// domain_id 域名ID，必选
// record_id 记录ID，必选
func (c DnspodClient) RemoveDomainRecordById(xl *xlog.Logger, domain_id, record_id string) (*RecordRemoveResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.removeDomainRecord(xl, values, record_id)
}

// 公共参数
// domain 域名，必选
// record_id 记录ID，必选
func (c DnspodClient) RemoveDomainRecord(xl *xlog.Logger, domain, record_id string) (*RecordRemoveResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.removeDomainRecord(xl, values, record_id)
}

type RecordInfoDomain struct {
	Id           string
	Domain       string
	Domain_grade string
}

type RecordInfo struct {
	Id             string
	Sub_domain     string
	Record_type    string
	Record_line    string
	Value          string
	Mx             string
	Ttl            string
	Enabled        string
	Monitor_status string
	Remark         string
	Updated_on     string
	Domain_id      string
}

type RecordInfoResp struct {
	Status Status
	Domain RecordInfoDomain
	Record RecordInfo
}

func (c DnspodClient) getDomainRecordInfo(values url.Values, record_id string) (*RecordInfoResp, error) {
	values.Add("record_id", record_id)

	response, err := c.post_data("/Record.Info", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret RecordInfoResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 请求参数：
// 公共参数
// domain_id 域名ID，必选
// record_id 记录ID，必选
func (c DnspodClient) GetDomainRecordInfoById(domain_id, record_id string) (*RecordInfoResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.getDomainRecordInfo(values, record_id)
}

// 请求参数：
// 公共参数
// domain 域名，必选
// record_id 记录ID，必选
func (c DnspodClient) GetDomainRecordInfo(domain, record_id string) (*RecordInfoResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.getDomainRecordInfo(values, record_id)
}

type RecordSSt struct {
	Id     string
	Name   string
	Status string
}

type RecordStatusResp struct {
	Status Status
	Record RecordSSt
}

func (c DnspodClient) setDomainRecordStatus(xl *xlog.Logger, values url.Values, record_id, status string) (*RecordStatusResp, error) {
	values.Add("record_id", record_id)
	values.Add("status", status)

	response, err := c.post_dataEx(xl, "/Record.Status", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret RecordStatusResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 请求参数：
// 公共参数
// domain_id 域名ID，必选
// record_id 记录ID，必选
// status {enable|disable} 新的状态，必选
func (c DnspodClient) SetDomainRecordStatusById(xl *xlog.Logger, domain_id, record_id, status string) (*RecordStatusResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.setDomainRecordStatus(xl, values, record_id, status)
}

// 请求参数：
// 公共参数
// domain 域名，必选
// record_id 记录ID，必选
// status {enable|disable} 新的状态，必选
func (c DnspodClient) SetDomainRecordStatus(xl *xlog.Logger, domain, record_id, status string) (*RecordStatusResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.setDomainRecordStatus(xl, values, record_id, status)
}

type RecordRemarkResp struct {
	Status Status
}

func (c DnspodClient) setDomainRecordRemark(values url.Values, record_id, remark string) (*RecordRemarkResp, error) {
	values.Add("record_id", record_id)
	values.Add("remark", remark)

	response, err := c.post_data("/Record.Remark", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret RecordRemarkResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 请求参数：
// 公共参数
// domain_id 域名ID，必选
// record_id 记录ID，必选
// remark 域名备注，删除备注请提交空内容，必选
func (c DnspodClient) SetDomainRecordRemarkById(domain_id, record_id, remark string) (*RecordRemarkResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.setDomainRecordRemark(values, record_id, remark)
}

// 请求参数：
// 公共参数
// domain 域名，必选
// record_id 记录ID，必选
// remark 域名备注，删除备注请提交空内容，必选
func (c DnspodClient) SetDomainRecordRemark(domain, record_id, remark string) (*RecordRemarkResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.setDomainRecordRemark(values, record_id, remark)
}
