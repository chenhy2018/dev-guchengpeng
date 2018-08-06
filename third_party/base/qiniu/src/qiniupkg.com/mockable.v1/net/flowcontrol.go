package net

import (
	"time"
)

var (
	FlowControlWindow = 50 * time.Millisecond
)

type speed struct {
	windowAt int
	Bps      int
}

type FlowControl struct {
	speeds  []speed
	startAt time.Time

	speedNow     speed
	lastWindow   int
	accumulation int64
}

func (c *FlowControl) Require(max int) (n int) {

	since := time.Since(c.startAt)
	window := int(since / FlowControlWindow)
	for {
		if len(c.speeds) == 0 {
			break
		}
		if window >= c.speeds[0].windowAt {
			// 会有跨越边界的问题，先简单处理
			c.speedNow = c.speeds[0]
			c.speeds = c.speeds[1:]
			continue
		}
		break
	}
	windowSize := int64(c.speedNow.Bps) * int64(FlowControlWindow) / int64(time.Second)

	if c.lastWindow != window {
		c.accumulation = 0
		c.lastWindow = window
	}
	if c.accumulation > windowSize {
		// 跨越边界的时候可能会出现累计值大于窗口大小
		c.accumulation = windowSize
	}

	n = int(windowSize) - int(c.accumulation)
	if n > max {
		n = max
	}

	return
}

func (c *FlowControl) Consume(n int) {

	c.accumulation += int64(n)
}

func NewFlowControl(s Speed) (c *FlowControl) {

	var speeds = make([]speed, 0, len(s))

	lastWindow := 0
	for _, item := range s {
		// NOTE: ignore lastItem.Duration
		speeds = append(speeds, speed{
			windowAt: lastWindow,
			Bps:      item.Bps,
		})
		lastWindow += int(time.Duration(item.Duration) * time.Millisecond / FlowControlWindow)
	}

	speedNow := speeds[0]
	c = &FlowControl{
		speeds:   speeds[1:],
		speedNow: speedNow,
		startAt:  time.Now(),
	}

	return

}
