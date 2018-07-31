package state

import (
	"sync"
	"time"
)

type Info struct {
	Count           int64 `json:"count"`
	FailCount       int64 `json:"failcount"`
	ConsumeTime     int64 `json:"consume"`
	FailConsumeTime int64 `json:"failconsume"`
	sync.Mutex      `json:"-"`
}

func (r *Info) add(result bool, consumeTime int64) {

	r.Lock()
	defer r.Unlock()

	if result {
		r.Count++
		r.ConsumeTime += consumeTime
	} else {
		r.FailCount++
		r.FailConsumeTime += consumeTime
	}
}

func (r *Info) Copy() *Info {

	r2 := new(Info)

	r.Lock()
	defer r.Unlock()

	r2.Count = r.Count
	r2.FailCount = r.FailCount
	r2.ConsumeTime = r.ConsumeTime
	r2.FailConsumeTime = r.FailConsumeTime

	return r2
}

//---------------------------------------------------------------------------//

type Unit struct {
	begin int64
	info  *Info
}

func (r *Unit) Leave(err *error) {
	end := time.Now().UnixNano()
	if err == nil || *err == nil {
		r.info.add(true, end-r.begin)
	} else {
		r.info.add(false, end-r.begin)
	}
}

//---------------------------------------------------------------------------//

type State struct {
	states map[string]*Info
	sync.Mutex
}

var state = &State{
	states: make(map[string]*Info),
}

func Enter(name string) *Unit {
	return state.Enter(name)
}

func (r *State) Enter(name string) *Unit {

	r.Lock()
	defer r.Unlock()

	info, ok := r.states[name]
	if !ok {
		info = new(Info)
		r.states[name] = info
	}

	return &Unit{time.Now().UnixNano(), info}
}

func Dump() map[string]*Info {
	return state.Dump()
}

func (r *State) Dump() map[string]*Info {

	r.Lock()
	defer r.Unlock()

	infos := make(map[string]*Info, len(r.states))
	for key, info := range r.states {
		infos[key] = info.Copy()
	}

	return infos
}
