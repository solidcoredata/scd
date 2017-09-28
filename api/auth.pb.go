// Code generated by protoc-gen-go. DO NOT EDIT.
// source: auth.proto

/*
Package api is a generated protocol buffer package.

It is generated from these files:
	auth.proto
	request.proto
	router.proto
	spa.proto

It has these top-level messages:
	ConfigureAuth
	RequestAuthResp
	RequestAuthReq
	LoginReq
	LoginResp
	LogoutReq
	LogoutResp
	NewPasswordReq
	NewPasswordResp
	ChangePasswordReq
	ChangePasswordResp
	ConfigureURL
	HTTPRequest
	HTTPResponse
	URL
	StringList
	KeyValueList
	TLSState
	NotifyReq
	UpdateReq
	UpdateResp
	ServiceConfigEndpoint
	ServiceConfig
	Resource
	LoginBundle
	ApplicationBundle
	ServiceBundle
	FetchUIRequest
	FetchUIResponse
	FetchUIItem
*/
package api

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/timestamp"

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

type LoginState int32

const (
	LoginState_Missing        LoginState = 0
	LoginState_Error          LoginState = 1
	LoginState_None           LoginState = 2
	LoginState_Granted        LoginState = 3
	LoginState_U2F            LoginState = 4
	LoginState_ChangePassword LoginState = 5
)

var LoginState_name = map[int32]string{
	0: "Missing",
	1: "Error",
	2: "None",
	3: "Granted",
	4: "U2F",
	5: "ChangePassword",
}
var LoginState_value = map[string]int32{
	"Missing":        0,
	"Error":          1,
	"None":           2,
	"Granted":        3,
	"U2F":            4,
	"ChangePassword": 5,
}

func (x LoginState) String() string {
	return proto.EnumName(LoginState_name, int32(x))
}
func (LoginState) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type ConfigureAuth_AreaType int32

const (
	ConfigureAuth_Unknown ConfigureAuth_AreaType = 0
	ConfigureAuth_System  ConfigureAuth_AreaType = 1
	ConfigureAuth_User    ConfigureAuth_AreaType = 2
)

var ConfigureAuth_AreaType_name = map[int32]string{
	0: "Unknown",
	1: "System",
	2: "User",
}
var ConfigureAuth_AreaType_value = map[string]int32{
	"Unknown": 0,
	"System":  1,
	"User":    2,
}

func (x ConfigureAuth_AreaType) String() string {
	return proto.EnumName(ConfigureAuth_AreaType_name, int32(x))
}
func (ConfigureAuth_AreaType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type ConfigureAuth struct {
	Area ConfigureAuth_AreaType `protobuf:"varint,1,opt,name=Area,enum=api.ConfigureAuth_AreaType" json:"Area,omitempty"`
	// Environment name.
	//   When AreaType=System, "QA" or "PROD".
	// /  When AreaType=User, "user-1" or "bobsmith".
	Environment string `protobuf:"bytes,2,opt,name=Environment" json:"Environment,omitempty"`
}

func (m *ConfigureAuth) Reset()                    { *m = ConfigureAuth{} }
func (m *ConfigureAuth) String() string            { return proto.CompactTextString(m) }
func (*ConfigureAuth) ProtoMessage()               {}
func (*ConfigureAuth) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *ConfigureAuth) GetArea() ConfigureAuth_AreaType {
	if m != nil {
		return m.Area
	}
	return ConfigureAuth_Unknown
}

func (m *ConfigureAuth) GetEnvironment() string {
	if m != nil {
		return m.Environment
	}
	return ""
}

type RequestAuthResp struct {
	LoginState    LoginState                 `protobuf:"varint,1,opt,name=LoginState,enum=api.LoginState" json:"LoginState,omitempty"`
	ID            int64                      `protobuf:"varint,2,opt,name=ID" json:"ID,omitempty"`
	Identity      string                     `protobuf:"bytes,3,opt,name=Identity" json:"Identity,omitempty"`
	Roles         []int64                    `protobuf:"varint,4,rep,packed,name=Roles" json:"Roles,omitempty"`
	ValidUntil    *google_protobuf.Timestamp `protobuf:"bytes,5,opt,name=ValidUntil" json:"ValidUntil,omitempty"`
	ElevatedUntil *google_protobuf.Timestamp `protobuf:"bytes,6,opt,name=ElevatedUntil" json:"ElevatedUntil,omitempty"`
	GivenName     string                     `protobuf:"bytes,7,opt,name=GivenName" json:"GivenName,omitempty"`
	FamilyName    string                     `protobuf:"bytes,8,opt,name=FamilyName" json:"FamilyName,omitempty"`
	Email         string                     `protobuf:"bytes,9,opt,name=Email" json:"Email,omitempty"`
	TokenKey      string                     `protobuf:"bytes,10,opt,name=TokenKey" json:"TokenKey,omitempty"`
	Secondary     *RequestAuthResp           `protobuf:"bytes,11,opt,name=Secondary" json:"Secondary,omitempty"`
}

