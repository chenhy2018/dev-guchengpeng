package stat

import (
	"net/http"
	"strconv"

	"qbox.us/rpc"
	"qbox.us/servend/account"
	"qbox.us/servend/oauth"
)

//---------------------------------------------------------------------------//

type TimeUnit string

const (
	TIME_UNIT_FIVE_MIN TimeUnit = "5min"
	TIME_UNIT_DAY      TimeUnit = "day"
	TIME_UNIT_MONTH    TimeUnit = "month"
)

type StatInfo struct {
	Space       int64 `json:"space"`
	Space_avg   int64 `json:"space_avg"`
	Bandwidth   int64 `json:"bandwidth"`
	Apicall_get int64 `json:"apicall_get"`
	Apicall_put int64 `json:"apicall_put"`
	Transfer    int64 `json:"transfer"`
	OvTransfer  int64 `json:"ov_transfer"`
}

func Info(c rpc.Client, host, uid, bucket string, emptybucket bool, month string) (info StatInfo, code int, err error) {
	code, err = c.CallWithForm(&info, host+"/info",
		map[string][]string{
			"uid":         []string{uid},
			"bucket":      []string{bucket},
			"emptybucket": []string{strconv.FormatBool(emptybucket)},
			"month":       []string{month},
		})
	return
}

type StatInfoDaily struct {
	Space             int64 `json:"space"`
	Apicall_get       int64 `json:"apicall_get"`
	Apicall_get_Month int64 `json:"apicall_get_month"`
	Apicall_put       int64 `json:"apicall_put"`
	Apicall_put_Month int64 `json:"apicall_put_month"`
	Transfer          int64 `json:"transfer"`
	Transfer_Month    int64 `json:"transfer_month"`
}

func DailyInfo(c rpc.Client, host, uid, bucket string, emptybucket bool, day string) (info StatInfoDaily, code int, err error) {
	code, err = c.CallWithForm(&info, host+"/info/day",
		map[string][]string{
			"uid":         []string{uid},
			"bucket":      []string{bucket},
			"emptybucket": []string{strconv.FormatBool(emptybucket)},
			"day":         []string{day},
		})
	return
}

func AllUser(c rpc.Client, host string, month string) (
	uids []uint32, code int, err error) {
	code, err = c.CallWithForm(&uids, host+"/admin/alluser",
		map[string][]string{
			"month": []string{month},
		})
	return
}

type DataResult struct {
	Time            []int64 `json:"time"`
	Data            []int64 `json:"data"`
	RealtimeIndexes []int   `json:"realtimeIndexes"`
}

// TODO p string --> p TimeUnit
func SelectSpace(c rpc.Client, host, uid, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	code, err = c.CallWithForm(&data, host+"/select/space",
		map[string][]string{
			"uid":         []string{uid},
			"bucket":      []string{bucket},
			"emptybucket": []string{strconv.FormatBool(emptybucket)},
			"from":        []string{from},
			"to":          []string{to},
			"p":           []string{p},
		})
	return
}

func SelectSpaceAdd(c rpc.Client, host, uid, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	code, err = c.CallWithForm(&data, host+"/select/space_add",
		map[string][]string{
			"uid":         []string{uid},
			"bucket":      []string{bucket},
			"emptybucket": []string{strconv.FormatBool(emptybucket)},
			"from":        []string{from},
			"to":          []string{to},
			"p":           []string{p},
		})
	return
}

func SelectApicall(c rpc.Client, host, uid, bucket string, emptybucket bool, apitype, from, to, p string) (
	data DataResult, code int, err error) {
	code, err = c.CallWithForm(&data, host+"/select/apicall",
		map[string][]string{
			"uid":         []string{uid},
			"bucket":      []string{bucket},
			"emptybucket": []string{strconv.FormatBool(emptybucket)},
			"type":        []string{apitype},
			"from":        []string{from},
			"to":          []string{to},
			"p":           []string{p},
		})
	return
}

func SelectTransfer(c rpc.Client, host, uid, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	code, err = c.CallWithForm(&data, host+"/select/transfer",
		map[string][]string{
			"uid":         []string{uid},
			"bucket":      []string{bucket},
			"emptybucket": []string{strconv.FormatBool(emptybucket)},
			"from":        []string{from},
			"to":          []string{to},
			"p":           []string{p},
		})
	return
}

func SelectOvTransfer(c rpc.Client, host, uid, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	code, err = c.CallWithForm(&data, host+"/select/ov_transfer",
		map[string][]string{
			"uid":         []string{uid},
			"bucket":      []string{bucket},
			"emptybucket": []string{strconv.FormatBool(emptybucket)},
			"from":        []string{from},
			"to":          []string{to},
			"p":           []string{p},
		})
	return
}

