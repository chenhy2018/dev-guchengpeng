package stun

import (
	"net"
)

const (
	NET_UDP = 0x0
	NET_TCP = 0x1
	NET_TLS = 0x2
)

type address struct {
	IP      net.IP
	Port    int
	Proto   byte
}

// -------------------------------------------------------------------------------------------------

func (this *Message) ProcessUDP(r *net.UDPAddr) (*Message, error) {

	addr := &address{
		IP:   r.IP,
		Port: r.Port,
		Proto: NET_UDP,
	}
	return this.process(addr)
}

func (this *Message) process(r *address) (*Message, error) {

	// special handlers
	switch this.method | this.encoding {
	case STUN_MSG_METHOD_ALLOCATE | STUN_MSG_REQUEST: return this.doAllocationRequest(r)
	case STUN_MSG_METHOD_BINDING | STUN_MSG_REQUEST:  return this.doBindingRequest(r)
	}

	if this.isRequest() {
		// common requests
		alloc, msg, err := this.generalRequestCheck(r)
		if err != nil {
			return msg, err
		}

		switch this.method {
		case STUN_MSG_METHOD_REFRESH:      return this.doRefreshRequest(alloc)
		case STUN_MSG_METHOD_CREATE_PERM:  return this.doCreatePermRequest(alloc)
		}

		return this.newErrorMessage(STUN_ERR_BAD_REQUEST, "not support")

	} else if this.isIndication() {
		// indications
		return nil, nil
	}

	return nil, nil // drop
}

func parseMessageType(method, encoding uint16) (m string, e string) {

	switch (method) {
	case STUN_MSG_METHOD_BINDING: m = "binding"
	case STUN_MSG_METHOD_ALLOCATE: m = "allocation"
	case STUN_MSG_METHOD_REFRESH: m = "refresh"
	case STUN_MSG_METHOD_SEND: m = "send"
	case STUN_MSG_METHOD_DATA: m = "data"
	case STUN_MSG_METHOD_CREATE_PERM: m = "create_permission"
	case STUN_MSG_METHOD_CHANNEL_BIND: m = "channel_bind"
	default: m = "unknown"
	}

	switch (encoding) {
	case STUN_MSG_REQUEST: e = "request"
	case STUN_MSG_INDICATION: e = "indication"
	case STUN_MSG_SUCCESS: e = "success_response"
	case STUN_MSG_ERROR: e = "error_response"
	}

	return
}

func parseAttributeType(db uint16) string {

	switch (db) {
	case STUN_ATTR_MAPPED_ADDR: return "MAPPED-ADDRESS"
	case STUN_ATTR_USERNAME: return "USERNAME"
	case STUN_ATTR_MESSAGE_INTEGRITY: return "MESSAGE-INTEGRITY"
	case STUN_ATTR_ERROR_CODE: return "ERROR-CODE"
	case STUN_ATTR_UNKNOWN_ATTR: return "UNKNOWN-ATTRIBUTES"
	case STUN_ATTR_REALM: return "REALM"
	case STUN_ATTR_NONCE: return "NONCE"
	case STUN_ATTR_XOR_MAPPED_ADDR: return "XOR-MAPPED-ADDRESS"
	case STUN_ATTR_SOFTWARE: return "SOFTWARE"
	case STUN_ATTR_ALTERNATE_SERVER: return "ALTERNATE-SERVER"
	case STUN_ATTR_FINGERPRINT: return "FINGERPRINT"
	case STUN_ATTR_CHANNEL_NUMBER: return "CHANNEL-NUMBER"
	case STUN_ATTR_LIFETIME: return "LIFETIME"
	case STUN_ATTR_XOR_PEER_ADDR: return "XOR-PEER-ADDRESS"
	case STUN_ATTR_DATA: return "DATA"
	case STUN_ATTR_XOR_RELAYED_ADDR: return "XOR-RELAYED-ADDRESS"
	case STUN_ATTR_EVENT_PORT: return "EVEN-PORT"
	case STUN_ATTR_REQUESTED_TRAN: return "REQUESTED-TRANSPORT"
	case STUN_ATTR_DONT_FRAGMENT: return "DONT-FRAGMENT"
	case STUN_ATTR_RESERVATION_TOKEN: return "RESERVATION-TOKEN"
	}
	return "RESERVED"
}
