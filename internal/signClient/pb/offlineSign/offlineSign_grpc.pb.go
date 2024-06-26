// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: offlineSign.proto

package offlineSign

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// OfflineSignClient is the client API for OfflineSign service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OfflineSignClient interface {
	BtcSign(ctx context.Context, in *BtcSignReq, opts ...grpc.CallOption) (*BtcSignResp, error)
	EthSign(ctx context.Context, in *EthSignReq, opts ...grpc.CallOption) (*EthSignResp, error)
}

type offlineSignClient struct {
	cc grpc.ClientConnInterface
}

func NewOfflineSignClient(cc grpc.ClientConnInterface) OfflineSignClient {
	return &offlineSignClient{cc}
}

func (c *offlineSignClient) BtcSign(ctx context.Context, in *BtcSignReq, opts ...grpc.CallOption) (*BtcSignResp, error) {
	out := new(BtcSignResp)
	err := c.cc.Invoke(ctx, "/OfflineSign/BtcSign", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *offlineSignClient) EthSign(ctx context.Context, in *EthSignReq, opts ...grpc.CallOption) (*EthSignResp, error) {
	out := new(EthSignResp)
	err := c.cc.Invoke(ctx, "/OfflineSign/EthSign", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OfflineSignServer is the server API for OfflineSign service.
// All implementations must embed UnimplementedOfflineSignServer
// for forward compatibility
type OfflineSignServer interface {
	BtcSign(context.Context, *BtcSignReq) (*BtcSignResp, error)
	EthSign(context.Context, *EthSignReq) (*EthSignResp, error)
	mustEmbedUnimplementedOfflineSignServer()
}

// UnimplementedOfflineSignServer must be embedded to have forward compatible implementations.
type UnimplementedOfflineSignServer struct {
}

func (UnimplementedOfflineSignServer) BtcSign(context.Context, *BtcSignReq) (*BtcSignResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BtcSign not implemented")
}
func (UnimplementedOfflineSignServer) EthSign(context.Context, *EthSignReq) (*EthSignResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EthSign not implemented")
}
func (UnimplementedOfflineSignServer) mustEmbedUnimplementedOfflineSignServer() {}

// UnsafeOfflineSignServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OfflineSignServer will
// result in compilation errors.
type UnsafeOfflineSignServer interface {
	mustEmbedUnimplementedOfflineSignServer()
}

func RegisterOfflineSignServer(s grpc.ServiceRegistrar, srv OfflineSignServer) {
	s.RegisterService(&OfflineSign_ServiceDesc, srv)
}

func _OfflineSign_BtcSign_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BtcSignReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OfflineSignServer).BtcSign(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/OfflineSign/BtcSign",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OfflineSignServer).BtcSign(ctx, req.(*BtcSignReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _OfflineSign_EthSign_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EthSignReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OfflineSignServer).EthSign(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/OfflineSign/EthSign",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OfflineSignServer).EthSign(ctx, req.(*EthSignReq))
	}
	return interceptor(ctx, in, info, handler)
}

// OfflineSign_ServiceDesc is the grpc.ServiceDesc for OfflineSign service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var OfflineSign_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "OfflineSign",
	HandlerType: (*OfflineSignServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BtcSign",
			Handler:    _OfflineSign_BtcSign_Handler,
		},
		{
			MethodName: "EthSign",
			Handler:    _OfflineSign_EthSign_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "offlineSign.proto",
}
