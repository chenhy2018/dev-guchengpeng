package mqrunner

import (
	"net/http"
	"net/url"
	"time"

	"github.com/qiniu/errors"
	"github.com/qiniu/xlog.v1"

	"qbox.us/qmq/qmqapi/mq2"
)

var ErrRetry = errors.New("retry")

// -----------------------------------------------------------

type msgItem struct {
	idx   int
	msgId string
	msg   []byte
}

type Service struct {
	MqHosts   []string
	MqId      string
	HandleMsg func(xl *xlog.Logger, query url.Values, err error) error // 如果返回 ErrRetry 则需要重试
	Transport http.RoundTripper
}

func msgFetcher(msgQ chan *msgItem, mq *mq2.Client, idx int, mqId string, lock *bool) {

	for {
		if *lock {
			break
		}

		xl := xlog.NewDummy()

		msg, msgId, err := mq.Get(xl, idx, mqId)
		if err != nil {
			xl.Info("mq.Get failed:", errors.Detail(err))
			time.Sleep(1e9)
			continue
		}

		msgQ <- &msgItem{idx, msgId, msg}
	}
}

func (p *Service) Run(done chan bool) (err error) {

	if len(p.MqHosts) != 2 {
		panic("len(MqHosts) must == 2")
	}

	if done == nil {
		done = make(chan bool)
	}

	mq := mq2.New(p.MqHosts, p.Transport)
	mqId, handleMsg := p.MqId, p.HandleMsg
	msgQ := make(chan *msgItem)
	lock := false

	go msgFetcher(msgQ, mq, 0, mqId, &lock)
	go msgFetcher(msgQ, mq, 1, mqId, &lock)

	for {
		var item *msgItem

		if lock {
			select {
			case item = <-msgQ:
			default:
				return nil
			}
		} else {
			select {
			case <-done:
				lock = true
				if len(msgQ) == 0 {
					time.Sleep(5e8)
				}
				continue
			case item = <-msgQ:
			}
		}

		xl := xlog.NewDummy()
		query, err := url.ParseQuery(string(item.msg))
		if err != nil {
			xl.Info("url.ParseQuery failed:", errors.Detail(err))
			//continue -- 需要让 handleMsg 来处理这个错误
		}

		err = handleMsg(xl, query, err)
		if err != nil {
			if err == ErrRetry {
				continue
			}
			xl.Info("handleMsg failed:", query, errors.Detail(err))
			//continue -- 只要不是ErrRetry，那么就认为这个错误不是可重试的
		}

		err = mq.Delete(xl, item.idx, mqId, item.msgId)
		if err != nil {
			xl.Info("mq.Delete failed:", item.idx, mqId, item.msgId, errors.Detail(err))
			continue
		}
	}
	return nil
}

// -----------------------------------------------------------
