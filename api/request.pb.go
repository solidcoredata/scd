// Code generated by protoc-gen-go. DO NOT EDIT.
// source: request.proto

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

type HTTPRequest struct {
	Host string `protobuf:"bytes,1,opt,name=Host" json:"Host,omitempty"`
	// Method specifies the HTTP method (GET, POST, PUT, etc.).
	Method string `protobuf:"bytes,2,opt,name=Method" json:"Method,omitempty"`
	// URL specifies either the URI being requested.
	URL *URL `protobuf:"bytes,3,opt,name=URL" json:"URL,omitempty"`
	// The protocol version for incoming server requests.
	ProtoMajor  int32         `protobuf:"varint,4,opt,name=ProtoMajor" json:"ProtoMajor,omitempty"`
	ProtoMinor  int32         `protobuf:"varint,5,opt,name=ProtoMinor" json:"ProtoMinor,omitempty"`
	Header      *KeyValueList `protobuf:"bytes,6,opt,name=Header" json:"Header,omitempty"`
	Body        []byte        `protobuf:"bytes,7,opt,name=Body,proto3" json:"Body,omitempty"`
	ContentType string        `protobuf:"bytes,8,opt,name=ContentType" json:"ContentType,omitempty"`
	// RemoteAddr allows HTTP servers and other software to record
	// the network address that sent the request, usually for
	// logging. This field is not filled in by ReadRequest and
	// has no defined format. The HTTP server in this package
	// sets RemoteAddr to an "IP:port" address before invoking a
	// handler.
	RemoteAddr string           `protobuf:"bytes,9,opt,name=RemoteAddr" json:"RemoteAddr,omitempty"`
	TLS        *TLSState        `protobuf:"bytes,10,opt,name=TLS" json:"TLS,omitempty"`
	Auth       *RequestAuthResp `protobuf:"bytes,11,opt,name=Auth" json:"Auth,omitempty"`
	Config     *ConfigureURL    `protobuf:"bytes,12,opt,name=Config" json:"Config,omitempty"`
}

func (m *HTTPRequest) Reset()                    { *m = HTTPRequest{} }
func (m *HTTPRequest) String() string            { return proto.CompactTextString(m) }
func (*HTTPRequest) ProtoMessage()               {}
func (*HTTPRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func (m *HTTPRequest) GetHost() string {
	if m != nil {
		return m.Host
	}
	return ""
}

func (m *HTTPRequest) GetMethod() string {
	if m != nil {
		return m.Method
	}
	return ""
}

func (m *HTTPRequest) GetURL() *URL {
	if m != nil {
		return m.URL
	}
	return nil
}

func (m *HTTPRequest) GetProtoMajor() int32 {
	if m != nil {
		return m.ProtoMajor
	}
	return 0
}

func (m *HTTPRequest) GetProtoMinor() int32 {
	if m != nil {
		return m.ProtoMinor
	}
	return 0
}

func (m *HTTPRequest) GetHeader() *KeyValueList {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *HTTPRequest) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

func (m *HTTPRequest) GetContentType() string {
	if m != nil {
		return m.ContentType
	}
	return ""
}

func (m *HTTPRequest) GetRemoteAddr() string {
	if m != nil {
		return m.RemoteAddr
	}
	return ""
}

func (m *HTTPRequest) GetTLS() *TLSState {
	if m != nil {
		return m.TLS
	}
	return nil
}

func (m *HTTPRequest) GetAuth() *RequestAuthResp {
	if m != nil {
		return m.Auth
	}
	return nil
}

func (m *HTTPRequest) GetConfig() *ConfigureURL {
	if m != nil {
		return m.Config
	}
	return nil
}

type HTTPResponse struct {
	Header *KeyValueList `protobuf:"bytes,1,opt,name=Header" json:"Header,omitempty"`
	// Content type of the body.
	ContentType string `protobuf:"bytes,2,opt,name=ContentType" json:"ContentType,omitempty"`
	// Encoding of the response. Often a compression method like "gzip" or "br".
	Encoding string `protobuf:"bytes,3,opt,name=Encoding" json:"Encoding,omitempty"`
	Body     []byte `protobuf:"bytes,4,opt,name=Body,proto3" json:"Body,omitempty"`
}

func (m *HTTPResponse) Reset()                    { *m = HTTPResponse{} }
func (m *HTTPResponse) String() string            { return proto.CompactTextString(m) }
func (*HTTPResponse) ProtoMessage()               {}
func (*HTTPResponse) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

func (m *HTTPResponse) GetHeader() *KeyValueList {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *HTTPResponse) GetContentType() string {
	if m != nil {
		return m.ContentType
	}
	return ""
}

func (m *HTTPResponse) GetEncoding() string {
	if m != nil {
		return m.Encoding
	}
	return ""
}

func (m *HTTPResponse) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

type URL struct {
	Host  string        `protobuf:"bytes,1,opt,name=Host" json:"Host,omitempty"`
	Path  string        `protobuf:"bytes,2,opt,name=Path" json:"Path,omitempty"`
	Query *KeyValueList `protobuf:"bytes,3,opt,name=Query" json:"Query,omitempty"`
}

func (m *URL) Reset()                    { *m = URL{} }
func (m *URL) String() string            { return proto.CompactTextString(m) }
func (*URL) ProtoMessage()               {}
func (*URL) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{2} }

