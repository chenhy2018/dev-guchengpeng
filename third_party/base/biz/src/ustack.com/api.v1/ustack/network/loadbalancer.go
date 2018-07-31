package network

import (
	"strings"

	"github.com/qiniu/rpc.v2"
)

// --------------------------------------------------
// 查看所有负载均衡器

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

func (p Client) ListLoadBalancers(l rpc.Logger) (ret LoadBalancersRet, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/lbaas/loadbalancers")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 创建负载均衡器

type CreateLoadBalancerArgs struct {
	LoadBalancer struct {
		VipNetworkId    string `json:"vip_network_id,omitempty"`   // must 该负载均衡器所处网络ID
		SecuritygroupId string `json:"securitygroup_id,omitempty"` // must 该负载均衡器所属安全组ID
		VipSubnetId     string `json:"vip_subnet_id,omitempty"`    // optional 该负载均衡器所处子网ID
		VipAddress      string `json:"vip_address,omitempty"`      // optional 该负载均衡器的VIP地址，如不指定则自动分配
		Name            string `json:"name,omitempty"`             // optional
		Desc            string `json:"description,omitempty"`      // optional
	} `json:"loadbalancer"`
}

func (p Client) CreateLoadBalancer(l rpc.Logger, args CreateLoadBalancerArgs) (ret LoadBalancerWrap, err error) {
	err = p.Conn.CallWithJson(l, &ret, "POST", "/v2.0/lbaas/loadbalancers", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 更新负载均衡器

type UpdateLoadBalancerArgs struct {
	LoadBalancer struct {
		SecuritygroupId string `json:"securitygroup_id,omitempty"` // optional
		Name            string `json:"name,omitempty"`             // optional
		Desc            string `json:"description,omitempty"`      // optional
	} `json:"loadbalancer"`
}

func (p Client) UpdateLoadBalancer(l rpc.Logger, args UpdateLoadBalancerArgs, loadbalancerId string) (ret LoadBalancerWrap, err error) {
	err = p.Conn.CallWithJson(l, &ret, "PUT", "/v2.0/lbaas/loadbalancers/"+loadbalancerId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看负载均衡器

func (p Client) LoadBalancerInfo(l rpc.Logger, loadbalancerId string) (ret LoadBalancerWrap, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/lbaas/loadbalancers/"+loadbalancerId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 删除负载均衡器

func (p Client) DeleteLoadBalancer(l rpc.Logger, loadbalancerId string) (err error) {
	err = p.Conn.Call(l, nil, "DELETE", "/v2.0/lbaas/loadbalancers/"+loadbalancerId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看所有监听器

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

func (p Client) ListListeners(l rpc.Logger) (ret ListenersRet, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/lbaas/listeners")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看指定负载均衡器下所有监听器

type LbaasListeners struct {
	Listeners []Listener `json:"lbaas_listeners"`
}

func (p Client) GetLbListeners(l rpc.Logger, loadbalancerId string) (ret LbaasListeners, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/lbaas/loadbalancers/"+loadbalancerId+"/lbaas_listeners")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 创建监听器

type CreateListenerArgs struct {
	Listener struct {
		LoadbalancerId string `json:"loadbalancer_id,omitempty"`  // must
		Protocol       string `json:"protocol,omitempty"`         // must
		ProtocolPort   int    `json:"protocol_port,omitempty"`    // must
		ConnLimit      int    `json:"connection_limit,omitempty"` // must
		DefaultPoolId  string `json:"default_pool_id,omitempty"`  // optional
		AdminStateUp   bool   `json:"admin_state_up,omitempty"`   // optional
		Name           string `json:"name,omitempty"`             // optional
		Desc           string `json:"description,omitempty"`      // optional
	} `json:"listener"`
}

func (p Client) CreateListener(l rpc.Logger, args CreateListenerArgs) (ret ListenerWrap, err error) {
	err = p.Conn.CallWithJson(l, &ret, "POST", "/v2.0/lbaas/listeners", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 更新监听器

type UpdateListenerArgs struct {
	Listener struct {
		ConnLimit     int    `json:"connection_limit,omitempty"` // optional
		DefaultPoolId string `json:"default_pool_id,omitempty"`  // optional
		Name          string `json:"name,omitempty"`             // optional
		Desc          string `json:"description,omitempty"`      // optional
	} `json:"listener"`
}

func (p Client) UpdateListener(l rpc.Logger, args UpdateListenerArgs, listenerId string) (ret ListenerWrap, err error) {
	err = p.Conn.CallWithJson(l, &ret, "PUT", "/v2.0/lbaas/listeners/"+listenerId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看监听器

func (p Client) ListenerInfo(l rpc.Logger, listenerId string) (ret ListenerWrap, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/lbaas/listeners/"+listenerId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 删除监听器

func (p Client) DeleteListener(l rpc.Logger, listenerId string) (err error) {
	err = p.Conn.Call(l, nil, "DELETE", "/v2.0/lbaas/listeners/"+listenerId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 启用或关闭监听器

type listenerSwitch struct {
	Listener struct {
		AdminStateUp bool `json:"admin_state_up"`
	} `json:"listener"`
}

func (p Client) ListenerSwitch(l rpc.Logger, listenerId string, adminStateUp bool) (err error) {
	var listener listenerSwitch
	listener.Listener.AdminStateUp = adminStateUp
	err = p.Conn.CallWithJson(l, nil, "PUT", "/v2.0/lbaas/listeners/"+listenerId, listener)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看所有7层策略

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
	AdminStateUp         bool     `json:"admin_state_up"`                    // 该7层策略的管理状态
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

func (p Client) ListL7policies(l rpc.Logger) (ret L7policiesRet, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/lbaas/l7policies")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看指定监听器的所有7层策略

type LbaasL7policies struct {
	L7policies []L7policy `json:"lbaas_l7policies"`
}

func (p Client) GetListenerL7policies(l rpc.Logger, listenerId string) (ret LbaasL7policies, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/lbaas/listeners/"+listenerId+"/lbaas_l7policies")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 创建7层策略

type CreateL7policyArgs struct {
	L7policy struct {
		ListenerId           string `json:"listener_id,omitempty"`
		Action               string `json:"action,omitempty"`
		RedirectPoolId       string `json:"redirect_pool_id,omitempty"`        // optional
		RedirectUrl          string `json:"redirect_url,omitempty"`            // optional
		RedirectUrlCode      int    `json:"redirect_url_code,omitempty"`       // optional
		RedirectUrlDropQuery bool   `json:"redirect_url_drop_query,omitempty"` // optional
		Position             int    `json:"position,omitempty"`                // optional
		//	AdminStateUp         bool   `json:"admin_state_up,omitempty"`          // optional
		Name string `json:"name,omitempty"`        // optional
		Desc string `json:"description,omitempty"` // optional
	} `json:"l7policy"`
}

func (args CreateL7policyArgs) toJsonMap() (ret map[string]interface{}) {
	params := make(map[string]interface{})
	policy := args.L7policy
	switch strings.ToLower(policy.Action) {
	case "redirect_to_pool":
		params["action"] = "REDIRECT_TO_POOL"
		params["redirect_pool_id"] = policy.RedirectPoolId
	case "redirect_to_url":
		params["action"] = "REDIRECT_TO_URL"
		params["redirect_url"] = policy.RedirectUrl
		params["redirect_url_code"] = policy.RedirectUrlCode
		params["redirect_url_drop_query"] = policy.RedirectUrlDropQuery
	case "reject":
		params["action"] = "REJECT"
	}
	params["listener_id"] = policy.ListenerId
	if policy.Position != 0 {
		params["position"] = policy.Position
	}
	if policy.Name != "" {
		params["name"] = policy.Name
	}
	if policy.Desc != "" {
		params["description"] = policy.Desc
	}

	ret = make(map[string]interface{})
	ret["l7policy"] = params
	return
}

func (p Client) CreateL7policy(l rpc.Logger, args CreateL7policyArgs) (ret L7policyWrap, err error) {
	err = p.Conn.CallWithJson(l, &ret, "POST", "/v2.0/lbaas/l7policies.json", args.toJsonMap())
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 更新7层策略

type UpdateL7policyArgs struct {
	L7policy struct {
		Action               string `json:"action,omitempty"`
		RedirectPoolId       string `json:"redirect_pool_id,omitempty"`
		RedirectUrl          string `json:"redirect_url,omitempty"`
		RedirectUrlCode      int    `json:"redirect_url_code,omitempty"`
		RedirectUrlDropQuery bool   `json:"redirect_url_drop_query"`
		Position             int    `json:"position,omitempty"`
		Name                 string `json:"name,omitempty"`
		Desc                 string `json:"description,omitempty"`
	} `json:"l7policy"`
}

func (args UpdateL7policyArgs) toJsonMap() (ret map[string]interface{}) {
	params := make(map[string]interface{})
	policy := args.L7policy
	switch strings.ToLower(policy.Action) {
	case "redirect_to_pool":
		params["action"] = "REDIRECT_TO_POOL"
		params["redirect_pool_id"] = policy.RedirectPoolId
	case "redirect_to_url":
		params["action"] = "REDIRECT_TO_URL"
		params["redirect_url"] = policy.RedirectUrl
		params["redirect_url_code"] = policy.RedirectUrlCode
		params["redirect_url_drop_query"] = policy.RedirectUrlDropQuery
	case "reject":
		params["action"] = "REJECT"
	}
	if policy.Position != 0 {
		params["position"] = policy.Position
	}
	if policy.Name != "" {
		params["name"] = policy.Name
	}
	if policy.Desc != "" {
		params["description"] = policy.Desc
	}

	ret = make(map[string]interface{})
	ret["l7policy"] = params
	return
}

func (p Client) UpdateL7policy(l rpc.Logger, args UpdateL7policyArgs, l7policyId string) (ret L7policyWrap, err error) {
	err = p.Conn.CallWithJson(l, &ret, "PUT", "/v2.0/lbaas/l7policies/"+l7policyId, args.toJsonMap())
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看7层策略

func (p Client) L7policyInfo(l rpc.Logger, l7policyId string) (ret L7policyWrap, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/lbaas/l7policies/"+l7policyId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 删除7层策略

func (p Client) DeleteL7policy(l rpc.Logger, l7policyId string) (err error) {
	err = p.Conn.Call(l, nil, "DELETE", "/v2.0/lbaas/l7policies/"+l7policyId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 启用或关闭7层策略

type L7policySwitchArgs struct {
	L7policy struct {
		Action               string `json:"action,omitempty"`
		RedirectPoolId       string `json:"redirect_pool_id,omitempty"`
		RedirectUrl          string `json:"redirect_url,omitempty"`
		RedirectUrlCode      int    `json:"redirect_url_code,omitempty"`
		RedirectUrlDropQuery bool   `json:"redirect_url_drop_query"`
		Position             int    `json:"position,omitempty"`
		Name                 string `json:"name,omitempty"`
		Desc                 string `json:"description,omitempty"`
		AdminStateUp         bool   `json:"admin_state_up"`
	} `json:"l7policy"`
}

func (args L7policySwitchArgs) toJsonMap() (ret map[string]interface{}) {
	params := make(map[string]interface{})
	policy := args.L7policy
	switch strings.ToLower(policy.Action) {
	case "redirect_to_pool":
		params["action"] = "REDIRECT_TO_POOL"
		params["redirect_pool_id"] = policy.RedirectPoolId
	case "redirect_to_url":
		params["action"] = "REDIRECT_TO_URL"
		params["redirect_url"] = policy.RedirectUrl
		params["redirect_url_code"] = policy.RedirectUrlCode
		params["redirect_url_drop_query"] = policy.RedirectUrlDropQuery
	case "reject":
		params["action"] = "REJECT"
	}
	if policy.Position != 0 {
		params["position"] = policy.Position
	}
	params["admin_state_up"] = policy.AdminStateUp
	if policy.Name != "" {
		params["name"] = policy.Name
	}
	if policy.Desc != "" {
		params["description"] = policy.Desc
	}

	ret = make(map[string]interface{})
	ret["l7policy"] = params
	return
}

func (p Client) L7policySwitch(l rpc.Logger, args L7policySwitchArgs, l7policyId string) (err error) {
	err = p.Conn.CallWithJson(l, nil, "PUT", "/v2.0/lbaas/l7policies/"+l7policyId, args.toJsonMap())
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看指定7层策略下所有7层规则

type L7rule struct {
	Id           string `json:"id,omitempty"`
	Type         string `json:"type,omitempty"`         // 该7层规则的类型  {HOST_NAME, PATH}
	CompareType  string `json:"compare_type,omitempty"` // 该7层规则的比较类型 {REGEX, EQUALS_TO}
	Key          string `json:"key,omitempty"`          // 该7层规则匹配的值
	AdminStateUp bool   `json:"admin_state_up"`         // 7层规则的管理状态
	Status       string `json:"status,omitempty"`       // 该7层规则的状态信息
	TenantId     string `json:"tenant_id,omitempty"`
}

type L7ruleWrap struct {
	Rule L7rule `json:"rule"`
}

type L7rulesRet struct {
	Rules []L7rule `json:"rules"`
}

func (p Client) ListL7policyL7rules(l rpc.Logger, l7policyId string) (ret L7rulesRet, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/lbaas/l7policies/"+l7policyId+"/rules")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 为指定的7层策略创建7层规则

type CreateL7ruleArgs struct {
	Rule struct {
		Type         string `json:"type,omitempty"`
		CompareType  string `json:"compare_type,omitempty"`
		Key          string `json:"key,omitempty"`
		AdminStateUp bool   `json:"admin_state_up,omitempty"` // optional
	} `json:"rule"`
}

func (p Client) CreateL7rule(l rpc.Logger, args CreateL7ruleArgs, l7policyId string) (ret L7ruleWrap, err error) {
	err = p.Conn.CallWithJson(l, &ret, "POST", "/v2.0/lbaas/l7policies/"+l7policyId+"/rules", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 更新7层规则

type UpdateL7ruleArgs struct {
	Rule struct {
		Key string `json:"key,omitempty"` // optional
	} `json:"rule"`
}

func (p Client) UpdateL7rule(l rpc.Logger, args UpdateL7ruleArgs, l7policyId, l7ruleId string) (ret L7ruleWrap, err error) {
	err = p.Conn.CallWithJson(l, &ret, "PUT", "/v2.0/lbaas/l7policies/"+l7policyId+"/rules/"+l7ruleId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看7层规则

func (p Client) L7ruleInfo(l rpc.Logger, l7policyId, l7ruleId string) (ret L7ruleWrap, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/lbaas/l7policies/"+l7policyId+"/rules/"+l7ruleId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 删除7层规则

func (p Client) DeleteL7rule(l rpc.Logger, l7policyId, l7ruleId string) (err error) {
	err = p.Conn.Call(l, nil, "DELETE", "/v2.0/lbaas/l7policies/"+l7policyId+"/rules/"+l7ruleId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 启用或关闭7层规则

type L7ruleSwitchArgs struct {
	Rule struct {
		AdminStateUp bool `json:"admin_state_up"`
	} `json:"rule"`
}

func (p Client) L7ruleSwitch(l rpc.Logger, adminStateUp bool, l7policyId, l7ruleId string) (err error) {
	var args L7ruleSwitchArgs
	args.Rule.AdminStateUp = adminStateUp
	err = p.Conn.CallWithJson(l, nil, "PUT", "/v2.0/lbaas/l7policies/"+l7policyId+"/rules/"+l7ruleId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看所有资源池

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

func (p Client) ListPools(l rpc.Logger) (ret PoolsRet, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/lbaas/pools")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 创建资源池

type CreatePoolArgs struct {
	Pool struct {
		Protocol           string             `json:"protocol,omitempty"`
		LbAlgorithm        string             `json:"lb_algorithm,omitempty"`
		HealthMonitor      HealthMonitor      `json:"healthmonitor,omitempty"`       // optional
		SessionPersistence SessionPersistence `json:"session_persistence,omitempty"` // optional
		NetworkId          string             `json:"network_id,omitempty"`
		SubnetId           string             `json:"subnet_id,omitempty"`      // optional
		AdminStateUp       bool               `json:"admin_state_up,omitempty"` // optional
		Name               string             `json:"name,omitempty"`           // optional
		Desc               string             `json:"description,omitempty"`    // optional
	} `json:"pool"`
}

func (p Client) CreatePool(l rpc.Logger, args CreatePoolArgs) (ret PoolWrap, err error) {
	err = p.Conn.CallWithJson(l, &ret, "POST", "/v2.0/lbaas/pools", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 更新资源池

type UpdatePoolArgs struct {
	Pool struct {
		HealthMonitor      HealthMonitor      `json:"healthmonitor,omitempty"`       // optional
		SessionPersistence SessionPersistence `json:"session_persistence,omitempty"` // optional
		LbAlgorithm        string             `json:"lb_algorithm,omitempty"`        // optional
		Name               string             `json:"name,omitempty"`                // optional
		Desc               string             `json:"description,omitempty"`         // optional
	} `json:"pool"`
}

func (p Client) UpdatePool(l rpc.Logger, args UpdatePoolArgs, poolId string) (ret PoolWrap, err error) {
	err = p.Conn.CallWithJson(l, &ret, "PUT", "/v2.0/lbaas/pools/"+poolId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看资源池

func (p Client) PoolInfo(l rpc.Logger, poolId string) (ret PoolWrap, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/lbaas/pools/"+poolId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 删除资源池

func (p Client) DeletePool(l rpc.Logger, poolId string) (err error) {
	err = p.Conn.Call(l, nil, "DELETE", "/v2.0/lbaas/pools/"+poolId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 启用或关闭资源池

type poolSwitchArgs struct {
	Pool struct {
		AdminStateUp bool `json:"admin_state_up"`
	} `json:"pool"`
}

func (p Client) PoolSwitch(l rpc.Logger, poolId string, adminStateUp bool) (err error) {
	var args poolSwitchArgs
	args.Pool.AdminStateUp = adminStateUp
	err = p.Conn.CallWithJson(l, nil, "PUT", "/v2.0/lbaas/pools/"+poolId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看指定资源池下所有成员

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

func (p Client) ListPoolMembers(l rpc.Logger, poolId string) (ret MembersRet, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/lbaas/pools/"+poolId+"/members")
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 创建成员

type CreateMemberArgs struct {
	Member struct {
		SubnetId     string `json:"subnet_id,omitempty"`      // 成员所在的子网ID,注意子网ID只是作为成员的信息,不具有“联动”作用
		Address      string `json:"address,omitempty"`        // 对成员的IP地址,注意IP地址只是作为成员的信息,不具有“联动”作用
		ProtocolPort int    `json:"protocol_port,omitempty"`  // 对成员的端口号
		InstanceId   string `json:"instance_id,omitempty"`    // 该成员关联虚拟机的UUID,,注意虚拟机UUID只是作为成员的信息,不具有“联动”作用
		Weight       int    `json:"weight,omitempty"`         // optional 该成员的权重
		AdminStateUp bool   `json:"admin_state_up,omitempty"` // optional 该成员的管理状态
	} `json:"member"`
}

func (p Client) CreateMember(l rpc.Logger, args CreateMemberArgs, poolId string) (ret MemberWrap, err error) {
	err = p.Conn.CallWithJson(l, &ret, "POST", "/v2.0/lbaas/pools/"+poolId+"/members", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 更新成员

type UpdateMemberArgs struct {
	Member struct {
		Weight int `json:"weight,omitempty"` // optional
	} `json:"member"`
}

func (p Client) UpdateMember(l rpc.Logger, args UpdateMemberArgs, poolId, memberId string) (ret MemberWrap, err error) {
	err = p.Conn.CallWithJson(l, &ret, "PUT", "/v2.0/lbaas/pools/"+poolId+"/members/"+memberId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查看成员

func (p Client) MemberInfo(l rpc.Logger, poolId, memberId string) (ret MemberWrap, err error) {
	err = p.Conn.Call(l, &ret, "GET", "/v2.0/lbaas/pools/"+poolId+"/members/"+memberId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 删除成员

func (p Client) DeleteMember(l rpc.Logger, poolId, memberId string) (err error) {
	err = p.Conn.Call(l, nil, "DELETE", "/v2.0/lbaas/pools/"+poolId+"/members/"+memberId)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 启用或关闭成员

type MemberSwtichArgs struct {
	Member struct {
		AdminStateUp bool `json:"admin_state_up"`
	} `json:"member"`
}

func (p Client) MemberSwitch(l rpc.Logger, poolId, memberId string, adminStateUp bool) (err error) {
	var args MemberSwtichArgs
	args.Member.AdminStateUp = adminStateUp
	err = p.Conn.CallWithJson(l, nil, "PUT", "/v2.0/lbaas/pools/"+poolId+"/members/"+memberId, args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}
