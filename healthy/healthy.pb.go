// Code generated by protoc-gen-go. DO NOT EDIT.
// source: healthy.proto

package cyclone_healthy

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type CycloneResponse_ResponseResult int32

const (
	CycloneResponse_Zombies CycloneResponse_ResponseResult = 0
	CycloneResponse_Sick    CycloneResponse_ResponseResult = -1
	CycloneResponse_Healthy CycloneResponse_ResponseResult = 1
)

var CycloneResponse_ResponseResult_name = map[int32]string{
	0:  "Zombies",
	-1: "Sick",
	1:  "Healthy",
}

var CycloneResponse_ResponseResult_value = map[string]int32{
	"Zombies": 0,
	"Sick":    -1,
	"Healthy": 1,
}

func (x CycloneResponse_ResponseResult) String() string {
	return proto.EnumName(CycloneResponse_ResponseResult_name, int32(x))
}

func (CycloneResponse_ResponseResult) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_741f22f95cb14d6b, []int{4, 0}
}

type CycloneRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CycloneRequest) Reset()         { *m = CycloneRequest{} }
func (m *CycloneRequest) String() string { return proto.CompactTextString(m) }
func (*CycloneRequest) ProtoMessage()    {}
func (*CycloneRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_741f22f95cb14d6b, []int{0}
}

func (m *CycloneRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CycloneRequest.Unmarshal(m, b)
}
func (m *CycloneRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CycloneRequest.Marshal(b, m, deterministic)
}
func (m *CycloneRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CycloneRequest.Merge(m, src)
}
func (m *CycloneRequest) XXX_Size() int {
	return xxx_messageInfo_CycloneRequest.Size(m)
}
func (m *CycloneRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CycloneRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CycloneRequest proto.InternalMessageInfo

type CycloneCloseResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CycloneCloseResponse) Reset()         { *m = CycloneCloseResponse{} }
func (m *CycloneCloseResponse) String() string { return proto.CompactTextString(m) }
func (*CycloneCloseResponse) ProtoMessage()    {}
func (*CycloneCloseResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_741f22f95cb14d6b, []int{1}
}

func (m *CycloneCloseResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CycloneCloseResponse.Unmarshal(m, b)
}
func (m *CycloneCloseResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CycloneCloseResponse.Marshal(b, m, deterministic)
}
func (m *CycloneCloseResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CycloneCloseResponse.Merge(m, src)
}
func (m *CycloneCloseResponse) XXX_Size() int {
	return xxx_messageInfo_CycloneCloseResponse.Size(m)
}
func (m *CycloneCloseResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_CycloneCloseResponse.DiscardUnknown(m)
}

var xxx_messageInfo_CycloneCloseResponse proto.InternalMessageInfo

type ApiInfo struct {
	Api                  string   `protobuf:"bytes,1,opt,name=api,proto3" json:"api,omitempty"`
	Error                string   `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ApiInfo) Reset()         { *m = ApiInfo{} }
func (m *ApiInfo) String() string { return proto.CompactTextString(m) }
func (*ApiInfo) ProtoMessage()    {}
func (*ApiInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_741f22f95cb14d6b, []int{2}
}

func (m *ApiInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ApiInfo.Unmarshal(m, b)
}
func (m *ApiInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ApiInfo.Marshal(b, m, deterministic)
}
func (m *ApiInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ApiInfo.Merge(m, src)
}
func (m *ApiInfo) XXX_Size() int {
	return xxx_messageInfo_ApiInfo.Size(m)
}
func (m *ApiInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ApiInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ApiInfo proto.InternalMessageInfo

func (m *ApiInfo) GetApi() string {
	if m != nil {
		return m.Api
	}
	return ""
}

func (m *ApiInfo) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

type ServiceStatus struct {
	Name                 string     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	ApiInfo              []*ApiInfo `protobuf:"bytes,2,rep,name=api_info,json=apiInfo,proto3" json:"api_info,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *ServiceStatus) Reset()         { *m = ServiceStatus{} }
func (m *ServiceStatus) String() string { return proto.CompactTextString(m) }
func (*ServiceStatus) ProtoMessage()    {}
func (*ServiceStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor_741f22f95cb14d6b, []int{3}
}

func (m *ServiceStatus) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServiceStatus.Unmarshal(m, b)
}
func (m *ServiceStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServiceStatus.Marshal(b, m, deterministic)
}
func (m *ServiceStatus) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServiceStatus.Merge(m, src)
}
func (m *ServiceStatus) XXX_Size() int {
	return xxx_messageInfo_ServiceStatus.Size(m)
}
func (m *ServiceStatus) XXX_DiscardUnknown() {
	xxx_messageInfo_ServiceStatus.DiscardUnknown(m)
}

