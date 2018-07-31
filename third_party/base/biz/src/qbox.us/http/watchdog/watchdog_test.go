package watchdog

import (
	"net/http"
	"testing"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/ts"
	"github.com/stretchr/testify/assert"
)

type TestServer struct {
	AppName string
}

func (this TestServer) GetUid(req *http.Request) (uid int, err error) {

	return 123456, nil
}

func (this TestServer) GetApiName(req *http.Request) (apiName string, err error) {
	return req.URL.Path, nil
}

type TestResponseWriter struct {
	status *int
}

func (this TestResponseWriter) Header() http.Header {
	return http.Header{}
}

func (this TestResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (this TestResponseWriter) WriteHeader(code int) {
	*this.status = code
}

func init() {
	log.SetOutputLevel(0)
}

func TestAppLimit(t *testing.T) {
	cfg := Config{
		AppCurThresholdNum: 1,
		UidApiLimiters: []UidApiLimiter{
			{
				ApiName:            "/api_1st",
				CurThresholdNum:    5000,
				TimeInterval:       5000,
				PeriodThresholdNum: 5000,
			},
			{
				ApiName:            "/api_2nd",
				CurThresholdNum:    5000,
				TimeInterval:       5000,
				PeriodThresholdNum: 5000,
			},
		},
	}

	wd, err := Open(cfg)
	if err != nil {
		ts.Fatal(t, "open watchdog failed. cfg:", cfg)
	}

	uid := uint32(123456)
	apiName := "/api_1st"

	code := wd.Check(uid, apiName)
	assert.Equal(t, 200, code)

	code = wd.Check(uid, apiName)
	assert.Equal(t, 571, code)
}

func TestApiUidLimit(t *testing.T) {
	cfg := Config{
		AppCurThresholdNum: 5000,
		UidApiLimiters: []UidApiLimiter{
			{
				ApiName:            "/api_1st",
				CurThresholdNum:    1,
				TimeInterval:       5000,
				PeriodThresholdNum: 5000,
			},
			{
				ApiName:            "/api_2nd",
				CurThresholdNum:    1,
				TimeInterval:       5000,
				PeriodThresholdNum: 5000,
			},
		},
	}

	wd, err := Open(cfg)
	if err != nil {
		ts.Fatal(t, "open watchdog failed. cfg:", cfg)
	}

	uid := uint32(123456)
	apiName := "/api_1st"

	//uid api cur check:
	code := wd.Check(uid, apiName)
	assert.Equal(t, 200, code)

	wd.DecreCurLoad(uid, apiName)

	code = wd.Check(uid, apiName)
	assert.Equal(t, 200, code)

	code = wd.Check(uid, apiName)
	assert.Equal(t, 571, code)

	//uid api period check:
	cfg = Config{
		AppCurThresholdNum: 5000,
		UidApiLimiters: []UidApiLimiter{
			{
				ApiName:            "/api_1st",
				CurThresholdNum:    5000,
				TimeInterval:       5000,
				PeriodThresholdNum: 1,
			},
			{
				ApiName:            "/api_2nd",
				CurThresholdNum:    5000,
				TimeInterval:       5000,
				PeriodThresholdNum: 1,
			},
		},
	}

	wd2, err := Open(cfg)
	code = wd2.Check(uid, apiName)
	assert.Equal(t, 200, code)

	code = wd2.Check(uid, apiName)
	assert.Equal(t, 571, code)
}
