package stun

import (
	"fmt"
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


func (this *Message) isAllocationRequest() bool {

	return (this.method | this.encoding) == (STUN_MSG_METHOD_ALLOCATE | STUN_MSG_REQUEST)
}

func (this *Message) doAllocationRequest() (*Message, error) {

	if err := this.checkAllocation(); err != nil {
		return this.newErrorMessage(STUN_ERR_BAD_REQUEST, "invalid alloc req: " + err.Error())
	}
	if token, err := this.getAttrReservToken(); err != nil {
		return this.newErrorMessage(STUN_ERR_INSUFFICIENT_CAP, "RESERVATION-TOKEN is not supported")
	} else {
		// TODO: not supported
	}

	msg := &Message{}
	
	return msg, nil
}

func (this *Message) checkAllocation() error {

	// check req tran attr
	found := false
	for _, attr := range this.attributes {
		if attr.typevalue == STUN_ATTR_REQUESTED_TRAN {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("missing REQUESTED-TRANSPORT attribute")
	}

	return nil
}

func (this *Message) getAttrReservToken() (string, error) {

	return "", nil
}
