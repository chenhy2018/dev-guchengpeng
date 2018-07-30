package now

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
)

type Message struct {
	When     int64  `bson:"when" json:"when"`
	Duration int64  `bson:"duration,omitempty" json:"duration,omitempty"`
	Hostname string `bson:"hostname,omitempty" json:"hostname,omitempty"`
	Status   Status `bson:"status" json:"status"`
}

func (m *Message) String() string {
	var str string
	for name, metric := range m.Status {
		str += fmt.Sprintf("%#v:count(%d),value(%d)  ", name, metric.Count, metric.Value)
	}
	return fmt.Sprintf("When:%d, Duration:%d, Hostname:%s, Status:(%v)",
		m.When, m.Duration, m.Hostname, str)
}

func (m *Message) IsValid() bool {
	if m.When == 0 || m.Duration == 0 || m.Hostname == "" || len(m.Status) == 0 {
		return false
	}
	return true
}

type Status map[string]*Metric

type Metric struct {
	Count uint64 `bson:"count,omitempty" json:"count,omitempty"`
	Value uint64 `bson:"value,omitempty" json:"value,omitempty"`
}

//---------------------------------------------------------
type Messages []*Message

func (msgs Messages) Len() int {
	return len(msgs)
}

func (msgs Messages) Less(i, j int) bool {
	return msgs[i].When < msgs[j].When
}

func (msgs Messages) Swap(i, j int) {
	msgs[i], msgs[j] = msgs[j], msgs[i]
}

//---------------------------------------------------------
type QueryArgs struct {
	From    int64 `json:"from"` // unix time
	To      int64 `json:"to"`
	Step    int64 `json:"step"`     // minute
	StepSec int64 `json:"step_sec"` // sec
}

type RawMessageArgs struct {
	From     int64    `json:"from"` // unix time
	To       int64    `json:"to"`
	Hostname []string `json:"hostname"`
}

type Client struct {
	Conn *lb.Client
}

func New(hosts []string, t http.RoundTripper) *Client {
	cfg := lb.Config{Hosts: hosts, TryTimes: uint32(len(hosts))}
	return &Client{Conn: lb.New(&cfg, t)}
}

func (c *Client) Report(l rpc.Logger, sid string, msg *Message) error {
	return c.Conn.CallWithJson(l, nil, "/report/"+sid, msg)
}

func (c *Client) Query(l rpc.Logger, ret interface{}, sid string, args *QueryArgs) error {
	parame := map[string][]string{
		"from":     {strconv.FormatInt(args.From, 10)},
		"to":       {strconv.FormatInt(args.To, 10)},
		"step":     {strconv.FormatInt(args.Step, 10)},
		"step_sec": {strconv.FormatInt(args.StepSec, 10)},
	}
	return c.Conn.CallWithForm(l, ret, "/query/"+sid, parame)
}

func (c *Client) RawMessage(l rpc.Logger, ret interface{}, sid string, args *RawMessageArgs) error {
	parame := map[string][]string{
		"from":     {strconv.FormatInt(args.From, 10)},
		"to":       {strconv.FormatInt(args.To, 10)},
		"hostname": args.Hostname,
	}
	return c.Conn.CallWithForm(l, ret, "/rawmessage/"+sid, parame)
}
