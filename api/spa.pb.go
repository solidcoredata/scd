// Code generated by protoc-gen-go. DO NOT EDIT.
// source: spa.proto

package api

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

type FetchUIAction int32

const (
	FetchUIAction_ActionMissing FetchUIAction = 0
	FetchUIAction_ActionStore   FetchUIAction = 1
	FetchUIAction_ActionExecute FetchUIAction = 2
)

var FetchUIAction_name = map[int32]string{
	0: "ActionMissing",
	1: "ActionStore",
	2: "ActionExecute",
}
var FetchUIAction_value = map[string]int32{
	"ActionMissing": 0,
	"ActionStore":   1,
	"ActionExecute": 2,
}

func (x FetchUIAction) String() string {
	return proto.EnumName(FetchUIAction_name, int32(x))
}
func (FetchUIAction) EnumDescriptor() ([]byte, []int) { return fileDescriptor3, []int{0} }

type FetchUIRequest struct {
	List []*FetchUICN `protobuf:"bytes,1,rep,name=List" json:"List,omitempty"`
}

func (m *FetchUIRequest) Reset()                    { *m = FetchUIRequest{} }
func (m *FetchUIRequest) String() string            { return proto.CompactTextString(m) }
func (*FetchUIRequest) ProtoMessage()               {}
func (*FetchUIRequest) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{0} }

func (m *FetchUIRequest) GetList() []*FetchUICN {
	if m != nil {
		return m.List
	}
	return nil
}

type FetchUICN struct {
	Category string `protobuf:"bytes,1,opt,name=Category" json:"Category,omitempty"`
	Name     string `protobuf:"bytes,2,opt,name=Name" json:"Name,omitempty"`
}

func (m *FetchUICN) Reset()                    { *m = FetchUICN{} }
func (m *FetchUICN) String() string            { return proto.CompactTextString(m) }
func (*FetchUICN) ProtoMessage()               {}
func (*FetchUICN) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{1} }

func (m *FetchUICN) GetCategory() string {
	if m != nil {
		return m.Category
	}
	return ""
}

func (m *FetchUICN) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type FetchUIResponse struct {
	List []*FetchUIItem `protobuf:"bytes,1,rep,name=List" json:"List,omitempty"`
}

func (m *FetchUIResponse) Reset()                    { *m = FetchUIResponse{} }
func (m *FetchUIResponse) String() string            { return proto.CompactTextString(m) }
func (*FetchUIResponse) ProtoMessage()               {}
func (*FetchUIResponse) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{2} }

func (m *FetchUIResponse) GetList() []*FetchUIItem {
	if m != nil {
		return m.List
	}
	return nil
}

type FetchUIItem struct {
	Action   FetchUIAction `protobuf:"varint,1,opt,name=Action,enum=api.FetchUIAction" json:"Action,omitempty"`
	Category string        `protobuf:"bytes,2,opt,name=Category" json:"Category,omitempty"`
	Name     string        `protobuf:"bytes,3,opt,name=Name" json:"Name,omitempty"`
	Require  []*FetchUICN  `protobuf:"bytes,4,rep,name=Require" json:"Require,omitempty"`
	Body     string        `protobuf:"bytes,5,opt,name=Body" json:"Body,omitempty"`
}

func (m *FetchUIItem) Reset()                    { *m = FetchUIItem{} }
func (m *FetchUIItem) String() string            { return proto.CompactTextString(m) }
func (*FetchUIItem) ProtoMessage()               {}
func (*FetchUIItem) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{3} }

func (m *FetchUIItem) GetAction() FetchUIAction {
	if m != nil {
		return m.Action
	}
	return FetchUIAction_ActionMissing
}

func (m *FetchUIItem) GetCategory() string {
	if m != nil {
		return m.Category
	}
	return ""
}

func (m *FetchUIItem) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *FetchUIItem) GetRequire() []*FetchUICN {
	if m != nil {
		return m.Require
	}
	return nil
}

func (m *FetchUIItem) GetBody() string {
	if m != nil {
		return m.Body
	}
	return ""
}

