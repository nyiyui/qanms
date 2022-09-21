// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.21.5
// source: api.proto

package api

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type AuthSQ struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Sq:
	//	*AuthSQ_S
	//	*AuthSQ_Q
	//	*AuthSQ_Token
	Sq isAuthSQ_Sq `protobuf_oneof:"sq"`
}

func (x *AuthSQ) Reset() {
	*x = AuthSQ{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthSQ) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthSQ) ProtoMessage() {}

func (x *AuthSQ) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthSQ.ProtoReflect.Descriptor instead.
func (*AuthSQ) Descriptor() ([]byte, []int) {
	return file_api_proto_rawDescGZIP(), []int{0}
}

func (m *AuthSQ) GetSq() isAuthSQ_Sq {
	if m != nil {
		return m.Sq
	}
	return nil
}

func (x *AuthSQ) GetS() *AuthS {
	if x, ok := x.GetSq().(*AuthSQ_S); ok {
		return x.S
	}
	return nil
}

func (x *AuthSQ) GetQ() *AuthQ {
	if x, ok := x.GetSq().(*AuthSQ_Q); ok {
		return x.Q
	}
	return nil
}

func (x *AuthSQ) GetToken() *AuthToken {
	if x, ok := x.GetSq().(*AuthSQ_Token); ok {
		return x.Token
	}
	return nil
}

type isAuthSQ_Sq interface {
	isAuthSQ_Sq()
}

type AuthSQ_S struct {
	S *AuthS `protobuf:"bytes,1,opt,name=s,proto3,oneof"`
}

type AuthSQ_Q struct {
	Q *AuthQ `protobuf:"bytes,2,opt,name=q,proto3,oneof"`
}

type AuthSQ_Token struct {
	Token *AuthToken `protobuf:"bytes,3,opt,name=token,proto3,oneof"`
}

func (*AuthSQ_S) isAuthSQ_Sq() {}

func (*AuthSQ_Q) isAuthSQ_Sq() {}

func (*AuthSQ_Token) isAuthSQ_Sq() {}

type AuthQ struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Network string `protobuf:"bytes,1,opt,name=network,proto3" json:"network,omitempty"`
	Me      string `protobuf:"bytes,2,opt,name=me,proto3" json:"me,omitempty"`
	You     string `protobuf:"bytes,4,opt,name=you,proto3" json:"you,omitempty"`
	Chall   []byte `protobuf:"bytes,5,opt,name=chall,proto3" json:"chall,omitempty"`
}

func (x *AuthQ) Reset() {
	*x = AuthQ{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthQ) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthQ) ProtoMessage() {}

func (x *AuthQ) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthQ.ProtoReflect.Descriptor instead.
func (*AuthQ) Descriptor() ([]byte, []int) {
	return file_api_proto_rawDescGZIP(), []int{1}
}

func (x *AuthQ) GetNetwork() string {
	if x != nil {
		return x.Network
	}
	return ""
}

func (x *AuthQ) GetMe() string {
	if x != nil {
		return x.Me
	}
	return ""
}

func (x *AuthQ) GetYou() string {
	if x != nil {
		return x.You
	}
	return ""
}

func (x *AuthQ) GetChall() []byte {
	if x != nil {
		return x.Chall
	}
	return nil
}

type AuthS struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ChallResp  []byte `protobuf:"bytes,1,opt,name=challResp,proto3" json:"challResp,omitempty"`
	ChallAdded []byte `protobuf:"bytes,2,opt,name=challAdded,proto3" json:"challAdded,omitempty"`
}

func (x *AuthS) Reset() {
	*x = AuthS{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthS) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthS) ProtoMessage() {}

func (x *AuthS) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthS.ProtoReflect.Descriptor instead.
func (*AuthS) Descriptor() ([]byte, []int) {
	return file_api_proto_rawDescGZIP(), []int{2}
}

func (x *AuthS) GetChallResp() []byte {
	if x != nil {
		return x.ChallResp
	}
	return nil
}

