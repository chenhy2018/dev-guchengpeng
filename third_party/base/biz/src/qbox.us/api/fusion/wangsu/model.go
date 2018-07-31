package wangsu

type Domain struct {
	DomainName     string `json:"domainName"`
	OriginIps      string `json:"originIps"`
	Cname          string `json:"cname"`
	Status         string `json:"status"`
	ConfigFormName string `json:"configFormName"`
	ServiceType    string `json:"serviceType"`
	ServiceFormId  string `json:"serviceFormId"`
	CreateTime     string `json:"createTime"`
	Operator       string `json:"operator"`
	CustomerName   string `json:"customerName"`
	Email          string `json:"email"`
	DetectUrl      string `json:"detectUrl"`
	CanChangeSrc   string `json:"canChangeSrc"`
	Version        string `json:"version"`
}
