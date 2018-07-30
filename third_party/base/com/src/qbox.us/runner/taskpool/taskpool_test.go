package taskpool

import (
	"math"
	"sync"
	"testing"
	"time"

	"qbox.us/runner/taskpool2"
)

// -------------------------------------------------------

func Test(t *testing.T) {

	runner := New(0, 0)
	ok := runner.TryRun(func() {
		t.Fatal("here")
	})
	if ok {
		t.Fatal("runner.Run failed")
	}

	runner = New(1, 0)
	time.Sleep(time.Millisecond * 100) // let the goroutine run
	ok = runner.TryRun(func() {
		time.Sleep(time.Second)
	})
	if !ok {
		t.Fatal("runner.Run failed")
	}
	ok = runner.TryRun(func() {
	})
	if ok {
		t.Fatal("runner.Run failed")
	}
}

const N = 1024

func BenchmarkGoroutine(b *testing.B) {

	var wg sync.WaitGroup
	var results [N]float64

	b.ResetTimer()
	for i := 0; i < N; i++ {
		wg.Add(1)
		idx := i
		go func() {
			for i := 0; i < 1024; i++ {
				results[(idx+i)%N] = math.Pow(3.14159, float64(i))
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkTaskPool(b *testing.B) {

	runner := New(N, 0)

	var wg sync.WaitGroup
	var results [N]float64

	b.ResetTimer()
	for i := 0; i < N; i++ {
		wg.Add(1)
		idx := i
		runner.Run(func() {
			for i := 0; i < 1024; i++ {
				results[(idx+i)%N] = math.Pow(3.14159, float64(i))
			}
			wg.Done()
		})
	}

	wg.Wait()
}

func BenchmarkTaskPool2(b *testing.B) {

	runner := taskpool2.New(N, 0)

	var wg sync.WaitGroup
	var results [N]float64

	b.ResetTimer()
	for i := 0; i < N; i++ {
		wg.Add(1)
		idx := i
		runner.Run(func() {
			for i := 0; i < 1024; i++ {
				results[(idx+i)%N] = math.Pow(3.14159, float64(i))
			}
			wg.Done()
		})
	}

	wg.Wait()
}

// -------------------------------------------------------