func (x *AuthS) GetChallAdded() []byte {
	if x != nil {
		return x.ChallAdded
	}
	return nil
}

type AuthToken struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token []byte `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
}

func (x *AuthToken) Reset() {
	*x = AuthToken{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthToken) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthToken) ProtoMessage() {}

func (x *AuthToken) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthToken.ProtoReflect.Descriptor instead.
func (*AuthToken) Descriptor() ([]byte, []int) {
	return file_api_proto_rawDescGZIP(), []int{3}
}

func (x *AuthToken) GetToken() []byte {
	if x != nil {
		return x.Token
	}
	return nil
}

type XchQ struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token  []byte `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	PubKey []byte `protobuf:"bytes,2,opt,name=pubKey,proto3" json:"pubKey,omitempty"`
	Psk    []byte `protobuf:"bytes,3,opt,name=psk,proto3" json:"psk,omitempty"`
}

func (x *XchQ) Reset() {
	*x = XchQ{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *XchQ) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*XchQ) ProtoMessage() {}

func (x *XchQ) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use XchQ.ProtoReflect.Descriptor instead.
func (*XchQ) Descriptor() ([]byte, []int) {
	return file_api_proto_rawDescGZIP(), []int{4}
}

func (x *XchQ) GetToken() []byte {
	if x != nil {
		return x.Token
	}
	return nil
}

func (x *XchQ) GetPubKey() []byte {
	if x != nil {
		return x.PubKey
	}
	return nil
}

func (x *XchQ) GetPsk() []byte {
	if x != nil {
		return x.Psk
	}
	return nil
}

type XchS struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PubKey []byte `protobuf:"bytes,1,opt,name=pubKey,proto3" json:"pubKey,omitempty"`
	Psk    []byte `protobuf:"bytes,2,opt,name=psk,proto3" json:"psk,omitempty"`
}

func (x *XchS) Reset() {
	*x = XchS{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *XchS) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*XchS) ProtoMessage() {}

func (x *XchS) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use XchS.ProtoReflect.Descriptor instead.
func (*XchS) Descriptor() ([]byte, []int) {
	return file_api_proto_rawDescGZIP(), []int{5}
}

func (x *XchS) GetPubKey() []byte {
	if x != nil {
		return x.PubKey
	}
	return nil
}

func (x *XchS) GetPsk() []byte {
	if x != nil {
		return x.Psk
	}
	return nil
}

type PingQS struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PingQS) Reset() {
	*x = PingQS{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingQS) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingQS) ProtoMessage() {}

func (x *PingQS) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingQS.ProtoReflect.Descriptor instead.
func (*PingQS) Descriptor() ([]byte, []int) {
	return file_api_proto_rawDescGZIP(), []int{6}
}

type PullQ struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token []byte `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
}

func (x *PullQ) Reset() {
	*x = PullQ{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PullQ) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PullQ) ProtoMessage() {}

func (x *PullQ) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PullQ.ProtoReflect.Descriptor instead.
func (*PullQ) Descriptor() ([]byte, []int) {
	return file_api_proto_rawDescGZIP(), []int{7}
}

func (x *PullQ) GetToken() []byte {
	if x != nil {
		return x.Token
	}
	return nil
}

type PullS struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cc *CentralConfig `protobuf:"bytes,1,opt,name=cc,proto3" json:"cc,omitempty"`
}

func (x *PullS) Reset() {
	*x = PullS{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PullS) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PullS) ProtoMessage() {}

func (x *PullS) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PullS.ProtoReflect.Descriptor instead.
func (*PullS) Descriptor() ([]byte, []int) {
	return file_api_proto_rawDescGZIP(), []int{8}
}

func (x *PullS) GetCc() *CentralConfig {
	if x != nil {
		return x.Cc
	}
	return nil
}

type CentralConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Networks map[string]*CentralNetwork `protobuf:"bytes,1,rep,name=networks,proto3" json:"networks,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *CentralConfig) Reset() {
	*x = CentralConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CentralConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CentralConfig) ProtoMessage() {}

