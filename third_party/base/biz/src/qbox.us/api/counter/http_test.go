package counter

import (
	"encoding/json"
	"fmt"
	"github.com/qiniu/log.v1"
	"github.com/stretchr/testify.v1/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	. "qbox.us/api/counter/common"
	"reflect"
	"testing"
	"time"
)

var httpBackend *HttpClient
var httpServer *httptest.Server

func testPostInc(t *testing.T) {
	key := "foo"
	counters := CounterMap{"c1": 1, "c2": 3}
	var expected = url.Values{}
	expected.Add("<url>", fmt.Sprintf("/inc/%s/", key))
	expected.Add("c1", "1")
	expected.Add("c2", "3")

	resp, err := httpBackend.postInc(key, counters)
	actual, err := parseResponse(resp, err)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	// test emtpy tags
	resp, err = httpBackend.postInc(key, nil)
	actual, err = parseResponse(resp, err)
	require.NoError(t, err)
	expected = url.Values{}
	expected.Add("<url>", fmt.Sprintf("/inc/%s/", key))
	require.Equal(t, expected, actual)
}

func testPostGet(t *testing.T) {
	key := "bar"
	tags := []string{"c0", "c3"}
	var expected = url.Values{}
	expected.Add("<url>", fmt.Sprintf("/get/%s/", key))
	expected.Add("tag", "c0")
	expected.Add("tag", "c3")

	resp, err := httpBackend.postGet(key, tags)
	actual, err := parseResponse(resp, err)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	// test emtpy tags
	resp, err = httpBackend.postGet(key, nil)
	actual, err = parseResponse(resp, err)
	require.NoError(t, err)
	expected = url.Values{}
	expected.Add("<url>", fmt.Sprintf("/get/%s/", key))
	require.Equal(t, expected, actual)
}

func testPostSet(t *testing.T) {
	key := "foo"
	counters := CounterMap{"c1": 1, "c2": 3}
	var expected = url.Values{}
	expected.Add("<url>", fmt.Sprintf("/set/%s/", key))
	expected.Add("c1", "1")
	expected.Add("c2", "3")

	resp, err := httpBackend.postSet(key, counters)
	actual, err := parseResponse(resp, err)
	if err != nil {
		t.Error(err)
		return
	}
	eq := reflect.DeepEqual(actual, expected)
	if !eq {
		t.Errorf("expected: %v\n actual: %v", expected, actual)
	}
}

func testPostListPrefix(t *testing.T) {
	prefix := "fo"
	var expected = url.Values{}
	expected.Add("<url>", fmt.Sprintf("/listprefix/"))
	expected.Add("prefix", "fo")

	resp, err := httpBackend.postListPrefix(prefix, nil)
	actual, err := parseResponse(resp, err)
	if err != nil {
		t.Error(err)
		return
	}
	eq := reflect.DeepEqual(actual, expected)
	if !eq {
		t.Errorf("expected: %v\n actual: %v", expected, actual)
	}
}

func testPostListRange(t *testing.T) {
	start := "foo"
	end := "bar"
	var expected = url.Values{}
	expected.Add("<url>", fmt.Sprintf("/listrange/"))
	expected.Add("start", "foo")
	expected.Add("end", "bar")

	resp, err := httpBackend.postListRange(start, end, nil)
	actual, err := parseResponse(resp, err)
	if err != nil {
		t.Error(err)
		return
	}
	eq := reflect.DeepEqual(actual, expected)
	if !eq {
		t.Errorf("expected: %v\n actual: %v", expected, actual)
	}
}

func testPostRemove(t *testing.T) {
	key := "foo"
	var expected = url.Values{}
	expected.Add("<url>", fmt.Sprintf("/remove/%s/", key))
	// FIXME: workaround for solving empty post body error

	resp, err := httpBackend.postRemove(key)
	actual, err := parseResponse(resp, err)
	if err != nil {
		t.Error(err)
		return
	}
	eq := reflect.DeepEqual(actual, expected)
	if !eq {
		t.Errorf("expected: %v\n actual: %v", expected, actual)
	}
}

func testSelectNextHost(t *testing.T) {
	conf := &HttpClientConfig{HostUrls: []string{"1", "2", "3"}}
	client, err := NewHttpClient(conf)
	require.NoError(t, err)
	first := client.selectNextHost(nil)
	require.Equal(t, "1", first.Url)
	second := client.selectNextHost(first)
	require.Equal(t, "2", second.Url)
	secondAgain := client.selectNextHost(first)
	require.Equal(t, second, secondAgain)
	third := client.selectNextHost(second)
	require.Equal(t, "3", third.Url)
	firstAgain := client.selectNextHost(third)
	require.Equal(t, "1", firstAgain.Url)
}

func testDisableHost(t *testing.T) {
	conf := &HttpClientConfig{HostUrls: []string{"1", "2", "3"},
		FailuresToDisable: 10, DisableMillis: 500}
	client, err := NewHttpClient(conf)
	require.NoError(t, err)
	first := client.selectNextHost(nil)
	require.Equal(t, "1", first.Url)
	second := client.selectNextHost(first)
	require.Equal(t, "2", second.Url)
	third := client.selectNextHost(second)
	require.Equal(t, "3", third.Url)
	for i := 0; i != conf.FailuresToDisable-1; i++ {
		recordFailure(first)
		host := client.selectNextHost(nil)
		require.Equal(t, first, host)
	}
	recordFailure(first)
	host := client.selectNextHost(nil)
	require.Equal(t, second, host)
	host = client.selectNextHost(third)
	require.Equal(t, second, host)
	for i := 0; i != 10*len(conf.HostUrls); i++ {
		host := client.selectRandomHost()
		require.NotEqual(t, host, first)
	}
	time.Sleep(conf.DisableMillis * time.Millisecond)
	host = client.selectNextHost(nil)
	require.Equal(t, first, host)
	for i := 0; i != conf.FailuresToDisable; i++ {
		recordFailure(first)
	}
	host = client.selectNextHost(nil)
	require.Equal(t, second, host)
	recordSuccess(first)
	host = client.selectNextHost(nil)
	require.Equal(t, first, host)
}

func parseResponse(resp *http.Response, oldErr error) (form url.Values, err error) {
	if oldErr != nil {
		log.Info("?")
		return nil, oldErr
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &form)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return form, err
}

func handleEcho(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	var form = req.Form
	form.Add("<url>", req.URL.String())
	bytes, err := json.Marshal(form)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

func RunEchoServer(port int) {
	http.HandleFunc("/", handleEcho)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleEcho)
	httpServer = httptest.NewServer(mux)
	defer httpServer.Close()
	conf := &HttpClientConfig{
		HostUrls: []string{
			"http://localhost:110",
			httpServer.URL,
		},
		TimeoutMillis:     1000,
		Retries:           1,
		FailuresToDisable: 10,
		DisableMillis:     5000,
	}
	var err error
	httpBackend, err = NewHttpClient(conf)
	if err != nil {
		log.Fatal("Fail to obtain http client: ", err)
	}
	testPostInc(t)
	testPostGet(t)
	testPostSet(t)
	testPostListPrefix(t)
	testPostListRange(t)
	testPostRemove(t)
	testSelectNextHost(t)
	testDisableHost(t)
}
