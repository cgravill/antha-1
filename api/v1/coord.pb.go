// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/antha-lang/antha/api/v1/coord.proto

/*
Package org_antha_lang_antha_v1 is a generated protocol buffer package.

It is generated from these files:
	github.com/antha-lang/antha/api/v1/coord.proto
	github.com/antha-lang/antha/api/v1/inventory.proto
	github.com/antha-lang/antha/api/v1/measurement.proto
	github.com/antha-lang/antha/api/v1/message.proto
	github.com/antha-lang/antha/api/v1/state.proto
	github.com/antha-lang/antha/api/v1/task.proto
	github.com/antha-lang/antha/api/v1/blob.proto
	github.com/antha-lang/antha/api/v1/polynomial.proto
	github.com/antha-lang/antha/api/v1/workflow.proto
	github.com/antha-lang/antha/api/v1/empty.proto
	github.com/antha-lang/antha/api/v1/element.proto
	github.com/antha-lang/antha/api/v1/device.proto

It has these top-level messages:
	OrdinalCoord
	PhysicalCoord
	InventoryItem
	None
	Tipbox
	Tipwaste
	DeckPosition
	Plate
	Well
	Component
	PlateType
	Measurement
	GrpcMessage
	GrpcCall
	Status
	Task
	OrderTask
	DeckLayoutTask
	PlatePrepTask
	PlatePrep
	DocumentTask
	MixerTask
	MixerState
	Placement
	ManualRunTask
	IncubateTask
	DataUploadTask
	Blob
	FromBytes
	FromHostFile
	Polynomial
	Workflow
	Element
	Connection
	ProcessAddress
	WorkflowParameters
	ElementParameters
	MixerOpt
	Empty
	ElementMetadata
	DeviceMetadata
*/
package org_antha_lang_antha_v1

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Zero-indexed coordinate system in ordinal space: ith item in X, Y, Z space.
// Origin is back, left, bottom (i.e., left-handed)
type OrdinalCoord struct {
	X int32 `protobuf:"varint,1,opt,name=x" json:"x,omitempty"`
	Y int32 `protobuf:"varint,2,opt,name=y" json:"y,omitempty"`
	Z int32 `protobuf:"varint,3,opt,name=z" json:"z,omitempty"`
}

func (m *OrdinalCoord) Reset()                    { *m = OrdinalCoord{} }
func (m *OrdinalCoord) String() string            { return proto.CompactTextString(m) }
func (*OrdinalCoord) ProtoMessage()               {}
func (*OrdinalCoord) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *OrdinalCoord) GetX() int32 {
	if m != nil {
		return m.X
	}
	return 0
}

func (m *OrdinalCoord) GetY() int32 {
	if m != nil {
		return m.Y
	}
	return 0
}

func (m *OrdinalCoord) GetZ() int32 {
	if m != nil {
		return m.Z
	}
	return 0
}

type PhysicalCoord struct {
	X *Measurement `protobuf:"bytes,1,opt,name=x" json:"x,omitempty"`
	Y *Measurement `protobuf:"bytes,2,opt,name=y" json:"y,omitempty"`
	Z *Measurement `protobuf:"bytes,3,opt,name=z" json:"z,omitempty"`
}

func (m *PhysicalCoord) Reset()                    { *m = PhysicalCoord{} }
func (m *PhysicalCoord) String() string            { return proto.CompactTextString(m) }
func (*PhysicalCoord) ProtoMessage()               {}
func (*PhysicalCoord) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *PhysicalCoord) GetX() *Measurement {
	if m != nil {
		return m.X
	}
	return nil
}

func (m *PhysicalCoord) GetY() *Measurement {
	if m != nil {
		return m.Y
	}
	return nil
}

func (m *PhysicalCoord) GetZ() *Measurement {
	if m != nil {
		return m.Z
	}
	return nil
}

func init() {
	proto.RegisterType((*OrdinalCoord)(nil), "org.antha_lang.antha.v1.OrdinalCoord")
	proto.RegisterType((*PhysicalCoord)(nil), "org.antha_lang.antha.v1.PhysicalCoord")
}

func init() { proto.RegisterFile("github.com/antha-lang/antha/api/v1/coord.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 193 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xd2, 0x4b, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0xcc, 0x2b, 0xc9, 0x48, 0xd4, 0xcd, 0x49, 0xcc,
	0x4b, 0x87, 0x30, 0xf5, 0x13, 0x0b, 0x32, 0xf5, 0xcb, 0x0c, 0xf5, 0x93, 0xf3, 0xf3, 0x8b, 0x52,
	0xf4, 0x0a, 0x8a, 0xf2, 0x4b, 0xf2, 0x85, 0xc4, 0xf3, 0x8b, 0xd2, 0xf5, 0xc0, 0xb2, 0xf1, 0x20,
	0x85, 0x10, 0xa6, 0x5e, 0x99, 0xa1, 0x94, 0x09, 0x11, 0x06, 0xe5, 0xa6, 0x26, 0x16, 0x97, 0x16,
	0xa5, 0xe6, 0xa6, 0xe6, 0x95, 0x40, 0x8c, 0x53, 0xb2, 0xe0, 0xe2, 0xf1, 0x2f, 0x4a, 0xc9, 0xcc,
	0x4b, 0xcc, 0x71, 0x06, 0x59, 0x22, 0xc4, 0xc3, 0xc5, 0x58, 0x21, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1,
	0x1a, 0xc4, 0x58, 0x01, 0xe2, 0x55, 0x4a, 0x30, 0x41, 0x78, 0x95, 0x20, 0x5e, 0x95, 0x04, 0x33,
	0x84, 0x57, 0xa5, 0xb4, 0x9a, 0x91, 0x8b, 0x37, 0x20, 0xa3, 0xb2, 0x38, 0x33, 0x19, 0xa6, 0xd7,
	0x08, 0xa6, 0x97, 0xdb, 0x48, 0x45, 0x0f, 0x87, 0x33, 0xf5, 0x7c, 0x11, 0x4e, 0x00, 0xd9, 0x60,
	0x04, 0xb3, 0x81, 0x68, 0x3d, 0x95, 0x20, 0x3d, 0x10, 0x77, 0x10, 0xad, 0xa7, 0x2a, 0x89, 0x0d,
	0xec, 0x5d, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0xcc, 0xfe, 0x10, 0x55, 0x6f, 0x01, 0x00,
	0x00,
}
