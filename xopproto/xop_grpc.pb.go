// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.7
// source: xop.proto

package xopproto

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

// IngestClient is the client API for Ingest service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IngestClient interface {
	Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error)
	UploadFragment(ctx context.Context, in *IngestFragment, opts ...grpc.CallOption) (*ErrorResponse, error)
}

type ingestClient struct {
	cc grpc.ClientConnInterface
}

func NewIngestClient(cc grpc.ClientConnInterface) IngestClient {
	return &ingestClient{cc}
}

func (c *ingestClient) Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/xop.Ingest/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ingestClient) UploadFragment(ctx context.Context, in *IngestFragment, opts ...grpc.CallOption) (*ErrorResponse, error) {
	out := new(ErrorResponse)
	err := c.cc.Invoke(ctx, "/xop.Ingest/UploadFragment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IngestServer is the server API for Ingest service.
// All implementations must embed UnimplementedIngestServer
// for forward compatibility
type IngestServer interface {
	Ping(context.Context, *Empty) (*Empty, error)
	UploadFragment(context.Context, *IngestFragment) (*ErrorResponse, error)
	mustEmbedUnimplementedIngestServer()
}

// UnimplementedIngestServer must be embedded to have forward compatible implementations.
type UnimplementedIngestServer struct {
}

func (UnimplementedIngestServer) Ping(context.Context, *Empty) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedIngestServer) UploadFragment(context.Context, *IngestFragment) (*ErrorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UploadFragment not implemented")
}
func (UnimplementedIngestServer) mustEmbedUnimplementedIngestServer() {}

// UnsafeIngestServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IngestServer will
// result in compilation errors.
type UnsafeIngestServer interface {
	mustEmbedUnimplementedIngestServer()
}

func RegisterIngestServer(s grpc.ServiceRegistrar, srv IngestServer) {
	s.RegisterService(&Ingest_ServiceDesc, srv)
}

func _Ingest_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IngestServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xop.Ingest/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IngestServer).Ping(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ingest_UploadFragment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IngestFragment)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IngestServer).UploadFragment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xop.Ingest/UploadFragment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IngestServer).UploadFragment(ctx, req.(*IngestFragment))
	}
	return interceptor(ctx, in, info, handler)
}

// Ingest_ServiceDesc is the grpc.ServiceDesc for Ingest service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Ingest_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "xop.Ingest",
	HandlerType: (*IngestServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Ingest_Ping_Handler,
		},
		{
			MethodName: "UploadFragment",
			Handler:    _Ingest_UploadFragment_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "xop.proto",
}
