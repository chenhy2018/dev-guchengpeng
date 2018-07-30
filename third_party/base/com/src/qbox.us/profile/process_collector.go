// https://github.com/prometheus/client_golang/blob/master/prometheus/process_collector.go
// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package profile

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
)

type processCollector struct {
	pid             int
	collectFn       func(chan<- prometheus.Metric)
	pidFn           func() (int, error)
	cpuTotal        *prometheus.Desc
	openFDs, maxFDs *prometheus.Desc
	vsize, rss      *prometheus.Desc
	startTime       *prometheus.Desc
}

// newProcessCollector returns a collector which exports the current state of
// process metrics including cpu, memory and file descriptor usage as well as
// the process start time for the given process id under the given namespace.
func newProcessCollector(pid int, namespace string) prometheus.Collector {
	return newProcessCollectorPIDFn(
		func() (int, error) { return pid, nil },
		namespace,
	)
}

// newProcessCollectorPIDFn returns a collector which exports the current state
// of process metrics including cpu, memory and file descriptor usage as well
// as the process start time under the given namespace. The given pidFn is
// called on each collect and is used to determine the process to export
// metrics for.
func newProcessCollectorPIDFn(
	pidFn func() (int, error),
	namespace string,
) prometheus.Collector {
	ns := ""
	if len(namespace) > 0 {
		ns = namespace + "_"
	}

	c := processCollector{
		pidFn:     pidFn,
		collectFn: func(chan<- prometheus.Metric) {},

		cpuTotal: prometheus.NewDesc(
			ns+"process_cpu_seconds_total",
			"Total user and system CPU time spent in seconds.",
			nil, constLabels,
		),
		openFDs: prometheus.NewDesc(
			ns+"process_open_fds",
			"Number of open file descriptors.",
			nil, constLabels,
		),
		maxFDs: prometheus.NewDesc(
			ns+"process_max_fds",
			"Maximum number of open file descriptors.",
			nil, constLabels,
		),
		vsize: prometheus.NewDesc(
			ns+"process_virtual_memory_bytes",
			"Virtual memory size in bytes.",
			nil, constLabels,
		),
		rss: prometheus.NewDesc(
			ns+"process_resident_memory_bytes",
			"Resident memory size in bytes.",
			nil, constLabels,
		),
		startTime: prometheus.NewDesc(
			ns+"process_start_time_seconds",
			"Start time of the process since unix epoch in seconds.",
			nil, constLabels,
		),
	}

	// Set up process metric collection if supported by the runtime.
	if _, err := procfs.NewStat(); err == nil {
		c.collectFn = c.processCollect
	}

	return &c
}

// Describe returns all descriptions of the collector.
func (c *processCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.cpuTotal
	ch <- c.openFDs
	ch <- c.maxFDs
	ch <- c.vsize
	ch <- c.rss
	ch <- c.startTime
}

// Collect returns the current state of all metrics of the collector.
func (c *processCollector) Collect(ch chan<- prometheus.Metric) {
	c.collectFn(ch)
}

// TODO(ts): Bring back error reporting by reverting 7faf9e7 as soon as the
// client allows users to configure the error behavior.
func (c *processCollector) processCollect(ch chan<- prometheus.Metric) {
	pid, err := c.pidFn()
	if err != nil {
		return
	}

	p, err := procfs.NewProc(pid)
	if err != nil {
		return
	}

	if stat, err := p.NewStat(); err == nil {
		ch <- prometheus.MustNewConstMetric(c.cpuTotal, prometheus.CounterValue, stat.CPUTime())
		ch <- prometheus.MustNewConstMetric(c.vsize, prometheus.GaugeValue, float64(stat.VirtualMemory()))
		ch <- prometheus.MustNewConstMetric(c.rss, prometheus.GaugeValue, float64(stat.ResidentMemory()))
		if startTime, err := stat.StartTime(); err == nil {
			ch <- prometheus.MustNewConstMetric(c.startTime, prometheus.GaugeValue, startTime)
		}
	}

	if fds, err := p.FileDescriptorsLen(); err == nil {
		ch <- prometheus.MustNewConstMetric(c.openFDs, prometheus.GaugeValue, float64(fds))
	}

	if limits, err := p.NewLimits(); err == nil {
		ch <- prometheus.MustNewConstMetric(c.maxFDs, prometheus.GaugeValue, float64(limits.OpenFiles))
	}
}
