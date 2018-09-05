package sugarcrm

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"qbox.us/api"
	"qbox.us/errors"
)

//--------------------------------------------------------------------

type Service struct {
	Host      string
	UserName  string
	Password  string
	SessionId string
	LoggedIn  bool
	Conn      *http.Client
}

func New(host, userName, password string, t http.RoundTripper) (svr *Service) {
	client := &http.Client{Transport: t}
	svr = &Service{
		Host:      host,
		UserName:  userName,
		Password:  password,
		Conn:      client,
		SessionId: "trerpvloepchf2s45gr32q9j97",
	}
	return
}

//--------------------------------------------------------------------
// query data
//The parameter naming is all wrong in these docs:
//http://developers.sugarcrm.com/docs/OS/6.0/-docs-Developer_Guides-Sugar_Developer_Guide_6.0-Chapter%202%20Application%20Framework.html#9000729
//http://support.sugarcrm.com/02_Documentation/04_Sugar_Developer/Sugar_Developer_Guide_6.5/02_Application_Framework/Web_Services/05_Method_Calls
//so Do not use named parameters when making REST calls in SugarCRM. i.e. Use positional parameters (JSON array) for 'rest_data' in API calls.
//to keep parameters's position I construct XxxRestData class instead of using map[string][]string when make post req.

type NameValue struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type SetEntryRestData struct {
	Session       string      `json:"session"`
	Module        string      `json:"module_name"`
	NameValueList []NameValue `json:"name_value_list"`
}

type SearchRestData struct {
	Session      string   `json:"session"`
	SearchString string   `json:"search_string"`
	Modules      []string `json:"modules"` // The list of modules to query ([]string)
	Offset       int      `json:"offset"`
	MaxResults   int      `json:"max_results"`
}

//--------------------------------------------------------------------
//response data

type Ret struct {
	Id          interface{} `json:"id"`
	EmailCount  NameValue   `json:"email_count"`
	PhoneWork   NameValue   `json:"phone_work"`
	PhoneMobile NameValue   `json:"phone_mobile"`
	PhoneOther  NameValue   `json:"phone_other"`
}

type SearchRetRecord struct {
	Id    NameValue `json:"id"`
	Name  NameValue `json:"name"`
	Email NameValue `json:"email"`
}

type SearchRet struct {
	Name    string            `name`
	Records []SearchRetRecord `records`
}

type ErrorRet struct {
	Number      int    `number`
	Name        string `name`
	Description string `description`
}

func (e *ErrorRet) Error() string {
	msg, _ := json.Marshal(e)
	return string(msg)
}

//--------------------------------------------------------------------
// func

func (p *Service) callRet(ret interface{}, resp *http.Response) (code int, err error) {

	defer resp.Body.Close()
	errRet := ErrorRet{}
	code = resp.StatusCode
	if code/100 == 2 {
		if ret != nil && resp.ContentLength != 0 {
			bs, _ := ioutil.ReadAll(resp.Body)
			sr := bytes.NewReader(bs)
			json.NewDecoder(sr).Decode(&errRet)
			if !(errRet.Number == 0 && errRet.Name == "" && errRet.Description == "") {
				code = api.UnexpectedResponse
				err = &errRet
				return
			}
			sr = bytes.NewReader(bs)
			err = json.NewDecoder(sr).Decode(ret)
			if err != nil {
				code = api.UnexpectedResponse
			}
		}
	} else {
		if resp.ContentLength != 0 {
			if ct, ok := resp.Header["Content-Type"]; ok && ct[0] == "application/json" {
				json.NewDecoder(resp.Body).Decode(err)
				return
			}
		}
		err = errors.Info(api.NewError(code), "jsonrpc.callRet")
	}
	return
}

func (p *Service) CallWithForm(ret interface{}, url_ string, param map[string][]string) (code int, err error) {
	resp, err := p.Conn.PostForm(url_, param)
	if err != nil {
		return api.NetworkError, err
	}
	return p.callRet(ret, resp)
}

func (p *Service) Call(method string, params, ret interface{}) (code int, err error) {
	encodedParams, _ := json.Marshal(params)
	post := map[string][]string{
		"method":        {method},
		"input_type":    {"JSON"},
		"response_type": {"JSON"},
		"rest_data":     {string(encodedParams)},
	}
	return p.CallWithForm(ret, p.Host, post)
}

func (p *Service) Login() bool {

	ps := md5.New()
	ps.Write([]byte(p.Password))
	password := fmt.Sprintf("%x", ps.Sum(nil))

	params := map[string]interface{}{
		"user_auth": map[string]string{
			"user_name": p.UserName,
			"password":  password,
			"version":   "1",
		},
	}
	ret := map[string]interface{}{}
	code, err := p.Call("login", params, &ret)
	if code != 200 || err != nil {
		return false
	}
	if sessionId, ok := ret["id"]; ok {
		p.SessionId = sessionId.(string)
		return true
	}
	return false
}

//http://.../set_entry
func (p *Service) SetEntry(module string, nv []NameValue) (ret Ret, err error) {
	params := SetEntryRestData{p.SessionId, module, nv}
	_, err = p.Call("set_entry", params, &ret)
	return
}

//http://.../search_by_module
func (p *Service) Search(searchString string, modules []string, offset, maxRet int) (searchRet []SearchRet, err error) {
	params := SearchRestData{
		Session:      p.SessionId,
		SearchString: searchString,
		Modules:      modules,
		Offset:       offset,
		MaxResults:   maxRet,
	}

	ret := map[string][]SearchRet{}
	_, err = p.Call("search_by_module", params, &ret)
	if entryList, ok := ret["entry_list"]; ok {
		searchRet = entryList
	}
	return
}

func (p *Service) LoginOut() {
	if p.LoggedIn {
		p.Call("logout", map[string]string{"session": p.SessionId}, nil)
		p.LoggedIn = false
	}
	return
}
