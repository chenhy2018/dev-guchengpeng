package cc

import (
	"github.com/qiniu/rpc.v2"
	"strconv"
)

type LoadBalancer struct {
	Id              string   `json:"id,omitempty"`
	VipNetworkId    string   `json:"vip_network_id,omitempty"` // 该负载均衡器的VIP所处网络ID
	VipSubnetId     string   `json:"vip_subnet_id,omitempty"`  // optional 该负载均衡器的VIP所处子网ID,当且仅当指定时具有返回值
	VipAddress      string   `json:"vip_address,omitempty"`    // 该负载均衡器的VIP地址
	VipPortId       string   `json:"vip_port_id,omitempty"`    // 该负载均衡器的PortID
	SecuritygroupId string   `json:"securitygroup_id,omitempty"`
	AdminStateUp    bool     `json:"admin_state_up,omitempty"` // 负载均衡器的管理状态
	Status          string   `json:"status,omitempty"`         // 该负载均衡器的状态信息
	Name            string   `json:"name,omitempty"`
	Desc            string   `json:"description,omitempty"`
	TenantId        string   `json:"tenant_id,omitempty"`
	ListenerIds     []string `json:"listener_ids,omitempty"` // 该负载均衡器下的所有监听器的ID列表
	CreatedAt       string   `json:"created_at,omitempty"`
}

type LoadBalancerWrap struct {
	LoadBalancer LoadBalancer `json:"loadbalancer"`
}

type LoadBalancersRet struct {
	LoadBalancers []LoadBalancer `json:"loadbalancers"`
}

// --------------------------------------------------
// 查看所有负载均衡器

func (r *Service) ListLoadBalancers(l rpc.Logger) (ret LoadBalancersRet, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/loadbalancers")
	return
}

// ------------------------------------------------------------
// 创建负载均衡器

type CreateLbArgs struct {
	VipNetworkId string // must
	SecGrpId     string // must
	VipSubnetId  string
	VipAddress   string
	Name         string
	Desc         string
}

func (r *Service) CreateLoadBalancer(l rpc.Logger, args CreateLbArgs) (ret LoadBalancerWrap, err error) {
	params := map[string][]string{
		"vip_network_id":   {args.VipNetworkId},
		"securitygroup_id": {args.SecGrpId},
		"vip_subnet_id":    {args.VipSubnetId},
		"vip_address":      {args.VipAddress},
		"name":             {args.Name},
		"description":      {args.Desc},
	}
	err = r.Conn.CallWithForm(l, &ret, "POST", r.Host+r.ApiPrefix+"/loadbalancers", params)
	return
}

// ------------------------------------------------------------
// 更新负载均衡器

type UpdateLbArgs struct {
	LoadbalancerId string
	SecGrpId       string // optional
	Name           string // optional
	Desc           string // optional
}

func (r *Service) UpdateLoadBalancer(l rpc.Logger, args UpdateLbArgs) (ret LoadBalancerWrap, err error) {
	params := map[string][]string{
		"securitygroup_id": {args.SecGrpId},
		"name":             {args.Name},
		"description":      {args.Desc},
	}
	err = r.Conn.CallWithForm(l, &ret, "PUT", r.Host+r.ApiPrefix+"/loadbalancers/"+args.LoadbalancerId, params)
	return
}

// ------------------------------------------------------------
// 查看负载均衡器

func (r *Service) GetLoadBalancer(l rpc.Logger, loadbalancerId string) (ret LoadBalancerWrap, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/loadbalancers/"+loadbalancerId)
	return
}

// ------------------------------------------------------------
// 删除负载均衡器

func (r *Service) DeleteLoadBalancer(l rpc.Logger, loadbalancerId string) (err error) {
	err = r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/loadbalancers/"+loadbalancerId)
	return
}

// ------------------------------------------------------------
// 监听器

