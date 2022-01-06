// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package onexdataflowapi

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

// OpenapiClient is the client API for Openapi service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OpenapiClient interface {
	SetConfig(ctx context.Context, in *SetConfigRequest, opts ...grpc.CallOption) (*SetConfigResponse, error)
	GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error)
	RunExperiment(ctx context.Context, in *RunExperimentRequest, opts ...grpc.CallOption) (*RunExperimentResponse, error)
	GetMetrics(ctx context.Context, in *GetMetricsRequest, opts ...grpc.CallOption) (*GetMetricsResponse, error)
}

type openapiClient struct {
	cc grpc.ClientConnInterface
}

func NewOpenapiClient(cc grpc.ClientConnInterface) OpenapiClient {
	return &openapiClient{cc}
}

func (c *openapiClient) SetConfig(ctx context.Context, in *SetConfigRequest, opts ...grpc.CallOption) (*SetConfigResponse, error) {
	out := new(SetConfigResponse)
	err := c.cc.Invoke(ctx, "/onexdataflowapi.Openapi/SetConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *openapiClient) GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error) {
	out := new(GetConfigResponse)
	err := c.cc.Invoke(ctx, "/onexdataflowapi.Openapi/GetConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *openapiClient) RunExperiment(ctx context.Context, in *RunExperimentRequest, opts ...grpc.CallOption) (*RunExperimentResponse, error) {
	out := new(RunExperimentResponse)
	err := c.cc.Invoke(ctx, "/onexdataflowapi.Openapi/RunExperiment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *openapiClient) GetMetrics(ctx context.Context, in *GetMetricsRequest, opts ...grpc.CallOption) (*GetMetricsResponse, error) {
	out := new(GetMetricsResponse)
	err := c.cc.Invoke(ctx, "/onexdataflowapi.Openapi/GetMetrics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OpenapiServer is the server API for Openapi service.
// All implementations must embed UnimplementedOpenapiServer
// for forward compatibility
type OpenapiServer interface {
	SetConfig(context.Context, *SetConfigRequest) (*SetConfigResponse, error)
	GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error)
	RunExperiment(context.Context, *RunExperimentRequest) (*RunExperimentResponse, error)
	GetMetrics(context.Context, *GetMetricsRequest) (*GetMetricsResponse, error)
	mustEmbedUnimplementedOpenapiServer()
}

// UnimplementedOpenapiServer must be embedded to have forward compatible implementations.
type UnimplementedOpenapiServer struct {
}

func (UnimplementedOpenapiServer) SetConfig(context.Context, *SetConfigRequest) (*SetConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetConfig not implemented")
}
func (UnimplementedOpenapiServer) GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetConfig not implemented")
}
func (UnimplementedOpenapiServer) RunExperiment(context.Context, *RunExperimentRequest) (*RunExperimentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RunExperiment not implemented")
}
func (UnimplementedOpenapiServer) GetMetrics(context.Context, *GetMetricsRequest) (*GetMetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMetrics not implemented")
}
func (UnimplementedOpenapiServer) mustEmbedUnimplementedOpenapiServer() {}

// UnsafeOpenapiServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OpenapiServer will
// result in compilation errors.
type UnsafeOpenapiServer interface {
	mustEmbedUnimplementedOpenapiServer()
}

func RegisterOpenapiServer(s grpc.ServiceRegistrar, srv OpenapiServer) {
	s.RegisterService(&Openapi_ServiceDesc, srv)
}

func _Openapi_SetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OpenapiServer).SetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/onexdataflowapi.Openapi/SetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OpenapiServer).SetConfig(ctx, req.(*SetConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Openapi_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OpenapiServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/onexdataflowapi.Openapi/GetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OpenapiServer).GetConfig(ctx, req.(*GetConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Openapi_RunExperiment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RunExperimentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OpenapiServer).RunExperiment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/onexdataflowapi.Openapi/RunExperiment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OpenapiServer).RunExperiment(ctx, req.(*RunExperimentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Openapi_GetMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OpenapiServer).GetMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/onexdataflowapi.Openapi/GetMetrics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OpenapiServer).GetMetrics(ctx, req.(*GetMetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Openapi_ServiceDesc is the grpc.ServiceDesc for Openapi service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Openapi_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "onexdataflowapi.Openapi",
	HandlerType: (*OpenapiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetConfig",
			Handler:    _Openapi_SetConfig_Handler,
		},
		{
			MethodName: "GetConfig",
			Handler:    _Openapi_GetConfig_Handler,
		},
		{
			MethodName: "RunExperiment",
			Handler:    _Openapi_RunExperiment_Handler,
		},
		{
			MethodName: "GetMetrics",
			Handler:    _Openapi_GetMetrics_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "onexdataflowapi.proto",
}
