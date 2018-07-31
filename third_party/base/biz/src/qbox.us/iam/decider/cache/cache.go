package cache

import (
	"sort"

	"qbox.us/iam/decider/match"
	"qbox.us/iam/decider/resource"
	"qbox.us/iam/enums"
)

func MakeActions(strs []string) match.Matchers {
	actions := make(match.Matchers, len(strs))
	for m, str := range strs {
		actions[m] = match.NewTermMatcher(str)
	}
	return actions
}

func MakeResources(strs []string) []match.Matchers {
	resources := make([]match.Matchers, len(strs))
	for m, str := range strs {
		// ignore the resource in policy if error occurred
		qrn, err := resource.ParseQRN(str)
		if err != nil {
			continue
		}
		resources[m] = match.MakeTermMatchers(qrn.Parts())
	}
	return resources
}

type Statement struct {
	Actions   match.Matchers
	Resources []match.Matchers
	Effect    enums.Effect
}

type Statements []Statement

func (s Statements) Less(i, j int) bool {
	return s[i].Effect.IsDeny() && s[j].Effect.IsAllow()
}

func (s Statements) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Statements) Len() int {
	return len(s)
}

func (s Statements) Sort() {
	sort.Sort(s)
}

type Item struct {
	Version    string
	Statements Statements
}

type CacheStore interface {
	Get(id uint32) (*Item, bool)
	Set(id uint32, item *Item)
}