func SelectBandwidth(c rpc.Client, host, uid, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	code, err = c.CallWithForm(&data, host+"/select/bandwidth",
		map[string][]string{
			"uid":         []string{uid},
			"bucket":      []string{bucket},
			"emptybucket": []string{strconv.FormatBool(emptybucket)},
			"from":        []string{from},
			"to":          []string{to},
			"p":           []string{p},
		})
	return
}

func SelectOvBandwidth(c rpc.Client, host, uid, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	code, err = c.CallWithForm(&data, host+"/select/ov_bandwidth",
		map[string][]string{
			"uid":         []string{uid},
			"bucket":      []string{bucket},
			"emptybucket": []string{strconv.FormatBool(emptybucket)},
			"from":        []string{from},
			"to":          []string{to},
			"p":           []string{p},
		})
	return
}

func SelectMoney(c rpc.Client, host, uid, type_, from, to, p string) (
	data DataResult, code int, err error) {
	code, err = c.CallWithForm(&data, host+"/select/money",
		map[string][]string{
			"uid":         []string{uid},
			"type":        []string{type_},
			"from":        []string{from},
			"emptybucket": []string{"true"},
			"to":          []string{to},
			"p":           []string{p},
		})
	return
}

//mps needs pay
//pipeline:pipeline name, if need all, please set "";
//mpskey, such as HD, SD, imageAve,vframe
func SelectMpsValue(c rpc.Client, host, uid string, pipeline string, mpskey string, from, to string, p TimeUnit) (
	data DataResult, code int, err error) {
	code, err = c.CallWithForm(&data, host+"/select/fopg",
		map[string][]string{
			"uid":      []string{uid},
			"pipename": []string{pipeline},
			"type":     []string{mpskey},
			"from":     []string{from},
			"to":       []string{to},
			"p":        []string{string(p)},
		})
	return
}

type BucketsResult struct {
	Bucket []string `json:"bucket"`
	Data   []int64  `json:"data"`
}

// 参数说明
//    p  时间粒度选择，可选day|month，为空默认选择day
func BucketsSpace(c rpc.Client, host, uid, from, to string, p TimeUnit) (
	data BucketsResult, code int, err error) {
	args := map[string][]string{
		"uid":  []string{uid},
		"from": []string{from},
		"to":   []string{to},
	}
	if p != "" {
		args["p"] = []string{string(p)}
	}
	code, err = c.CallWithForm(&data, host+"/buckets/space", args)
	return
}

// 参数说明
//    p  时间粒度选择，可选5min|day|month，为空默认选择day
func BucketsTransfer(c rpc.Client, host, uid, from, to string, p TimeUnit) (
	data BucketsResult, code int, err error) {
	args := map[string][]string{
		"uid":  []string{uid},
		"from": []string{from},
		"to":   []string{to},
	}
	if p != "" {
		args["p"] = []string{string(p)}
	}
	code, err = c.CallWithForm(&data, host+"/buckets/transfer", args)
	return
}

// 参数说明
//    p  时间粒度选择，可选5min|day|month，为空默认选择day
func BucketsApicall(c rpc.Client, host, uid, apitype, from, to string, p TimeUnit) (
	data BucketsResult, code int, err error) {
	args := map[string][]string{
		"uid":  []string{uid},
		"type": []string{apitype},
		"from": []string{from},
		"to":   []string{to},
	}
	if p != "" {
		args["p"] = []string{string(p)}
	}
	code, err = c.CallWithForm(&data, host+"/buckets/apicall", args)
	return
}

//---------------------------------------------------------------------------//

type ServiceIn struct {
	host string
	acc  account.InterfaceEx
}

func NewServiceIn(host string, acc account.InterfaceEx) *ServiceIn {
	return &ServiceIn{host: host, acc: acc}
}

func (r *ServiceIn) getClient(user account.UserInfo) rpc.Client {
	token := r.acc.MakeAccessToken(user)
	return rpc.Client{oauth.NewClient(token, nil)}
}

func (r *ServiceIn) Info(user account.UserInfo, bucket string, emptybucket bool, month string) (
	info StatInfo, code int, err error) {
	return Info(r.getClient(user), r.host, "", bucket, emptybucket, month)
}

func (r *ServiceIn) DailyInfo(user account.UserInfo, bucket string, emptybucket bool, month string) (
	info StatInfoDaily, code int, err error) {
	return DailyInfo(r.getClient(user), r.host, "", bucket, emptybucket, month)
}

func (r *ServiceIn) AdminInfo(user account.UserInfo, uid, bucket string, emptybucket bool, month string) (
	info StatInfo, code int, err error) {
	return Info(r.getClient(user), r.host, uid, bucket, emptybucket, month)
}

