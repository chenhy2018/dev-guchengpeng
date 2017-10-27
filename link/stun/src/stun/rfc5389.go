package stun

import (
	"fmt"
	"encoding/binary"
	"net"
)

const (
	STUN_MSG_HEADER_SIZE  = 20
	STUN_MSG_MAGIC_COOKIE = 0x2112A442
)

const (
	STUN_MSG_TYPE_METHOD_MASK    = 0x3eef   // 0b 0011 1110 1110 1111
	STUN_MSG_TYPE_ENCODING_MASK  = 0x0110   // 0b 0000 0001 0001 0000

	STUN_MSG_METHOD_BINDING      = 0x0001
	STUN_MSG_BINDING_REQUEST     = 0x0000
	STUN_MSG_BINDING_INDICATION  = 0x0010
	STUN_MSG_BINDING_SUCCESS     = 0x0100
	STUN_MSG_BINDING_ERROR       = 0x0110
)

const (
	STUN_ATTR_MAPPED_ADDR       = 0x0001
	STUN_ATTR_USERNAME          = 0x0006
	STUN_ATTR_MESSAGE_INTEGRITY = 0x0008
	STUN_ATTR_ERROR_CODE        = 0x0009
	STUN_ATTR_UNKNOWN_ATTR      = 0x000a
	STUN_ATTR_REALM             = 0x0014
	STUN_ATTR_NONCE             = 0x0015
	STUN_ATTR_XOR_MAPPED_ADDR   = 0x0020

	STUN_ATTR_SOFTWARE          = 0x8022
	STUN_ATTR_ALTERNATE_SERVER  = 0x8023
	STUN_ATTR_FINGERPRINT       = 0x8028
)

type attribute struct {
	typename          string
	typevalue         uint16
	length            int
	value             []byte
}

type Message struct {
	methodName        string
	encodingName      string
	method            uint16
	encoding          uint16
	length            int
	transactionID     []byte
	attributes        []attribute
}

func (this *Message) Buffer() []byte {

	payload := make([]byte, 20)

	// message type
	binary.BigEndian.PutUint16(payload[0:], uint16(this.method | this.encoding))
	
	// message length
	binary.BigEndian.PutUint16(payload[2:], uint16(this.length))

	// put magic cookie
	binary.BigEndian.PutUint32(payload[4:], uint32(STUN_MSG_MAGIC_COOKIE))

	// put transaction ID
	copy(payload[8:], this.transactionID)

	// append attributes
	for _, attr := range this.attributes {
		bytes := make([]byte, 4)
		binary.BigEndian.PutUint16(bytes[0:], attr.typevalue)
		binary.BigEndian.PutUint16(bytes[2:], uint16(attr.length))
		payload = append(payload, bytes...)
		payload = append(payload, attr.value...)
	}

	return payload
}

func (this *Message) ProcessUDP(r *net.UDPAddr) (*Message, error) {

	if this.isBindingRequest() {
		return this.doBindingRequest(r)
	}

	return nil, nil
}

func (this *Message) isBindingRequest() bool {

	return (this.method | this.encoding) == (STUN_MSG_METHOD_BINDING | STUN_MSG_BINDING_REQUEST)
}

func (this *Message) doBindingRequest(r *net.UDPAddr) (*Message, error) {

	msg := &Message{}
	msg.method = STUN_MSG_METHOD_BINDING
	msg.transactionID = append(msg.transactionID, this.transactionID...)

	// add xor port and address
	len := msg.addXorMappedAddr(r)

	msg.length = len
	msg.encoding = STUN_MSG_BINDING_SUCCESS
	msg.methodName, msg.encodingName = parseMessageType(msg.method, msg.encoding)
	return msg, nil
}