func (m *RequestAuthResp) Reset()                    { *m = RequestAuthResp{} }
func (m *RequestAuthResp) String() string            { return proto.CompactTextString(m) }
func (*RequestAuthResp) ProtoMessage()               {}
func (*RequestAuthResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *RequestAuthResp) GetLoginState() LoginState {
	if m != nil {
		return m.LoginState
	}
	return LoginState_Missing
}

func (m *RequestAuthResp) GetID() int64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *RequestAuthResp) GetIdentity() string {
	if m != nil {
		return m.Identity
	}
	return ""
}

func (m *RequestAuthResp) GetRoles() []int64 {
	if m != nil {
		return m.Roles
	}
	return nil
}

func (m *RequestAuthResp) GetValidUntil() *google_protobuf.Timestamp {
	if m != nil {
		return m.ValidUntil
	}
	return nil
}

func (m *RequestAuthResp) GetElevatedUntil() *google_protobuf.Timestamp {
	if m != nil {
		return m.ElevatedUntil
	}
	return nil
}

func (m *RequestAuthResp) GetGivenName() string {
	if m != nil {
		return m.GivenName
	}
	return ""
}

func (m *RequestAuthResp) GetFamilyName() string {
	if m != nil {
		return m.FamilyName
	}
	return ""
}

func (m *RequestAuthResp) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *RequestAuthResp) GetTokenKey() string {
	if m != nil {
		return m.TokenKey
	}
	return ""
}

func (m *RequestAuthResp) GetSecondary() *RequestAuthResp {
	if m != nil {
		return m.Secondary
	}
	return nil
}

type RequestAuthReq struct {
	Token         string         `protobuf:"bytes,1,opt,name=Token" json:"Token,omitempty"`
	Configuration *ConfigureAuth `protobuf:"bytes,2,opt,name=Configuration" json:"Configuration,omitempty"`
}

func (m *RequestAuthReq) Reset()                    { *m = RequestAuthReq{} }
func (m *RequestAuthReq) String() string            { return proto.CompactTextString(m) }
func (*RequestAuthReq) ProtoMessage()               {}
func (*RequestAuthReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *RequestAuthReq) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *RequestAuthReq) GetConfiguration() *ConfigureAuth {
	if m != nil {
		return m.Configuration
	}
	return nil
}

type LoginReq struct {
	Identity string `protobuf:"bytes,1,opt,name=Identity" json:"Identity,omitempty"`
	Password string `protobuf:"bytes,2,opt,name=Password" json:"Password,omitempty"`
}

func (m *LoginReq) Reset()                    { *m = LoginReq{} }
func (m *LoginReq) String() string            { return proto.CompactTextString(m) }
func (*LoginReq) ProtoMessage()               {}
func (*LoginReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *LoginReq) GetIdentity() string {
	if m != nil {
		return m.Identity
	}
	return ""
}

func (m *LoginReq) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

type LoginResp struct {
	SessionTokenValue string `protobuf:"bytes,1,opt,name=SessionTokenValue" json:"SessionTokenValue,omitempty"`
}

func (m *LoginResp) Reset()                    { *m = LoginResp{} }
func (m *LoginResp) String() string            { return proto.CompactTextString(m) }
func (*LoginResp) ProtoMessage()               {}
func (*LoginResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *LoginResp) GetSessionTokenValue() string {
	if m != nil {
		return m.SessionTokenValue
	}
	return ""
}

type LogoutReq struct {
	// Types that are valid to be assigned to Value:
	//	*LogoutReq_SessionTokenValue
	//	*LogoutReq_Identity
	Value isLogoutReq_Value `protobuf_oneof:"Value"`
}