func (r *ServiceIn) AdminDailyInfo(user account.UserInfo, uid, bucket string, emptybucket bool, day string) (
	info StatInfoDaily, code int, err error) {
	return DailyInfo(r.getClient(user), r.host, uid, bucket, emptybucket, day)
}

func (r *ServiceIn) AllUse(user account.UserInfo, month string) (
	uids []uint32, code int, err error) {
	return AllUser(r.getClient(user), r.host, month)
}

func (r *ServiceIn) SelectSpace(user account.UserInfo, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectSpace(r.getClient(user), r.host, "", bucket, emptybucket, from, to, p)
}

func (r *ServiceIn) SelectSpaceAdd(user account.UserInfo, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectSpaceAdd(r.getClient(user), r.host, "", bucket, emptybucket, from, to, p)
}

func (r *ServiceIn) SelectApicall(user account.UserInfo, bucket string, emptybucket bool, apitype, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectApicall(r.getClient(user), r.host, "", bucket, emptybucket, apitype, from, to, p)
}

func (r *ServiceIn) SelectTransfer(user account.UserInfo, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectTransfer(r.getClient(user), r.host, "", bucket, emptybucket, from, to, p)
}

func (r *ServiceIn) SelectOvTransfer(user account.UserInfo, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectOvTransfer(r.getClient(user), r.host, "", bucket, emptybucket, from, to, p)
}

func (r *ServiceIn) SelectBandwidth(user account.UserInfo, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectBandwidth(r.getClient(user), r.host, "", bucket, emptybucket, from, to, p)
}

func (r *ServiceIn) SelectOvBandwidth(user account.UserInfo, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectOvBandwidth(r.getClient(user), r.host, "", bucket, emptybucket, from, to, p)
}

func (r *ServiceIn) SelectMoney(user account.UserInfo, type_, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectMoney(r.getClient(user), r.host, "", type_, from, to, p)
}

func (r *ServiceIn) BucketsSpace(user account.UserInfo, from, to string, p TimeUnit) (
	data BucketsResult, code int, err error) {
	return BucketsSpace(r.getClient(user), r.host, "", from, to, p)
}

func (r *ServiceIn) BucketsTransfer(user account.UserInfo, from, to string, p TimeUnit) (
	data BucketsResult, code int, err error) {
	return BucketsTransfer(r.getClient(user), r.host, "", from, to, p)
}

func (r *ServiceIn) BucketsApicall(user account.UserInfo, apitype, from, to string, p TimeUnit) (
	data BucketsResult, code int, err error) {
	return BucketsApicall(r.getClient(user), r.host, "", apitype, from, to, p)
}

//---------------------------------------------------------------------------//

type Service struct {
	Conn rpc.Client
}

func New(t http.RoundTripper) Service {
	client := &http.Client{Transport: t}
	return Service{rpc.Client{client}}
}

func (r Service) Info(host, bucket string, emptybucket bool, month string) (
	info StatInfo, code int, err error) {
	return Info(r.Conn, host, "", bucket, emptybucket, month)
}

func (r Service) DailyInfo(host, bucket string, emptybucket bool, day string) (
	info StatInfoDaily, code int, err error) {
	return DailyInfo(r.Conn, host, "", bucket, emptybucket, day)
}

func (r Service) SelectSpace(host, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectSpace(r.Conn, host, "", bucket, emptybucket, from, to, p)
}

func (r Service) SelectSpaceAdd(host, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectSpaceAdd(r.Conn, host, "", bucket, emptybucket, from, to, p)
}

func (r Service) SelectApicall(host string, bucket string, emptybucket bool, apitype, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectApicall(r.Conn, host, "", bucket, emptybucket, apitype, from, to, p)
}

func (r Service) SelectTransfer(host, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectTransfer(r.Conn, host, "", bucket, emptybucket, from, to, p)
}

func (r Service) SelectOvTransfer(host, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectOvTransfer(r.Conn, host, "", bucket, emptybucket, from, to, p)
}

func (r Service) SelectBandwidth(host, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectBandwidth(r.Conn, host, "", bucket, emptybucket, from, to, p)
}

func (r Service) SelectOvBandwidth(host, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectOvBandwidth(r.Conn, host, "", bucket, emptybucket, from, to, p)
}

func (r Service) SelectMoney(host, type_, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectMoney(r.Conn, host, "", type_, from, to, p)
}

func (r Service) BucketsSpace(host, uid, from, to string, p TimeUnit) (
	data BucketsResult, code int, err error) {
	return BucketsSpace(r.Conn, host, uid, from, to, p)
}