func (x *CentralConfig) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CentralConfig.ProtoReflect.Descriptor instead.
func (*CentralConfig) Descriptor() ([]byte, []int) {
	return file_api_proto_rawDescGZIP(), []int{9}
}

func (x *CentralConfig) GetNetworks() map[string]*CentralNetwork {
	if x != nil {
		return x.Networks
	}
	return nil
}

type CentralNetwork struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ips       []*IPNet                `protobuf:"bytes,1,rep,name=ips,proto3" json:"ips,omitempty"`
	Me        string                  `protobuf:"bytes,2,opt,name=me,proto3" json:"me,omitempty"`
	Keepalive *durationpb.Duration    `protobuf:"bytes,3,opt,name=keepalive,proto3" json:"keepalive,omitempty"`
	Peers     map[string]*CentralPeer `protobuf:"bytes,5,rep,name=peers,proto3" json:"peers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *CentralNetwork) Reset() {
	*x = CentralNetwork{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CentralNetwork) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CentralNetwork) ProtoMessage() {}

func (x *CentralNetwork) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CentralNetwork.ProtoReflect.Descriptor instead.
func (*CentralNetwork) Descriptor() ([]byte, []int) {
	return file_api_proto_rawDescGZIP(), []int{10}
}

func (x *CentralNetwork) GetIps() []*IPNet {
	if x != nil {
		return x.Ips
	}
	return nil
}

func (x *CentralNetwork) GetMe() string {
	if x != nil {
		return x.Me
	}
	return ""
}

func (x *CentralNetwork) GetKeepalive() *durationpb.Duration {
	if x != nil {
		return x.Keepalive
	}
	return nil
}

func (x *CentralNetwork) GetPeers() map[string]*CentralPeer {
	if x != nil {
		return x.Peers
	}
	return nil
}

type CentralPeer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Host       string     `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	AllowedIPs []*IPNet   `protobuf:"bytes,2,rep,name=allowedIPs,proto3" json:"allowedIPs,omitempty"`
	PublicKey  *PublicKey `protobuf:"bytes,3,opt,name=publicKey,proto3" json:"publicKey,omitempty"`
}

func (x *CentralPeer) Reset() {
	*x = CentralPeer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CentralPeer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CentralPeer) ProtoMessage() {}

func (x *CentralPeer) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CentralPeer.ProtoReflect.Descriptor instead.
func (*CentralPeer) Descriptor() ([]byte, []int) {
	return file_api_proto_rawDescGZIP(), []int{11}
}

func (x *CentralPeer) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *CentralPeer) GetAllowedIPs() []*IPNet {
	if x != nil {
		return x.AllowedIPs
	}
	return nil
}

func (x *CentralPeer) GetPublicKey() *PublicKey {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

type PublicKey struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Raw []byte `protobuf:"bytes,1,opt,name=raw,proto3" json:"raw,omitempty"`
}

func (x *PublicKey) Reset() {
	*x = PublicKey{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PublicKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublicKey) ProtoMessage() {}

func (x *PublicKey) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublicKey.ProtoReflect.Descriptor instead.
func (*PublicKey) Descriptor() ([]byte, []int) {
	return file_api_proto_rawDescGZIP(), []int{12}
}

func (x *PublicKey) GetRaw() []byte {
	if x != nil {
		return x.Raw
	}
	return nil
}

type IPNet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cidr string `protobuf:"bytes,1,opt,name=cidr,proto3" json:"cidr,omitempty"`
}

func (x *IPNet) Reset() {
	*x = IPNet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IPNet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IPNet) ProtoMessage() {}

func (x *IPNet) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IPNet.ProtoReflect.Descriptor instead.
func (*IPNet) Descriptor() ([]byte, []int) {
	return file_api_proto_rawDescGZIP(), []int{13}
}

