package lib

import (
	"net"
	"net/http"

	papanet "qiniu.com/papa/papa-golang/net"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/qiniu/log.v1"
)

func ServeGoMetrics(address string) (err error) {
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", prometheus.Handler())

	proto, addr := papanet.ParseIPCAddr(address, "tcp4")
	lis, err := net.Listen(proto, addr)
	if err != nil {
		log.Error("Papa go metrics listen failed: ", err)
		return
	}

	goSilent(func() {
		log.Info("Metrics Start serving at", address)

		server := &http.Server{Handler: metricsMux}

		log.Error("Ping service failed", server.Serve(lis))
	}, nil)
	return
}

func goSilent(f func(), d func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("Panic silenced:", r)
				if d != nil {
					d()
				}
			}
		}()
		f()
	}()
}