func (r Service) BucketsTransfer(host, uid, from, to string, p TimeUnit) (
	data BucketsResult, code int, err error) {
	return BucketsTransfer(r.Conn, host, uid, from, to, p)
}

func (r Service) BucketsApicall(host, uid, apitype, from, to string, p TimeUnit) (
	data BucketsResult, code int, err error) {
	return BucketsApicall(r.Conn, host, uid, apitype, from, to, p)
}

func (r Service) AdminInfo(host string, uid uint32, bucket string, emptybucket bool, month string) (
	info StatInfo, code int, err error) {
	return Info(r.Conn, host, strconv.FormatUint(uint64(uid), 10), bucket, emptybucket, month)
}

func (r Service) AdminDailyInfo(host string, uid uint32, bucket string, emptybucket bool, day string) (
	info StatInfoDaily, code int, err error) {
	return DailyInfo(r.Conn, host, strconv.FormatUint(uint64(uid), 10), bucket, emptybucket, day)
}

func (r Service) AllUser(host string, month string) (
	uids []uint32, code int, err error) {
	return AllUser(r.Conn, host, month)
}

func (r Service) AdminSelectSpace(host string, uid uint32, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectSpace(r.Conn, host, strconv.FormatUint(uint64(uid), 10), bucket, emptybucket, from, to, p)
}

func (r Service) AdminSelectSpaceAdd(host string, uid uint32, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectSpaceAdd(r.Conn, host, strconv.FormatUint(uint64(uid), 10), bucket, emptybucket, from, to, p)
}

func (r Service) AdminSelectApicall(host string, uid uint32, bucket string, emptybucket bool, apitype, from, to, p string) (data DataResult, code int, err error) {

	return SelectApicall(r.Conn, host, strconv.FormatUint(uint64(uid), 10), bucket, emptybucket, apitype, from, to, p)
}

func (r Service) AdminSelectTransfer(host string, uid uint32, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectTransfer(r.Conn, host, strconv.FormatUint(uint64(uid), 10),
		bucket, emptybucket, from, to, p)
}

func (r Service) AdminSelectOvTransfer(host string, uid uint32, bucket string, emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectOvTransfer(r.Conn, host, strconv.FormatUint(uint64(uid), 10),
		bucket, emptybucket, from, to, p)
}

func (r Service) AdminSelectBandwidth(host string, uid uint32, bucket string,
	emptybucket bool, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectBandwidth(r.Conn, host, strconv.FormatUint(uint64(uid), 10),
		bucket, emptybucket, from, to, p)
}

func (r Service) AdminSelectOvBandwidth(host string, uid uint32, bucket string, emptybucket bool, from, to, p string) (data DataResult, code int, err error) {
	return SelectOvBandwidth(r.Conn, host, strconv.FormatUint(uint64(uid),
		10), bucket, emptybucket, from, to, p)
}

func (r Service) AdminSelectMoney(host string, uid uint32, type_, from, to, p string) (
	data DataResult, code int, err error) {
	return SelectMoney(r.Conn, host, strconv.FormatUint(uint64(uid), 10), type_, from, to, p)
}

func (r Service) BandwidthAdjustment(host string, uid uint32, bucket string,
	emptybucket bool, from, to string, limit uint32) (
	data DataResult, code int, err error) {
	code, err = r.Conn.CallWithForm(&data, host+"/bandwidth/adjustment",
		map[string][]string{
			"uid":         []string{strconv.FormatUint(uint64(uid), 10)},
			"bucket":      []string{bucket},
			"emptybucket": []string{strconv.FormatBool(emptybucket)},
			"from":        []string{from},
			"to":          []string{to},
			"limit":       []string{strconv.FormatUint(uint64(limit), 10)},
		})
	return
}

func (r Service) OvBandwidthAdjustment(host string, uid uint32, bucket string,
	emptybucket bool, from, to string, limit uint32) (
	data DataResult, code int, err error) {
	code, err = r.Conn.CallWithForm(&data, host+"/ov-bandwidth/adjustment",
		map[string][]string{
			"uid":         []string{strconv.FormatUint(uint64(uid), 10)},
			"bucket":      []string{bucket},
			"emptybucket": []string{strconv.FormatBool(emptybucket)},
			"from":        []string{from},
			"to":          []string{to},
			"limit":       []string{strconv.FormatUint(uint64(limit), 10)},
		})
	return
}

func (r Service) AdminSelectMpsValue(host string, uid uint32, pipeline, mpskey string, from, to string, p TimeUnit) (
	data DataResult, code int, err error) {
	return SelectMpsValue(r.Conn, host, strconv.FormatUint(uint64(uid), 10), pipeline, mpskey, from, to, p)
}