type Listener struct {
	Id             string   `json:"id,omitempty"`
	LoadBalancerId string   `json:"loadbalancer_id,omitempty"`
	Protocol       string   `json:"protocol,omitempty"`         // 该监听器监听的协议，目前支持"TCP"、"HTTP"，大小写敏感
	ProtocolPort   int      `json:"protocol_port,omitempty"`    // 该监听器监听的端口号
	ConnLimit      int      `json:"connection_limit,omitempty"` // 该监听器的最大连接数限制，默认值5000. {5000,10000,20000,40000}
	DefaultPoolId  string   `json:"default_pool_id,omitempty"`  // 该监听器的默认资源池ID
	L7PolicyIds    []string `json:"l7policy_ids,omitempty"`     // 属于该监听器的7层策略ID列表
	AdminStateUp   bool     `json:"admin_state_up,omitempty"`   // 监听器的管理状态
	Status         string   `json:"status,omitempty"`           // 该监听器的状态信息
	Name           string   `json:"name,omitempty"`
	Desc           string   `json:"description,omitempty"`
	TenantId       string   `json:"tenant_id,omitempty"`
	CreatedAt      string   `json:"created_at,omitempty"`
}

type ListenerWrap struct {
	Listener Listener `json:"listener"`
}

type ListenersRet struct {
	Listeners []Listener `json:"listeners"`
}

type LbaasListeners struct {
	Listeners []Listener `json:"lbaas_listeners"`
}

// --------------------------------------------------
// 查看所有监听器

func (r *Service) ListListeners(l rpc.Logger) (ret ListenersRet, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/listeners")
	return
}

// ------------------------------------------------------------
// 查看指定负载均衡器下所有监听器

func (r *Service) ListLbListeners(l rpc.Logger, loadbalancerId string) (ret LbaasListeners, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/loadbalancers/"+loadbalancerId+"/listeners")
	return
}

// ------------------------------------------------------------
// 创建监听器

type CreateListenerArgs struct {
	LoadbalancerId string
	Protocol       string
	ProtocolPort   int
	ConnLimit      int
	DefaultPoolId  string // optional
	AdminStateUp   bool   // * 建议显式设置，因为bool型默认false *
	Name           string // optional
	Desc           string // optional
}

func (r *Service) CreateListener(l rpc.Logger, args CreateListenerArgs) (ret ListenerWrap, err error) {
	params := map[string][]string{
		"loadbalancer_id":  {args.LoadbalancerId},
		"protocol":         {args.Protocol},
		"protocol_port":    {strconv.Itoa(args.ProtocolPort)},
		"connection_limit": {strconv.Itoa(args.ConnLimit)},
		"default_pool_id":  {args.DefaultPoolId},
		"admin_state_up":   {strconv.FormatBool(args.AdminStateUp)},
		"name":             {args.Name},
		"description":      {args.Desc},
	}
	err = r.Conn.CallWithForm(l, &ret, "POST", r.Host+r.ApiPrefix+"/listeners", params)
	return
}

// ------------------------------------------------------------
// 更新监听器

type UpdateListenerArgs struct {
	ListenerId    string
	ConnLimit     int    // optional
	DefaultPoolId string // optional
	Name          string // optional
	Desc          string // optional
}

func (r *Service) UpdateListener(l rpc.Logger, args UpdateListenerArgs) (ret ListenerWrap, err error) {
	params := map[string][]string{
		"connection_limit": {strconv.Itoa(args.ConnLimit)},
		"default_pool_id":  {args.DefaultPoolId},
		"name":             {args.Name},
		"description":      {args.Desc},
	}
	err = r.Conn.CallWithForm(l, &ret, "PUT", r.Host+r.ApiPrefix+"/listeners/"+args.ListenerId, params)
	return
}

// ------------------------------------------------------------
// 启用或关闭监听器

func (r *Service) ListenerSwitch(l rpc.Logger, listenerId string, adminStateUp bool) (err error) {
	params := map[string][]string{
		"admin_state_up": {strconv.FormatBool(adminStateUp)},
	}
	err = r.Conn.CallWithForm(l, nil, "PUT", r.Host+r.ApiPrefix+"/listeners/"+listenerId+"/switch", params)
	return
}

// ------------------------------------------------------------
// 查看监听器

