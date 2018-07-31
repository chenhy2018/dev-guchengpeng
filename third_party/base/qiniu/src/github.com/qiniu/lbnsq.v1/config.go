package lbnsq

import (
	"strings"
	"time"

	"github.com/nsqio/go-nsq"
	"qbox.us/errors"
)

var (
	EEmptyLookupdAddrs = errors.New("empty lookupd addrs")
	EEmptyTopicChannel = errors.New("empty topic or channel")
)

type ProducerConfig struct {
	NsqLookupdAddrs       []string `json:"nsq_lookupd_addrs"`
	RefreshNsqdIntervalMs int      `json:"refresh_nsqd_interval_ms"`
	ConcurrencyCount      int      `json:"concurrency_count"`
	ClientTimeoutMs       int      `json:"client_timeout_ms"`
	DialTimeoutMs         int      `json:"dial_timeout_ms"`
}

func MakeProducerConfig(cfg *ProducerConfig) (err error) {
	if len(cfg.NsqLookupdAddrs) < 1 {
		return EEmptyLookupdAddrs
	}
	for idx, v := range cfg.NsqLookupdAddrs {
		if !strings.HasPrefix(v, "http://") {
			cfg.NsqLookupdAddrs[idx] = "http://" + v
		}
	}
	if cfg.RefreshNsqdIntervalMs < 5000 {
		cfg.RefreshNsqdIntervalMs = 5000
	}
	if cfg.ConcurrencyCount < 1 {
		cfg.ConcurrencyCount = 1
	}
	if cfg.ClientTimeoutMs < 1 {
		cfg.ClientTimeoutMs = 1000
	}
	if cfg.DialTimeoutMs < 1 {
		cfg.DialTimeoutMs = 500
	}
	return
}

type ConsumerConfig struct {
	Topic              string   `json:"topic"`
	Channel            string   `json:"channel"`
	NsqLookupdAddrs    []string `json:"nsq_lookupd_addrs"`
	RequeueDelayMs     int      `json:"requeue_delay_ms"`
	MsgTimeoutS        int      `json:"msg_timeout_s"`
	MaxAttempts        int      `json:"max_attempts"`
	MaxInFlight        int      `json:"max_in_flight"`
	ConsumeConcurrency int      `json:"consume_concurrency"`
	DialTimeoutMs      int      `json:"dial_timeout_ms"`
	WriteTimeoutMs     int      `json:"write_timeout_ms"`
	ReadTimeoutMs      int      `json:"read_timeout_ms"`
}

func MakeConsumerConfig(cfg *ConsumerConfig) (nsqCfg *nsq.Config, err error) {
	if len(cfg.NsqLookupdAddrs) < 1 {
		return nil, EEmptyLookupdAddrs
	}
	if cfg.Topic == "" || cfg.Channel == "" {
		return nil, EEmptyTopicChannel
	}
	nsqCfg = nsq.NewConfig()
	if cfg.MsgTimeoutS > 0 {
		nsqCfg.MsgTimeout = time.Second * time.Duration(cfg.MsgTimeoutS)
	} else {
		nsqCfg.MsgTimeout = time.Second * 60
	}
	if cfg.MaxInFlight > 0 {
		nsqCfg.MaxInFlight = cfg.MaxInFlight
	} else {
		nsqCfg.MaxInFlight = 16
	}
	if cfg.MaxAttempts > 0 {
		nsqCfg.MaxAttempts = uint16(cfg.MaxAttempts)
	} else {
		nsqCfg.MaxAttempts = 65535
	}
	if cfg.RequeueDelayMs > 0 {
		nsqCfg.DefaultRequeueDelay = time.Millisecond * time.Duration(cfg.RequeueDelayMs)
	} else {
		nsqCfg.DefaultRequeueDelay = time.Second * 60
	}
	if cfg.DialTimeoutMs > 0 {
		nsqCfg.DialTimeout = time.Millisecond * time.Duration(cfg.DialTimeoutMs)
	} else {
		nsqCfg.DialTimeout = time.Second * 5
	}
	if cfg.WriteTimeoutMs > 0 {
		nsqCfg.WriteTimeout = time.Millisecond * time.Duration(cfg.WriteTimeoutMs)
	} else {
		nsqCfg.WriteTimeout = time.Second * 5
	}
	if cfg.ReadTimeoutMs > 0 {
		nsqCfg.ReadTimeout = time.Millisecond * time.Duration(cfg.ReadTimeoutMs)
	} else {
		nsqCfg.ReadTimeout = time.Second * 60
	}
	if cfg.ConsumeConcurrency < 1 {
		cfg.ConsumeConcurrency = 1
	}
	nsqCfg.MaxRequeueDelay = time.Hour
	return
}
