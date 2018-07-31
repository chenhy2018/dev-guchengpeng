package counter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v2"
	. "qbox.us/api/counter/common"
)

type HttpClientConfig struct {
	AccessKey         string        `json:"access_key"`
	SecretKey         string        `json:"secret_key"`
	HostUrls          []string      `json:"urls"`
	TimeoutMillis     time.Duration `json:"timeout_millis"`
	Retries           int           `json:"retries"`
	FailuresToDisable int           `json:"failures_to_disable"`
	DisableMillis     time.Duration `json:"disable_millis"`
}

type Host struct {
	Url            string
	LastFailedTime time.Time
	Failures       int
	mutex          sync.RWMutex
}

type HttpClient struct {
	hosts             []*Host
	timeoutDuration   time.Duration
	retries           int
	client            *http.Client
	failuresToDisable int
	disableDuration   time.Duration
}

func NewHttpClient(conf *HttpClientConfig) (*HttpClient, error) {

	if len(conf.HostUrls) == 0 {
		return nil, errors.New("You should specify at least one backendUrls")
	}
	timeoutDuration := conf.TimeoutMillis * time.Millisecond
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   timeoutDuration,
			KeepAlive: 30 * time.Second,
		}).Dial,
		ResponseHeaderTimeout: timeoutDuration,
	}
	client := digest.NewClient(&digest.Mac{
		conf.AccessKey, []byte(conf.SecretKey),
	}, transport)
	hosts := []*Host{}
	for _, url := range conf.HostUrls {
		hosts = append(hosts, &Host{url, time.Time{}, 0, sync.RWMutex{}})
	}
	return &HttpClient{hosts, timeoutDuration, conf.Retries, client,
		conf.FailuresToDisable, conf.DisableMillis * time.Millisecond}, nil
}

func (client *HttpClient) selectRandomHost() *Host {
	i := rand.Intn(len(client.hosts))
	return client.selectHostAtIndex(i)
}

func (client *HttpClient) selectNextHost(host *Host) *Host {
	for i := 0; i != len(client.hosts); i++ {
		if client.hosts[i] == host {
			return client.selectHostAtIndex((i + 1) % len(client.hosts))
		}
	}
	// not found, so just select starting from first
	return client.selectHostAtIndex(0)
}

func (client *HttpClient) selectHostAtIndex(index int) *Host {
	hostLen := len(client.hosts)
	for i := index; i < len(client.hosts); i++ {
		host := client.hosts[i]
		if client.hostAvailable(host) {
			return host
		}
	}
	for i := 0; i < index && i < hostLen; i++ {
		host := client.hosts[i]
		if client.hostAvailable(host) {
			return host
		}
	}
	// no hosts available, so just return host at current index
	return client.hosts[index]
}

func recordFailure(host *Host) {
	host.mutex.Lock()
	defer host.mutex.Unlock()
	host.Failures++
	host.LastFailedTime = time.Now()

}

func recordSuccess(host *Host) {
	host.mutex.Lock()
	defer host.mutex.Unlock()
	host.Failures = 0
}

func (client *HttpClient) hostAvailable(host *Host) bool {
	now := time.Now()
	host.mutex.Lock()
	defer host.mutex.Unlock()
	if client.failuresToDisable == 0 { // do not disable hosts
		return true
	}
	if host.Failures < client.failuresToDisable {
		return true
	}
	if now.Sub(host.LastFailedTime) > client.disableDuration {
		// disable should be canceled
		host.Failures = host.Failures / 2
		return true
	}
	return false
}

func (client *HttpClient) Inc(key string, counters CounterMap) (err error) {
	if len(counters) == 0 {
		return httputil.NewError(400, "counters should not be empty.")
	}
	resp, err := client.postInc(key, counters)
	bytes, err := readResponse(resp, err)
	_ = bytes
	return err
}

func (client *HttpClient) Get(key string, tags []string) (counters CounterMap, err error) {
	resp, err := client.postGet(key, tags)
	bytes, err := readResponse(resp, err)
	if err != nil {
		return nil, err
	}
	var result map[string]CounterMap
	err = json.Unmarshal(bytes, &result)
	if err != nil || result[key] == nil {
		return nil, httputil.NewError(500, "incorrect values for returned counters")
	}
	return result[key], nil
}

func (client *HttpClient) MGet(keys []string, tags []string) (counters map[string]CounterMap, err error) {
	resp, err := client.postMGet(keys, tags)
	bytes, err := readResponse(resp, err)
	if err != nil {
		return nil, err
	}
	var result map[string]CounterMap
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, httputil.NewError(500, "incorrect values for returned counters")
	}
	return result, nil
}

func (client *HttpClient) Set(key string, counters CounterMap) (err error) {
	if len(counters) == 0 {
		return httputil.NewError(400, "counters should not be empty.")
	}
	resp, err := client.postSet(key, counters)
	_, err = readResponse(resp, err)
	return err
}

func (client *HttpClient) ListPrefix(prefix string, tags []string) (
	counterGroups []CounterGroup, err error) {
	resp, err := client.postListPrefix(prefix, tags)
	bytes, err := readResponse(resp, err)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &counterGroups)
	return counterGroups, err
}