func (r *Service) GetListener(l rpc.Logger, listenerId string) (ret ListenerWrap, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/listeners/"+listenerId)
	return
}

// ------------------------------------------------------------
// 删除监听器

func (r *Service) DeleteListener(l rpc.Logger, listenerId string) (err error) {
	err = r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/listeners/"+listenerId)
	return
}

// ------------------------------------------------------------
// 7层策略

type L7policy struct {
	Id                   string   `json:"id,omitempty"`
	ListenerId           string   `json:"listener_id,omitempty"`             // 该7层策略所属监听器的ID
	Action               string   `json:"action,omitempty"`                  // 该7层策略的动作, REDIRECT_TO_POOL/REDIRECT_TO_URL/REJECT
	RedirectPoolId       string   `json:"redirect_pool_id,omitempty"`        // optional 该7层策略重定向到的资源池ID(动作是REDIRECT_TO_POOL时有效),没有该值则不返回
	RedirectUrl          string   `json:"redirect_url,omitempty"`            // optional 该7层策略重定向URL时的URL(动作是REDIRECT_TO_URL时有效),没有该值则不返回
	RedirectUrlCode      int      `json:"redirect_url_code,omitempty"`       // optional 该7层策略重定向URL时的返回码(动作是REDIRECT_TO_URL时有效),没有该值则不返回
	RedirectUrlDropQuery bool     `json:"redirect_url_drop_query,omitempty"` // optional 该7层策略重定向URL时是否丢弃查询字符串(动作是REDIRECT_TO_URL时有效),没有该值则不返回
	Position             int      `json:"position,omitempty"`                // 该7层策略在所属监听器的7层策略列表中的序号
	Rules                []L7rule `json:"rules,omitempty"`                   // 该7层策略施加作用所匹配的规则列表
	AdminStateUp         bool     `json:"admin_state_up,omitempty"`          // 该7层策略的管理状态
	Status               string   `json:"status,omitempty"`                  // 该7层策略的状态信息
	Name                 string   `json:"name,omitempty"`
	Desc                 string   `json:"description,omitempty"`
	TenantId             string   `json:"tenant_id,omitempty"`
	CreatedAt            string   `json:"created_at,omitempty"`
}

type L7policyWrap struct {
	L7policy L7policy `json:"l7policy"`
}

type L7policiesRet struct {
	L7policies []L7policy `json:"l7policies"`
}

type LbaasL7policies struct {
	L7policies []L7policy `json:"lbaas_l7policies"`
}

// ------------------------------------------------------------
// 查看所有7层策略

func (r *Service) ListL7policies(l rpc.Logger) (ret L7policiesRet, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/l7policies")
	return
}

// ------------------------------------------------------------
// 查看指定监听器的所有7层策略

func (r *Service) ListListenerL7policies(l rpc.Logger, listenerId string) (ret LbaasL7policies, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/listeners/"+listenerId+"/l7policies")
	return
}

// ------------------------------------------------------------
// 创建7层策略

type CreateL7policyArgs struct {
	ListenerId           string // must
	Action               string // must
	RedirectPoolId       string
	RedirectUrl          string
	RedirectUrlCode      int
	RedirectUrlDropQuery bool
	Position             int
	AdminStateUp         bool // * 建议显式设置，因为bool型默认false *
	Name                 string
	Desc                 string
}

func (r *Service) CreateL7policy(l rpc.Logger, args CreateL7policyArgs) (ret L7policyWrap, err error) {
	params := map[string][]string{
		"listener_id":             {args.ListenerId},
		"action":                  {args.Action},
		"redirect_pool_id":        {args.RedirectPoolId},
		"redirect_url":            {args.RedirectUrl},
		"redirect_url_code":       {strconv.Itoa(args.RedirectUrlCode)},
		"redirect_url_drop_query": {strconv.FormatBool(args.RedirectUrlDropQuery)},
		"position":                {strconv.Itoa(args.Position)},
		"admin_state_up":          {strconv.FormatBool(args.AdminStateUp)},
		"name":                    {args.Name},
		"description":             {args.Desc},
	}
	err = r.Conn.CallWithForm(l, &ret, "POST", r.Host+r.ApiPrefix+"/l7policies", params)
	return
}

