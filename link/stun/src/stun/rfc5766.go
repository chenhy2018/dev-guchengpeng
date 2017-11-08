package stun

import (
	"fmt"
	"encoding/binary"
	"conf"
	"net"
	"time"
	"sync"
)

const (
	TURN_MAX_LIFETIME            = 600
	TURN_SRV_MIN_PORT            = 49152
	TURN_SRV_MAX_PORT            = 65535
	TURN_PERM_LIFETIME           = 300   // this is fixed (https://tools.ietf.org/html/rfc5766#section-8)
	TURN_PERM_LIMIT              = 10
)

const (
	TURN_RELAY_NEW                 = 0
	TURN_RELAY_BINDED              = 1
	TURN_RELAY_LISTENING           = 2
	TURN_RELAY_CLOSED              = 3
)

const  (
	STUN_MSG_METHOD_ALLOCATE     = 0x0003
	STUN_MSG_METHOD_REFRESH      = 0x0004
	STUN_MSG_METHOD_SEND         = 0x0006
	STUN_MSG_METHOD_DATA         = 0x0007
	STUN_MSG_METHOD_CREATE_PERM  = 0x0008
	STUN_MSG_METHOD_CHANNEL_BIND = 0x0009
)

const (
	STUN_ATTR_CHANNEL_NUMBER     = 0x000C
	STUN_ATTR_LIFETIME           = 0x000d
	STUN_ATTR_XOR_PEER_ADDR      = 0x0012
	STUN_ATTR_DATA               = 0x0013
	STUN_ATTR_XOR_RELAYED_ADDR   = 0x0016
	STUN_ATTR_EVENT_PORT         = 0x0018
	STUN_ATTR_REQUESTED_TRAN     = 0x0019
	STUN_ATTR_DONT_FRAGMENT      = 0x001a
	STUN_ATTR_RESERVATION_TOKEN  = 0x0022
)

const (
	STUN_ERR_FORBIDDEN           = 403
	STUN_ERR_ALLOC_MISMATCH      = 437
	STUN_ERR_WRONG_CRED          = 441
	STUN_ERR_UNSUPPORTED_PROTO   = 442
	STUN_ERR_ALLOC_QUOTA         = 486
	STUN_ERR_INSUFFICIENT_CAP    = 508
)

type turnpool struct {
	// allocation struct map
	table      map[string]*allocation
	tableLck   *sync.Mutex

	// available port number cursor
	availPort  int
	portLck    *sync.Mutex
}

type relayserver struct {
	// relay server
	conn        *net.UDPConn

	// sync on exit
	wg          *sync.WaitGroup

	// server status
	status      int

	// server status
	svrLck      *sync.Mutex

	// alloc reference
	allocRef    *allocation
}

type allocation struct {
	// hash key to find the alloc struct in pool
	key         string

	// keep alive time
	lifetime    uint32

	// expired time computed by the recent refresh req
	expiry      time.Time

	// reservation token
	token       []byte

	// client ip and port info
	source      address

	// relayed transport address
	relay       address

	// permission list
	perms       map[string]time.Time
	permsLck    *sync.Mutex

	// server
	server      *relayserver
}

var (
	allocPool = &turnpool{
		table: map[string]*allocation{},
		tableLck: &sync.Mutex{},
		availPort: TURN_SRV_MIN_PORT,
		portLck: &sync.Mutex{},
	}
)

// -------------------------------------------------------------------------------------------------

func keygen(r *address) string {

	return fmt.Sprintf("%d:%s:%d", r.Proto, r.IP.String(), r.Port)
}

// -------------------------------------------------------------------------------------------------

// general behavior
func (this *Message) generalRequestCheck(r *address) (*allocation, *Message, error) {

	alloc, ok := allocPool.find(keygen(r))
	if !ok {
		msg, err := this.newErrorMessage(STUN_ERR_ALLOC_MISMATCH, "allocation not found")
		return nil, msg, err
	}
	return alloc, nil, nil
}

