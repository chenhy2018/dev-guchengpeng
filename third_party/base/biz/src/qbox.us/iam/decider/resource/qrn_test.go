package resource_test

import (
	"testing"

	"github.com/stretchr/testify.v2/assert"
	"qbox.us/iam/decider/resource"
)

func TestSplitQRN(t *testing.T) {
	testCases := []struct {
		qrn           string
		expectedErr   error
		expectedParts []string
	}{
		{"qrn:kodo:z0:123456:bucket/qiniublog", nil, []string{"qrn", "kodo", "z0", "123456", "bucket/qiniublog"}},
		{"qrn:kodo:123456:bucket/qiniublog", resource.ErrBadQRN, nil},
		{"qrn::z0:123456:bucket/qiniublog", resource.ErrBadQRN, nil},
		{"qrn:kodo::123456:bucket/qiniublog", nil, []string{"qrn", "kodo", "", "123456", "bucket/qiniublog"}},
		{"qrn:kodo:z0::bucket/qiniublog", nil, []string{"qrn", "kodo", "z0", "", "bucket/qiniublog"}},
		{"qrn:kodo:z0:123456:", resource.ErrBadQRN, nil},
		{"qrn:kodo:::bucket/qiniublog", nil, []string{"qrn", "kodo", "", "", "bucket/qiniublog"}},
		{":kodo:::bucket/qiniublog", resource.ErrBadQRN, nil},
		{"*", nil, []string{"qrn", "*", "*", "*", "*"}},
		{"abc*", resource.ErrBadQRN, nil},
		{"qrn:*", resource.ErrBadQRN, nil},
	}

	for _, testCase := range testCases {
		q, err := resource.ParseQRN(testCase.qrn)
		assert.Equal(t, testCase.expectedErr, err, "assert error, testCase: %+v", testCase)
		if err == nil {
			assert.Equal(t, testCase.expectedParts, q.Parts(), "assert subs, testCase: %+v", testCase)
		}
	}
}