// ------------------------------------------------------------
// 更新7层策略

type UpdateL7policyArgs struct {
	L7policyId           string
	Action               string // optional
	RedirectPoolId       string // optional
	RedirectUrl          string // optional
	RedirectUrlCode      int    // optional
	RedirectUrlDropQuery bool   // optional
	Position             int    // optional
	Name                 string // optional
	Desc                 string // optional
}

func (r *Service) UpdateL7policy(l rpc.Logger, args UpdateL7policyArgs) (ret L7policyWrap, err error) {
	params := map[string][]string{
		"action":                  {args.Action},
		"redirect_pool_id":        {args.RedirectPoolId},
		"redirect_url":            {args.RedirectUrl},
		"redirect_url_code":       {strconv.Itoa(args.RedirectUrlCode)},
		"redirect_url_drop_query": {strconv.FormatBool(args.RedirectUrlDropQuery)},
		"position":                {strconv.Itoa(args.Position)},
		"name":                    {args.Name},
		"description":             {args.Desc},
	}
	err = r.Conn.CallWithForm(l, &ret, "PUT", r.Host+r.ApiPrefix+"/l7policies/"+args.L7policyId, params)
	return
}

// ------------------------------------------------------------
// 启用或关闭7层策略

type L7policySwitchArgs struct {
	L7policyId           string
	Action               string `json:"action"`                  // optional
	RedirectPoolId       string `json:"redirect_pool_id"`        // optional
	RedirectUrl          string `json:"redirect_url"`            // optional
	RedirectUrlCode      int    `json:"redirect_url_code"`       // optional
	RedirectUrlDropQuery bool   `json:"redirect_url_drop_query"` // optional
	Position             int    `json:"position"`                // optional
	Name                 string `json:"name"`                    // optional
	Desc                 string `json:"description"`             // optional
	AdminStateUp         bool   `json:"admin_state_up"`
}

func (r *Service) L7policySwitch(l rpc.Logger, args L7policySwitchArgs) (err error) {
	params := map[string][]string{
		"action":                  {args.Action},
		"redirect_pool_id":        {args.RedirectPoolId},
		"redirect_url":            {args.RedirectUrl},
		"redirect_url_code":       {strconv.Itoa(args.RedirectUrlCode)},
		"redirect_url_drop_query": {strconv.FormatBool(args.RedirectUrlDropQuery)},
		"position":                {strconv.Itoa(args.Position)},
		"admin_state_up":          {strconv.FormatBool(args.AdminStateUp)},
		"name":                    {args.Name},
		"description":             {args.Desc},
	}
	err = r.Conn.CallWithForm(l, nil, "PUT", r.Host+r.ApiPrefix+"/l7policies/"+args.L7policyId+"/switch", params)
	return
}

// ------------------------------------------------------------
// 查看7层策略

func (r *Service) GetL7policy(l rpc.Logger, l7policyId string) (ret L7policyWrap, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/l7policies/"+l7policyId)
	return
}

// ------------------------------------------------------------
// 删除7层策略

func (r *Service) DeleteL7policy(l rpc.Logger, l7policyId string) (err error) {
	err = r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/l7policies/"+l7policyId)
	return
}

// ------------------------------------------------------------
// 7层规则

type L7rule struct {
	Id           string `json:"id,omitempty"`
	Type         string `json:"type,omitempty"`           // 该7层规则的类型  {HOST_NAME, PATH}
	CompareType  string `json:"compare_type,omitempty"`   // 该7层规则的比较类型 {REGEX, EQUALS_TO}
	Key          string `json:"key,omitempty"`            // 该7层规则匹配的值
	AdminStateUp bool   `json:"admin_state_up,omitempty"` // 7层规则的管理状态
	Status       string `json:"status,omitempty"`         // 该7层规则的状态信息
	TenantId     string `json:"tenant_id,omitempty"`
}

type L7ruleWrap struct {
	Rule L7rule `json:"rule"`
}

