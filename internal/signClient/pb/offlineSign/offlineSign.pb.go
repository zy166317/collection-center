// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.4
// source: offlineSign.proto

package offlineSign

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type BtcInput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TxId         string  `protobuf:"bytes,1,opt,name=txId,proto3" json:"txId,omitempty"`
	VOut         int64   `protobuf:"varint,2,opt,name=vOut,proto3" json:"vOut,omitempty"`
	WIF          string  `protobuf:"bytes,3,opt,name=WIF,proto3" json:"WIF,omitempty"`
	RedeemScript string  `protobuf:"bytes,4,opt,name=redeemScript,proto3" json:"redeemScript,omitempty"`
	SegWit       bool    `protobuf:"varint,5,opt,name=SegWit,proto3" json:"SegWit,omitempty"`
	Amount       float64 `protobuf:"fixed64,6,opt,name=amount,proto3" json:"amount,omitempty"`
}

func (x *BtcInput) Reset() {
	*x = BtcInput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_offlineSign_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BtcInput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BtcInput) ProtoMessage() {}

func (x *BtcInput) ProtoReflect() protoreflect.Message {
	mi := &file_offlineSign_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BtcInput.ProtoReflect.Descriptor instead.
func (*BtcInput) Descriptor() ([]byte, []int) {
	return file_offlineSign_proto_rawDescGZIP(), []int{0}
}

func (x *BtcInput) GetTxId() string {
	if x != nil {
		return x.TxId
	}
	return ""
}

func (x *BtcInput) GetVOut() int64 {
	if x != nil {
		return x.VOut
	}
	return 0
}

func (x *BtcInput) GetWIF() string {
	if x != nil {
		return x.WIF
	}
	return ""
}

func (x *BtcInput) GetRedeemScript() string {
	if x != nil {
		return x.RedeemScript
	}
	return ""
}

func (x *BtcInput) GetSegWit() bool {
	if x != nil {
		return x.SegWit
	}
	return false
}

func (x *BtcInput) GetAmount() float64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

type BtcSignReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BtcInputs []*BtcInput `protobuf:"bytes,1,rep,name=BtcInputs,proto3" json:"BtcInputs,omitempty"`
	Script    string      `protobuf:"bytes,2,opt,name=script,proto3" json:"script,omitempty"`
}

func (x *BtcSignReq) Reset() {
	*x = BtcSignReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_offlineSign_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BtcSignReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BtcSignReq) ProtoMessage() {}

func (x *BtcSignReq) ProtoReflect() protoreflect.Message {
	mi := &file_offlineSign_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BtcSignReq.ProtoReflect.Descriptor instead.
func (*BtcSignReq) Descriptor() ([]byte, []int) {
	return file_offlineSign_proto_rawDescGZIP(), []int{1}
}

func (x *BtcSignReq) GetBtcInputs() []*BtcInput {
	if x != nil {
		return x.BtcInputs
	}
	return nil
}

func (x *BtcSignReq) GetScript() string {
	if x != nil {
		return x.Script
	}
	return ""
}

type BtcSignResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Signed string `protobuf:"bytes,1,opt,name=signed,proto3" json:"signed,omitempty"`
}

func (x *BtcSignResp) Reset() {
	*x = BtcSignResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_offlineSign_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BtcSignResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BtcSignResp) ProtoMessage() {}

func (x *BtcSignResp) ProtoReflect() protoreflect.Message {
	mi := &file_offlineSign_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BtcSignResp.ProtoReflect.Descriptor instead.
func (*BtcSignResp) Descriptor() ([]byte, []int) {
	return file_offlineSign_proto_rawDescGZIP(), []int{2}
}

func (x *BtcSignResp) GetSigned() string {
	if x != nil {
		return x.Signed
	}
	return ""
}

type EthSignReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TxBinaryText string `protobuf:"bytes,1,opt,name=txBinaryText,proto3" json:"txBinaryText,omitempty"`
	ChainID      int64  `protobuf:"varint,2,opt,name=chainID,proto3" json:"chainID,omitempty"`
}

func (x *EthSignReq) Reset() {
	*x = EthSignReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_offlineSign_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EthSignReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EthSignReq) ProtoMessage() {}

func (x *EthSignReq) ProtoReflect() protoreflect.Message {
	mi := &file_offlineSign_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EthSignReq.ProtoReflect.Descriptor instead.
func (*EthSignReq) Descriptor() ([]byte, []int) {
	return file_offlineSign_proto_rawDescGZIP(), []int{3}
}

func (x *EthSignReq) GetTxBinaryText() string {
	if x != nil {
		return x.TxBinaryText
	}
	return ""
}

func (x *EthSignReq) GetChainID() int64 {
	if x != nil {
		return x.ChainID
	}
	return 0
}

type EthSignResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SignedTxBinary string `protobuf:"bytes,1,opt,name=signedTxBinary,proto3" json:"signedTxBinary,omitempty"`
}

func (x *EthSignResp) Reset() {
	*x = EthSignResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_offlineSign_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EthSignResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EthSignResp) ProtoMessage() {}

