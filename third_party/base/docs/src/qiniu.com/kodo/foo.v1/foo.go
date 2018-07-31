package foo

import (
	"github.com/qiniu/http/rpcutil.v1"
	"qiniu.com/auth/authstub.v1"
	"qiniupkg.com/http/httputil.v2"
)

var (
	ErrEntryNotFound = httputil.NewError(612, "entry not found")
)

// ---------------------------------------------------------------------------
// POST /v1/foos - 创建foo对象

type postFoosRet struct {
	Id string `json:"id"`
}

func (p *Service) PostFoos(args *fooInfo, env *authstub.Env) (ret postFoosRet, err error) {

	ret.Id, err = genId()
	if err != nil {
		return
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.foos[ret.Id] = args
	return
}

// ---------------------------------------------------------------------------
// GET /v1/foos - 列出foo对象

type getFoosItem struct {
	Id  string `json:"id"`
	A   int    `json:"a"`
	Bar string `json:"bar"`
}

func (p *Service) GetFoos(env *authstub.Env) (items []getFoosItem, err error) {

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	items = make([]getFoosItem, 0, len(p.foos))
	for k, v := range p.foos {
		items = append(items, getFoosItem{k, v.A, v.Bar})
	}
	return
}

// ---------------------------------------------------------------------------
// DELETE /v1/foos/<FooId> - 删除foo对象

func (p *Service) DeleteFoos_(args *cmdArgs, env *authstub.Env) (err error) {

	id := args.CmdArgs[0]
	if _, ok := p.foos[id]; ok {

		p.mutex.Lock()
		defer p.mutex.Unlock()

		delete(p.foos, id)
		return
	}
	err = ErrEntryNotFound
	return
}

// ---------------------------------------------------------------------------
// GET /v1/foos/<FooId> - 获取foo对象

func (p *Service) GetFoos_(args *cmdArgs, env *authstub.Env) (ret *fooInfo, err error) {

	id := args.CmdArgs[0]
	if _, ok := p.foos[id]; ok {

		p.mutex.RLock()
		defer p.mutex.RUnlock()

		info := *p.foos[id] // 需要复制内容，否则会有竞争问题
		return &info, nil
	}
	err = ErrEntryNotFound
	return
}

// ---------------------------------------------------------------------------
// POST /v1/foos/<FooId>/bar - 修改foo的bar属性

type postFoosBarArgs struct {
	CmdArgs []string
	Bar     string `json:"val"`
}

func (p *Service) PostFoos_Bar(args *postFoosBarArgs, env *authstub.Env) (err error) {

	id := args.CmdArgs[0]
	if _, ok := p.foos[id]; ok {

		p.mutex.Lock()
		defer p.mutex.Unlock()

		p.foos[id].Bar = args.Bar
		return
	}
	err = ErrEntryNotFound
	return
}

// ---------------------------------------------------------------------------
// GET /v1/foos/<FooId>/bar - 获取foo的bar属性（可匿名获取）
// 注意: env 类型是 *rpcutil.Env，不是 *authstub.Env

type getFoosBarRet struct {
	Bar string `json:"val"`
}

func (p *Service) GetFoos_Bar(args *cmdArgs, env *rpcutil.Env) (ret getFoosBarRet, err error) {

	id := args.CmdArgs[0]
	if _, ok := p.foos[id]; ok {
	
		p.mutex.Lock()
		defer p.mutex.Unlock()

		ret.Bar = p.foos[id].Bar
		return
	}
	err = ErrEntryNotFound
	return
}

// ---------------------------------------------------------------------------