type L7rulesRet struct {
	Rules []L7rule `json:"rules"`
}

// ------------------------------------------------------------
// 查看指定7层策略下所有7层规则

func (r *Service) ListL7policyL7rules(l rpc.Logger, l7policyId string) (ret L7rulesRet, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/l7policies/"+l7policyId+"/l7rules")
	return
}

// ------------------------------------------------------------
// 为指定的7层策略创建7层规则

type CreateL7ruleArgs struct {
	L7policyId   string
	Type         string `json:"type"`
	CompareType  string `json:"compare_type"`
	Key          string `json:"key"`
	AdminStateUp bool   `json:"admin_state_up"` // optional
}

func (r *Service) CreateL7rule(l rpc.Logger, args CreateL7ruleArgs) (ret L7ruleWrap, err error) {
	params := map[string][]string{
		"type":           {args.Type},
		"compare_type":   {args.CompareType},
		"key":            {args.Key},
		"admin_state_up": {strconv.FormatBool(args.AdminStateUp)},
	}
	err = r.Conn.CallWithForm(l, &ret, "POST", r.Host+r.ApiPrefix+"/l7policies/"+args.L7policyId+"/l7rules", params)
	return
}

// ------------------------------------------------------------
// 更新7层规则

type UpdateL7ruleArgs struct {
	L7policyId string
	L7ruleId   string
	Key        string // optional
}

func (r *Service) UpdateL7rule(l rpc.Logger, args UpdateL7ruleArgs) (ret L7ruleWrap, err error) {
	params := map[string][]string{
		"key": {args.Key},
	}
	err = r.Conn.CallWithForm(l, &ret, "PUT", r.Host+r.ApiPrefix+"/l7policies/"+args.L7policyId+"/l7rules/"+args.L7ruleId, params)
	return
}

// ------------------------------------------------------------
// 启用或关闭7层规则

func (r *Service) L7ruleSwitch(l rpc.Logger, l7policyId, l7ruleId string, adminStateUp bool) (err error) {
	params := map[string][]string{
		"admin_state_up": {strconv.FormatBool(adminStateUp)},
	}
	err = r.Conn.CallWithForm(l, nil, "PUT", r.Host+r.ApiPrefix+"/l7policies/"+l7policyId+"/l7rules/"+l7ruleId+"/switch", params)
	return
}

// ------------------------------------------------------------
// 查看7层规则

func (r *Service) GetL7rule(l rpc.Logger, l7policyId, l7ruleId string) (ret L7ruleWrap, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/l7policies/"+l7policyId+"/l7rules/"+l7ruleId)
	return
}

// ------------------------------------------------------------
// 删除7层规则

func (r *Service) DeleteL7rule(l rpc.Logger, l7policyId, l7ruleId string) (err error) {
	err = r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/l7policies/"+l7policyId+"/l7rules/"+l7ruleId)
	return
}

// ------------------------------------------------------------
// 资源池

type HealthMonitor struct { // 资源池使用的健康检查信息
	Type          string `json:"type,omitempty"`           // 健康检查方式 TCP/HTTP
	Delay         int    `json:"delay,omitempty"`          // 间隔时间 范围[2,60]s
	Timeout       int    `json:"timeout,omitempty"`        // 超时时间 范围[5,300]s
	MaxRetries    int    `json:"max_retries,omitempty"`    // 尝试次数 范围[1,10]
	HttpMethod    string `json:"http_method,omitempty"`    // HTTP方法 GET/POST/PUT/DELETE HTTP模式下有值;TCP模式下该字段不返回
	UrlPath       string `json:"url_path,omitempty"`       // URL路径 HTTP模式下有值;TCP模式下该字段不返回
	ExpectedCodes string `json:"expected_codes,omitempty"` // 期待返回码 HTTP模式下有值;TCP模式下该字段不返回 可以是诸如以下的方式,方式1: 200 方式2: 200,202,404 方式3: 200-202
	Id            string `json:"id,omitempty"`             // 健康检查器的id，自动生成，请忽略
}

