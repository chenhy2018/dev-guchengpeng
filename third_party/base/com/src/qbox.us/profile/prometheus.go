package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	host, _     = os.Hostname()
	constLabels = map[string]string{
		"pid":       fmt.Sprint(os.Getpid()),
		"service":   filepath.Base(os.Args[0]),
		"goversion": runtime.Version(),
		"host":      host,
	}
	prometheusRegistry = prometheus.NewRegistry(prometheus.NewRegistryOption{})
)

func RegisterToPrometheus(c prometheus.Collector) error {
	_, err := prometheusRegistry.Register(c)
	return err
}
func UnRegisterFromPrometheus(c prometheus.Collector) bool {
	return prometheusRegistry.Unregister(c)
}

func GetConstLabels() map[string]string {
	<-initDone
	m := make(map[string]string)
	for k, v := range constLabels {
		m[k] = v
	}
	return m
}

func registerAll() {
	prometheusRegistry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace:   "profile",
		Subsystem:   "process",
		Name:        "uptime_seconds",
		Help:        "process uptime seconds",
		ConstLabels: constLabels,
	}, func() (v float64) {
		return time.Since(startTime).Seconds()
	}))

	prometheusRegistry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace:   "profile",
		Subsystem:   "go",
		Name:        "maxprocs",
		Help:        "go runtime.GOMAXPROCS",
		ConstLabels: constLabels,
	}, func() (v float64) {
		return float64(runtime.GOMAXPROCS(-1))
	}))

	prometheusRegistry.MustRegister(newGoCollector())
	prometheusRegistry.MustRegister(newProcessCollector(os.Getpid(), "profile"))
}
