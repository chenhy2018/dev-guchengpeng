package decider

import (
	"qbox.us/iam/decider/cache"
	"qbox.us/iam/decider/match"
	"qbox.us/iam/decider/resource"
	"qbox.us/iam/entity"
)

type SimpleDecider struct {
	cache cache.CacheStore
}

var _ Decider = &SimpleDecider{}

func NewSimpleDecider(cache cache.CacheStore) *SimpleDecider {
	return &SimpleDecider{
		cache: cache,
	}
}

func (d *SimpleDecider) matchResources(resources []match.Matchers, qrnSubs []string) bool {
	for _, matchers := range resources {
		if matchers.MatchAll(qrnSubs) {
			return true
		}
	}
	return false
}

func (d *SimpleDecider) matchActions(matchers match.Matchers, action string) (matched bool) {
	matchers.Each(func(_ int, matcher match.Matcher) bool {
		matched = matcher.Match(action)
		return matched
	})
	return
}

func (d *SimpleDecider) buildCacheItem(version string, stmts []entity.Statement) *cache.Item {
	cacheItem := &cache.Item{
		Version:    version,
		Statements: make(cache.Statements, len(stmts)),
	}
	for i, stmt := range stmts {
		cacheItem.Statements[i] = cache.Statement{
			Actions:   cache.MakeActions(stmt.Action),
			Resources: cache.MakeResources(stmt.Resource),
			Effect:    stmt.Effect,
		}
	}
	cacheItem.Statements.Sort()
	return cacheItem
}

func (d *SimpleDecider) getCacheItem(iuid uint32, version string, stmts []entity.Statement) *cache.Item {
	if d.cache == nil {
		return d.buildCacheItem(version, stmts)
	}
	cacheItem, ok := d.cache.Get(iuid)
	if !ok || cacheItem.Version != version {
		cacheItem = d.buildCacheItem(version, stmts)
		d.cache.Set(iuid, cacheItem)
	}
	return cacheItem
}

func (d *SimpleDecider) Verify(policy Policy, action string, qrn *resource.QRN) bool {
	if !policy.IsEnabled() {
		return false
	}
	cacheItem := d.getCacheItem(policy.GetIUID(), policy.GetVersion(), policy.GetStatments())
	for _, stmt := range cacheItem.Statements {
		if !d.matchActions(stmt.Actions, action) {
			continue
		}
		if qrn == nil {
			return true
		}
		if d.matchResources(stmt.Resources, qrn.Parts()) {
			return stmt.Effect.IsAllow()
		}
	}
	return false
}