type SessionPersistence struct { // 资源池使用的会话保持信息,仅HTTP模式下可以配置
	Type       string `json:"type,omitempty"`        // 类型 APP_COOKIE/HTTP_COOKIE/SOURCE_IP
	CookieName string `json:"cookie_name,omitempty"` // COOKIE名 type为APP_COOKIE下有效
	PoolId     string `json:"pool_id,omitempty"`     // 该会话保持所属资源池，自动生成，请忽略
}

type Pool struct {
	Id                 string             `json:"id,omitempty"`
	Protocol           string             `json:"protocol,omitempty"`            // 该资源池使用的协议,目前支持TCP、HTTP
	LbAlgorithm        string             `json:"lb_algorithm,omitempty"`        // 资源调度采用的负载均衡算法, 目前支持ROUND_ROBIN、LEAST_CONNECTIONS、SOURCE_IP
	HealthMonitor      HealthMonitor      `json:"healthmonitor,omitempty"`       // optional 该资源池使用的健康检查信息，如果没有配置，则不返回
	SessionPersistence SessionPersistence `json:"session_persistence,omitempty"` // optional 该资源池使用的会话保持信息,仅HTTP模式下可以配置，如果没有配置，则不返回
	NetworkId          string             `json:"network_id,omitempty"`
	SubnetId           string             `json:"subnet_id,omitempty"` // optional 该资源池所在子网ID，当且仅当指定时有返回值
	Members            []Member           `json:"members,omitempty"`   // 该资源池包含的资源信息的列表
	AdminStateUp       bool               `json:"admin_state_up,omitempty"`
	Status             string             `json:"status,omitempty"`
	Name               string             `json:"name,omitempty"`
	Desc               string             `json:"description,omitempty"`
	TenantId           string             `json:"tenant_id,omitempty"`
	CreatedAt          string             `json:"created_at,omitempty"`
}

type PoolWrap struct {
	Pool Pool `json:"pool"`
}

type PoolsRet struct {
	Pools []Pool `json:"pools"`
}

// ------------------------------------------------------------
// 查看所有资源池

func (r *Service) GetPools(l rpc.Logger) (ret PoolsRet, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/pools")
	return
}

// ------------------------------------------------------------
// 创建资源池  *Content-Type: application/json*

type CreatePoolArgs struct {
	Protocol           string             `json:"protocol"`
	LbAlgorithm        string             `json:"lb_algorithm"`
	HealthMonitor      HealthMonitor      `json:"health_monitor"`      // optional
	SessionPersistence SessionPersistence `json:"session_persistence"` // optional
	NetworkId          string             `json:"network_id"`
	SubnetId           string             `json:"subnet_id"`      // optional
	AdminStateUp       bool               `json:"admin_state_up"` // * 建议显式设置，因为bool型默认false *
	Name               string             `json:"name"`           // optional
	Desc               string             `json:"description"`    // optional
}

func (r *Service) CreatePool(l rpc.Logger, args CreatePoolArgs) (ret PoolWrap, err error) {
	err = r.Conn.CallWithJson(l, &ret, "POST", r.Host+r.ApiPrefix+"/pools", args)
	return
}

// ------------------------------------------------------------
// 更新资源池  *Content-Type: application/json*

type UpdatePoolArgs struct {
	HealthMonitor      HealthMonitor      `json:"health_monitor"`      // optional
	SessionPersistence SessionPersistence `json:"session_persistence"` // optional
	LbAlgorithm        string             `json:"lb_algorithm"`        // optional
	Name               string             `json:"name"`                // optional
	Desc               string             `json:"description"`         // optional
}

func (r *Service) UpdatePool(l rpc.Logger, args UpdatePoolArgs, poolId string) (ret PoolWrap, err error) {
	err = r.Conn.CallWithJson(l, &ret, "PUT", r.Host+r.ApiPrefix+"/pools/"+poolId, args)
	return
}

// ------------------------------------------------------------
// 启用或关闭资源池

func (r *Service) PoolSwitch(l rpc.Logger, poolId string, adminStateUp bool) (err error) {
	params := map[string][]string{
		"admin_state_up": {strconv.FormatBool(adminStateUp)},
	}
	err = r.Conn.CallWithForm(l, nil, "PUT", r.Host+r.ApiPrefix+"/pools/"+poolId+"/switch", params)
	return
}