func (x *EthSignResp) ProtoReflect() protoreflect.Message {
	mi := &file_offlineSign_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EthSignResp.ProtoReflect.Descriptor instead.
func (*EthSignResp) Descriptor() ([]byte, []int) {
	return file_offlineSign_proto_rawDescGZIP(), []int{4}
}

func (x *EthSignResp) GetSignedTxBinary() string {
	if x != nil {
		return x.SignedTxBinary
	}
	return ""
}

var File_offlineSign_proto protoreflect.FileDescriptor

var file_offlineSign_proto_rawDesc = []byte{
	0x0a, 0x11, 0x6f, 0x66, 0x66, 0x6c, 0x69, 0x6e, 0x65, 0x53, 0x69, 0x67, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x98, 0x01, 0x0a, 0x08, 0x42, 0x74, 0x63, 0x49, 0x6e, 0x70, 0x75, 0x74,
	0x12, 0x12, 0x0a, 0x04, 0x74, 0x78, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x74, 0x78, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x76, 0x4f, 0x75, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x04, 0x76, 0x4f, 0x75, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x57, 0x49, 0x46, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x57, 0x49, 0x46, 0x12, 0x22, 0x0a, 0x0c, 0x72, 0x65,
	0x64, 0x65, 0x65, 0x6d, 0x53, 0x63, 0x72, 0x69, 0x70, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x72, 0x65, 0x64, 0x65, 0x65, 0x6d, 0x53, 0x63, 0x72, 0x69, 0x70, 0x74, 0x12, 0x16,
	0x0a, 0x06, 0x53, 0x65, 0x67, 0x57, 0x69, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06,
	0x53, 0x65, 0x67, 0x57, 0x69, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x01, 0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x4d,
	0x0a, 0x0a, 0x42, 0x74, 0x63, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x65, 0x71, 0x12, 0x27, 0x0a, 0x09,
	0x42, 0x74, 0x63, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x09, 0x2e, 0x42, 0x74, 0x63, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x52, 0x09, 0x42, 0x74, 0x63, 0x49,
	0x6e, 0x70, 0x75, 0x74, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x22, 0x25, 0x0a,
	0x0b, 0x42, 0x74, 0x63, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x12, 0x16, 0x0a, 0x06,
	0x73, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x69,
	0x67, 0x6e, 0x65, 0x64, 0x22, 0x4a, 0x0a, 0x0a, 0x45, 0x74, 0x68, 0x53, 0x69, 0x67, 0x6e, 0x52,
	0x65, 0x71, 0x12, 0x22, 0x0a, 0x0c, 0x74, 0x78, 0x42, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x54, 0x65,
	0x78, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x74, 0x78, 0x42, 0x69, 0x6e, 0x61,
	0x72, 0x79, 0x54, 0x65, 0x78, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x49,
	0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x44,
	0x22, 0x35, 0x0a, 0x0b, 0x45, 0x74, 0x68, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x12,
	0x26, 0x0a, 0x0e, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x54, 0x78, 0x42, 0x69, 0x6e, 0x61, 0x72,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x54,
	0x78, 0x42, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x32, 0x59, 0x0a, 0x0b, 0x4f, 0x66, 0x66, 0x6c, 0x69,
	0x6e, 0x65, 0x53, 0x69, 0x67, 0x6e, 0x12, 0x24, 0x0a, 0x07, 0x42, 0x74, 0x63, 0x53, 0x69, 0x67,
	0x6e, 0x12, 0x0b, 0x2e, 0x42, 0x74, 0x63, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x65, 0x71, 0x1a, 0x0c,
	0x2e, 0x42, 0x74, 0x63, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x12, 0x24, 0x0a, 0x07,
	0x45, 0x74, 0x68, 0x53, 0x69, 0x67, 0x6e, 0x12, 0x0b, 0x2e, 0x45, 0x74, 0x68, 0x53, 0x69, 0x67,
	0x6e, 0x52, 0x65, 0x71, 0x1a, 0x0c, 0x2e, 0x45, 0x74, 0x68, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x65,
	0x73, 0x70, 0x42, 0x10, 0x5a, 0x0e, 0x70, 0x62, 0x2f, 0x6f, 0x66, 0x66, 0x6c, 0x69, 0x6e, 0x65,
	0x53, 0x69, 0x67, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_offlineSign_proto_rawDescOnce sync.Once
	file_offlineSign_proto_rawDescData = file_offlineSign_proto_rawDesc
)

func file_offlineSign_proto_rawDescGZIP() []byte {
	file_offlineSign_proto_rawDescOnce.Do(func() {
		file_offlineSign_proto_rawDescData = protoimpl.X.CompressGZIP(file_offlineSign_proto_rawDescData)
	})
	return file_offlineSign_proto_rawDescData
}

var file_offlineSign_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_offlineSign_proto_goTypes = []interface{}{
	(*BtcInput)(nil),    // 0: BtcInput
	(*BtcSignReq)(nil),  // 1: BtcSignReq
	(*BtcSignResp)(nil), // 2: BtcSignResp
	(*EthSignReq)(nil),  // 3: EthSignReq
	(*EthSignResp)(nil), // 4: EthSignResp
}
var file_offlineSign_proto_depIdxs = []int32{
	0, // 0: BtcSignReq.BtcInputs:type_name -> BtcInput
	1, // 1: OfflineSign.BtcSign:input_type -> BtcSignReq
	3, // 2: OfflineSign.EthSign:input_type -> EthSignReq
	2, // 3: OfflineSign.BtcSign:output_type -> BtcSignResp
	4, // 4: OfflineSign.EthSign:output_type -> EthSignResp
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_offlineSign_proto_init() }
func file_offlineSign_proto_init() {
	if File_offlineSign_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_offlineSign_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BtcInput); i {
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
		file_offlineSign_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BtcSignReq); i {
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
		file_offlineSign_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BtcSignResp); i {
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
		file_offlineSign_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EthSignReq); i {
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
		file_offlineSign_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EthSignResp); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_offlineSign_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_offlineSign_proto_goTypes,
		DependencyIndexes: file_offlineSign_proto_depIdxs,
		MessageInfos:      file_offlineSign_proto_msgTypes,
	}.Build()
	File_offlineSign_proto = out.File
	file_offlineSign_proto_rawDesc = nil
	file_offlineSign_proto_goTypes = nil
	file_offlineSign_proto_depIdxs = nil
}
