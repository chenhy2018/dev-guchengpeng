package entity_test

import (
	"testing"

	"github.com/stretchr/testify.v2/assert"
	"qbox.us/iam/entity"
)

func TestIAMKey(t *testing.T) {
	testCases := []struct {
		key      string
		isIAMKey bool
	}{
		{entity.MakeAccessKey(), true},
		{entity.MakeSecretKey(), false},
	}

	for _, testCase := range testCases {
		assert.Len(t, testCase.key, entity.KeyLen*4/3, "make qiam access key: %s", testCase.key)
		ok := entity.IsIAMKey(testCase.key)
		assert.Equal(t, testCase.isIAMKey, ok, "assert qiam access key result, access key: %s", testCase.key)
	}
}