// ------------------------------------------------------------
// 查看资源池

func (r *Service) GetPool(l rpc.Logger, poolId string) (ret PoolWrap, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/pools/"+poolId)
	return
}

// ------------------------------------------------------------
// 删除资源池

func (r *Service) DeletePool(l rpc.Logger, poolId string) (err error) {
	err = r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/pools/"+poolId)
	return
}

// ------------------------------------------------------------
// 成员

type Member struct {
	Id           string `json:"id,omitempty"`
	SubnetId     string `json:"subnet_id,omitempty"`
	Address      string `json:"address,omitempty"`       // 对成员的IP地址
	ProtocolPort int    `json:"protocol_port,omitempty"` // 对成员的端口号
	Weight       int    `json:"weight,omitempty"`        // 该成员的权重
	AdminStateUp bool   `json:"admin_state_up,omitempty"`
	Status       string `json:"status,omitempty"`
	TenantId     string `json:"tenant_id,omitempty"`
	InstanceId   string `json:"instance_id,omitempty"` // 该成员关联虚拟机的UUID
}

type MemberWrap struct {
	Member Member `json:"member"`
}

type MembersRet struct {
	Members []Member `json:"members"`
}

// ------------------------------------------------------------
// 查看指定资源池下所有成员

func (r *Service) GetPoolMembers(l rpc.Logger, poolId string) (ret MembersRet, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/pools/"+poolId+"/members")
	return
}

// ------------------------------------------------------------
// 创建成员

type CreateMemberArgs struct {
	PoolId       string
	SubnetId     string
	Address      string
	ProtocolPort int
	InstanceId   string
	Weight       int  // optional
	AdminStateUp bool // * 建议显式设置，因为bool型默认false *
}

func (r *Service) CreateMember(l rpc.Logger, args CreateMemberArgs) (ret MemberWrap, err error) {
	params := map[string][]string{
		"subnet_id":      {args.SubnetId},
		"address":        {args.Address},
		"protocol_port":  {strconv.Itoa(args.ProtocolPort)},
		"instance_id":    {args.InstanceId},
		"weight":         {strconv.Itoa(args.Weight)},
		"admin_state_up": {strconv.FormatBool(args.AdminStateUp)},
	}
	err = r.Conn.CallWithForm(l, &ret, "POST", r.Host+r.ApiPrefix+"/pools/"+args.PoolId+"/members", params)
	return
}

// ------------------------------------------------------------
// 更新成员

type UpdateMemberArgs struct {
	PoolId   string
	MemberId string
	Weight   int // optional
}

func (r *Service) UpdateMember(l rpc.Logger, args UpdateMemberArgs) (ret MemberWrap, err error) {
	params := map[string][]string{
		"weight": {strconv.Itoa(args.Weight)},
	}
	err = r.Conn.CallWithForm(l, &ret, "PUT", r.Host+r.ApiPrefix+"/pools/"+args.PoolId+"/members/"+args.MemberId, params)
	return
}

// ------------------------------------------------------------
// 启用或关闭成员

func (r *Service) MemberSwitch(l rpc.Logger, poolId, memberId string, adminStateUp bool) (err error) {
	params := map[string][]string{
		"admin_state_up": {strconv.FormatBool(adminStateUp)},
	}
	err = r.Conn.CallWithForm(l, nil, "PUT", r.Host+r.ApiPrefix+"/pools/"+poolId+"/members/"+memberId+"/switch", params)
	return
}

// ------------------------------------------------------------
// 查看成员

func (r *Service) GetMember(l rpc.Logger, poolId, memberId string) (ret MemberWrap, err error) {
	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/pools/"+poolId+"/members/"+memberId)
	return
}

// ------------------------------------------------------------
// 删除成员

func (r *Service) DeleteMember(l rpc.Logger, poolId, memberId string) (err error) {
	err = r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/pools/"+poolId+"/members/"+memberId)
	return
}
