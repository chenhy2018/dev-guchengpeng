package enums

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpTypeClusterMap(t *testing.T) {
	clusterMap := OpTypeClusterMap()

	assert.Equal(t, clusterMap, map[OpTypeCluster]string{
		OpTypeClusterAll:     "全部",
		OpTypeClusterAccount: "账户",
		OpTypeClusterBucket:  "空间",
	})
}

func TestOpTypeClusterAll(t *testing.T) {
	clusterAll := OpTypeClusterAll

	assert.Equal(t, clusterAll.Humanize(), "全部")
}

func TestOpTypeClusterAccount(t *testing.T) {
	clusterAccount := OpTypeClusterAccount

	assert.Equal(t, clusterAccount.Humanize(), "账户")
}

func TestOpTypeClusterBucket(t *testing.T) {
	clusterBucket := OpTypeClusterBucket

	assert.Equal(t, clusterBucket.Humanize(), "空间")
}
