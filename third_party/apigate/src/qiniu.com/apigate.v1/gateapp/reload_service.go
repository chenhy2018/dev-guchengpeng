package gateapp

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/qiniu/log.v1"
)

type graceHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	CloseAndWait()
}

type reloadService struct {
	realService *graceHandler
	creator     func() graceHandler
}

func newReloadService(creator func() graceHandler) *reloadService {
	s := creator()
	return &reloadService{realService: &s, creator: creator}
}

func (p *reloadService) ServeHTTP(res http.ResponseWriter, r *http.Request) {

	(*p.realService).ServeHTTP(res, r)
}

func (p *reloadService) Close() {

	(*p.realService).CloseAndWait()
}

func (p *reloadService) Reload() {

	oriS := p.realService
	newS := p.creator()
	p.realService = &newS
	(*oriS).CloseAndWait()
}

func (p *reloadService) ProcessSignals() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGUSR2, os.Interrupt, os.Kill)

	for {
		sig := <-c
		log.Info("receive signal:", sig.String())
		switch sig {
		case syscall.SIGUSR2:
			p.Reload()
		default:
			p.Close()
			if sysSig, ok := sig.(syscall.Signal); ok {
				os.Exit(128 + int(sysSig))
			}
			os.Exit(255)
		}
	}
}
