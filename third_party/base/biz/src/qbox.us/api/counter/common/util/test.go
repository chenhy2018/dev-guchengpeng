package util

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/log.v1"
	"github.com/stretchr/testify.v1/require"
	. "qbox.us/api/counter/common"
)

var bar = CounterGroup{"bar", CounterMap{"b1": 11, "b2": 12, "b3": 13}}
var foo = CounterGroup{"foo", CounterMap{"f1": 1}}
var foobar = CounterGroup{"foobar", CounterMap{"f1": 111, "b2": 112}}
var testingGroups = []CounterGroup{bar, foo, foobar}
var testingPrefixes = []string{"fo", "foo", "foob", "b", "bar", "bar1", "quick"}
var testingRanges = [][]string{{"foo", "bar"}}
var testingTags = [][]string{{"f1"}, {"b3"}, {"f1", "b2"}, {"f1", "f2"}, {"b1", "b2", "b3"}, nil}
var nonexist = "thisisnonexistfortest"

func selectPrefix(groups []CounterGroup, prefix string, tags []string) []CounterGroup {
	results := []CounterGroup{}
	for _, group := range groups {
		if strings.HasPrefix(group.Key, prefix) {
			counterGroup := CounterGroup{group.Key, filterCounters(group.Counters, tags)}
			for _, tag := range tags {
				if _, ok := counterGroup.Counters[tag]; !ok {
					counterGroup.Counters[tag] = 0
				}
			}
			results = append(results, counterGroup)
		}
	}
	return results
}

func selectRange(groups []CounterGroup, start string, end string, tags []string) []CounterGroup {
	results := []CounterGroup{}
	for _, group := range groups {
		if group.Key >= start && group.Key < end {
			counterGroup := CounterGroup{group.Key, filterCounters(group.Counters, tags)}
			for _, tag := range tags {
				if _, ok := counterGroup.Counters[tag]; !ok {
					counterGroup.Counters[tag] = 0
				}
			}
			results = append(results, counterGroup)
		}
	}
	return results
}

func clearTestingKeys(counter Counter) {
	for _, group := range testingGroups {
		counter.Remove(group.Key)
	}
}

func initTestingKeys(counter Counter) {
	clearTestingKeys(counter)
	for _, group := range testingGroups {
		counter.Set(group.Key, group.Counters)
	}
}

func allTags(counters CounterMap) []string {
	tags := make([]string, 0)
	for tag, _ := range counters {
		tags = append(tags, tag)
	}
	return tags
}

func filterCounters(counters CounterMap, tags []string) CounterMap {
	filtered := CounterMap{}
	if len(tags) == 0 {
		for tag, value := range counters {
			filtered[tag] = value
		}
		return filtered
	} else {
		for _, tag := range tags {
			if value, ok := counters[tag]; ok {
				filtered[tag] = value
			} else {
				filtered[tag] = 0
			}
		}
		return filtered
	}
}

func addCounters(countersList ...CounterMap) CounterMap {
	sum := CounterMap{}
	for _, counters := range countersList {
		for tag, value := range counters {
			sum[tag] += value
		}
	}
	return sum
}

func sample(list []string, count int) []string {
	if count > len(list) {
		count = len(list)
	}
	perms := rand.Perm(count)
	sample := make([]string, 0)
	for _, i := range perms {
		sample = append(sample, list[i])
	}
	return sample
}

func TestCounterGet(t *testing.T, counter Counter) {
	initTestingKeys(counter)
	for _, group := range testingGroups {
		// get all tags
		counters, err := counter.Get(group.Key, nil)
		require.NoError(t, err, "counter: %#v, group: %#v", counter, group)
		require.Equal(t, group.Counters, counters, "counter: %#v, group: %#v", counter, group)
		tags := allTags(group.Counters)
		// get specified tags
		for i := 0; i <= len(tags); i++ {
			selectedTags := sample(tags, i)
			expected := filterCounters(group.Counters, selectedTags)
			counters, err := counter.Get(group.Key, selectedTags)
			require.NoError(t, err)
			require.Equal(t, expected, counters)
		}

		// get non-exist tag
		selectedTags := []string{nonexist}
		counters, err = counter.Get(group.Key, selectedTags)
		expected := filterCounters(group.Counters, selectedTags)
		require.NoError(t, err, "counter: %#v, group: %#v", counter, group)
		require.Equal(t, expected, counters, "counter: %#v, group: %#v", counter, group)

		// get specified tags with non-exist tag
		for i := 0; i <= len(tags); i++ {
			selectedTags := append(sample(tags, i), nonexist)
			expected := filterCounters(group.Counters, selectedTags)
			counters, err := counter.Get(group.Key, selectedTags)
			require.NoError(t, err)
			require.Equal(t, expected, counters)
		}
	}
	// get non-exist key
	_, err := counter.Get(nonexist, nil)

	code, _ := httputil.DetectError(err)
	require.Equal(t, 404, code)
}

