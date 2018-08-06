package ios

import "sync"

// -------------------------------------------------------------------------

type Channels struct {
	free  []Channel
	Io    *Service
	mutex sync.Mutex
	avail int
}

func NewChannels(io *Service, max int) *Channels {
	return &Channels{Io: io, free: make([]Channel, max)}
}

func (r *Channels) Cap() int {
	return len(r.free)
}

func (r *Channels) Alloc() (channel Channel, code int, err error) {

	r.mutex.Lock()
	if r.avail > 0 {
		r.avail--
		channel = r.free[r.avail]
		r.mutex.Unlock()
		code = 200
		return
	}
	r.mutex.Unlock()

	return r.Io.Mkchan()
}

func (r *Channels) Free(channel Channel) {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.avail < len(r.free) {
		r.free[r.avail] = channel
		r.avail++
	}
}

// -------------------------------------------------------------------------
