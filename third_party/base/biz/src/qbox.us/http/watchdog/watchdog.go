package watchdog

import (
	"strconv"
	"sync"
	"time"

	"github.com/qiniu/http/httputil.v1"
	clog "github.com/qiniu/log.v1"
)

type UidApiLimiter struct {
	ApiName            string `json:"api_name"`
	CurThresholdNum    int64  `json:"cur_threshold_num"`
	TimeInterval       int64  `json:"time_interval"` //unit: ms
	PeriodThresholdNum int64  `json:"period_threshold_num"`
}

type UidApiLoad struct {
	CurReqNum    int64
	PeriodReqNum int64
	LastStatTime int64
}

type WatchDog struct {
	mutex              sync.Mutex
	AppCurThresholdNum int64
	AppCurReqNum       int64

	MapUidApiLimiter map[string]UidApiLimiter
	MapApiLoads      map[string]*UidApiLoad
}

type Config struct {
	AppCurThresholdNum int64           `json:"app_cur_threshold_num"`
	UidApiLimiters     []UidApiLimiter `json:"uid_api_limit_cfgs"`
}

func Open(cfg Config) (dog *WatchDog, err error) {

	clog.Infof("watchdog cfg:%v", cfg)

	mapUidApiLimiter := make(map[string]UidApiLimiter, 0)
	for _, limiter := range cfg.UidApiLimiters {
		mapUidApiLimiter[limiter.ApiName] = limiter
	}

	return &WatchDog{
		AppCurThresholdNum: cfg.AppCurThresholdNum,
		AppCurReqNum:       0,
		MapUidApiLimiter:   mapUidApiLimiter,
		MapApiLoads:        make(map[string]*UidApiLoad, 0),
	}, nil
}

func (this *WatchDog) DecreCurLoad(Uid uint32, apiName string) {

	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.AppCurReqNum--

	if Uid == 0 || apiName == "" {
		return
	}

	key := apiName + "-" + strconv.FormatUint(uint64(Uid), 10)
	curApiLoad, ok := this.MapApiLoads[key]
	if ok {
		curApiLoad.CurReqNum--
	}
}

func (this *WatchDog) getUidApiLoad(Uid uint32, apiName string, curTime int64) (load *UidApiLoad, limiter UidApiLimiter) {

	//if Uid or apiName unknown, not check and just return:
	if Uid == 0 || apiName == "" {
		return
	}

	uaLimiter, ok := this.MapUidApiLimiter[apiName]
	if !ok {
		clog.Debugf("not find api:%v's limiter cfg, so just do this req.", apiName)
		return
	}

	key := apiName + "-" + strconv.FormatUint(uint64(Uid), 10)
	curApiLoad, ok := this.MapApiLoads[key]
	if !ok {
		apiLoad := UidApiLoad{
			LastStatTime: curTime,
			CurReqNum:    0,
		}
		curApiLoad = &apiLoad
		this.MapApiLoads[key] = curApiLoad
	}

	return curApiLoad, uaLimiter
}

func (this *WatchDog) Check(Uid uint32, apiName string) (code int) {

	clog.Debugf("in watchdog func. Uid:%v apiName:%v", Uid, apiName)

	//svr level limit:
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.AppCurReqNum += 1

	//先找到uidApi级别load,将load值+1再判断是否超载,否则后面的DecreLoad会导致uidApi级别load值为负数
	curTime := time.Now().UnixNano() / 1e6
	curApiLoad, uaLimiter := this.getUidApiLoad(Uid, apiName, curTime)
	if curApiLoad != nil {
		curApiLoad.CurReqNum += 1
		curApiLoad.PeriodReqNum += 1
	}

	clog.Debugf("app CurReqNum:%v AppCurThresholdNum:%v", this.AppCurReqNum, this.AppCurThresholdNum)
	if this.AppCurReqNum > this.AppCurThresholdNum {
		clog.Warnf("svr level overload. this.AppCurLoad:%v > this.AppCurThresholdNum:%v", this.AppCurReqNum, this.AppCurThresholdNum)
		return httputil.StatusOverload
	}

	if curApiLoad == nil {
		return 200
	}

	clog.Debugf("apiname Uid CurReqNum:%v PeriodReqNum:%v uaLimiter.CurThresholdNum:%v uaLimiter.PeriodThresholdNum:%v",
		curApiLoad.CurReqNum, curApiLoad.PeriodReqNum, uaLimiter.CurThresholdNum, uaLimiter.PeriodThresholdNum)

	//当前请求数的检查:
	if curApiLoad.CurReqNum > uaLimiter.CurThresholdNum {
		clog.Warnf("apiName and uid level curload limit. cur api load:%v > limit load:%v", curApiLoad.CurReqNum, uaLimiter.CurThresholdNum)
		return httputil.StatusOverload
	}

	//过去这段时间的请求数检查:
	if curTime-curApiLoad.LastStatTime <= uaLimiter.TimeInterval {
		if curApiLoad.PeriodReqNum > uaLimiter.PeriodThresholdNum {
			//deny
			clog.Warnf("apiName and Uid level period load limit. period load:%v > limit load:%v", curApiLoad.PeriodReqNum, uaLimiter.PeriodThresholdNum)
			return httputil.StatusOverload
		}
	} else {
		curApiLoad.LastStatTime = curTime
		curApiLoad.PeriodReqNum = 1
	}
	return 200
}