func TestCounterSet(t *testing.T, counter Counter) {
	clearTestingKeys(counter)
	// set when empty
	for _, group := range testingGroups {
		err := counter.Set(group.Key, group.Counters)
		require.NoError(t, err)
		counters, err := counter.Get(group.Key, nil)
		require.NoError(t, err)
		require.Equal(t, group.Counters, counters)
		counter.Remove(group.Key)
	}

	// set where both modify and add new tags
	for _, group := range testingGroups {
		tags := allTags(group.Counters)
		for i := 0; i <= len(tags); i++ {
			err := counter.Set(group.Key, group.Counters)
			require.NoError(t, err)
			counters, err := counter.Get(group.Key, nil)
			require.NoError(t, err)
			require.Equal(t, group.Counters, counters)

			selectedTags := sample(tags, i)
			modifiedCounters := filterCounters(group.Counters, selectedTags)
			doubledCounters := addCounters(modifiedCounters, modifiedCounters)
			newCounters := CounterMap{nonexist: rand.Int63()}
			setCounters := addCounters(doubledCounters, newCounters)

			err = counter.Set(group.Key, setCounters)
			counters, err = counter.Get(group.Key, nil)
			expected := addCounters(group.Counters, modifiedCounters, newCounters)
			require.NoError(t, err)
			require.Equal(t, expected, counters)
			counter.Remove(group.Key)
		}
	}
	// set with empty tags
	for _, group := range testingGroups {
		err := counter.Set(group.Key, CounterMap{})
		code, _ := httputil.DetectError(err)
		require.Equal(t, 400, code)
	}
}

func TestCounterInc(t *testing.T, counter Counter) {
	clearTestingKeys(counter)
	// inc when empty
	for _, group := range testingGroups {
		err := counter.Inc(group.Key, group.Counters)
		require.NoError(t, err)
		counters, err := counter.Get(group.Key, nil)
		require.NoError(t, err)
		require.Equal(t, group.Counters, counters)
		counter.Remove(group.Key)
	}

	// inc where both modify and add new tags
	for _, group := range testingGroups {
		tags := allTags(group.Counters)
		for i := 0; i <= len(tags); i++ {
			err := counter.Inc(group.Key, group.Counters)
			require.NoError(t, err)
			counters, err := counter.Get(group.Key, nil)
			require.NoError(t, err)
			require.Equal(t, group.Counters, counters)

			selectedTags := sample(tags, i)
			modifiedCounters := filterCounters(group.Counters, selectedTags)
			newCounters := CounterMap{nonexist: rand.Int63()}
			incCounters := addCounters(modifiedCounters, newCounters)

			err = counter.Inc(group.Key, incCounters)
			counters, err = counter.Get(group.Key, nil)
			expected := addCounters(group.Counters, modifiedCounters, newCounters)
			require.NoError(t, err)
			require.Equal(t, expected, counters)
			counter.Remove(group.Key)
		}
	}
	// inc with empty tags
	for _, group := range testingGroups {
		err := counter.Set(group.Key, CounterMap{})
		valueError, ok := err.(*httputil.ErrorInfo)
		require.True(t, ok)
		require.Equal(t, 400, valueError.Code)
	}
}

func TestCounterListPrefix(t *testing.T, counter Counter) {
	initTestingKeys(counter)
	for _, prefix := range testingPrefixes {
		for _, tags := range testingTags {
			groups, err := counter.ListPrefix(prefix, tags)
			expected := selectPrefix(testingGroups, prefix, tags)
			require.NoError(t, err)
			require.Equal(t, expected, groups)
		}
	}
}

func TestCounterListRange(t *testing.T, counter Counter) {
	initTestingKeys(counter)
	for _, keyRange := range testingRanges {
		for _, tags := range testingTags {
			start := keyRange[0]
			end := keyRange[1]
			groups, err := counter.ListRange(start, end, tags)
			expected := selectRange(testingGroups, start, end, tags)
			require.NoError(t, err)
			require.Equal(t, expected, groups)
		}
	}
}

func TestCounterRemove(t *testing.T, counter Counter) {
	clearTestingKeys(counter)
	// set and remove
	for _, group := range testingGroups {
		err := counter.Set(group.Key, group.Counters)
		require.NoError(t, err)
		counters, err := counter.Get(group.Key, nil)
		require.NoError(t, err)
		require.Equal(t, group.Counters, counters)
		err = counter.Remove(group.Key)
		require.NoError(t, err)
		err = counter.Remove(group.Key)
		code, _ := httputil.DetectError(err)
		require.Equal(t, 404, code)
		counters, err = counter.Get(group.Key, nil)
		code, _ = httputil.DetectError(err)
		require.Equal(t, 404, code)
	}
}

func TestCounterClose(t *testing.T, counter Counter) {
	counter.Close()
}

func TestCounterAll(t *testing.T, counter Counter) {
	log.Info("Test counter All!")
	TestCounterGet(t, counter)
	TestCounterSet(t, counter)
	TestCounterInc(t, counter)
	TestCounterListPrefix(t, counter)
	TestCounterListRange(t, counter)
	TestCounterRemove(t, counter)
	TestCounterClose(t, counter)
}
