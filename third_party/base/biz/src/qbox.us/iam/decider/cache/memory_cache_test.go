package cache_test

import (
	"testing"

	"github.com/stretchr/testify.v2/assert"
	"qbox.us/iam/decider/cache"
)

func TestMemoryCache(t *testing.T) {
	cacheItem := cacheItemData()

	id := uint32(123)
	c := cache.NewMemoryCache()

	_, ok := c.Get(id)
	assert.False(t, ok, "before set item")

	c.Set(id, cacheItem)

	item, ok := c.Get(id)
	assert.True(t, ok, "after set item")
	assert.Equal(t, cacheItem, item)
}
