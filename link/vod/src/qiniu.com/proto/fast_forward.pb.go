// Code generated by protoc-gen-go. DO NOT EDIT.
// source: fast_forward.proto

package fastforward

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type FastForwardInfo struct {
	Baseurl              string   `protobuf:"bytes,1,opt,name=baseurl,proto3" json:"baseurl,omitempty"`
	From                 int64    `protobuf:"varint,2,opt,name=from,proto3" json:"from,omitempty"`
	To                   int64    `protobuf:"varint,3,opt,name=to,proto3" json:"to,omitempty"`
	Expire               int64    `protobuf:"varint,4,opt,name=expire,proto3" json:"expire,omitempty"`
	Token                string   `protobuf:"bytes,5,opt,name=token,proto3" json:"token,omitempty"`
	Speed                int32    `protobuf:"varint,6,opt,name=speed,proto3" json:"speed,omitempty"`
	ApiVerion            string   `protobuf:"bytes,7,opt,name=apiVerion,proto3" json:"apiVerion,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FastForwardInfo) Reset()         { *m = FastForwardInfo{} }
func (m *FastForwardInfo) String() string { return proto.CompactTextString(m) }
func (*FastForwardInfo) ProtoMessage()    {}
func (*FastForwardInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_fast_forward_4a113861de0bc42e, []int{0}
}
func (m *FastForwardInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FastForwardInfo.Unmarshal(m, b)
}
func (m *FastForwardInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FastForwardInfo.Marshal(b, m, deterministic)
}
func (dst *FastForwardInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FastForwardInfo.Merge(dst, src)
}
func (m *FastForwardInfo) XXX_Size() int {
	return xxx_messageInfo_FastForwardInfo.Size(m)
}
func (m *FastForwardInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_FastForwardInfo.DiscardUnknown(m)
}

var xxx_messageInfo_FastForwardInfo proto.InternalMessageInfo

func (m *FastForwardInfo) GetBaseurl() string {
	if m != nil {
		return m.Baseurl
	}
	return ""
}

func (m *FastForwardInfo) GetFrom() int64 {
	if m != nil {
		return m.From
	}
	return 0
}

func (m *FastForwardInfo) GetTo() int64 {
	if m != nil {
		return m.To
	}
	return 0
}

func (m *FastForwardInfo) GetExpire() int64 {
	if m != nil {
		return m.Expire
	}
	return 0
}

func (m *FastForwardInfo) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *FastForwardInfo) GetSpeed() int32 {
	if m != nil {
		return m.Speed
	}
	return 0
}

func (m *FastForwardInfo) GetApiVerion() string {
	if m != nil {
		return m.ApiVerion
	}
	return ""
}

type FastForwardStream struct {
	Stream               []byte   `protobuf:"bytes,1,opt,name=stream,proto3" json:"stream,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FastForwardStream) Reset()         { *m = FastForwardStream{} }
func (m *FastForwardStream) String() string { return proto.CompactTextString(m) }
func (*FastForwardStream) ProtoMessage()    {}
func (*FastForwardStream) Descriptor() ([]byte, []int) {
	return fileDescriptor_fast_forward_4a113861de0bc42e, []int{1}
}
func (m *FastForwardStream) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FastForwardStream.Unmarshal(m, b)
}
func (m *FastForwardStream) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FastForwardStream.Marshal(b, m, deterministic)
}
func (dst *FastForwardStream) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FastForwardStream.Merge(dst, src)
}
func (m *FastForwardStream) XXX_Size() int {
	return xxx_messageInfo_FastForwardStream.Size(m)
}
func (m *FastForwardStream) XXX_DiscardUnknown() {
	xxx_messageInfo_FastForwardStream.DiscardUnknown(m)
}

var xxx_messageInfo_FastForwardStream proto.InternalMessageInfo

func (m *FastForwardStream) GetStream() []byte {
	if m != nil {
		return m.Stream
	}
	return nil
}

func init() {
	proto.RegisterType((*FastForwardInfo)(nil), "fastforward.FastForwardInfo")
	proto.RegisterType((*FastForwardStream)(nil), "fastforward.FastForwardStream")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// FastForwardClient is the client API for FastForward service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type FastForwardClient interface {
	GetTsStream(ctx context.Context, in *FastForwardInfo, opts ...grpc.CallOption) (FastForward_GetTsStreamClient, error)
}

type fastForwardClient struct {
	cc *grpc.ClientConn
}

func NewFastForwardClient(cc *grpc.ClientConn) FastForwardClient {
	return &fastForwardClient{cc}
}

func (c *fastForwardClient) GetTsStream(ctx context.Context, in *FastForwardInfo, opts ...grpc.CallOption) (FastForward_GetTsStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &_FastForward_serviceDesc.Streams[0], "/fastforward.FastForward/GetTsStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &fastForwardGetTsStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type FastForward_GetTsStreamClient interface {
	Recv() (*FastForwardStream, error)
	grpc.ClientStream
}

