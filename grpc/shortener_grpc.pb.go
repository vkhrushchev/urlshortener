// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.1
// source: grpc/shortener.proto

package grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	ShortenerService_CreateShortURL_FullMethodName             = "/grpc.ShortenerService/CreateShortURL"
	ShortenerService_GetShortURL_FullMethodName                = "/grpc.ShortenerService/GetShortURL"
	ShortenerService_CreateShortURLBatch_FullMethodName        = "/grpc.ShortenerService/CreateShortURLBatch"
	ShortenerService_GetShortURLByUserID_FullMethodName        = "/grpc.ShortenerService/GetShortURLByUserID"
	ShortenerService_DeleteShortURLsByShortURIs_FullMethodName = "/grpc.ShortenerService/DeleteShortURLsByShortURIs"
	ShortenerService_Ping_FullMethodName                       = "/grpc.ShortenerService/Ping"
	ShortenerService_GetStats_FullMethodName                   = "/grpc.ShortenerService/GetStats"
)

// ShortenerServiceClient is the client API for ShortenerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ShortenerServiceClient interface {
	CreateShortURL(ctx context.Context, in *CreateShortURLRequest, opts ...grpc.CallOption) (*CreateShortURLResponse, error)
	GetShortURL(ctx context.Context, in *GetShortURLRequest, opts ...grpc.CallOption) (*GetShortURLResponse, error)
	CreateShortURLBatch(ctx context.Context, in *CreateShortURLBatchRequest, opts ...grpc.CallOption) (*CreateShortURLBatchResponse, error)
	GetShortURLByUserID(ctx context.Context, in *GetShortURLsByUserIDRequest, opts ...grpc.CallOption) (*GetShortURLsByUserIDResponse, error)
	DeleteShortURLsByShortURIs(ctx context.Context, in *DeleteShortURLsByShortURIsRequest, opts ...grpc.CallOption) (*DeleteShortURLsByShortURIsResponse, error)
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	GetStats(ctx context.Context, in *GetStatsRequest, opts ...grpc.CallOption) (*GetStatsResponse, error)
}

type shortenerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewShortenerServiceClient(cc grpc.ClientConnInterface) ShortenerServiceClient {
	return &shortenerServiceClient{cc}
}