func (m *URL) GetHost() string {
	if m != nil {
		return m.Host
	}
	return ""
}

func (m *URL) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *URL) GetQuery() *KeyValueList {
	if m != nil {
		return m.Query
	}
	return nil
}

type StringList struct {
	Value []string `protobuf:"bytes,1,rep,name=Value" json:"Value,omitempty"`
}

func (m *StringList) Reset()                    { *m = StringList{} }
func (m *StringList) String() string            { return proto.CompactTextString(m) }
func (*StringList) ProtoMessage()               {}
func (*StringList) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{3} }

func (m *StringList) GetValue() []string {
	if m != nil {
		return m.Value
	}
	return nil
}

type KeyValueList struct {
	Values map[string]*StringList `protobuf:"bytes,1,rep,name=Values" json:"Values,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *KeyValueList) Reset()                    { *m = KeyValueList{} }
func (m *KeyValueList) String() string            { return proto.CompactTextString(m) }
func (*KeyValueList) ProtoMessage()               {}
func (*KeyValueList) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{4} }

func (m *KeyValueList) GetValues() map[string]*StringList {
	if m != nil {
		return m.Values
	}
	return nil
}

type TLSState struct {
	Version           uint32 `protobuf:"varint,1,opt,name=Version" json:"Version,omitempty"`
	HandshakeComplete bool   `protobuf:"varint,2,opt,name=HandshakeComplete" json:"HandshakeComplete,omitempty"`
	DidResume         bool   `protobuf:"varint,3,opt,name=DidResume" json:"DidResume,omitempty"`
	CipherSuite       uint32 `protobuf:"varint,4,opt,name=CipherSuite" json:"CipherSuite,omitempty"`
	ServerName        string `protobuf:"bytes,5,opt,name=ServerName" json:"ServerName,omitempty"`
}

func (m *TLSState) Reset()                    { *m = TLSState{} }
func (m *TLSState) String() string            { return proto.CompactTextString(m) }
func (*TLSState) ProtoMessage()               {}
func (*TLSState) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{5} }

func (m *TLSState) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *TLSState) GetHandshakeComplete() bool {
	if m != nil {
		return m.HandshakeComplete
	}
	return false
}

func (m *TLSState) GetDidResume() bool {
	if m != nil {
		return m.DidResume
	}
	return false
}

func (m *TLSState) GetCipherSuite() uint32 {
	if m != nil {
		return m.CipherSuite
	}
	return 0
}

func (m *TLSState) GetServerName() string {
	if m != nil {
		return m.ServerName
	}
	return ""
}

func init() {
	proto.RegisterType((*HTTPRequest)(nil), "api.HTTPRequest")
	proto.RegisterType((*HTTPResponse)(nil), "api.HTTPResponse")
	proto.RegisterType((*URL)(nil), "api.URL")
	proto.RegisterType((*StringList)(nil), "api.StringList")
	proto.RegisterType((*KeyValueList)(nil), "api.KeyValueList")
	proto.RegisterType((*TLSState)(nil), "api.TLSState")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for HTTP service

type HTTPClient interface {
	ServeHTTP(ctx context.Context, in *HTTPRequest, opts ...grpc.CallOption) (*HTTPResponse, error)
}

type hTTPClient struct {
	cc *grpc.ClientConn
}

func NewHTTPClient(cc *grpc.ClientConn) HTTPClient {
	return &hTTPClient{cc}
}

func (c *hTTPClient) ServeHTTP(ctx context.Context, in *HTTPRequest, opts ...grpc.CallOption) (*HTTPResponse, error) {
	out := new(HTTPResponse)
	err := grpc.Invoke(ctx, "/api.HTTP/ServeHTTP", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for HTTP service

type HTTPServer interface {
	ServeHTTP(context.Context, *HTTPRequest) (*HTTPResponse, error)
}

func RegisterHTTPServer(s *grpc.Server, srv HTTPServer) {
	s.RegisterService(&_HTTP_serviceDesc, srv)
}

func _HTTP_ServeHTTP_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HTTPRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HTTPServer).ServeHTTP(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.HTTP/ServeHTTP",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HTTPServer).ServeHTTP(ctx, req.(*HTTPRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _HTTP_serviceDesc = grpc.ServiceDesc{
	ServiceName: "api.HTTP",
	HandlerType: (*HTTPServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ServeHTTP",
			Handler:    _HTTP_ServeHTTP_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "request.proto",
}

func init() { proto.RegisterFile("request.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 571 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x94, 0x6d, 0x6b, 0xd3, 0x50,
	0x14, 0xc7, 0x49, 0xd3, 0x76, 0xcd, 0x49, 0x8b, 0xdb, 0x65, 0xc8, 0x25, 0xf8, 0x10, 0x02, 0x62,
	0x04, 0x29, 0x32, 0x11, 0x86, 0xef, 0xe6, 0x1c, 0x14, 0xed, 0x64, 0xde, 0x74, 0x7b, 0x1f, 0xcd,
	0x71, 0x89, 0x5b, 0x73, 0xe3, 0xcd, 0xcd, 0x20, 0x9f, 0x41, 0xdf, 0xf8, 0x55, 0xfc, 0x84, 0x72,
	0x4f, 0xb2, 0xf5, 0x62, 0x07, 0xbe, 0xbb, 0xe7, 0x77, 0xfe, 0x3d, 0xcf, 0x29, 0xcc, 0x14, 0xfe,
	0x68, 0xb0, 0xd6, 0xf3, 0x4a, 0x49, 0x2d, 0x99, 0x9b, 0x56, 0x45, 0x00, 0x69, 0xa3, 0xf3, 0x0e,
	0x04, 0x53, 0x25, 0x1b, 0x8d, 0xaa, 0xb3, 0xa2, 0x5f, 0x2e, 0xf8, 0x8b, 0xd5, 0xea, 0x4c, 0x74,
	0x3f, 0x62, 0x0c, 0x86, 0x0b, 0x59, 0x6b, 0xee, 0x84, 0x4e, 0xec, 0x09, 0x7a, 0xb3, 0x87, 0x30,
	0x3e, 0x45, 0x9d, 0xcb, 0x8c, 0x0f, 0x88, 0xf6, 0x16, 0x0b, 0xc0, 0x3d, 0x17, 0x4b, 0xee, 0x86,
	0x4e, 0xec, 0x1f, 0x4c, 0xe6, 0x69, 0x55, 0xcc, 0xcf, 0xc5, 0x52, 0x18, 0xc8, 0x9e, 0x00, 0x9c,
	0x99, 0x04, 0xa7, 0xe9, 0x77, 0xa9, 0xf8, 0x30, 0x74, 0xe2, 0x91, 0xb0, 0xc8, 0xc6, 0x5f, 0x94,
	0x52, 0xf1, 0x91, 0xed, 0x37, 0x84, 0xbd, 0x80, 0xf1, 0x02, 0xd3, 0x0c, 0x15, 0x1f, 0x53, 0xf8,
	0x3d, 0x0a, 0xff, 0x11, 0xdb, 0x8b, 0xf4, 0xba, 0xc1, 0x65, 0x51, 0x6b, 0xd1, 0x0b, 0x4c, 0xc9,
	0xef, 0x64, 0xd6, 0xf2, 0x9d, 0xd0, 0x89, 0xa7, 0x82, 0xde, 0x2c, 0x04, 0xff, 0x58, 0x96, 0x1a,
	0x4b, 0xbd, 0x6a, 0x2b, 0xe4, 0x13, 0xaa, 0xdb, 0x46, 0xa6, 0x00, 0x81, 0x6b, 0xa9, 0xf1, 0x28,
	0xcb, 0x14, 0xf7, 0x48, 0x60, 0x11, 0xf6, 0x14, 0xdc, 0xd5, 0x32, 0xe1, 0x40, 0xd9, 0x67, 0x94,
	0x7d, 0xb5, 0x4c, 0x12, 0x9d, 0x6a, 0x14, 0xc6, 0xc3, 0x62, 0x18, 0x1e, 0x35, 0x3a, 0xe7, 0x3e,
	0x29, 0xf6, 0x49, 0xd1, 0x4f, 0xd1, 0x70, 0x81, 0x75, 0x25, 0x48, 0x61, 0x7a, 0x39, 0x96, 0xe5,
	0xb7, 0xe2, 0x92, 0x4f, 0xad, 0x5e, 0x3a, 0xd4, 0x28, 0x34, 0x33, 0xeb, 0x05, 0xd1, 0x4f, 0x07,
	0xa6, 0xdd, 0x3a, 0xea, 0x4a, 0x96, 0x35, 0x5a, 0x73, 0x70, 0xfe, 0x37, 0x87, 0x7f, 0x7a, 0x1e,
	0x6c, 0xf7, 0x1c, 0xc0, 0xe4, 0xa4, 0xfc, 0x2a, 0xb3, 0xa2, 0xbc, 0xa4, 0xad, 0x79, 0xe2, 0xce,
	0xbe, 0x9b, 0xe2, 0x70, 0x33, 0xc5, 0xe8, 0x82, 0x16, 0x7c, 0xef, 0x4d, 0x30, 0x18, 0x9e, 0xa5,
	0x3a, 0xef, 0xb3, 0xd0, 0x9b, 0x3d, 0x87, 0xd1, 0xe7, 0x06, 0x55, 0xdb, 0x5f, 0xc4, 0x3d, 0xa5,
	0x76, 0xfe, 0x28, 0x02, 0x48, 0xb4, 0x2a, 0xca, 0x4b, 0x03, 0xd9, 0x3e, 0x8c, 0x48, 0xc1, 0x9d,
	0xd0, 0x8d, 0x3d, 0xd1, 0x19, 0xd1, 0x6f, 0x07, 0xa6, 0xf6, 0x6f, 0xd9, 0x1b, 0x18, 0x93, 0x51,
	0x93, 0xce, 0x3f, 0x78, 0xbc, 0x15, 0x7e, 0xde, 0xf9, 0x4f, 0x4a, 0xad, 0x5a, 0xd1, 0x8b, 0x83,
	0x0f, 0xe0, 0x5b, 0x98, 0xed, 0x82, 0x7b, 0x85, 0x6d, 0xdf, 0x8a, 0x79, 0xb2, 0x67, 0x30, 0xba,
	0xa1, 0xf4, 0x03, 0xaa, 0xfa, 0x01, 0x85, 0xdd, 0x94, 0x27, 0x3a, 0xef, 0xdb, 0xc1, 0xa1, 0x13,
	0xfd, 0x71, 0x60, 0x72, 0x7b, 0x04, 0x8c, 0xc3, 0xce, 0x05, 0xaa, 0xba, 0x90, 0x25, 0x45, 0x9b,
	0x89, 0x5b, 0x93, 0xbd, 0x84, 0xbd, 0x45, 0x5a, 0x66, 0x75, 0x9e, 0x5e, 0xe1, 0xb1, 0x5c, 0x57,
	0xd7, 0xa8, 0xbb, 0xe8, 0x13, 0xb1, 0xed, 0x60, 0x8f, 0xc0, 0x7b, 0x5f, 0x64, 0x02, 0xeb, 0x66,
	0x8d, 0x34, 0xb9, 0x89, 0xd8, 0x00, 0x5a, 0x6a, 0x51, 0xe5, 0xa8, 0x92, 0xa6, 0xd0, 0x48, 0xdb,
	0x99, 0x09, 0x1b, 0x99, 0x43, 0x4e, 0x50, 0xdd, 0xa0, 0xfa, 0x94, 0xae, 0x91, 0xbe, 0x24, 0x4f,
	0x58, 0xe4, 0xe0, 0x10, 0x86, 0xe6, 0xa2, 0xd8, 0x2b, 0xf0, 0x88, 0x92, 0xb1, 0x4b, 0x5d, 0x5a,
	0x1f, 0x7e, 0xb0, 0x67, 0x91, 0xee, 0xf6, 0xbe, 0x8c, 0xe9, 0x2f, 0xe2, 0xf5, 0xdf, 0x00, 0x00,
	0x00, 0xff, 0xff, 0xcb, 0xf9, 0x5b, 0x93, 0x52, 0x04, 0x00, 0x00,
}