func (this *Message) addXorMappedAddr(r *net.UDPAddr) int {

	attr := attribute{}
	attr.typevalue = STUN_ATTR_XOR_MAPPED_ADDR
	attr.typename = parseAttributeType(attr.typevalue)

	if r.IP.To4() == nil {
		attr.value = make([]byte, 20)
	} else {
		attr.value = make([]byte, 8)
	}
	attr.length = len(attr.value)

	// first byte is 0
	attr.value[0] = 0x00

	// family
	attr.value[1] = 0x01
	if r.IP.To4() == nil {
		attr.value[1] = 0x02
	}

	// x-port
	port16 := uint16(r.Port)
	xor := port16 ^ (STUN_MSG_MAGIC_COOKIE >> 16)
	binary.BigEndian.PutUint16(attr.value[2:], xor)

	// x-address
	if r.IP.To4() == nil {
		// TODO ipv6 x-address
	} else {
		addr32 := binary.BigEndian.Uint32(r.IP)
		xor := addr32 ^ STUN_MSG_MAGIC_COOKIE
		binary.BigEndian.PutUint32(attr.value[4:], xor)
	}

	this.attributes = append(this.attributes, attr)
	return 4 + len(attr.value)
}

func (this *Message) Print(title string) {

	str := fmt.Sprintf("========== %s ==========\n", title)
	str += fmt.Sprintf("method=%s %s, length=%d bytes\n", this.methodName, this.encodingName, this.length)
	str += fmt.Sprintf("  transactionID=")
	for _, v := range this.transactionID {
		str += fmt.Sprintf("0x%02x ", v)
	}
	str += "\n"
	str += fmt.Sprintf("  attributes:\n")
	for _, v := range this.attributes {
		str += fmt.Sprintf("    type=0x%04x(%s), len=%d, value=%v\n", v.typevalue, v.typename, v.length, v.value)
	}
	fmt.Println(str)
}

func NewMessage(buf []byte) (*Message, error) {

	if err := checkMessage(buf); err != nil {
		return nil, fmt.Errorf("invalid stun msg: %s", err)
	}

	msg := &Message{}

	// get method and encoding
	msg.method = binary.BigEndian.Uint16(buf[0:]) & STUN_MSG_TYPE_METHOD_MASK
	msg.encoding = binary.BigEndian.Uint16(buf[0:]) & STUN_MSG_TYPE_ENCODING_MASK
	msg.methodName, msg.encodingName = parseMessageType(msg.method, msg.encoding)

	// get length
	msg.length = len(buf) - STUN_MSG_HEADER_SIZE

	// get transactionID
	msg.transactionID = append(msg.transactionID, buf[8:20]...)

	// get attributes
	for i := 0; i < msg.length; {
		attr := attribute{}

		// first 2 bytes are Type and Length
		attr.typevalue = binary.BigEndian.Uint16(buf[20+i:])
		attr.typename = parseAttributeType(attr.typevalue)
		len := int(binary.BigEndian.Uint16(buf[20+i+2:]))
		attr.length = len

		// following bytes are attributes
		attr.value = append(attr.value, buf[20+i+4:20+i+4+len]...)
		msg.attributes = append(msg.attributes, attr)

		// padding for 4 bytes per attribute item
		if len % 4 != 0 {
			i += len + (4 - len % 4) + 4
		} else {
			i += len + 4
		}
	}

	return msg, nil
}

func checkMessage(buf []byte) error {

	// STUN message len does not meet the min requirement
	if len(buf) < STUN_MSG_HEADER_SIZE {
		return fmt.Errorf("stun msg is too short: size=%d", len(buf))
	}

	// first byte should be 0x0 or 0x1
	if buf[0] != 0x00 && buf[0] != 0x01 {
		return fmt.Errorf("invalid stun msg type: first_byte=0x%02x", buf[0])
	}

	// check STUN message length
	msgLen := int(binary.BigEndian.Uint16(buf[2:]))
	if msgLen + 20 != len(buf) {
		return fmt.Errorf("msg length is not correct: len=%d, actual=%d", msgLen, len(buf))
	}

	// STUN message is always padded to a multiple of 4 bytes
	if msgLen & 0x03 != 0 {
		return fmt.Errorf("stun message is not aligned")
	}

	return nil
}

func parseMessageType(method, encoding uint16) (m string, e string) {

	switch (method) {
	case STUN_MSG_METHOD_BINDING:
		m = "binding"
		switch (encoding) {
		case STUN_MSG_BINDING_REQUEST: e = "request"
		case STUN_MSG_BINDING_INDICATION: e = "indication"
		case STUN_MSG_BINDING_SUCCESS: e = "success_response"
		case STUN_MSG_BINDING_ERROR: e = "error_response"
		}
	default:
		m = "user"
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
	}
	return "RESERVED"
}