var xxx_messageInfo_ServiceStatus proto.InternalMessageInfo

func (m *ServiceStatus) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ServiceStatus) GetApiInfo() []*ApiInfo {
	if m != nil {
		return m.ApiInfo
	}
	return nil
}

type CycloneResponse struct {
	Code                 CycloneResponse_ResponseResult `protobuf:"varint,1,opt,name=code,proto3,enum=cyclone.healthy.CycloneResponse_ResponseResult" json:"code,omitempty"`
	Response             *ServiceStatus                 `protobuf:"bytes,2,opt,name=response,proto3" json:"response,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                       `json:"-"`
	XXX_unrecognized     []byte                         `json:"-"`
	XXX_sizecache        int32                          `json:"-"`
}

func (m *CycloneResponse) Reset()         { *m = CycloneResponse{} }
func (m *CycloneResponse) String() string { return proto.CompactTextString(m) }
func (*CycloneResponse) ProtoMessage()    {}
func (*CycloneResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_741f22f95cb14d6b, []int{4}
}

func (m *CycloneResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CycloneResponse.Unmarshal(m, b)
}
func (m *CycloneResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CycloneResponse.Marshal(b, m, deterministic)
}
func (m *CycloneResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CycloneResponse.Merge(m, src)
}
func (m *CycloneResponse) XXX_Size() int {
	return xxx_messageInfo_CycloneResponse.Size(m)
}
func (m *CycloneResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_CycloneResponse.DiscardUnknown(m)
}

var xxx_messageInfo_CycloneResponse proto.InternalMessageInfo

func (m *CycloneResponse) GetCode() CycloneResponse_ResponseResult {
	if m != nil {
		return m.Code
	}
	return CycloneResponse_Zombies
}

func (m *CycloneResponse) GetResponse() *ServiceStatus {
	if m != nil {
		return m.Response
	}
	return nil
}

func init() {
	proto.RegisterEnum("cyclone.healthy.CycloneResponse_ResponseResult", CycloneResponse_ResponseResult_name, CycloneResponse_ResponseResult_value)
	proto.RegisterType((*CycloneRequest)(nil), "cyclone.healthy.CycloneRequest")
	proto.RegisterType((*CycloneCloseResponse)(nil), "cyclone.healthy.CycloneCloseResponse")
	proto.RegisterType((*ApiInfo)(nil), "cyclone.healthy.ApiInfo")
	proto.RegisterType((*ServiceStatus)(nil), "cyclone.healthy.ServiceStatus")
	proto.RegisterType((*CycloneResponse)(nil), "cyclone.healthy.CycloneResponse")
}

func init() { proto.RegisterFile("healthy.proto", fileDescriptor_741f22f95cb14d6b) }

var fileDescriptor_741f22f95cb14d6b = []byte{
	// 319 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x52, 0x4d, 0x4b, 0xc3, 0x40,
	0x10, 0x6d, 0xfa, 0x61, 0xeb, 0x94, 0xb6, 0x71, 0x28, 0x12, 0x3c, 0x68, 0x59, 0x10, 0x7a, 0x8a,
	0x98, 0xde, 0x04, 0x0f, 0xd2, 0x8b, 0x5e, 0x04, 0xb7, 0x17, 0xf1, 0x22, 0x69, 0x9c, 0xd2, 0xc5,
	0x34, 0x1b, 0xb3, 0x5b, 0xa1, 0xbf, 0xcb, 0x5f, 0xe2, 0x2f, 0x52, 0xba, 0xbb, 0x0d, 0xb4, 0xa5,
	0x34, 0xa7, 0x97, 0x37, 0x33, 0x6f, 0xe7, 0xf1, 0x06, 0x3a, 0x73, 0x8a, 0x53, 0x3d, 0x5f, 0x85,
	0x79, 0x21, 0xb5, 0xc4, 0x5e, 0xb2, 0x4a, 0x52, 0x99, 0x51, 0xe8, 0x68, 0xe6, 0x43, 0x77, 0x6c,
	0x29, 0x4e, 0x5f, 0x4b, 0x52, 0x9a, 0x9d, 0x43, 0xdf, 0x31, 0xe3, 0x54, 0x2a, 0xe2, 0xa4, 0x72,
	0x99, 0x29, 0x62, 0xb7, 0xd0, 0x7c, 0xc8, 0xc5, 0x53, 0x36, 0x93, 0xe8, 0x43, 0x2d, 0xce, 0x45,
	0xe0, 0x0d, 0xbc, 0xe1, 0x29, 0x5f, 0x43, 0xec, 0x43, 0x83, 0x8a, 0x42, 0x16, 0x41, 0xd5, 0x70,
	0xf6, 0x87, 0xbd, 0x42, 0x67, 0x42, 0xc5, 0xb7, 0x48, 0x68, 0xa2, 0x63, 0xbd, 0x54, 0x88, 0x50,
	0xcf, 0xe2, 0x05, 0xb9, 0x49, 0x83, 0x71, 0x04, 0xad, 0x38, 0x17, 0xef, 0x22, 0x9b, 0xc9, 0xa0,
	0x3a, 0xa8, 0x0d, 0xdb, 0x51, 0x10, 0xee, 0x6c, 0x19, 0xba, 0x87, 0x79, 0x33, 0xb6, 0x80, 0xfd,
	0x7a, 0xd0, 0x2b, 0xf7, 0xb6, 0x0b, 0xe2, 0x18, 0xea, 0x89, 0xfc, 0xb0, 0xe2, 0xdd, 0xe8, 0x66,
	0x4f, 0x64, 0xa7, 0x3f, 0xdc, 0x00, 0x4e, 0x6a, 0x99, 0x6a, 0x6e, 0x86, 0xf1, 0x0e, 0x5a, 0x85,
	0xe3, 0x8d, 0x97, 0x76, 0x74, 0xb9, 0x27, 0xb4, 0xe5, 0x89, 0x97, 0xfd, 0xec, 0x1e, 0xba, 0xdb,
	0x9a, 0xd8, 0x86, 0xe6, 0x9b, 0x5c, 0x4c, 0x05, 0x29, 0xbf, 0x82, 0x67, 0x50, 0x9f, 0x88, 0xe4,
	0xd3, 0xff, 0xdb, 0x7c, 0xde, 0xba, 0xfe, 0x68, 0x45, 0x7d, 0x2f, 0xfa, 0xf1, 0xca, 0x2c, 0x1c,
	0x89, 0xcf, 0x65, 0x1d, 0xaf, 0x0e, 0xfb, 0x31, 0xb9, 0x5d, 0x0c, 0x8e, 0x19, 0x66, 0x15, 0x7c,
	0x81, 0x86, 0x09, 0xf5, 0xb8, 0xda, 0xf5, 0xa1, 0x86, 0xed, 0xa3, 0xa8, 0x4c, 0x4f, 0xcc, 0x61,
	0x8d, 0xfe, 0x03, 0x00, 0x00, 0xff, 0xff, 0x37, 0x8b, 0x8f, 0x3f, 0x69, 0x02, 0x00, 0x00,
}
