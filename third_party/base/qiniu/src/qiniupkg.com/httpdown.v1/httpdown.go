package httpdown

import (
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/facebookgo/clock"
	"github.com/facebookgo/httpdown"
	"github.com/facebookgo/stats"
	"github.com/kavu/go_reuseport"
	"qiniupkg.com/x/log.v7"
)

var (
	ErrReusePortNotSupport = errors.New("SO_REUSEPORT is not supported")
)

type Server httpdown.Server

type HTTP struct {
	// StopTimeout is the duration before we begin force closing connections.
	// Defaults to 1 minute.
	StopTimeout time.Duration

	// KillTimeout is the duration before which we completely give up and abort
	// even though we still have connected clients. This is useful when a large
	// number of client connections exist and closing them can take a long time.
	// Note, this is in addition to the StopTimeout. Defaults to 1 minute.
	KillTimeout time.Duration

	// Stats is optional. If provided, it will be used to record various metrics.
	Stats stats.Client

	// Clock allows for testing timing related functionality. Do not specify this
	// in production code.
	Clock clock.Clock

	DontReusePort  bool
	ForceReusePort bool
}

func (p *HTTP) Serve(s *http.Server, l net.Listener) Server {

	return (httpdown.HTTP{
		StopTimeout: p.StopTimeout,
		KillTimeout: p.KillTimeout,
		Stats:       p.Stats,
		Clock:       p.Clock,
	}).Serve(s, l)
}

func (p *HTTP) ListenAndServe(s *http.Server) (Server, error) {

	if p.DontReusePort {
		return (httpdown.HTTP{
			StopTimeout: p.StopTimeout,
			KillTimeout: p.KillTimeout,
			Stats:       p.Stats,
			Clock:       p.Clock,
		}).ListenAndServe(s)
	}

	l, err := reuseport.NewReusablePortListener("tcp4", s.Addr)
	if err != nil {
		if strings.Contains(err.Error(), "protocol not available") {
			log.Println("SO_REUSEPORT is not supported in this kernal")

			// force SO_REUSEPORT
			if p.ForceReusePort {
				return nil, ErrReusePortNotSupport
			}

			// listen without SO_REUSEPORT
			l, err = net.Listen("tcp", s.Addr)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return (httpdown.HTTP{
		StopTimeout: p.StopTimeout,
		KillTimeout: p.KillTimeout,
		Stats:       p.Stats,
		Clock:       p.Clock,
	}).Serve(s, l), nil
}

var (
	defaultHTTP = &HTTP{}
)

// ListenAndServe is a convenience function to serve and wait for a SIGTERM
// or SIGINT before shutting down.
func ListenAndServe(s *http.Server, hd *HTTP) error {
	if hd == nil {
		hd = defaultHTTP
	}
	hs, err := hd.ListenAndServe(s)
	if err != nil {
		return err
	}
	log.Printf("serving on http://%s/ with pid %d\n", s.Addr, os.Getpid())

	waiterr := make(chan error, 1)
	go func() {
		defer close(waiterr)
		waiterr <- hs.Wait()
	}()

	signals := make(chan os.Signal, 10)
	signal.Notify(signals, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-waiterr:
		if err != nil {
			return err
		}
	case s := <-signals:
		signal.Stop(signals)
		log.Printf("signal received: %s\n", s)
		if err := hs.Stop(); err != nil {
			return err
		}
		if err := <-waiterr; err != nil {
			return err
		}
	}
	log.Println("exiting")
	return nil
}
