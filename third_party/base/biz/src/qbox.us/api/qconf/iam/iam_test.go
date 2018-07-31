package iam_test

import (
	"testing"

	"github.com/stretchr/testify.v2/assert"
	"qbox.us/api/qconf/iam"
)

func TestMakeAndParseId(t *testing.T) {
	accessKey := "AccessKey"
	id := iam.MakeId(accessKey)
	ak, err := iam.ParseId(id)
	if assert.Nil(t, err, "iam.ParseId(%s)", id) {
		assert.Equal(t, accessKey, ak, "unexpected ak: %s", ak)
	}

	_, err = iam.ParseId("InvalidKey")
	assert.NotNil(t, err)
}