type fastForwardGetTsStreamClient struct {
	grpc.ClientStream
}

func (x *fastForwardGetTsStreamClient) Recv() (*FastForwardStream, error) {
	m := new(FastForwardStream)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// FastForwardServer is the server API for FastForward service.
type FastForwardServer interface {
	GetTsStream(*FastForwardInfo, FastForward_GetTsStreamServer) error
}

func RegisterFastForwardServer(s *grpc.Server, srv FastForwardServer) {
	s.RegisterService(&_FastForward_serviceDesc, srv)
}

func _FastForward_GetTsStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(FastForwardInfo)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(FastForwardServer).GetTsStream(m, &fastForwardGetTsStreamServer{stream})
}

type FastForward_GetTsStreamServer interface {
	Send(*FastForwardStream) error
	grpc.ServerStream
}

type fastForwardGetTsStreamServer struct {
	grpc.ServerStream
}

func (x *fastForwardGetTsStreamServer) Send(m *FastForwardStream) error {
	return x.ServerStream.SendMsg(m)
}

var _FastForward_serviceDesc = grpc.ServiceDesc{
	ServiceName: "fastforward.FastForward",
	HandlerType: (*FastForwardServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetTsStream",
			Handler:       _FastForward_GetTsStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "fast_forward.proto",
}

func init() { proto.RegisterFile("fast_forward.proto", fileDescriptor_fast_forward_4a113861de0bc42e) }

var fileDescriptor_fast_forward_4a113861de0bc42e = []byte{
	// 237 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x90, 0xc1, 0x4a, 0xc4, 0x40,
	0x0c, 0x86, 0x9d, 0xee, 0xb6, 0xcb, 0xa6, 0xa2, 0x18, 0x44, 0x06, 0x59, 0xa4, 0xf4, 0x54, 0x10,
	0x8a, 0xe8, 0x3b, 0xac, 0x78, 0x12, 0xaa, 0x78, 0x54, 0x66, 0xd9, 0x14, 0x8a, 0xb6, 0x29, 0x33,
	0x11, 0x7d, 0x2d, 0xdf, 0x50, 0x9a, 0x56, 0x2c, 0x82, 0xb7, 0x7c, 0xdf, 0x24, 0x93, 0xf0, 0x03,
	0xd6, 0x2e, 0xc8, 0x4b, 0xcd, 0xfe, 0xc3, 0xf9, 0x7d, 0xd9, 0x7b, 0x16, 0xc6, 0x74, 0x70, 0x93,
	0xca, 0xbf, 0x0c, 0x1c, 0x6f, 0x5d, 0x90, 0xed, 0xc8, 0x77, 0x5d, 0xcd, 0x68, 0x61, 0xb5, 0x73,
	0x81, 0xde, 0xfd, 0x9b, 0x35, 0x99, 0x29, 0xd6, 0xd5, 0x0f, 0x22, 0xc2, 0xb2, 0xf6, 0xdc, 0xda,
	0x28, 0x33, 0xc5, 0xa2, 0xd2, 0x1a, 0x8f, 0x20, 0x12, 0xb6, 0x0b, 0x35, 0x91, 0x30, 0x9e, 0x41,
	0x42, 0x9f, 0x7d, 0xe3, 0xc9, 0x2e, 0xd5, 0x4d, 0x84, 0xa7, 0x10, 0x0b, 0xbf, 0x52, 0x67, 0x63,
	0xfd, 0x73, 0x84, 0xc1, 0x86, 0x9e, 0x68, 0x6f, 0x93, 0xcc, 0x14, 0x71, 0x35, 0x02, 0x6e, 0x60,
	0xed, 0xfa, 0xe6, 0x89, 0x7c, 0xc3, 0x9d, 0x5d, 0x69, 0xff, 0xaf, 0xc8, 0x2f, 0xe1, 0x64, 0x76,
	0xf2, 0x83, 0x78, 0x72, 0xed, 0xb0, 0x36, 0x68, 0xa5, 0x37, 0x1f, 0x56, 0x13, 0x5d, 0x3f, 0x43,
	0x3a, 0x6b, 0xc6, 0x7b, 0x48, 0x6f, 0x49, 0x1e, 0xc3, 0x34, 0xb5, 0x29, 0x67, 0x61, 0x94, 0x7f,
	0x82, 0x38, 0xbf, 0xf8, 0xef, 0x75, 0x9c, 0xce, 0x0f, 0xae, 0xcc, 0x2e, 0xd1, 0x50, 0x6f, 0xbe,
	0x03, 0x00, 0x00, 0xff, 0xff, 0x93, 0xe6, 0x4f, 0xd0, 0x6a, 0x01, 0x00, 0x00,
}