func (this *Message) doCreatePermRequest(alloc *allocation) (*Message, error) {

	addrs, err := this.getAttrXorPeerAddresses()
	if err != nil {
		return this.newErrorMessage(STUN_ERR_BAD_REQUEST, "peer addresses: " + err.Error())
	}

	if err := alloc.addPerms(addrs); err != nil {
		return this.newErrorMessage(STUN_ERR_INSUFFICIENT_CAP, err.Error())
	}

	msg := &Message{}
	msg.method = this.method
	msg.encoding = STUN_MSG_SUCCESS
	msg.methodName, msg.encodingName = parseMessageType(msg.method, msg.encoding)
	msg.transactionID = append(msg.transactionID, this.transactionID...)

	return msg, nil
}

func (this *Message) doRefreshRequest(alloc *allocation) (*Message, error) {

	msg := &Message{}
	msg.method = this.method
	msg.encoding = STUN_MSG_SUCCESS
	msg.methodName, msg.encodingName = parseMessageType(msg.method, msg.encoding)
	msg.transactionID = append(msg.transactionID, this.transactionID...)

	// get lifetime attribute from stun message
	lifetime, err := this.getAttrLifetime()
	if err != nil {
		lifetime = TURN_MAX_LIFETIME
	} else {
		if lifetime == 0 {
			alloc.free()
			msg.length += msg.addAttrLifetime(0)
			return msg, nil
		}
	}
	alloc.refresh(lifetime)
	msg.length += msg.addAttrLifetime(alloc.lifetime)
	return msg, nil
}

func (this *Message) doAllocationRequest(r *address) (msg *Message, err error) {

	// TODO 1. long-term credential

	// 2. find existing allocations
	alloc, err := newAllocation(r)
	if err != nil {
		return this.newErrorMessage(STUN_ERR_ALLOC_MISMATCH, err.Error())
	}

	// 3. check allocation
	if err = this.checkAllocation(); err != nil {
		return this.newErrorMessage(STUN_ERR_BAD_REQUEST, "invalid alloc req: " + err.Error())
	}

	// 4. TODO handle DONT-FRAGMENT attribute

	// 5. get reservation token
	alloc.token, err = this.getAttrReservToken()
	if err == nil {
		return this.newErrorMessage(STUN_ERR_INSUFFICIENT_CAP, "RESERVATION-TOKEN is not supported")
	}

	// 6. TODO get even port

	// 7. TODO check quota

	// 8. TODO handle ALTERNATE attribute

	// set lifetime
	lifetime, err := this.getAttrLifetime()
	if err != nil {
		lifetime = TURN_MAX_LIFETIME
	}
	alloc.refresh(lifetime)

	// save allocation and reply to the client
	err = alloc.save()
	if err != nil {
		return this.newErrorMessage(STUN_ERR_SERVER_ERROR, "alloc failed: " + err.Error())
	}
	return this.replyAllocationRequest(alloc)
}

func (this *Message) checkAllocation() error {

	// check req tran attr
	if this.findAttr(STUN_ATTR_REQUESTED_TRAN) == nil {
		return fmt.Errorf("missing REQUESTED-TRANSPORT attribute")
	}

	return nil
}

// same as XOR_MAPPED_ADDRESS
func (this *Message) addAttrXorRelayedAddr(r *address) int {

	return this.addAttrXorAddr(r, STUN_ATTR_XOR_RELAYED_ADDR)
}

func (this *Message) addAttrLifetime(t uint32) int {

	attr := &attribute{}
	attr.typevalue = STUN_ATTR_LIFETIME
	attr.typename = parseAttributeType(attr.typevalue)
	attr.length = 4
	attr.value = make([]byte, 4)
	binary.BigEndian.PutUint32(attr.value[0:], t)

	this.attributes = append(this.attributes, attr)
	return 8 // fixed 4 bytes
}

func (this *Message) getAttrReservToken() ([]byte, error) {

	attr := this.findAttr(STUN_ATTR_RESERVATION_TOKEN)
	if attr != nil {
		return attr.value, nil
	}
	return nil, fmt.Errorf("not found")
}