func init() {
	proto.RegisterType((*FetchUIRequest)(nil), "api.FetchUIRequest")
	proto.RegisterType((*FetchUICN)(nil), "api.FetchUICN")
	proto.RegisterType((*FetchUIResponse)(nil), "api.FetchUIResponse")
	proto.RegisterType((*FetchUIItem)(nil), "api.FetchUIItem")
	proto.RegisterEnum("api.FetchUIAction", FetchUIAction_name, FetchUIAction_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for SPA service

type SPAClient interface {
	// TODO: RequestAuth and Login should both take some additional features
	// about where the request is coming from (HTTPS info, remote address).
	FetchUI(ctx context.Context, in *FetchUIRequest, opts ...grpc.CallOption) (*FetchUIResponse, error)
}

type sPAClient struct {
	cc *grpc.ClientConn
}

func NewSPAClient(cc *grpc.ClientConn) SPAClient {
	return &sPAClient{cc}
}

func (c *sPAClient) FetchUI(ctx context.Context, in *FetchUIRequest, opts ...grpc.CallOption) (*FetchUIResponse, error) {
	out := new(FetchUIResponse)
	err := grpc.Invoke(ctx, "/api.SPA/FetchUI", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for SPA service

type SPAServer interface {
	// TODO: RequestAuth and Login should both take some additional features
	// about where the request is coming from (HTTPS info, remote address).
	FetchUI(context.Context, *FetchUIRequest) (*FetchUIResponse, error)
}

func RegisterSPAServer(s *grpc.Server, srv SPAServer) {
	s.RegisterService(&_SPA_serviceDesc, srv)
}

func _SPA_FetchUI_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FetchUIRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SPAServer).FetchUI(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.SPA/FetchUI",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SPAServer).FetchUI(ctx, req.(*FetchUIRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _SPA_serviceDesc = grpc.ServiceDesc{
	ServiceName: "api.SPA",
	HandlerType: (*SPAServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FetchUI",
			Handler:    _SPA_FetchUI_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "spa.proto",
}

func init() { proto.RegisterFile("spa.proto", fileDescriptor3) }

var fileDescriptor3 = []byte{
	// 289 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x51, 0x4f, 0x4b, 0xfb, 0x40,
	0x10, 0xfd, 0xe5, 0xcf, 0xaf, 0x35, 0x13, 0x9a, 0xc6, 0xd1, 0x43, 0xe8, 0xa9, 0x04, 0x0f, 0xa1,
	0x87, 0x1c, 0x62, 0xc1, 0x43, 0x4f, 0xb5, 0x58, 0x28, 0x68, 0x91, 0x14, 0x3f, 0x40, 0x8c, 0x43,
	0xdd, 0x43, 0xb3, 0x31, 0xbb, 0x05, 0xfb, 0x89, 0xfc, 0x9a, 0x92, 0xec, 0x9a, 0x26, 0xa0, 0xb7,
	0xd9, 0x37, 0xef, 0xed, 0x7b, 0xbc, 0x01, 0x47, 0x94, 0x59, 0x5c, 0x56, 0x5c, 0x72, 0xb4, 0xb2,
	0x92, 0x85, 0x73, 0xf0, 0xd6, 0x24, 0xf3, 0xf7, 0x97, 0x4d, 0x4a, 0x1f, 0x47, 0x12, 0x12, 0x43,
	0xb0, 0x1f, 0x99, 0x90, 0x81, 0x31, 0xb5, 0x22, 0x37, 0xf1, 0xe2, 0xac, 0x64, 0xb1, 0xa6, 0xac,
	0xb6, 0x69, 0xb3, 0x0b, 0x17, 0xe0, 0xb4, 0x10, 0x4e, 0xe0, 0x62, 0x95, 0x49, 0xda, 0xf3, 0xea,
	0x14, 0x18, 0x53, 0x23, 0x72, 0xd2, 0xf6, 0x8d, 0x08, 0xf6, 0x36, 0x3b, 0x50, 0x60, 0x36, 0x78,
	0x33, 0x87, 0x77, 0x30, 0x6e, 0x2d, 0x45, 0xc9, 0x0b, 0x41, 0x78, 0xd3, 0xf3, 0xf4, 0xbb, 0x9e,
	0x1b, 0x49, 0x07, 0xed, 0xfa, 0x65, 0x80, 0xdb, 0x41, 0x71, 0x06, 0x83, 0x65, 0x2e, 0x19, 0x2f,
	0x1a, 0x5b, 0x2f, 0xc1, 0xae, 0x4e, 0x6d, 0x52, 0xcd, 0xe8, 0x85, 0x34, 0xff, 0x08, 0x69, 0x9d,
	0x43, 0x62, 0x04, 0xc3, 0xba, 0x10, 0x56, 0x51, 0x60, 0xff, 0x5a, 0xc4, 0xcf, 0xba, 0x56, 0xdf,
	0xf3, 0xb7, 0x53, 0xf0, 0x5f, 0xa9, 0xeb, 0x79, 0xb6, 0x86, 0x51, 0x2f, 0x06, 0x5e, 0xc2, 0x48,
	0x4d, 0x4f, 0x4c, 0x08, 0x56, 0xec, 0xfd, 0x7f, 0x38, 0x06, 0x57, 0x41, 0x3b, 0xc9, 0x2b, 0xf2,
	0x8d, 0x33, 0xe7, 0xe1, 0x93, 0xf2, 0xa3, 0x24, 0xdf, 0x4c, 0x16, 0x60, 0xed, 0x9e, 0x97, 0x38,
	0x87, 0xa1, 0xfe, 0x0e, 0xaf, 0xba, 0x31, 0xf4, 0xc9, 0x26, 0xd7, 0x7d, 0x50, 0x95, 0xfa, 0x3a,
	0x68, 0xce, 0x7c, 0xfb, 0x1d, 0x00, 0x00, 0xff, 0xff, 0x58, 0x5a, 0x6a, 0x7d, 0xf3, 0x01, 0x00,
	0x00,
}