func (c *shortenerServiceClient) CreateShortURL(ctx context.Context, in *CreateShortURLRequest, opts ...grpc.CallOption) (*CreateShortURLResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateShortURLResponse)
	err := c.cc.Invoke(ctx, ShortenerService_CreateShortURL_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) GetShortURL(ctx context.Context, in *GetShortURLRequest, opts ...grpc.CallOption) (*GetShortURLResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetShortURLResponse)
	err := c.cc.Invoke(ctx, ShortenerService_GetShortURL_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) CreateShortURLBatch(ctx context.Context, in *CreateShortURLBatchRequest, opts ...grpc.CallOption) (*CreateShortURLBatchResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateShortURLBatchResponse)
	err := c.cc.Invoke(ctx, ShortenerService_CreateShortURLBatch_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) GetShortURLByUserID(ctx context.Context, in *GetShortURLsByUserIDRequest, opts ...grpc.CallOption) (*GetShortURLsByUserIDResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetShortURLsByUserIDResponse)
	err := c.cc.Invoke(ctx, ShortenerService_GetShortURLByUserID_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) DeleteShortURLsByShortURIs(ctx context.Context, in *DeleteShortURLsByShortURIsRequest, opts ...grpc.CallOption) (*DeleteShortURLsByShortURIsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteShortURLsByShortURIsResponse)
	err := c.cc.Invoke(ctx, ShortenerService_DeleteShortURLsByShortURIs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, ShortenerService_Ping_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) GetStats(ctx context.Context, in *GetStatsRequest, opts ...grpc.CallOption) (*GetStatsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetStatsResponse)
	err := c.cc.Invoke(ctx, ShortenerService_GetStats_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ShortenerServiceServer is the server API for ShortenerService service.
// All implementations must embed UnimplementedShortenerServiceServer
// for forward compatibility.
type ShortenerServiceServer interface {
	CreateShortURL(context.Context, *CreateShortURLRequest) (*CreateShortURLResponse, error)
	GetShortURL(context.Context, *GetShortURLRequest) (*GetShortURLResponse, error)
	CreateShortURLBatch(context.Context, *CreateShortURLBatchRequest) (*CreateShortURLBatchResponse, error)
	GetShortURLByUserID(context.Context, *GetShortURLsByUserIDRequest) (*GetShortURLsByUserIDResponse, error)
	DeleteShortURLsByShortURIs(context.Context, *DeleteShortURLsByShortURIsRequest) (*DeleteShortURLsByShortURIsResponse, error)
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	GetStats(context.Context, *GetStatsRequest) (*GetStatsResponse, error)
	mustEmbedUnimplementedShortenerServiceServer()
}

// UnimplementedShortenerServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedShortenerServiceServer struct{}

func (UnimplementedShortenerServiceServer) CreateShortURL(context.Context, *CreateShortURLRequest) (*CreateShortURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateShortURL not implemented")
}
func (UnimplementedShortenerServiceServer) GetShortURL(context.Context, *GetShortURLRequest) (*GetShortURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetShortURL not implemented")
}
func (UnimplementedShortenerServiceServer) CreateShortURLBatch(context.Context, *CreateShortURLBatchRequest) (*CreateShortURLBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateShortURLBatch not implemented")
}
func (UnimplementedShortenerServiceServer) GetShortURLByUserID(context.Context, *GetShortURLsByUserIDRequest) (*GetShortURLsByUserIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetShortURLByUserID not implemented")
}
func (UnimplementedShortenerServiceServer) DeleteShortURLsByShortURIs(context.Context, *DeleteShortURLsByShortURIsRequest) (*DeleteShortURLsByShortURIsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteShortURLsByShortURIs not implemented")
}
func (UnimplementedShortenerServiceServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedShortenerServiceServer) GetStats(context.Context, *GetStatsRequest) (*GetStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStats not implemented")
}
func (UnimplementedShortenerServiceServer) mustEmbedUnimplementedShortenerServiceServer() {}
func (UnimplementedShortenerServiceServer) testEmbeddedByValue()                          {}

// UnsafeShortenerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ShortenerServiceServer will
// result in compilation errors.
type UnsafeShortenerServiceServer interface {
	mustEmbedUnimplementedShortenerServiceServer()
}

func RegisterShortenerServiceServer(s grpc.ServiceRegistrar, srv ShortenerServiceServer) {
	// If the following call pancis, it indicates UnimplementedShortenerServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ShortenerService_ServiceDesc, srv)
}

func _ShortenerService_CreateShortURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateShortURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).CreateShortURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_CreateShortURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).CreateShortURL(ctx, req.(*CreateShortURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_GetShortURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetShortURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).GetShortURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_GetShortURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).GetShortURL(ctx, req.(*GetShortURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_CreateShortURLBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateShortURLBatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).CreateShortURLBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_CreateShortURLBatch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).CreateShortURLBatch(ctx, req.(*CreateShortURLBatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_GetShortURLByUserID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetShortURLsByUserIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).GetShortURLByUserID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_GetShortURLByUserID_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).GetShortURLByUserID(ctx, req.(*GetShortURLsByUserIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_DeleteShortURLsByShortURIs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteShortURLsByShortURIsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).DeleteShortURLsByShortURIs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_DeleteShortURLsByShortURIs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).DeleteShortURLsByShortURIs(ctx, req.(*DeleteShortURLsByShortURIsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_GetStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).GetStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_GetStats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).GetStats(ctx, req.(*GetStatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ShortenerService_ServiceDesc is the grpc.ServiceDesc for ShortenerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ShortenerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.ShortenerService",
	HandlerType: (*ShortenerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateShortURL",
			Handler:    _ShortenerService_CreateShortURL_Handler,
		},
		{
			MethodName: "GetShortURL",
			Handler:    _ShortenerService_GetShortURL_Handler,
		},
		{
			MethodName: "CreateShortURLBatch",
			Handler:    _ShortenerService_CreateShortURLBatch_Handler,
		},
		{
			MethodName: "GetShortURLByUserID",
			Handler:    _ShortenerService_GetShortURLByUserID_Handler,
		},
		{
			MethodName: "DeleteShortURLsByShortURIs",
			Handler:    _ShortenerService_DeleteShortURLsByShortURIs_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _ShortenerService_Ping_Handler,
		},
		{
			MethodName: "GetStats",
			Handler:    _ShortenerService_GetStats_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc/shortener.proto",
}
