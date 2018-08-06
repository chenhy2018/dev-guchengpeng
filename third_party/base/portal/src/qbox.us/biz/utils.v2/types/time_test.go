package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHunderdNanoSecondTime(t *testing.T) {
	assert.Equal(t, HundredNanoSecond(1e8+1).Time(), time.Unix(10, 100))
	assert.Equal(t, HundredNanoSecond(1e8-1).Time(), time.Unix(9, 999999900))
}

func TestMakeHunderedNanoSecondFromTime(t *testing.T) {
	a := MakeHunderedNanoSecondFromTime(time.Now())
	assert.Equal(t, a, MakeHunderedNanoSecondFromTime(a.Time()))
}

func TestDuration(t *testing.T) {
	type durationExpectedStruct struct {
		Layout string
		Result string
	}

	durationExpected := []durationExpectedStruct{
		durationExpectedStruct{"1ns", "1纳秒"},
		durationExpectedStruct{"1000ns", "1微秒"},
		durationExpectedStruct{"1000µs", "1毫秒"},
		durationExpectedStruct{"1000ms", "1秒"},
		durationExpectedStruct{"60s", "1分钟"},

		durationExpectedStruct{"0h0m1s0ns", "1秒"},
		durationExpectedStruct{"0h0m61s0ns", "1分钟 1秒"},
		durationExpectedStruct{"0h61m0s0ns", "1小时 1分钟"},
		durationExpectedStruct{"25h0m0s0ns", "25小时"},
	}

	for _, item := range durationExpected {
		d, _ := time.ParseDuration(item.Layout)
		str := Duration(d).Localize()
		assert.Equal(t, item.Result, str)
	}

}
