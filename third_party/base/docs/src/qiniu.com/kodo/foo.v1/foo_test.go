package foo

import (
	"testing"

	"github.com/qiniu/http/restrpc.v1"
	"qiniupkg.com/qiniutest/httptest.v1"
	"qiniupkg.com/x/mockhttp.v7"
)

func TestService(t *testing.T) {

	transport := mockhttp.NewTransport()

	cfg := &Config{}
	svr, err := New(cfg)
	if err != nil {
		t.Fatal("New service failed:", err)
	}

	router := restrpc.Router{
		PatternPrefix: "/v1",
		Mux: restrpc.NewServeMux(),
	}
	transport.ListenAndServe("foo.com", router.Register(svr))

	ctx := httptest.New(t)
	ctx.SetTransport(transport)

	ctx.Exec(`

	# 构建一个测试用的授权对象（auth object），取名 qiniutest
	#
	auth qiniutest |authstub -uid 1 -utype 4|

	# 测试匿名创建对象，预期会失败(bad token)
	#
	post http://foo.com/v1/foos
	json '{
		"a": 1,
		"bar": "hello, world!"
	}'
	ret 401
	json '{
		"error": "bad token"
	}'

	# 测试创建对象
	#
	post http://foo.com/v1/foos
	auth qiniutest
	json '{
		"a": 1,
		"bar": "hello, world!"
	}'
	ret 200
	json '{
		"id": $(id1)
	}'

	# 测试匿名获取对象，预期会失败(bad token)
	#
	get http://foo.com/v1/foos/$(id1)
	ret 401
	json '{
		"error": "bad token"
	}'

	# 测试获取对象
	#
	get http://foo.com/v1/foos/$(id1)
	auth qiniutest
	ret 200
	json '{
		"a": 1,
		"bar": "hello, world!"
	}'

	# 测试匿名修改对象属性，预期会失败(bad token)
	#
	post http://foo.com/v1/foos/$(id1)/bar
	json '{
		"val": "hello, xsw"
	}'
	ret 401
	json '{
		"error": "bad token"
	}'

	# 测试修改对象属性
	#
	post http://foo.com/v1/foos/$(id1)/bar
	auth qiniutest
	json '{
		"val": "hello, xsw"
	}'
	ret 200

	# 获取对象，以确认刚才的修改的确生效了
	#
	get http://foo.com/v1/foos/$(id1)
	auth qiniutest
	ret 200
	json '{
		"a": 1,
		"bar": "hello, xsw"
	}'

	# 测试匿名获取对象的bar属性，预期会成功，因为我们API定义了该属性可以匿名获取
	#
	get http://foo.com/v1/foos/$(id1)/bar
	ret 200
	json '{
		"val": "hello, xsw"
	}'

	# 为了测试列出对象，先再插入一个foo对象
	#
	post http://foo.com/v1/foos
	auth qiniutest
	json '{
		"a": 2,
		"bar": "qiniu"
	}'
	ret 200
	json '{
		"id": $(id2)
	}'

	# 测试匿名列出对象，预期会失败(bad token)
	#
	get http://foo.com/v1/foos
	ret 401
	json '{
		"error": "bad token"
	}'

	# 测试列出对象，因为无法预期返回的对象列表次序，我们用 equalSet 而不是常规的 match
	#
	get http://foo.com/v1/foos
	auth qiniutest
	ret 200
	equalSet $(resp.body) '[
		{
			"id": $(id1),
			"a": 1,
			"bar": "hello, xsw"
		},
		{
			"id": $(id2),
			"a": 2,
			"bar": "qiniu"
		}
	]'

	# 测试匿名删除对象，预期会失败(bad token)
	#
	delete http://foo.com/v1/foos/$(id1)
	ret 401
	json '{
		"error": "bad token"
	}'

	# 测试删除对象
	#
	delete http://foo.com/v1/foos/$(id1)
	auth qiniutest
	ret 200

	# 获取对象，预期会失败（从而确认刚才的删除的确生效了）
	#
	get http://foo.com/v1/foos/$(id1)
	auth qiniutest
	ret 612
	json '{
		"error": "entry not found"
	}'
	`)
}

