package serverstat

import (
	"io/ioutil"
	"net/http"
	"qbox.us/dyn"
	"qbox.us/errors"
	"qbox.us/net/httputil"
	"qbox.us/servestk"
)

// ----------------------------------------------------------

func Get(url string) (data []byte, err error) {

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	return
}

// ----------------------------------------------------------

type Interface interface {
	ServiceStat() interface{}
}

// ----------------------------------------------------------

type Service struct {
	Interface
}

func New(i Interface) Service {
	return Service{i}
}

//
// GET /stat														- 获得完整的 ServiceStat 信息
// GET /stat?q=<Key>												- 获得 ServiceStat 某个 prop 的值
// GET /stat?dividend=<Dividend>&divisor=<Divisor> 				- 获得 ServiceStat 2个 prop 的相除结果
//
func (r Service) Stat(w http.ResponseWriter, req *http.Request) {

	doc := r.ServiceStat()

	req.ParseForm()

	if len(req.Form) == 0 {
		httputil.Reply(w, 200, doc)
		return
	}

	if q, ok := req.Form["q"]; ok {
		val, ok2 := dyn.Get(doc, q[0])
		if ok2 {
			httputil.Reply(w, 200, val)
			return
		}
	} else if dividend, ok := req.Form["dividend"]; ok {
		if divisor, ok2 := req.Form["divisor"]; ok2 {
			dividend1, ok3 := dyn.GetFloat(doc, dividend[0])
			divisor1, ok4 := dyn.GetFloat(doc, divisor[0])
			if ok3 && ok4 {
				httputil.Reply(w, 200, dividend1/divisor1)
				return
			}
		}
	}
	httputil.ReplyError(w, "Not Found", 400)
}

func (r Service) RegisterHandlers(mux1 *http.ServeMux) (err error) {

	mux := servestk.New(mux1, servestk.SafeHandler)
	mux.HandleFunc("/stat", func(w http.ResponseWriter, req *http.Request) { r.Stat(w, req) })
	return nil
}

func Run(addr string, i Interface) (err error) {

	r := Service{i}
	mux := http.NewServeMux()
	err = r.RegisterHandlers(mux)
	if err != nil {
		err = errors.Info(err, "serverstat.Run", addr).Detail(err).Warn()
		return
	}
	return http.ListenAndServe(addr, mux)
}

// ----------------------------------------------------------
