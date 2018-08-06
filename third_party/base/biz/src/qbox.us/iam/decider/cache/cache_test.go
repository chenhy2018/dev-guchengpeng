package cache_test

import (
	"testing"

	"github.com/stretchr/testify.v2/assert"
	"qbox.us/iam/decider/cache"
	"qbox.us/iam/enums"
)

func cacheItemData() *cache.Item {
	return &cache.Item{
		Version: "the_version",
		Statements: cache.Statements{
			{
				Actions:   cache.MakeActions([]string{"Action11", "Action12"}),
				Resources: cache.MakeResources([]string{"bad_qrn"}),
				Effect:    enums.EffectAllow,
			},
			{
				Actions:   cache.MakeActions([]string{"Action21", "Action22"}),
				Resources: cache.MakeResources([]string{"qrn:fusion:z0:user2:test_domain"}),
				Effect:    enums.EffectDeny,
			},
			{
				Actions:   cache.MakeActions([]string{"Action31", "Action32"}),
				Resources: cache.MakeResources([]string{"qrn:kodo:z1:user3:test_file"}),
				Effect:    enums.EffectAllow,
			},
			{
				Actions:   cache.MakeActions([]string{"Action41", "Action42"}),
				Resources: cache.MakeResources([]string{"qrn:fusion:z3:user4:test_cert"}),
				Effect:    enums.EffectDeny,
			},
			{
				Actions:   cache.MakeActions([]string{"Action31", "Action32"}),
				Resources: cache.MakeResources([]string{"qrn:kodo:z4:user5:test_file"}),
				Effect:    enums.EffectAllow,
			},
			{
				Actions:   cache.MakeActions([]string{"Action41", "Action42"}),
				Resources: cache.MakeResources([]string{"qrn:fusion:z5:user6:test_cert"}),
				Effect:    enums.EffectDeny,
			},
		},
	}
}

func TestSortStatements(t *testing.T) {
	cacheItem := cacheItemData()
	denyCount := 0
	for _, stmt := range cacheItem.Statements {
		if !stmt.Effect.IsAllow() {
			denyCount++
		}
	}
	cacheItem.Statements.Sort()
	for i, testCase := range cacheItem.Statements {
		assert.Equal(t, i < denyCount, testCase.Effect.IsDeny(), "testCase: %+v", testCase)
	}
}
