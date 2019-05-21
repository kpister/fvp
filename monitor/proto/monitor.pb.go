// Code generated by protoc-gen-go. DO NOT EDIT.
// source: monitor.proto

package monitor

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

func init() { proto.RegisterFile("monitor.proto", fileDescriptor_44174b7b2a306b71) }

var fileDescriptor_44174b7b2a306b71 = []byte{
	// 54 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xcd, 0xcd, 0xcf, 0xcb,
	0x2c, 0xc9, 0x2f, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x87, 0x72, 0x8d, 0x38, 0xb9,
	0xd8, 0x7d, 0x21, 0xcc, 0x24, 0x36, 0xb0, 0x94, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x1e, 0xec,
	0x7e, 0xf1, 0x2b, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MonitorClient is the client API for Monitor service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MonitorClient interface {
}

type monitorClient struct {
	cc *grpc.ClientConn
}

func NewMonitorClient(cc *grpc.ClientConn) MonitorClient {
	return &monitorClient{cc}
}

// MonitorServer is the server API for Monitor service.
type MonitorServer interface {
}

// UnimplementedMonitorServer can be embedded to have forward compatible implementations.
type UnimplementedMonitorServer struct {
}

func RegisterMonitorServer(s *grpc.Server, srv MonitorServer) {
	s.RegisterService(&_Monitor_serviceDesc, srv)
}

var _Monitor_serviceDesc = grpc.ServiceDesc{
	ServiceName: "monitor.Monitor",
	HandlerType: (*MonitorServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "monitor.proto",
}