func (this *Message) getAttrLifetime() (uint32, error) {

	attr := this.findAttr(STUN_ATTR_LIFETIME)
	if attr != nil {
		lifetime := binary.BigEndian.Uint32(attr.value)
		return lifetime, nil
	}
	return 0, fmt.Errorf("not found")
}

func (this *Message) getAttrXorPeerAddresses() ([]*address, error) {

	results := []*address{}

	list := this.findAttrAll(STUN_ATTR_XOR_PEER_ADDR)
	if len(list) == 0 {
		return nil, fmt.Errorf("not found")
	}

	for _, attr := range list {
		addr, err := this.getAttrXorAddr(attr)
		if err != nil {
			return nil, fmt.Errorf("value invalid: %s", err)
		}
		results = append(results, addr)
	}

	return results, nil
}

func (this *Message) replyAllocationRequest(alloc *allocation) (*Message, error) {

	msg := &Message{}
	msg.method = STUN_MSG_METHOD_ALLOCATE
	msg.encoding = STUN_MSG_SUCCESS
	msg.methodName, msg.encodingName = parseMessageType(msg.method, msg.encoding)
	msg.transactionID = append(msg.transactionID, this.transactionID...)
	msg.length += msg.addAttrXorRelayedAddr(&alloc.relay)
	msg.length += msg.addAttrLifetime(alloc.lifetime)
	msg.length += msg.addAttrXorMappedAddr(&alloc.source)

	return msg, nil
}

// -------------------------------------------------------------------------------------------------

func newAllocation(r *address) (*allocation, error) {

	key := keygen(r)
	if _, ok := allocPool.find(key); ok {
		return nil, fmt.Errorf("allocation already exists")
	}

	return &allocation{
		key:      key,
		source:   *r,
		perms:    map[string]time.Time{},
		permsLck: &sync.Mutex{},
	}, nil
}

func (alloc *allocation) save() error {

	// insert allocation struct to global pool
	if ok := alloc.addToPool(); !ok {
		return  fmt.Errorf("already allocated")
	}

	// create relay service
	alloc.server = newRelay(alloc)
	port, err := alloc.server.bind()
	if err != nil {
		alloc.removeFromPool()
		return err
	}

	// save relay address
	alloc.relay.IP = net.ParseIP(*conf.Args.IP).To4() // use default IP in args
	alloc.relay.Port = port

	// spawn a thread to listen UDP channel
	if err := alloc.server.spawn(); err != nil {
		return err
	}

	return nil
}

func (alloc *allocation) free() error {

	alloc.server.kill() // may block for a while
	alloc.removeFromPool()

	return nil
}

func (alloc *allocation) addToPool() bool {

	return allocPool.insert(alloc)
}

func (alloc *allocation) removeFromPool() {

	allocPool.remove(alloc.key)
}

func (alloc *allocation) refresh(lifetime uint32) {

	alloc.lifetime = lifetime
	if lifetime > TURN_MAX_LIFETIME {
		alloc.lifetime = TURN_MAX_LIFETIME
	}
	alloc.expiry = time.Now().Add(time.Second * time.Duration(alloc.lifetime))
}

func (alloc *allocation) getRestLife() (int, error) {

	t := int(alloc.expiry.Unix() - time.Now().Unix())
	if t <= 0 {
		return 0, fmt.Errorf("expired.")
	} else {
		return t, nil
	}
}

func (alloc *allocation) addPerms(addrs []*address) (err error) {

	err = nil
	now := time.Now()

	alloc.permsLck.Lock()
	defer alloc.permsLck.Unlock()

	// clear expired permissions
	for ip, expiry := range alloc.perms {
		if now.After(expiry) {
			delete(alloc.perms, ip)
		}
	}

	// add/refresh permission entry
	for _, addr := range addrs {
		key := addr.IP.String()
		if _, ok := alloc.perms[key]; !ok {
			// check maximum capacity of permissions
			if len(alloc.perms) >= TURN_PERM_LIMIT {
				err = fmt.Errorf("maximum permissions reached")
			}
		}
		alloc.perms[key] = now.Add(time.Second * time.Duration(TURN_PERM_LIFETIME))
	}

	return err
}