func (x *IPNet) GetCidr() string {
	if x != nil {
		return x.Cidr
	}
	return ""
}

var File_api_proto protoreflect.FileDescriptor

var file_api_proto_rawDesc = []byte{
	0x0a, 0x09, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x62, 0x0a, 0x06, 0x41,
	0x75, 0x74, 0x68, 0x53, 0x51, 0x12, 0x16, 0x0a, 0x01, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x06, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x53, 0x48, 0x00, 0x52, 0x01, 0x73, 0x12, 0x16, 0x0a,
	0x01, 0x71, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x06, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x51,
	0x48, 0x00, 0x52, 0x01, 0x71, 0x12, 0x22, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x54, 0x6f, 0x6b, 0x65, 0x6e,
	0x48, 0x00, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x42, 0x04, 0x0a, 0x02, 0x73, 0x71, 0x22,
	0x59, 0x0a, 0x05, 0x41, 0x75, 0x74, 0x68, 0x51, 0x12, 0x18, 0x0a, 0x07, 0x6e, 0x65, 0x74, 0x77,
	0x6f, 0x72, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x12, 0x0e, 0x0a, 0x02, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x6d, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x79, 0x6f, 0x75, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x79, 0x6f, 0x75, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x68, 0x61, 0x6c, 0x6c, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x05, 0x63, 0x68, 0x61, 0x6c, 0x6c, 0x22, 0x45, 0x0a, 0x05, 0x41, 0x75,
	0x74, 0x68, 0x53, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x68, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x63, 0x68, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x73,
	0x70, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x68, 0x61, 0x6c, 0x6c, 0x41, 0x64, 0x64, 0x65, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x63, 0x68, 0x61, 0x6c, 0x6c, 0x41, 0x64, 0x64, 0x65,
	0x64, 0x22, 0x21, 0x0a, 0x09, 0x41, 0x75, 0x74, 0x68, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x14,
	0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x74,
	0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x46, 0x0a, 0x04, 0x58, 0x63, 0x68, 0x51, 0x12, 0x14, 0x0a, 0x05,
	0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x74, 0x6f, 0x6b,
	0x65, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x06, 0x70, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x70, 0x73,
	0x6b, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x70, 0x73, 0x6b, 0x22, 0x30, 0x0a, 0x04,
	0x58, 0x63, 0x68, 0x53, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x70, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x70, 0x73, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x70, 0x73, 0x6b, 0x22, 0x08,
	0x0a, 0x06, 0x50, 0x69, 0x6e, 0x67, 0x51, 0x53, 0x22, 0x1d, 0x0a, 0x05, 0x50, 0x75, 0x6c, 0x6c,
	0x51, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x27, 0x0a, 0x05, 0x50, 0x75, 0x6c, 0x6c, 0x53,
	0x12, 0x1e, 0x0a, 0x02, 0x63, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x43,
	0x65, 0x6e, 0x74, 0x72, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x02, 0x63, 0x63,
	0x22, 0x97, 0x01, 0x0a, 0x0d, 0x43, 0x65, 0x6e, 0x74, 0x72, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x12, 0x38, 0x0a, 0x08, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x43, 0x65, 0x6e, 0x74, 0x72, 0x61, 0x6c, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x08, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x1a, 0x4c, 0x0a, 0x0d,
	0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x25, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f,
	0x2e, 0x43, 0x65, 0x6e, 0x74, 0x72, 0x61, 0x6c, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xed, 0x01, 0x0a, 0x0e, 0x43,
	0x65, 0x6e, 0x74, 0x72, 0x61, 0x6c, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x12, 0x18, 0x0a,
	0x03, 0x69, 0x70, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x06, 0x2e, 0x49, 0x50, 0x4e,
	0x65, 0x74, 0x52, 0x03, 0x69, 0x70, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x6d, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x6d, 0x65, 0x12, 0x37, 0x0a, 0x09, 0x6b, 0x65, 0x65, 0x70, 0x61,
	0x6c, 0x69, 0x76, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x09, 0x6b, 0x65, 0x65, 0x70, 0x61, 0x6c, 0x69, 0x76, 0x65,
	0x12, 0x30, 0x0a, 0x05, 0x70, 0x65, 0x65, 0x72, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x43, 0x65, 0x6e, 0x74, 0x72, 0x61, 0x6c, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x2e, 0x50, 0x65, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x70, 0x65, 0x65,
	0x72, 0x73, 0x1a, 0x46, 0x0a, 0x0a, 0x50, 0x65, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x22, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0c, 0x2e, 0x43, 0x65, 0x6e, 0x74, 0x72, 0x61, 0x6c, 0x50, 0x65, 0x65, 0x72, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x73, 0x0a, 0x0b, 0x43, 0x65,
	0x6e, 0x74, 0x72, 0x61, 0x6c, 0x50, 0x65, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x73,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x26, 0x0a,
	0x0a, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x49, 0x50, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x06, 0x2e, 0x49, 0x50, 0x4e, 0x65, 0x74, 0x52, 0x0a, 0x61, 0x6c, 0x6c, 0x6f, 0x77,
	0x65, 0x64, 0x49, 0x50, 0x73, 0x12, 0x28, 0x0a, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b,
	0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x50, 0x75, 0x62, 0x6c, 0x69,
	0x63, 0x4b, 0x65, 0x79, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x22,
	0x1d, 0x0a, 0x09, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x72, 0x61, 0x77, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x72, 0x61, 0x77, 0x22, 0x1b,
	0x0a, 0x05, 0x49, 0x50, 0x4e, 0x65, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x69, 0x64, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x69, 0x64, 0x72, 0x32, 0x53, 0x0a, 0x04, 0x4e,
	0x6f, 0x64, 0x65, 0x12, 0x1c, 0x0a, 0x04, 0x61, 0x75, 0x74, 0x68, 0x12, 0x07, 0x2e, 0x41, 0x75,
	0x74, 0x68, 0x53, 0x51, 0x1a, 0x07, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x53, 0x51, 0x28, 0x01, 0x30,
	0x01, 0x12, 0x13, 0x0a, 0x03, 0x78, 0x63, 0x68, 0x12, 0x05, 0x2e, 0x58, 0x63, 0x68, 0x51, 0x1a,
	0x05, 0x2e, 0x58, 0x63, 0x68, 0x53, 0x12, 0x18, 0x0a, 0x04, 0x70, 0x69, 0x6e, 0x67, 0x12, 0x07,
	0x2e, 0x50, 0x69, 0x6e, 0x67, 0x51, 0x53, 0x1a, 0x07, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x51, 0x53,
	0x32, 0x45, 0x0a, 0x0d, 0x43, 0x65, 0x6e, 0x74, 0x72, 0x61, 0x6c, 0x53, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x12, 0x1c, 0x0a, 0x04, 0x61, 0x75, 0x74, 0x68, 0x12, 0x07, 0x2e, 0x41, 0x75, 0x74, 0x68,
	0x53, 0x51, 0x1a, 0x07, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x53, 0x51, 0x28, 0x01, 0x30, 0x01, 0x12,
	0x16, 0x0a, 0x04, 0x70, 0x75, 0x6c, 0x6c, 0x12, 0x06, 0x2e, 0x50, 0x75, 0x6c, 0x6c, 0x51, 0x1a,
	0x06, 0x2e, 0x50, 0x75, 0x6c, 0x6c, 0x53, 0x42, 0x22, 0x5a, 0x20, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6e, 0x79, 0x69, 0x79, 0x75, 0x69, 0x2f, 0x71, 0x61, 0x6e,
	0x6d, 0x73, 0x2f, 0x6e, 0x6f, 0x64, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_api_proto_rawDescOnce sync.Once
	file_api_proto_rawDescData = file_api_proto_rawDesc
)

