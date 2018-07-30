package decider_test

import (
	"testing"

	"github.com/stretchr/testify.v2/assert"
	"qbox.us/iam/decider"
	"qbox.us/iam/decider/cache"
	"qbox.us/iam/decider/resource"
	"qbox.us/iam/entity"
)

var (
	testPolicy0 = &entity.QConfInfo{
		IUID:    1,
		Enabled: true,
		Statement: []entity.Statement{
			{
				Action: []string{
					"cdn/CreateDomain",
					"cdn/UpdateDomain",
				},
				Resource: []string{
					"qrn:fusion:::domain/www.qiniu.*",
				},
				Effect: "Allow",
			},
			{
				Action: []string{
					"bucket/Create*",
					"bucket/Delete*",
				},
				Resource: []string{
					"qrn:kodo:z0::bucket/qbox",
					"qrn:kodo:z0::bucket/qiniu*",
				},
				Effect: "Allow",
			},
		},
	}
	testPolicy1 = &entity.QConfInfo{
		IUID:    2,
		Enabled: true,
		Statement: []entity.Statement{
			{
				Action: []string{"*"},
				Resource: []string{
					"qrn:fusion:::domain/www.qiniu.*",
				},
				Effect: "Allow",
			},
		},
	}
	testPolicy2 = &entity.QConfInfo{
		IUID:    3,
		Enabled: true,
		Statement: []entity.Statement{
			{
				Action: []string{"cdn/CreateDomain"},
				Resource: []string{
					"*",
				},
				Effect: "Allow",
			},
		},
	}
	testPolicy3 = &entity.QConfInfo{
		IUID: 5,
	}
	testPolicy4 = &entity.QConfInfo{
		IUID:    3,
		Enabled: false,
		Statement: []entity.Statement{
			{
				Action: []string{"cdn/CreateDomain"},
				Resource: []string{
					"*",
				},
				Effect: "Allow",
			},
		},
	}
	testCases = []struct {
		action     string
		resource   string
		expectErr  error
		expectPass bool
	}{
		{"cdn/CreateDomain", "qrn:fusion:::domain/www.qiniu.com", nil, true},
		{"cdn/UpdateDomain", "qrn:fusion:::domain/www.qiniu.io", nil, true},
		{"cdn/UpdateDomain", "qrn:fusion:z0::domain/www.qiniu.io", nil, true},
		{"cdn/UpdateDomain", "qrn:fusion::123456:domain/www.qiniu.io", nil, true},
		{"cdn/UpdateDomain", "qrn:fusion:z0:123456:domain/www.qiniu.io", nil, true},
		{"bucket/CreateBucket", "qrn:kodo:::bucket/qbox", nil, false},
		{"bucket/CreateBucket", "qrn:kodo:z0::bucket/qbox", nil, true},
		{"bucket/CreateBucket", "qrn:kodo:z0:123456:bucket/qbox", nil, true},
		{"bucket/DeleteBucket", "qrn:kodo:z0::bucket/qiniublog", nil, true},
		{"bucket/notexist", "qrn:kodo:z0::bucket/qiniublog", nil, false},
		{"bucket/DeleteBucket", "qrn:kodo:z0:123456:bucket/qiniublog", nil, true},
		{"notexist/notexist", "qrn:kodo:z0:123456:bucket/qiniublog", nil, false},
		{"notexist/notexist", "qrn:notexist:::notexist/notexist", nil, false},
		{"bucket/DeleteBucket", "qrn:kodo:z0::", resource.ErrBadQRN, false},
		{"notexist/notexist", ":notexist:::notexist/notexist", resource.ErrBadQRN, false},
		{"notexist/notexist", "qrn:notexist::notexist/notexist", resource.ErrBadQRN, false},
		{"bucket/DeleteBucket", "qrn:z0:123456:four_parts", resource.ErrBadQRN, false},
		{"bucket/DeleteBucket", "qrn:z0:three_parts", resource.ErrBadQRN, false},
		{"cdn/UpdateDomain", "", resource.ErrBadQRN, false},
	}
	testCases1 = []struct {
		action     string
		expectPass bool
	}{
		{"cdn/CreateDomain", true},
		{"bucket/notexist", false},
	}
	simpleDeciderWithCache    decider.Decider
	simpleDeciderWithoutCache decider.Decider
)