func (m *LogoutReq) Reset()                    { *m = LogoutReq{} }
func (m *LogoutReq) String() string            { return proto.CompactTextString(m) }
func (*LogoutReq) ProtoMessage()               {}
func (*LogoutReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type isLogoutReq_Value interface {
	isLogoutReq_Value()
}

type LogoutReq_SessionTokenValue struct {
	SessionTokenValue string `protobuf:"bytes,1,opt,name=SessionTokenValue,oneof"`
}
type LogoutReq_Identity struct {
	Identity string `protobuf:"bytes,2,opt,name=Identity,oneof"`
}

func (*LogoutReq_SessionTokenValue) isLogoutReq_Value() {}
func (*LogoutReq_Identity) isLogoutReq_Value()          {}

func (m *LogoutReq) GetValue() isLogoutReq_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *LogoutReq) GetSessionTokenValue() string {
	if x, ok := m.GetValue().(*LogoutReq_SessionTokenValue); ok {
		return x.SessionTokenValue
	}
	return ""
}

func (m *LogoutReq) GetIdentity() string {
	if x, ok := m.GetValue().(*LogoutReq_Identity); ok {
		return x.Identity
	}
	return ""
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*LogoutReq) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _LogoutReq_OneofMarshaler, _LogoutReq_OneofUnmarshaler, _LogoutReq_OneofSizer, []interface{}{
		(*LogoutReq_SessionTokenValue)(nil),
		(*LogoutReq_Identity)(nil),
	}
}

func _LogoutReq_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*LogoutReq)
	// Value
	switch x := m.Value.(type) {
	case *LogoutReq_SessionTokenValue:
		b.EncodeVarint(1<<3 | proto.WireBytes)
		b.EncodeStringBytes(x.SessionTokenValue)
	case *LogoutReq_Identity:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		b.EncodeStringBytes(x.Identity)
	case nil:
	default:
		return fmt.Errorf("LogoutReq.Value has unexpected type %T", x)
	}
	return nil
}

func _LogoutReq_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*LogoutReq)
	switch tag {
	case 1: // Value.SessionTokenValue
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeStringBytes()
		m.Value = &LogoutReq_SessionTokenValue{x}
		return true, err
	case 2: // Value.Identity
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeStringBytes()
		m.Value = &LogoutReq_Identity{x}
		return true, err
	default:
		return false, nil
	}
}