func file_api_proto_rawDescGZIP() []byte {
	file_api_proto_rawDescOnce.Do(func() {
		file_api_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_rawDescData)
	})
	return file_api_proto_rawDescData
}

var file_api_proto_msgTypes = make([]protoimpl.MessageInfo, 16)
var file_api_proto_goTypes = []interface{}{
	(*AuthSQ)(nil),              // 0: AuthSQ
	(*AuthQ)(nil),               // 1: AuthQ
	(*AuthS)(nil),               // 2: AuthS
	(*AuthToken)(nil),           // 3: AuthToken
	(*XchQ)(nil),                // 4: XchQ
	(*XchS)(nil),                // 5: XchS
	(*PingQS)(nil),              // 6: PingQS
	(*PullQ)(nil),               // 7: PullQ
	(*PullS)(nil),               // 8: PullS
	(*CentralConfig)(nil),       // 9: CentralConfig
	(*CentralNetwork)(nil),      // 10: CentralNetwork
	(*CentralPeer)(nil),         // 11: CentralPeer
	(*PublicKey)(nil),           // 12: PublicKey
	(*IPNet)(nil),               // 13: IPNet
	nil,                         // 14: CentralConfig.NetworksEntry
	nil,                         // 15: CentralNetwork.PeersEntry
	(*durationpb.Duration)(nil), // 16: google.protobuf.Duration
}
var file_api_proto_depIdxs = []int32{
	2,  // 0: AuthSQ.s:type_name -> AuthS
	1,  // 1: AuthSQ.q:type_name -> AuthQ
	3,  // 2: AuthSQ.token:type_name -> AuthToken
	9,  // 3: PullS.cc:type_name -> CentralConfig
	14, // 4: CentralConfig.networks:type_name -> CentralConfig.NetworksEntry
	13, // 5: CentralNetwork.ips:type_name -> IPNet
	16, // 6: CentralNetwork.keepalive:type_name -> google.protobuf.Duration
	15, // 7: CentralNetwork.peers:type_name -> CentralNetwork.PeersEntry
	13, // 8: CentralPeer.allowedIPs:type_name -> IPNet
	12, // 9: CentralPeer.publicKey:type_name -> PublicKey
	10, // 10: CentralConfig.NetworksEntry.value:type_name -> CentralNetwork
	11, // 11: CentralNetwork.PeersEntry.value:type_name -> CentralPeer
	0,  // 12: Node.auth:input_type -> AuthSQ
	4,  // 13: Node.xch:input_type -> XchQ
	6,  // 14: Node.ping:input_type -> PingQS
	0,  // 15: CentralSource.auth:input_type -> AuthSQ
	7,  // 16: CentralSource.pull:input_type -> PullQ
	0,  // 17: Node.auth:output_type -> AuthSQ
	5,  // 18: Node.xch:output_type -> XchS
	6,  // 19: Node.ping:output_type -> PingQS
	0,  // 20: CentralSource.auth:output_type -> AuthSQ
	8,  // 21: CentralSource.pull:output_type -> PullS
	17, // [17:22] is the sub-list for method output_type
	12, // [12:17] is the sub-list for method input_type
	12, // [12:12] is the sub-list for extension type_name
	12, // [12:12] is the sub-list for extension extendee
	0,  // [0:12] is the sub-list for field type_name
}

func init() { file_api_proto_init() }
func file_api_proto_init() {
	if File_api_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthSQ); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthQ); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthS); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthToken); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*XchQ); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*XchS); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingQS); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PullQ); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PullS); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CentralConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CentralNetwork); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CentralPeer); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PublicKey); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IPNet); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_api_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*AuthSQ_S)(nil),
		(*AuthSQ_Q)(nil),
		(*AuthSQ_Token)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   16,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_api_proto_goTypes,
		DependencyIndexes: file_api_proto_depIdxs,
		MessageInfos:      file_api_proto_msgTypes,
	}.Build()
	File_api_proto = out.File
	file_api_proto_rawDesc = nil
	file_api_proto_goTypes = nil
	file_api_proto_depIdxs = nil
}