func init() {
	simpleDeciderWithCache = decider.NewSimpleDecider(cache.NewMemoryCache())
	simpleDeciderWithoutCache = decider.NewSimpleDecider(nil)
}

func TestSimpleVerify(t *testing.T) {
	for i, testCase := range testCases {
		qrn, err := resource.ParseQRN(testCase.resource)
		assert.Equal(t, err, testCase.expectErr, "err unexpect, i: %d, testCase: %+v", i, testCase)
		if err == nil {
			pass := simpleDeciderWithCache.Verify(testPolicy0, testCase.action, qrn)
			assert.Equal(t, testCase.expectPass, pass, "pass unexpect, i: %d, testCase: %+v", i, testCase)

			pass = simpleDeciderWithoutCache.Verify(testPolicy0, testCase.action, qrn)
			assert.Equal(t, testCase.expectPass, pass, "pass unexpect, i: %d, testCase: %+v", i, testCase)

			pass = simpleDeciderWithoutCache.Verify(testPolicy3, testCase.action, qrn)
			assert.False(t, pass, "pass unexpect, i: %d, testCase: %+v", i, testCase)

			pass = simpleDeciderWithoutCache.Verify(testPolicy4, testCase.action, qrn)
			assert.False(t, pass, "pass unexpect, i: %d, testCase: %+v", i, testCase)
		}
	}

	for i, testCase := range testCases1 {
		pass := simpleDeciderWithCache.Verify(testPolicy0, testCase.action, nil)
		assert.Equal(t, testCase.expectPass, pass, "pass unexpect, i: %d, testCase: %+v", i, testCase)
	}

	{
		qrn, err := resource.ParseQRN("qrn:fusion:::domain/www.qiniu.com")
		if assert.NoError(t, err, "error: %s", err) {
			pass := simpleDeciderWithCache.Verify(testPolicy1, "cdn/UpdateDomain", qrn)
			assert.True(t, pass)
		}
	}
	{
		qrn, err := resource.ParseQRN("qrn:fusion:::domain/www.qiniu.com")
		if assert.NoError(t, err, "error: %s", err) {
			pass := simpleDeciderWithCache.Verify(testPolicy2, "cdn/CreateDomain", qrn)
			assert.True(t, pass)
		}
	}
}

func BenchmarkSimpleVerifyWithCacheAND1IUID(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i, testCase := range testCases {
			qrn, err := resource.ParseQRN(testCase.resource)
			if err != nil {
				continue
			}
			pass := simpleDeciderWithCache.Verify(testPolicy0, testCase.action, qrn)
			if testCase.expectPass != pass {
				b.Logf("pass unexpect, i: %d, testCase: %+v", i, testCase)
				b.Fail()
			}
		}
	}
}
func BenchmarkSimpleVerifyWithCacheAND50IUIDs(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testPolicy0.IUID = uint32(n % 50)
		for i, testCase := range testCases {
			qrn, err := resource.ParseQRN(testCase.resource)
			if err != nil {
				continue
			}
			pass := simpleDeciderWithCache.Verify(testPolicy0, testCase.action, qrn)
			if testCase.expectPass != pass {
				b.Logf("pass unexpect, i: %d, testCase: %+v", i, testCase)
				b.Fail()
			}
		}
	}
}

func BenchmarkSimpleVerifyWithoutCacheAND1IUID(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i, testCase := range testCases {
			qrn, err := resource.ParseQRN(testCase.resource)
			if err != nil {
				continue
			}
			pass := simpleDeciderWithoutCache.Verify(testPolicy0, testCase.action, qrn)
			if err != testCase.expectErr {
				b.Logf("err unexpect, i: %d, testCase: %+v", i, testCase)
				b.Fail()
				continue
			}
			if testCase.expectPass != pass {
				b.Logf("pass unexpect, i: %d, testCase: %+v", i, testCase)
				b.Fail()
			}
		}
	}
}