// -------------------------------------------------------------------------------------------------

func newRelay(alloc *allocation) *relayserver {

	return &relayserver{
		status:   TURN_RELAY_NEW,
		svrLck:   &sync.Mutex{},
		allocRef: alloc,
		wg:       &sync.WaitGroup{},
	}
}

func (svr *relayserver) bind() (p int, _ error) {

	svr.svrLck.Lock()
	defer svr.svrLck.Unlock()
	if svr.status != TURN_RELAY_NEW {
		return -1, fmt.Errorf("relay server has already started")
	}

	// try 40 times, NEVER ASK WHY 40
	for i := 0; i < 40; i++ {

		p = allocPool.nextPort()
		addr := fmt.Sprintf("%s:%d", *conf.Args.IP, p)

		udp, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			continue
		}

		svr.conn, err = net.ListenUDP("udp", udp)
		if err != nil {
			continue
		}

		svr.status = TURN_RELAY_BINDED
		return p, nil
	}
	return -1, fmt.Errorf("could not bind local address")
}

// TODO connection needs retry
func (svr *relayserver) spawn() error {

	svr.svrLck.Lock()
	defer svr.svrLck.Unlock()
	if svr.status != TURN_RELAY_BINDED {
		return fmt.Errorf("could not listen relay address")
	}
	svr.status = TURN_RELAY_LISTENING

	go func(svr *relayserver) {

		// read from UDP socket
		dch := make(chan []byte) // data channel
		ech := make(chan error)  // error channel

		// spawn listening thread
		svr.wg.Add(1)
		go func(conn *net.UDPConn, wg *sync.WaitGroup, dch chan []byte, ech chan error) {
			defer wg.Done()
			defer conn.Close()

			for ;; {
				buf := make([]byte, 1024)
				nr, _, err := conn.ReadFromUDP(buf)
				if err != nil {
					ech <- err
					break
				} else {
					dch <- buf[:nr]
				}
			}
		}(svr.conn, svr.wg, dch, ech)

		// poll fds
		timer := time.NewTimer(time.Second * time.Duration(svr.allocRef.lifetime))
		for quit := false; !quit; {
			select {
			case <-timer.C:
				if seconds, err := svr.allocRef.getRestLife(); err == nil {
					timer = time.NewTimer(time.Second * time.Duration(seconds))
					break
				}
				svr.conn.SetDeadline(time.Now())
				svr.allocRef.removeFromPool()
			case <-dch:
				// TODO
			case <-ech:
				quit = true
			}
		}

		// wait for listening thread
		svr.wg.Wait()
		svr.status = TURN_RELAY_CLOSED
	}(svr)

	return nil
}

func (svr *relayserver) kill() {

	svr.svrLck.Lock()
	defer svr.svrLck.Unlock()
	svr.conn.SetDeadline(time.Now())
	svr.wg.Wait()
}

// -------------------------------------------------------------------------------------------------

func (pool *turnpool) insert(alloc *allocation) bool {

	pool.tableLck.Lock()
	defer pool.tableLck.Unlock()

	if _, ok := pool.table[alloc.key]; !ok {
		pool.table[alloc.key] = alloc
		return true
	}
	return false
}

func (pool *turnpool) remove(key string) {

	pool.tableLck.Lock()
	defer pool.tableLck.Unlock()
	delete(pool.table, key)
}

func (pool *turnpool) find(key string) (alloc *allocation, ok bool) {

	pool.tableLck.Lock()
	defer pool.tableLck.Unlock()
	alloc, ok = pool.table[key]
	return
}

func (pool *turnpool) nextPort() (p int) {

	pool.portLck.Lock()
	defer pool.portLck.Unlock()

	p = pool.availPort
	if pool.availPort == TURN_SRV_MAX_PORT {
		pool.availPort = TURN_SRV_MIN_PORT
	} else {
		pool.availPort++
	}
	return
}
