package dnspodapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
)

type Config struct {
	Login_email       string      `json:"login_email"`
	Login_password    string      `json:"login_password"`
	Login_token       string      `json:"login_token"`
	Login_code        string      `json:"login_code"`
	Login_code_cookie http.Cookie `json:"login_code_cookie"`
}

type DnspodClient struct {
	conf Config
}

func NewDnspodClient(conf Config) (dc *DnspodClient, err error) {
	return &DnspodClient{conf}, nil
}

func (c DnspodClient) post_data(url string, content url.Values) (*http.Response, error) {
	client := &http.Client{}
	values := c.generate_header(content)

	req, _ := http.NewRequest("POST", "https://dnsapi.cn"+url, strings.NewReader(values.Encode()))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req.Header.Set("User-Agent", fmt.Sprintf("QNGoDNS/0.1 (%s)", c.conf.Login_email))

	req.AddCookie(&(c.conf.Login_code_cookie))
	return client.Do(req)
}

func (c DnspodClient) post_dataEx(xl *xlog.Logger, url string, content url.Values) (*http.Response, error) {
	client := &http.Client{}
	values := c.generate_header(content)

	req, _ := http.NewRequest("POST", "https://dnsapi.cn"+url, strings.NewReader(values.Encode()))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req.Header.Set("User-Agent", fmt.Sprintf("QNGoDNS/0.1 (%s)", c.conf.Login_email))

	req.AddCookie(&(c.conf.Login_code_cookie))

	return client.Do(req)
}

func (c DnspodClient) generate_header(content url.Values) url.Values {
	header := url.Values{}
	if c.conf.Login_email != "" && c.conf.Login_password != "" {
		header.Add("login_email", c.conf.Login_email)
		header.Add("login_password", c.conf.Login_password)
	} else {
		if c.conf.Login_token != "" {
			header.Add("login_token", c.conf.Login_token)
		}
	}
	header.Add("format", "json")
	header.Add("lang", "en")
	header.Add("error_on_empty", "no")

	if content != nil {
		for k, _ := range content {
			header.Add(k, content.Get(k))
		}
	}

	return header
}

type CommonResp struct {
	Status Status
}

