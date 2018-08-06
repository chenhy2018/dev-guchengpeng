package net

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFlowControl(t *testing.T) {

	ast := assert.New(t)

	speedCases := []Speed{
		{{Bps: 1000}},
		{{Bps: 500}},
		{{Bps: 15000}},
		{{Bps: 600, Duration: 100}, {Bps: 1000}},
		{{Bps: 600, Duration: 100}, {Duration: 150, Bps: 1000}, {Bps: 1}},
		{{Bps: 100, Duration: 100}, {Duration: 150, Bps: 1000}, {Bps: 1}},
	}

	done := make(chan bool, len(speedCases))
	for _, ss := range speedCases {

		speeds := ss
		go func() {
			fc := NewFlowControl(speeds)
			startAt := time.Now()
			for _, speed := range speeds {
				duration := speed.Duration
				if duration == 0 {
					duration = 200
				}
				total := 0
				for {
					since := time.Since(startAt)
					if since < time.Duration(duration)*time.Millisecond {
						n := fc.Require(100)
						fc.Consume(n)
						total += n
						if n == 0 {
							time.Sleep(FlowControlWindow / 2)
							continue
						}
					} else {
						startAt = startAt.Add(time.Duration(duration) * time.Millisecond)
						ast.Equal(speed.Bps*duration/1000, total, "%+v of %+v", speed, speeds)
						break
					}
				}
				if speed.Duration == 0 {
					break
				}
			}
			done <- true
		}()
	}

	for range speedCases {
		<-done
	}
}
