package conn

import (
	"net"
	"time"
)

// Listener wraps a net.Listener, and gives a place to store the timeout
// parameters. On Accept, it will wrap the net.Conn with our own Conn for us.
type TimeoutListener struct {
	net.Listener
	Timeout time.Duration
}

func (l *TimeoutListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	tc := &TimeoutConn{
		Conn:    c,
		Timeout: l.Timeout,
	}
	err = tc.SetDeadline(time.Now().Add(l.Timeout))
	return tc, err
}

// Conn wraps a net.Conn, and sets a deadline for every read
// and write operation.
type TimeoutConn struct {
	net.Conn
	Timeout  time.Duration
	Deadline time.Time
}

type Dialer interface {
	Dial(network, address string) (net.Conn, error)
}

type TimeoutDialer struct {
	Dialer
	Timeout     time.Duration
	genDeadline func() time.Time
}

func (d *TimeoutDialer) Dial(network, address string) (net.Conn, error) {
	c, err := d.Dialer.Dial(network, address)
	if err != nil {
		return c, err
	}

	tc := &TimeoutConn{Conn: c, Timeout: d.Timeout}
	if d.genDeadline != nil {
		tc.SetDeadline(d.genDeadline())
	}
	return tc, nil
}

func NewDailer(d Dialer, timeout time.Duration) Dialer {

	return &TimeoutDialer{Dialer: d, Timeout: timeout}
}

func NewDialerWithDeadline(d Dialer, timeout time.Duration, genDeadline func() time.Time) Dialer {

	return &TimeoutDialer{Dialer: d, Timeout: timeout, genDeadline: genDeadline}
}

func (c *TimeoutConn) Read(b []byte) (count int, e error) {
	count, e = c.Conn.Read(b)
	if e != nil {
		return
	}
	if c.Timeout > 0 {
		deadline := time.Now().Add(c.Timeout)
		if !c.Deadline.IsZero() && deadline.After(c.Deadline) {
			deadline = c.Deadline
		}
		e = c.Conn.SetDeadline(deadline)
	}
	return
}

func (c *TimeoutConn) Write(b []byte) (count int, e error) {
	count, e = c.Conn.Write(b)
	if e != nil {
		return
	}
	if c.Timeout > 0 {
		deadline := time.Now().Add(c.Timeout)
		if !c.Deadline.IsZero() && deadline.After(c.Deadline) {
			deadline = c.Deadline
		}
		e = c.Conn.SetDeadline(deadline)
	}
	return

}

func (c *TimeoutConn) SetDeadline(deadline time.Time) error {
	c.Deadline = deadline
	return c.Conn.SetDeadline(deadline)
}

func (c *TimeoutConn) Close() error {
	return c.Conn.Close()
}
