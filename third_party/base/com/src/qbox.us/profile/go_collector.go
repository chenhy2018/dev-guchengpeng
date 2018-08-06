// https://github.com/prometheus/client_golang/blob/master/prometheus/go_collector.go
package profile

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type goCollector struct {
	goroutinesDesc *prometheus.Desc
	threadsDesc    *prometheus.Desc
	gcDesc         *prometheus.Desc

	// metrics to describe and collect
	metrics memStatsMetrics
}

// NewGoCollector returns a collector which exports metrics about the current
// go process.
func newGoCollector() prometheus.Collector {
	return &goCollector{
		goroutinesDesc: prometheus.NewDesc(
			"profile_go_goroutines",
			"Number of goroutines that currently exist.",
			nil, constLabels),
		threadsDesc: prometheus.NewDesc(
			"profile_go_threads",
			"Number of OS threads created.",
			nil, constLabels),
		gcDesc: prometheus.NewDesc(
			"profile_go_gc_duration_seconds",
			"A summary of the GC invocation durations.",
			nil, constLabels),
		metrics: memStatsMetrics{
			{
				desc: prometheus.NewDesc(
					memstatNamespace("alloc_bytes"),
					"Number of bytes allocated and still in use.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.Alloc) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("alloc_bytes_total"),
					"Total number of bytes allocated, even if freed.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.TotalAlloc) },
				valType: prometheus.CounterValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("sys_bytes"),
					"Number of bytes obtained from system.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.Sys) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("lookups_total"),
					"Total number of pointer lookups.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.Lookups) },
				valType: prometheus.CounterValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("mallocs_total"),
					"Total number of mallocs.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.Mallocs) },
				valType: prometheus.CounterValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("frees_total"),
					"Total number of frees.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.Frees) },
				valType: prometheus.CounterValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("heap_alloc_bytes"),
					"Number of heap bytes allocated and still in use.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.HeapAlloc) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("heap_sys_bytes"),
					"Number of heap bytes obtained from system.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.HeapSys) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("heap_idle_bytes"),
					"Number of heap bytes waiting to be used.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.HeapIdle) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("heap_inuse_bytes"),
					"Number of heap bytes that are in use.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.HeapInuse) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("heap_released_bytes"),
					"Number of heap bytes released to OS.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.HeapReleased) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("heap_objects"),
					"Number of allocated objects.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.HeapObjects) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("stack_inuse_bytes"),
					"Number of bytes in use by the stack allocator.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.StackInuse) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("stack_sys_bytes"),
					"Number of bytes obtained from system for stack allocator.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.StackSys) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("mspan_inuse_bytes"),
					"Number of bytes in use by mspan structures.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.MSpanInuse) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("mspan_sys_bytes"),
					"Number of bytes used for mspan structures obtained from system.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.MSpanSys) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("mcache_inuse_bytes"),
					"Number of bytes in use by mcache structures.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.MCacheInuse) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("mcache_sys_bytes"),
					"Number of bytes used for mcache structures obtained from system.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.MCacheSys) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("buck_hash_sys_bytes"),
					"Number of bytes used by the profiling bucket hash table.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.BuckHashSys) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("gc_sys_bytes"),
					"Number of bytes used for garbage collection system metadata.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.GCSys) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("other_sys_bytes"),
					"Number of bytes used for other system allocations.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.OtherSys) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("next_gc_bytes"),
					"Number of heap bytes when next garbage collection will take place.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.NextGC) },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("last_gc_time_seconds"),
					"Number of seconds since 1970 of last garbage collection.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return float64(ms.LastGC) / 1e9 },
				valType: prometheus.GaugeValue,
			}, {
				desc: prometheus.NewDesc(
					memstatNamespace("gc_cpu_fraction"),
					"The fraction of this program's available CPU time used by the GC since the program started.",
					nil, constLabels,
				),
				eval:    func(ms *runtime.MemStats) float64 { return ms.GCCPUFraction },
				valType: prometheus.GaugeValue,
			},
		},
	}
}

func memstatNamespace(s string) string {
	return fmt.Sprintf("profile_go_memstats_%s", s)
}

// Describe returns all descriptions of the collector.
func (c *goCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.goroutinesDesc
	ch <- c.threadsDesc
	ch <- c.gcDesc
	for _, i := range c.metrics {
		ch <- i.desc
	}
}

// Collect returns the current state of all metrics of the collector.
func (c *goCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.goroutinesDesc, prometheus.GaugeValue, float64(runtime.NumGoroutine()))
	n, _ := runtime.ThreadCreateProfile(nil)
	ch <- prometheus.MustNewConstMetric(c.threadsDesc, prometheus.GaugeValue, float64(n))

	var stats debug.GCStats
	stats.PauseQuantiles = make([]time.Duration, 5)
	debug.ReadGCStats(&stats)

	quantiles := make(map[float64]float64)
	for idx, pq := range stats.PauseQuantiles[1:] {
		quantiles[float64(idx+1)/float64(len(stats.PauseQuantiles)-1)] = pq.Seconds()
	}
	quantiles[0.0] = stats.PauseQuantiles[0].Seconds()
	ch <- prometheus.MustNewConstSummary(c.gcDesc, uint64(stats.NumGC), float64(stats.PauseTotal.Seconds()), quantiles)

	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)
	for _, i := range c.metrics {
		ch <- prometheus.MustNewConstMetric(i.desc, i.valType, i.eval(ms))
	}
}

// memStatsMetrics provide description, value, and value type for memstat metrics.
type memStatsMetrics []struct {
	desc    *prometheus.Desc
	eval    func(*runtime.MemStats) float64
	valType prometheus.ValueType
}