func _LogoutReq_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*LogoutReq)
	// Value
	switch x := m.Value.(type) {
	case *LogoutReq_SessionTokenValue:
		n += proto.SizeVarint(1<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(len(x.SessionTokenValue)))
		n += len(x.SessionTokenValue)
	case *LogoutReq_Identity:
		n += proto.SizeVarint(2<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(len(x.Identity)))
		n += len(x.Identity)
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type LogoutResp struct {
}

func (m *LogoutResp) Reset()                    { *m = LogoutResp{} }
func (m *LogoutResp) String() string            { return proto.CompactTextString(m) }
func (*LogoutResp) ProtoMessage()               {}
func (*LogoutResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type NewPasswordReq struct {
	Identity string `protobuf:"bytes,1,opt,name=Identity" json:"Identity,omitempty"`
}

func (m *NewPasswordReq) Reset()                    { *m = NewPasswordReq{} }
func (m *NewPasswordReq) String() string            { return proto.CompactTextString(m) }
func (*NewPasswordReq) ProtoMessage()               {}
func (*NewPasswordReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *NewPasswordReq) GetIdentity() string {
	if m != nil {
		return m.Identity
	}
	return ""
}

type NewPasswordResp struct {
}

func (m *NewPasswordResp) Reset()                    { *m = NewPasswordResp{} }
func (m *NewPasswordResp) String() string            { return proto.CompactTextString(m) }
func (*NewPasswordResp) ProtoMessage()               {}
func (*NewPasswordResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

type ChangePasswordReq struct {
	// SessionTokenValue must be valid for the password to be changed.
	SessionTokenValue string `protobuf:"bytes,1,opt,name=SessionTokenValue" json:"SessionTokenValue,omitempty"`
	// CurrentPassword is not used if the login state is Change Password.
	// In that state it is assumed the client has just authenticated and
	// entering the current password again would be redundant.
	CurrentPassword string `protobuf:"bytes,2,opt,name=CurrentPassword" json:"CurrentPassword,omitempty"`
	// NewPassword to set. If the password is too weak it may be rejected.
	NewPassword string `protobuf:"bytes,3,opt,name=NewPassword" json:"NewPassword,omitempty"`
}

func (m *ChangePasswordReq) Reset()                    { *m = ChangePasswordReq{} }
func (m *ChangePasswordReq) String() string            { return proto.CompactTextString(m) }
func (*ChangePasswordReq) ProtoMessage()               {}
func (*ChangePasswordReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *ChangePasswordReq) GetSessionTokenValue() string {
	if m != nil {
		return m.SessionTokenValue
	}
	return ""
}

func (m *ChangePasswordReq) GetCurrentPassword() string {
	if m != nil {
		return m.CurrentPassword
	}
	return ""
}

func (m *ChangePasswordReq) GetNewPassword() string {
	if m != nil {
		return m.NewPassword
	}
	return ""
}

type ChangePasswordResp struct {
	// Changed is true when the password was changed.
	// If false the InvalidNewPasswordMessage text should
	// be displayed to the user.
	Changed bool `protobuf:"varint,1,opt,name=Changed" json:"Changed,omitempty"`
	// InvalidNewPasswordMessage is is set when Changed is true.
	InvalidNewPasswordMessage string `protobuf:"bytes,2,opt,name=InvalidNewPasswordMessage" json:"InvalidNewPasswordMessage,omitempty"`
}

func (m *ChangePasswordResp) Reset()                    { *m = ChangePasswordResp{} }
func (m *ChangePasswordResp) String() string            { return proto.CompactTextString(m) }
func (*ChangePasswordResp) ProtoMessage()               {}
func (*ChangePasswordResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *ChangePasswordResp) GetChanged() bool {
	if m != nil {
		return m.Changed
	}
	return false
}

func (m *ChangePasswordResp) GetInvalidNewPasswordMessage() string {
	if m != nil {
		return m.InvalidNewPasswordMessage
	}
	return ""
}

func init() {
	proto.RegisterType((*ConfigureAuth)(nil), "api.ConfigureAuth")
	proto.RegisterType((*RequestAuthResp)(nil), "api.RequestAuthResp")
	proto.RegisterType((*RequestAuthReq)(nil), "api.RequestAuthReq")
	proto.RegisterType((*LoginReq)(nil), "api.LoginReq")
	proto.RegisterType((*LoginResp)(nil), "api.LoginResp")
	proto.RegisterType((*LogoutReq)(nil), "api.LogoutReq")
	proto.RegisterType((*LogoutResp)(nil), "api.LogoutResp")
	proto.RegisterType((*NewPasswordReq)(nil), "api.NewPasswordReq")
	proto.RegisterType((*NewPasswordResp)(nil), "api.NewPasswordResp")
	proto.RegisterType((*ChangePasswordReq)(nil), "api.ChangePasswordReq")
	proto.RegisterType((*ChangePasswordResp)(nil), "api.ChangePasswordResp")
	proto.RegisterEnum("api.LoginState", LoginState_name, LoginState_value)
	proto.RegisterEnum("api.ConfigureAuth_AreaType", ConfigureAuth_AreaType_name, ConfigureAuth_AreaType_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Auth service

type AuthClient interface {
	// TODO: RequestAuth and Login should both take some additional features
	// about where the request is coming from (HTTPS info, remote address).
	RequestAuth(ctx context.Context, in *RequestAuthReq, opts ...grpc.CallOption) (*RequestAuthResp, error)
}

type authClient struct {
	cc *grpc.ClientConn
}

func NewAuthClient(cc *grpc.ClientConn) AuthClient {
	return &authClient{cc}
}

func (c *authClient) RequestAuth(ctx context.Context, in *RequestAuthReq, opts ...grpc.CallOption) (*RequestAuthResp, error) {
	out := new(RequestAuthResp)
	err := grpc.Invoke(ctx, "/api.Auth/RequestAuth", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Auth service

type AuthServer interface {
	// TODO: RequestAuth and Login should both take some additional features
	// about where the request is coming from (HTTPS info, remote address).
	RequestAuth(context.Context, *RequestAuthReq) (*RequestAuthResp, error)
}

func RegisterAuthServer(s *grpc.Server, srv AuthServer) {
	s.RegisterService(&_Auth_serviceDesc, srv)
}

func _Auth_RequestAuth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestAuthReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).RequestAuth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.Auth/RequestAuth",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).RequestAuth(ctx, req.(*RequestAuthReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Auth_serviceDesc = grpc.ServiceDesc{
	ServiceName: "api.Auth",
	HandlerType: (*AuthServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RequestAuth",
			Handler:    _Auth_RequestAuth_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "auth.proto",
}

func init() { proto.RegisterFile("auth.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 670 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0xcd, 0x4e, 0xdb, 0x40,
	0x10, 0xc6, 0x71, 0x42, 0x92, 0x49, 0x49, 0xcc, 0x96, 0x83, 0x9b, 0xa2, 0x36, 0xf2, 0x29, 0xaa,
	0xa8, 0x23, 0xa5, 0x17, 0x5a, 0xf5, 0xc0, 0x5f, 0xa0, 0x51, 0x0b, 0xaa, 0x1c, 0x82, 0x7a, 0xec,
	0xd2, 0x0c, 0x66, 0x85, 0xbd, 0x6b, 0xbc, 0xeb, 0xa0, 0x3c, 0x44, 0x0f, 0x7d, 0x89, 0x3e, 0x67,
	0xe5, 0xb5, 0x43, 0x9c, 0x00, 0x42, 0x3d, 0xce, 0x37, 0xdf, 0xcc, 0x7c, 0x3b, 0x3f, 0x0b, 0x40,
	0x13, 0x75, 0xed, 0x46, 0xb1, 0x50, 0x82, 0x98, 0x34, 0x62, 0xed, 0xb7, 0xbe, 0x10, 0x7e, 0x80,
	0x3d, 0x0d, 0x5d, 0x26, 0x57, 0x3d, 0xc5, 0x42, 0x94, 0x8a, 0x86, 0x51, 0xc6, 0x72, 0xfe, 0x18,
	0xb0, 0x71, 0x28, 0xf8, 0x15, 0xf3, 0x93, 0x18, 0xf7, 0x13, 0x75, 0x4d, 0x7a, 0x50, 0xde, 0x8f,
	0x91, 0xda, 0x46, 0xc7, 0xe8, 0x36, 0xfb, 0xaf, 0x5d, 0x1a, 0x31, 0x77, 0x89, 0xe1, 0xa6, 0xee,
	0xf3, 0x59, 0x84, 0x9e, 0x26, 0x92, 0x0e, 0x34, 0x06, 0x7c, 0xca, 0x62, 0xc1, 0x43, 0xe4, 0xca,
	0x2e, 0x75, 0x8c, 0x6e, 0xdd, 0x2b, 0x42, 0xce, 0x7b, 0xa8, 0xcd, 0x63, 0x48, 0x03, 0xaa, 0x63,
	0x7e, 0xc3, 0xc5, 0x1d, 0xb7, 0xd6, 0x08, 0xc0, 0xfa, 0x68, 0x26, 0x15, 0x86, 0x96, 0x41, 0x6a,
	0x50, 0x1e, 0x4b, 0x8c, 0xad, 0x92, 0xf3, 0xd7, 0x84, 0x96, 0x87, 0xb7, 0x09, 0x4a, 0x95, 0xd6,
	0xf3, 0x50, 0x46, 0xa4, 0x07, 0xf0, 0x4d, 0xf8, 0x8c, 0x8f, 0x14, 0x55, 0x98, 0x6b, 0x6b, 0x69,
	0x6d, 0x0b, 0xd8, 0x2b, 0x50, 0x48, 0x13, 0x4a, 0xc3, 0x23, 0x2d, 0xc6, 0xf4, 0x4a, 0xc3, 0x23,
	0xd2, 0x86, 0xda, 0x70, 0x82, 0x5c, 0x31, 0x35, 0xb3, 0x4d, 0x2d, 0xf1, 0xde, 0x26, 0x5b, 0x50,
	0xf1, 0x44, 0x80, 0xd2, 0x2e, 0x77, 0xcc, 0xae, 0xe9, 0x65, 0x06, 0xf9, 0x04, 0x70, 0x41, 0x03,
	0x36, 0x19, 0x73, 0xc5, 0x02, 0xbb, 0xd2, 0x31, 0xba, 0x8d, 0x7e, 0xdb, 0xcd, 0x1a, 0xea, 0xce,
	0x1b, 0xea, 0x9e, 0xcf, 0x1b, 0xea, 0x15, 0xd8, 0x64, 0x0f, 0x36, 0x06, 0x01, 0x4e, 0xa9, 0xc2,
	0x3c, 0x7c, 0xfd, 0xd9, 0xf0, 0xe5, 0x00, 0xb2, 0x0d, 0xf5, 0x13, 0x36, 0x45, 0x7e, 0x46, 0x43,
	0xb4, 0xab, 0x5a, 0xf0, 0x02, 0x20, 0x6f, 0x00, 0x8e, 0x69, 0xc8, 0x82, 0x99, 0x76, 0xd7, 0xb4,
	0xbb, 0x80, 0xa4, 0x2f, 0x1a, 0x84, 0x94, 0x05, 0x76, 0x5d, 0xbb, 0x32, 0x23, 0xed, 0xc1, 0xb9,
	0xb8, 0x41, 0xfe, 0x15, 0x67, 0x36, 0x64, 0x3d, 0x98, 0xdb, 0xa4, 0x0f, 0xf5, 0x11, 0xfe, 0x12,
	0x7c, 0x42, 0xe3, 0x99, 0xdd, 0xd0, 0x6a, 0xb7, 0x74, 0x7f, 0x57, 0x26, 0xe1, 0x2d, 0x68, 0xce,
	0x4f, 0x68, 0x2e, 0x79, 0x6f, 0xd3, 0xba, 0x3a, 0xa3, 0x9e, 0x50, 0xdd, 0xcb, 0x0c, 0xb2, 0xbb,
	0xd8, 0x31, 0xaa, 0x98, 0xe0, 0x7a, 0x2c, 0x8d, 0x3e, 0x79, 0xb8, 0x5b, 0xde, 0x32, 0xd1, 0x39,
	0x80, 0x9a, 0x9e, 0x69, 0x9a, 0xbb, 0x38, 0x41, 0x63, 0x65, 0x82, 0x6d, 0xa8, 0x7d, 0xa7, 0x52,
	0xde, 0x89, 0x78, 0x92, 0x2f, 0xe0, 0xbd, 0xed, 0x7c, 0x84, 0x7a, 0x9e, 0x43, 0x46, 0x64, 0x07,
	0x36, 0x47, 0x28, 0x25, 0x13, 0x5c, 0x4b, 0xbb, 0xa0, 0x41, 0x82, 0x79, 0xb6, 0x87, 0x0e, 0xe7,
	0x52, 0x87, 0x8a, 0x44, 0xa5, 0xf5, 0xdd, 0x27, 0x43, 0xbf, 0xac, 0x3d, 0x12, 0x4c, 0xb6, 0x0b,
	0x7a, 0x4b, 0x39, 0xed, 0x1e, 0x39, 0xa8, 0x42, 0x25, 0xab, 0xf1, 0x42, 0x6f, 0xb6, 0xae, 0x21,
	0x23, 0x67, 0x07, 0x9a, 0x67, 0x78, 0x37, 0xd7, 0xfe, 0xcc, 0xb3, 0x9d, 0x4d, 0x68, 0x2d, 0xb1,
	0x65, 0xe4, 0xfc, 0x36, 0x60, 0xf3, 0xf0, 0x9a, 0x72, 0x1f, 0x8b, 0x49, 0xfe, 0xeb, 0xd9, 0xa4,
	0x0b, 0xad, 0xc3, 0x24, 0x8e, 0x91, 0xab, 0x95, 0xa6, 0xae, 0xc2, 0xe9, 0xed, 0x17, 0x04, 0xe4,
	0x87, 0x55, 0x84, 0x9c, 0x00, 0xc8, 0xaa, 0x1c, 0x19, 0x11, 0x1b, 0xaa, 0x19, 0x3a, 0xd1, 0x2a,
	0x6a, 0xde, 0xdc, 0x24, 0x9f, 0xe1, 0xd5, 0x90, 0x4f, 0xd3, 0x4b, 0x2a, 0x64, 0x39, 0x45, 0x29,
	0xa9, 0x8f, 0xb9, 0x8a, 0xa7, 0x09, 0xef, 0x7e, 0x14, 0xbf, 0x89, 0xf4, 0xaf, 0x39, 0x65, 0x52,
	0x32, 0xee, 0x5b, 0x6b, 0xa4, 0x0e, 0x95, 0x41, 0x1c, 0x8b, 0x38, 0xfb, 0x6a, 0xce, 0x04, 0x47,
	0xab, 0x94, 0x32, 0x4e, 0x62, 0xca, 0x15, 0x4e, 0x2c, 0x93, 0x54, 0xc1, 0x1c, 0xf7, 0x8f, 0xad,
	0x32, 0x21, 0xd0, 0x5c, 0xd6, 0x6c, 0x55, 0xfa, 0x7b, 0x50, 0xd6, 0xdf, 0xe3, 0x2e, 0x34, 0x0a,
	0x3b, 0x4f, 0x5e, 0x3e, 0xbc, 0x91, 0xdb, 0xf6, 0xa3, 0x87, 0x73, 0xb9, 0xae, 0x8f, 0xfe, 0xc3,
	0xbf, 0x00, 0x00, 0x00, 0xff, 0xff, 0xc9, 0x92, 0x09, 0x18, 0xa5, 0x05, 0x00, 0x00,
}