func (c DnspodClient) ApiVersion() (*CommonResp, error) {
	response, err := c.post_data("/Info.Version", nil)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret CommonResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

/*
 域名相关
*/

type domain struct {
	Id       string
	Punycode string
	Domain   string
}
type CreateDomainResp struct {
	Status Status
	Domain domain
}

// 请求参数：
// 公共参数
// domain 域名, 没有 www, 如 dnspod.com
// group_id 域名分组ID, 可选参数
// is_mark {yes|no} 是否星标域名, 可选参数
func (c DnspodClient) CreateDomain(domain, group_id, is_mark string) (*CreateDomainResp, error) {
	values := url.Values{}
	values.Add("domain", domain)
	if group_id != "" {
		values.Add("group_id", group_id)
	}
	if is_mark != "" {
		values.Add("is_mark", is_mark)
	}
	response, err := c.post_data("/Domain.Create", values)
	if err != nil {
		log.Error("post data err!")
		return nil, err
	}
	defer response.Body.Close()

	var ret CreateDomainResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

type DLInfo struct {
	Domain_total    int64
	All_total       int64
	Mine_total      int64
	Share_total     int64
	Vip_total       int64
	Ismark_total    int64
	Pause_total     int64
	Error_total     int64
	Lock_total      int64
	Spam_total      int64
	Vip_expire      int64
	Share_out_total int64
}

type DomainListItem struct {
	Id                int64
	Status            string
	Grade             string
	Group_id          string
	Searchengine_push string
	Is_mark           string
	Ttl               string
	Cname_speedup     string
	Remark            bool
	Created_on        string
	Updated_on        string
	Punycode          string
	Ext_status        string
	Name              string
	Grade_title       string
	Is_vip            string
	Owner             string
	Records           string
	Auth_to_anquanbao bool
}

type DomainListResp struct {
	Status  Status
	Info    DLInfo
	Domains []DomainListItem
	//Domains []interface{}
}

// 请求参数：
// 公共参数

// dtype 域名权限种类, 可选参数, 默认为’all’. 包含以下类型：
// all：所有域名
// mine：我的域名
// share：共享给我的域名
// ismark：星标域名
// pause：暂停域名
// vip：VIP域名
// recent：最近操作过的域名
// share_out：我共享出去的域名

// offset 记录开始的偏移, 第一条记录为 0, 依次类推, 可选参数
// length 共要获取的记录的数量, 比如获取20条, 则为20, 可选参数
// group_id 分组ID, 获取指定分组的域名, 可选参数
// keyword, 搜索的关键字, 如果指定则只返回符合该关键字的域名, 可选参数
func (c DnspodClient) GetDomainList(dtype string, offset, length int64, group_id, keyword string) (*DomainListResp, error) {
	values := url.Values{}
	if dtype != "" {
		values.Add("type", dtype)
	} else {
		values.Add("type", "all")
	}
	if offset >= 0 {
		values.Add("offset", strconv.FormatInt(offset, 10))
	}
	if length > 0 {
		values.Add("length", strconv.FormatInt(length, 10))
	}
	if group_id != "" {
		values.Add("group_id", group_id)
	}
	if keyword != "" {
		values.Add("keyword", keyword)
	}

	response, err := c.post_data("/Domain.List", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret DomainListResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

func (c DnspodClient) removeDomain(values url.Values) (*CommonResp, error) {
	response, err := c.post_data("/Domain.Remove", values)
	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	var ret CommonResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 请求参数：
// 公共参数
// domain_id 域名ID
func (c DnspodClient) RemoveDomainByID(domain_id string) (*CommonResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.removeDomain(values)
}

// 请求参数：
// 公共参数
// domain 域名
func (c DnspodClient) RemoveDomain(domain string) (*CommonResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.removeDomain(values)
}

type DomainInfo struct {
	Id                string
	Name              string
	Punycode          string
	Grade             string
	Grade_title       string
	Status            string
	Ext_status        string
	Records           string
	Group_id          string
	Is_mark           string
	Remark            bool
	Is_vip            string
	Searchengine_push string
	User_id           string
	Created_on        string
	Updated_on        string
	Ttl               string
	Cname_speedup     string
	Owner             string
	Auth_to_anquanbao bool
}

type DomainInfoResp struct {
	Status Status
	Domain DomainInfo
}

func (c DnspodClient) getDomainInfo(values url.Values) (*DomainInfoResp, error) {
	response, err := c.post_data("/Domain.Info", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret DomainInfoResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 请求参数：
// 公共参数
// domain_id 域名ID
func (c DnspodClient) GetDomainInfoByID(domain_id string) (*DomainInfoResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.getDomainInfo(values)
}

// 请求参数：
// 公共参数
// domain 域名
func (c DnspodClient) GetDomainInfo(domain string) (*DomainInfoResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.getDomainInfo(values)
}

func (c DnspodClient) setDomainStatus(values url.Values, status string) (*CommonResp, error) {
	values.Add("status", status)

	response, err := c.post_data("/Domain.Status", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret CommonResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 请求参数：
// 公共参数
// domain_id 域名ID
// status {enable, disable} 域名状态
func (c DnspodClient) SetDomainStatusByID(domain_id, status string) (*CommonResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.setDomainStatus(values, status)
}

// 请求参数：
// 公共参数
// domain 域名
// status {enable, disable} 域名状态
func (c DnspodClient) SetDomainStatus(domain, status string) (*CommonResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.setDomainStatus(values, status)
}

type LogInfo struct {
	Count int64
	Size  int64
}

type DomainLogResp struct {
	Status Status
	Log    []string
	Info   LogInfo
}

func (c DnspodClient) getDomainLog(values url.Values, offset, length int64) (*DomainLogResp, error) {
	if offset >= 0 {
		values.Add("offset", strconv.FormatInt(offset, 10))
	}
	if length > 0 {
		values.Add("length", strconv.FormatInt(length, 10))
	}

	response, err := c.post_data("/Domain.Log", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret DomainLogResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 公共参数
// domain_id 或 domain，分别对应域名ID和域名，提交其中一个即可
// offset 记录开始的偏移，第一条记录为 0，依次类推，可选参数
// length 共要获取的日志条数，比如获取20条，则为20，可选参数。默认为500条，最大值为500
func (c DnspodClient) GetDomainLogByID(domain_id string, offset, length int64) (*DomainLogResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.getDomainLog(values, offset, length)
}

// 公共参数
// domain 域名
// offset 记录开始的偏移，第一条记录为 0，依次类推，可选参数
// length 共要获取的日志条数，比如获取20条，则为20，可选参数。默认为500条，最大值为500
func (c DnspodClient) GetDomainLog(domain string, offset, length int64) (*DomainLogResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.getDomainLog(values, offset, length)
}

type DomainAlias struct {
	Id     string
	Domain string
}

type DomainAliasListResp struct {
	Status Status
	Alias  []DomainAlias
}

func (c DnspodClient) getDomainaliasList(values url.Values) (*DomainAliasListResp, error) {
	response, err := c.post_data("/Domainalias.List", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()
	var ret DomainAliasListResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 公共参数
// domain_id 域名ID
func (c DnspodClient) GetDomainaliasListById(domain_id string) (*DomainAliasListResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.getDomainaliasList(values)
}

// 公共参数
// domain 域名
func (c DnspodClient) GetDomainaliasList(domain string) (*DomainAliasListResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.getDomainaliasList(values)
}

type AliasInfo struct {
	Id       string
	Punycode string
}
type DomainAliasResp struct {
	Status Status
	Alias  AliasInfo
}

func (c DnspodClient) createDomainalias(values url.Values, domainalias string) (*DomainAliasResp, error) {
	values.Add("domain", domainalias)

	response, err := c.post_data("/Domainalias.Create", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret DomainAliasResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 公共参数
// domain_id , 域名ID
// domainalias 要绑定的域名, 不带www.
func (c DnspodClient) CreateDomainaliasById(domain_id, domainalias string) (*DomainAliasResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.createDomainalias(values, domainalias)
}

func (c DnspodClient) removeDomainalias(values url.Values, alias_id string) (*CommonResp, error) {
	values.Add("alias_id", alias_id)

	response, err := c.post_data("/Domainalias.Remove", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret CommonResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 公共参数
// domain_id 域名ID
// alias_id 绑定ID, 绑定域名的时候会返回
func (c DnspodClient) RemoveDomainaliasById(domain_id, alias_id string) (*CommonResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.removeDomainalias(values, alias_id)
}

// 公共参数
// domain 域名
// alias_id 绑定ID, 绑定域名的时候会返回
func (c DnspodClient) RemoveDomainalias(domain, alias_id string) (*CommonResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.removeDomainalias(values, alias_id)
}

type PurviewInfoItem struct {
	Name  string
	Value int64
}
type PurviewInfoResp struct {
	Status  Status
	Purview []PurviewInfoItem
}

func (c DnspodClient) getDomainPurview(values url.Values) (*PurviewInfoResp, error) {
	response, err := c.post_data("/Domain.Purview", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret PurviewInfoResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 公共参数
// domain_id 域名ID
func (c DnspodClient) GetDomainPurviewById(domain_id string) (*PurviewInfoResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.getDomainPurview(values)
}

// 公共参数
// domain 域名
func (c DnspodClient) GetDomainPurview(domain string) (*PurviewInfoResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.getDomainPurview(values)
}

/*
{
    "status": {
        "code": "1",
        "message": "Action completed successful",
        "created_at": "2015-01-18 18:23:40"
    },
    "types": [
        "A",
        "CNAME",
        "MX",
        "TXT",
        "NS",
        "AAAA",
        "SRV",
        "URL"
    ]
}
*/

type RecordTypeResp struct {
	Status Status
	Types  []string
}

// 公共参数
// domain_grade 域名等级, 分别为：D_Free, D_Plus, D_Extra, D_Expert, D_Ultra, 分别对应免费套餐、个人豪华、企业1、企业2、企业3
// 新套餐：DP_Free DP_Plus DP_Extra DP_Expert DP_Ultra, 分别对应新免费、个人专业版、企业创业版、企业标准版、企业旗舰版
func (c DnspodClient) GetRecordType(domain_grade string) (*RecordTypeResp, error) {
	values := url.Values{}
	values.Add("domain_grade", domain_grade)

	response, err := c.post_data("/Record.Type", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret RecordTypeResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

type RecordLineResp struct {
	Status Status
	Lines  []string
}

func (c DnspodClient) getRecordLine(xl *xlog.Logger, values url.Values, domain_grade string) (*RecordLineResp, error) {
	values.Add("domain_grade", domain_grade)

	response, err := c.post_dataEx(xl, "/Record.Line", values)

	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret RecordLineResp
	json.NewDecoder(response.Body).Decode(&ret)

	return &ret, err
}

// 公共参数
// domain_grade 域名等级, 分别为：D_Free, D_Plus, D_Extra, D_Expert, D_Ultra, 分别对应免费套餐、个人豪华、企业Ⅰ、企业Ⅱ、企业Ⅲ.
// 新套餐：DP_Free, DP_Plus, DP_Extra, DP_Expert, DP_Ultra, 分别对应新免费、个人专业版、企业创业版、企业标准版、企业旗舰版
// domain_id 域名ID
func (c DnspodClient) GetRecordLineById(xl *xlog.Logger, domain_id, domain_grade string) (*RecordLineResp, error) {
	values := url.Values{}
	values.Add("domain_id", domain_id)

	return c.getRecordLine(xl, values, domain_grade)
}

// 公共参数
// domain_grade 域名等级, 分别为：D_Free, D_Plus, D_Extra, D_Expert, D_Ultra, 分别对应免费套餐、个人豪华、企业Ⅰ、企业Ⅱ、企业Ⅲ.
// 新套餐：DP_Free, DP_Plus, DP_Extra, DP_Expert, DP_Ultra, 分别对应新免费、个人专业版、企业创业版、企业标准版、企业旗舰版
// domain 域名
func (c DnspodClient) GetRecordLine(xl *xlog.Logger, domain, domain_grade string) (*RecordLineResp, error) {
	values := url.Values{}
	values.Add("domain", domain)

	return c.getRecordLine(xl, values, domain_grade)
}

type RecordListDomain struct {
	Id       string
	Name     string
	Punycode string
	Grade    string
	Owner    string
}

type RecordListInfo struct {
	Sub_domains  string
	Record_total json.Number
}

type RecordListItem struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Line           string `json:"line"`
	Type           string `json:"type"`
	Ttl            string `json:"ttl"`
	Value          string `json:"value"`
	Mx             string `json:"mx"`
	Enabled        string `json:"enabled"`
	Status         string `json:"status"`
	Monitor_status string `json:"monitor_status"`
	Remark         string `json:"remark"`
	Update_on      string `json:"updated_on"`
	Use_aqb        string `json:"use_aqb"`
	Weight         int    `json:"weight"`
}

type RecordListResp struct {
	Status  Status
	Domain  RecordListDomain
	Info    RecordListInfo
	Records []RecordListItem `json:"records"`
}

type Status struct {
	Code       string
	Message    string
	Created_at string
}

func (c DnspodClient) getRecordList(xl *xlog.Logger, values url.Values, offset, length int64, sub_domain, keyword string) (*RecordListResp, error) {
	if offset >= 0 {
		values.Add("offset", strconv.FormatInt(offset, 10))
	}
	if length > 0 {
		values.Add("length", strconv.FormatInt(length, 10))
	}
	if sub_domain != "" {
		values.Add("sub_domain", sub_domain)
	}
	if keyword != "" {
		values.Add("keyword", keyword)
	}

	response, err := c.post_dataEx(xl, "/Record.List", values)
	if err != nil {
		log.Error("post data err!")
		return nil, err
	}

	defer response.Body.Close()

	var ret RecordListResp
	dec := json.NewDecoder(response.Body)
	dec.UseNumber()
	if err = dec.Decode(&ret); err != nil {
		log.Info(err)
	}

	return &ret, err
}