func (client *HttpClient) ListRange(start string, end string, tags []string) (
	counterGroups []CounterGroup, err error) {

	return client.ListRanges([]StartEnd{StartEnd{Start: start, End: end}}, tags)
}

type StartEnd struct {
	Start string
	End   string
}

func (client *HttpClient) ListRanges(r []StartEnd, tags []string) (
	counterGroups []CounterGroup, err error) {

	var ranges []url.Values
	for _, se := range r {
		u := url.Values{}
		u.Set("start", se.Start)
		u.Set("end", se.End)
		u["tag"] = tags
		ranges = append(ranges, u)
	}
	resp, err := client.postForm("listrange/", ranges)
	bytes, err := readResponse(resp, err)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &counterGroups)
	return counterGroups, err
}

func (client *HttpClient) Close() {
	// do nothing
}

func (client *HttpClient) Remove(key string) (err error) {
	resp, err := client.postRemove(key)
	_, err = readResponse(resp, err)
	return err
}

func (client *HttpClient) postInc(key string, counters CounterMap) (
	resp *http.Response, err error) {

	formValues := url.Values{}
	for tag, value := range counters {
		formValues.Add(tag, strconv.FormatInt(value, 10))
	}
	incUrl := fmt.Sprintf("inc/%s/", key)
	return client.postForm(incUrl, formValues)
}

func shouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	return resp.StatusCode >= http.StatusInternalServerError
}

func (client *HttpClient) postGet(key string, tags []string) (
	resp *http.Response, err error) {

	getUrl := fmt.Sprintf("get/%s/", key)
	formValues := url.Values{}
	for _, tag := range tags {
		formValues.Add("tag", tag)
	}
	return client.postForm(getUrl, formValues)
}

func (client *HttpClient) postMGet(keys []string, tags []string) (
	resp *http.Response, err error) {

	formValues := url.Values{}
	for _, tag := range tags {
		formValues.Add("tag", tag)
	}
	for _, key := range keys {
		formValues.Add("key", key)
	}
	return client.postForm("mget/", formValues)
}

func (client *HttpClient) postSet(key string, counters CounterMap) (
	resp *http.Response, err error) {

	formValues := url.Values{}
	for tag, value := range counters {
		formValues.Add(tag, strconv.FormatInt(value, 10))
	}
	setUrl := fmt.Sprintf("set/%s/", key)
	return client.postForm(setUrl, formValues)
}

func (client *HttpClient) postListPrefix(prefix string, tags []string) (
	resp *http.Response, err error) {

	formValues := url.Values{}
	formValues.Add("prefix", prefix)
	formValues["tag"] = tags
	listPrefixUrl := "listprefix/"
	return client.postForm(listPrefixUrl, formValues)
}

func (client *HttpClient) postListRange(
	start string, end string, tags []string) (
	resp *http.Response, err error) {

	formValues := url.Values{}
	formValues.Add("start", start)
	formValues.Add("end", end)
	formValues["tag"] = tags
	listRangeUrl := "listrange/"
	return client.postForm(listRangeUrl, formValues)
}

func (client *HttpClient) postRemove(key string) (
	resp *http.Response, err error) {

	removeUrl := fmt.Sprintf("remove/%s/", key)
	formValues := url.Values{}
	return client.postForm(removeUrl, formValues)
}

func (client *HttpClient) postWithOptionalForm(
	url string, formValues []url.Values) (resp *http.Response, err error) {

	var forms []string
	for _, v := range formValues {
		forms = append(forms, v.Encode())
	}
	data := strings.TrimSpace(strings.Join(forms, "\n"))
	if len(data) == 0 {
		return client.client.Post(
			url,
			"text/plain",
			strings.NewReader(""),
		)
	}
	return client.client.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
}

func (client *HttpClient) postForm(path string, formValues interface{}) (
	resp *http.Response, err error) {

	var form []url.Values
	switch f := formValues.(type) {
	case url.Values:
		form = append(form, f)
	case []url.Values:
		form = f
	default:
		return nil, errors.New("not support")
	}
	host := client.selectRandomHost()
	fullUrl := fmt.Sprintf("%s/%s", host.Url, path)
	resp, err = client.postWithOptionalForm(fullUrl, form)
	remainingRetries := client.retries
	for shouldRetry(resp, err) && remainingRetries > 0 {
		if resp != nil {
			resp.Body.Close()
		}
		recordFailure(host)
		host = client.selectNextHost(host)
		fullUrl = fmt.Sprintf("%s/%s", host.Url, path)
		resp, err = client.postWithOptionalForm(fullUrl, form)
		remainingRetries--
	}
	if shouldRetry(resp, err) {
		recordFailure(host)
	} else {
		recordSuccess(host)
	}
	return resp, err
}

func readResponse(resp *http.Response, httpError error) (bytes []byte, err error) {
	if httpError != nil {
		return nil, httpError
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, rpc.ResponseError(resp)
	}
	bytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
