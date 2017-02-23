// Code generated by protoc-gen-go.
// source: github.com/antha-lang/antha/api/v1/inventory.proto
// DO NOT EDIT!

package org_antha_lang_antha_v1

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/any"
import google_protobuf1 "github.com/golang/protobuf/ptypes/timestamp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type InventoryItem struct {
	// Inventory id
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	// Metadata
	Metadata map[string]*google_protobuf.Any `protobuf:"bytes,2,rep,name=metadata" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// Time this inventory item was created at
	CreatedAt *google_protobuf1.Timestamp `protobuf:"bytes,3,opt,name=created_at,json=createdAt" json:"created_at,omitempty"`
	// History of this inventory item
	FromItems []*InventoryItem `protobuf:"bytes,4,rep,name=from_items,json=fromItems" json:"from_items,omitempty"`
	// Types that are valid to be assigned to Item:
	//	*InventoryItem_Tipbox
	//	*InventoryItem_Tipwaste
	//	*InventoryItem_Plate
	//	*InventoryItem_DeckPosition
	//	*InventoryItem_Component
	//	*InventoryItem_PlateType
	//	*InventoryItem_None
	Item isInventoryItem_Item `protobuf_oneof:"item"`
}

func (m *InventoryItem) Reset()                    { *m = InventoryItem{} }
func (m *InventoryItem) String() string            { return proto.CompactTextString(m) }
func (*InventoryItem) ProtoMessage()               {}
func (*InventoryItem) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

type isInventoryItem_Item interface {
	isInventoryItem_Item()
}

type InventoryItem_Tipbox struct {
	Tipbox *Tipbox `protobuf:"bytes,5,opt,name=tipbox,oneof"`
}
type InventoryItem_Tipwaste struct {
	Tipwaste *Tipwaste `protobuf:"bytes,6,opt,name=tipwaste,oneof"`
}
type InventoryItem_Plate struct {
	Plate *Plate `protobuf:"bytes,7,opt,name=plate,oneof"`
}
type InventoryItem_DeckPosition struct {
	DeckPosition *DeckPosition `protobuf:"bytes,8,opt,name=deck_position,json=deckPosition,oneof"`
}
type InventoryItem_Component struct {
	Component *Component `protobuf:"bytes,9,opt,name=component,oneof"`
}
type InventoryItem_PlateType struct {
	PlateType *PlateType `protobuf:"bytes,10,opt,name=plate_type,json=plateType,oneof"`
}
type InventoryItem_None struct {
	None *None `protobuf:"bytes,11,opt,name=none,oneof"`
}

func (*InventoryItem_Tipbox) isInventoryItem_Item()       {}
func (*InventoryItem_Tipwaste) isInventoryItem_Item()     {}
func (*InventoryItem_Plate) isInventoryItem_Item()        {}
func (*InventoryItem_DeckPosition) isInventoryItem_Item() {}
func (*InventoryItem_Component) isInventoryItem_Item()    {}
func (*InventoryItem_PlateType) isInventoryItem_Item()    {}
func (*InventoryItem_None) isInventoryItem_Item()         {}

func (m *InventoryItem) GetItem() isInventoryItem_Item {
	if m != nil {
		return m.Item
	}
	return nil
}

func (m *InventoryItem) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *InventoryItem) GetMetadata() map[string]*google_protobuf.Any {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func (m *InventoryItem) GetCreatedAt() *google_protobuf1.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

func (m *InventoryItem) GetFromItems() []*InventoryItem {
	if m != nil {
		return m.FromItems
	}
	return nil
}

func (m *InventoryItem) GetTipbox() *Tipbox {
	if x, ok := m.GetItem().(*InventoryItem_Tipbox); ok {
		return x.Tipbox
	}
	return nil
}

func (m *InventoryItem) GetTipwaste() *Tipwaste {
	if x, ok := m.GetItem().(*InventoryItem_Tipwaste); ok {
		return x.Tipwaste
	}
	return nil
}

func (m *InventoryItem) GetPlate() *Plate {
	if x, ok := m.GetItem().(*InventoryItem_Plate); ok {
		return x.Plate
	}
	return nil
}

func (m *InventoryItem) GetDeckPosition() *DeckPosition {
	if x, ok := m.GetItem().(*InventoryItem_DeckPosition); ok {
		return x.DeckPosition
	}
	return nil
}

func (m *InventoryItem) GetComponent() *Component {
	if x, ok := m.GetItem().(*InventoryItem_Component); ok {
		return x.Component
	}
	return nil
}

func (m *InventoryItem) GetPlateType() *PlateType {
	if x, ok := m.GetItem().(*InventoryItem_PlateType); ok {
		return x.PlateType
	}
	return nil
}

func (m *InventoryItem) GetNone() *None {
	if x, ok := m.GetItem().(*InventoryItem_None); ok {
		return x.None
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*InventoryItem) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _InventoryItem_OneofMarshaler, _InventoryItem_OneofUnmarshaler, _InventoryItem_OneofSizer, []interface{}{
		(*InventoryItem_Tipbox)(nil),
		(*InventoryItem_Tipwaste)(nil),
		(*InventoryItem_Plate)(nil),
		(*InventoryItem_DeckPosition)(nil),
		(*InventoryItem_Component)(nil),
		(*InventoryItem_PlateType)(nil),
		(*InventoryItem_None)(nil),
	}
}

func _InventoryItem_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*InventoryItem)
	// item
	switch x := m.Item.(type) {
	case *InventoryItem_Tipbox:
		b.EncodeVarint(5<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Tipbox); err != nil {
			return err
		}
	case *InventoryItem_Tipwaste:
		b.EncodeVarint(6<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Tipwaste); err != nil {
			return err
		}
	case *InventoryItem_Plate:
		b.EncodeVarint(7<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Plate); err != nil {
			return err
		}
	case *InventoryItem_DeckPosition:
		b.EncodeVarint(8<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.DeckPosition); err != nil {
			return err
		}
	case *InventoryItem_Component:
		b.EncodeVarint(9<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Component); err != nil {
			return err
		}
	case *InventoryItem_PlateType:
		b.EncodeVarint(10<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.PlateType); err != nil {
			return err
		}
	case *InventoryItem_None:
		b.EncodeVarint(11<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.None); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("InventoryItem.Item has unexpected type %T", x)
	}
	return nil
}

func _InventoryItem_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*InventoryItem)
	switch tag {
	case 5: // item.tipbox
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Tipbox)
		err := b.DecodeMessage(msg)
		m.Item = &InventoryItem_Tipbox{msg}
		return true, err
	case 6: // item.tipwaste
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Tipwaste)
		err := b.DecodeMessage(msg)
		m.Item = &InventoryItem_Tipwaste{msg}
		return true, err
	case 7: // item.plate
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Plate)
		err := b.DecodeMessage(msg)
		m.Item = &InventoryItem_Plate{msg}
		return true, err
	case 8: // item.deck_position
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(DeckPosition)
		err := b.DecodeMessage(msg)
		m.Item = &InventoryItem_DeckPosition{msg}
		return true, err
	case 9: // item.component
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Component)
		err := b.DecodeMessage(msg)
		m.Item = &InventoryItem_Component{msg}
		return true, err
	case 10: // item.plate_type
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(PlateType)
		err := b.DecodeMessage(msg)
		m.Item = &InventoryItem_PlateType{msg}
		return true, err
	case 11: // item.none
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(None)
		err := b.DecodeMessage(msg)
		m.Item = &InventoryItem_None{msg}
		return true, err
	default:
		return false, nil
	}
}

func _InventoryItem_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*InventoryItem)
	// item
	switch x := m.Item.(type) {
	case *InventoryItem_Tipbox:
		s := proto.Size(x.Tipbox)
		n += proto.SizeVarint(5<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *InventoryItem_Tipwaste:
		s := proto.Size(x.Tipwaste)
		n += proto.SizeVarint(6<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *InventoryItem_Plate:
		s := proto.Size(x.Plate)
		n += proto.SizeVarint(7<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *InventoryItem_DeckPosition:
		s := proto.Size(x.DeckPosition)
		n += proto.SizeVarint(8<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *InventoryItem_Component:
		s := proto.Size(x.Component)
		n += proto.SizeVarint(9<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *InventoryItem_PlateType:
		s := proto.Size(x.PlateType)
		n += proto.SizeVarint(10<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *InventoryItem_None:
		s := proto.Size(x.None)
		n += proto.SizeVarint(11<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type None struct {
}

func (m *None) Reset()                    { *m = None{} }
func (m *None) String() string            { return proto.CompactTextString(m) }
func (*None) ProtoMessage()               {}
func (*None) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

// Pipette tips in a box
type Tipbox struct {
	// Tipbox type
	Type string `protobuf:"bytes,1,opt,name=type" json:"type,omitempty"`
}

func (m *Tipbox) Reset()                    { *m = Tipbox{} }
func (m *Tipbox) String() string            { return proto.CompactTextString(m) }
func (*Tipbox) ProtoMessage()               {}
func (*Tipbox) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{2} }

func (m *Tipbox) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

// Disposal for used pipette tips
type Tipwaste struct {
	// Tipwaste type
	Type string `protobuf:"bytes,1,opt,name=type" json:"type,omitempty"`
}

func (m *Tipwaste) Reset()                    { *m = Tipwaste{} }
func (m *Tipwaste) String() string            { return proto.CompactTextString(m) }
func (*Tipwaste) ProtoMessage()               {}
func (*Tipwaste) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{3} }

func (m *Tipwaste) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

// Synthetic inventory item to represent position on deck
type DeckPosition struct {
	// Position
	Position string `protobuf:"bytes,1,opt,name=position" json:"position,omitempty"`
}

func (m *DeckPosition) Reset()                    { *m = DeckPosition{} }
func (m *DeckPosition) String() string            { return proto.CompactTextString(m) }
func (*DeckPosition) ProtoMessage()               {}
func (*DeckPosition) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{4} }

func (m *DeckPosition) GetPosition() string {
	if m != nil {
		return m.Position
	}
	return ""
}

// Plate
type Plate struct {
	// Plate type
	Type  string  `protobuf:"bytes,1,opt,name=type" json:"type,omitempty"`
	Wells []*Well `protobuf:"bytes,2,rep,name=wells" json:"wells,omitempty"`
}

func (m *Plate) Reset()                    { *m = Plate{} }
func (m *Plate) String() string            { return proto.CompactTextString(m) }
func (*Plate) ProtoMessage()               {}
func (*Plate) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{5} }

func (m *Plate) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *Plate) GetWells() []*Well {
	if m != nil {
		return m.Wells
	}
	return nil
}

// Well in plate
type Well struct {
	Position  *OrdinalCoord  `protobuf:"bytes,1,opt,name=position" json:"position,omitempty"`
	Component *InventoryItem `protobuf:"bytes,2,opt,name=component" json:"component,omitempty"`
}

func (m *Well) Reset()                    { *m = Well{} }
func (m *Well) String() string            { return proto.CompactTextString(m) }
func (*Well) ProtoMessage()               {}
func (*Well) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{6} }

func (m *Well) GetPosition() *OrdinalCoord {
	if m != nil {
		return m.Position
	}
	return nil
}

func (m *Well) GetComponent() *InventoryItem {
	if m != nil {
		return m.Component
	}
	return nil
}

// Physical component, typically a liquid
type Component struct {
	// Component type
	Type string `protobuf:"bytes,1,opt,name=type" json:"type,omitempty"`
	// Name
	Name string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	// Volume
	Volume *Measurement `protobuf:"bytes,3,opt,name=volume" json:"volume,omitempty"`
	// Viscosity
	Viscosity *Measurement `protobuf:"bytes,4,opt,name=viscosity" json:"viscosity,omitempty"`
	// Mass
	Mass *Measurement `protobuf:"bytes,5,opt,name=mass" json:"mass,omitempty"`
	// Amount (moles)
	Amount *Measurement `protobuf:"bytes,6,opt,name=amount" json:"amount,omitempty"`
	// If non-atomic component, this is what we are comprised of
	Components []*Component `protobuf:"bytes,7,rep,name=components" json:"components,omitempty"`
}

func (m *Component) Reset()                    { *m = Component{} }
func (m *Component) String() string            { return proto.CompactTextString(m) }
func (*Component) ProtoMessage()               {}
func (*Component) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{7} }

func (m *Component) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *Component) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Component) GetVolume() *Measurement {
	if m != nil {
		return m.Volume
	}
	return nil
}

func (m *Component) GetViscosity() *Measurement {
	if m != nil {
		return m.Viscosity
	}
	return nil
}

func (m *Component) GetMass() *Measurement {
	if m != nil {
		return m.Mass
	}
	return nil
}

func (m *Component) GetAmount() *Measurement {
	if m != nil {
		return m.Amount
	}
	return nil
}

func (m *Component) GetComponents() []*Component {
	if m != nil {
		return m.Components
	}
	return nil
}

// PlateType describes the properties of a class of plates
type PlateType struct {
	// Dimensions of bounding box of plate
	Dim *PhysicalCoord `protobuf:"bytes,1,opt,name=dim" json:"dim,omitempty"`
	// Dimensions of a bounding box of well including well bottom
	WellDim *PhysicalCoord `protobuf:"bytes,2,opt,name=well_dim,json=wellDim" json:"well_dim,omitempty"`
	// Number of wells
	NumWells *OrdinalCoord `protobuf:"bytes,3,opt,name=num_wells,json=numWells" json:"num_wells,omitempty"`
	// Maximum volume of a well
	MaxVolume *Measurement `protobuf:"bytes,4,opt,name=max_volume,json=maxVolume" json:"max_volume,omitempty"`
	// Shape of well in x-y plane
	WellShape string `protobuf:"bytes,5,opt,name=well_shape,json=wellShape" json:"well_shape,omitempty"`
	// Function relating volume of well (uL) to x-y area (mm^2)
	VolumeUlToAreaMm2 *Polynomial `protobuf:"bytes,6,opt,name=volume_ul_to_area_mm2,json=volumeUlToAreaMm2" json:"volume_ul_to_area_mm2,omitempty"`
	// Function relating volume of well (ul) to its z height (mm)
	VolumeUlToHeightMm *Polynomial `protobuf:"bytes,7,opt,name=volume_ul_to_height_mm,json=volumeUlToHeightMm" json:"volume_ul_to_height_mm,omitempty"`
	// Residual volume of a well
	ResidualVolume *Measurement `protobuf:"bytes,8,opt,name=residual_volume,json=residualVolume" json:"residual_volume,omitempty"`
	// Distance between well centers in (x,y) dimension
	WellOffset *PhysicalCoord `protobuf:"bytes,9,opt,name=well_offset,json=wellOffset" json:"well_offset,omitempty"`
	// Distance from origin of plate to bottom, center of first well (including
	// well bottom)
	WellOrigin *PhysicalCoord `protobuf:"bytes,10,opt,name=well_origin,json=wellOrigin" json:"well_origin,omitempty"`
	// Manufacturer
	Mnfr string `protobuf:"bytes,11,opt,name=mnfr" json:"mnfr,omitempty"`
	// Height of well bottom. Deprecated in favor of VolumeUlToAreaMm2 and
	// VolumeUlToHeightMm.
	WellBottomHeight *Measurement `protobuf:"bytes,1000,opt,name=well_bottom_height,json=wellBottomHeight" json:"well_bottom_height,omitempty"`
	// Shape of well bottom. Deprecated in favor of VolumeUlToAreaMm2 and
	// VolumeUlToHeightMm.
	WellBottomShape string `protobuf:"bytes,1001,opt,name=well_bottom_shape,json=wellBottomShape" json:"well_bottom_shape,omitempty"`
}

func (m *PlateType) Reset()                    { *m = PlateType{} }
func (m *PlateType) String() string            { return proto.CompactTextString(m) }
func (*PlateType) ProtoMessage()               {}
func (*PlateType) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{8} }

func (m *PlateType) GetDim() *PhysicalCoord {
	if m != nil {
		return m.Dim
	}
	return nil
}

func (m *PlateType) GetWellDim() *PhysicalCoord {
	if m != nil {
		return m.WellDim
	}
	return nil
}

func (m *PlateType) GetNumWells() *OrdinalCoord {
	if m != nil {
		return m.NumWells
	}
	return nil
}

func (m *PlateType) GetMaxVolume() *Measurement {
	if m != nil {
		return m.MaxVolume
	}
	return nil
}

func (m *PlateType) GetWellShape() string {
	if m != nil {
		return m.WellShape
	}
	return ""
}

func (m *PlateType) GetVolumeUlToAreaMm2() *Polynomial {
	if m != nil {
		return m.VolumeUlToAreaMm2
	}
	return nil
}

func (m *PlateType) GetVolumeUlToHeightMm() *Polynomial {
	if m != nil {
		return m.VolumeUlToHeightMm
	}
	return nil
}

func (m *PlateType) GetResidualVolume() *Measurement {
	if m != nil {
		return m.ResidualVolume
	}
	return nil
}

func (m *PlateType) GetWellOffset() *PhysicalCoord {
	if m != nil {
		return m.WellOffset
	}
	return nil
}

func (m *PlateType) GetWellOrigin() *PhysicalCoord {
	if m != nil {
		return m.WellOrigin
	}
	return nil
}

func (m *PlateType) GetMnfr() string {
	if m != nil {
		return m.Mnfr
	}
	return ""
}

func (m *PlateType) GetWellBottomHeight() *Measurement {
	if m != nil {
		return m.WellBottomHeight
	}
	return nil
}

func (m *PlateType) GetWellBottomShape() string {
	if m != nil {
		return m.WellBottomShape
	}
	return ""
}

func init() {
	proto.RegisterType((*InventoryItem)(nil), "org.antha_lang.antha.v1.InventoryItem")
	proto.RegisterType((*None)(nil), "org.antha_lang.antha.v1.None")
	proto.RegisterType((*Tipbox)(nil), "org.antha_lang.antha.v1.Tipbox")
	proto.RegisterType((*Tipwaste)(nil), "org.antha_lang.antha.v1.Tipwaste")
	proto.RegisterType((*DeckPosition)(nil), "org.antha_lang.antha.v1.DeckPosition")
	proto.RegisterType((*Plate)(nil), "org.antha_lang.antha.v1.Plate")
	proto.RegisterType((*Well)(nil), "org.antha_lang.antha.v1.Well")
	proto.RegisterType((*Component)(nil), "org.antha_lang.antha.v1.Component")
	proto.RegisterType((*PlateType)(nil), "org.antha_lang.antha.v1.PlateType")
}

func init() { proto.RegisterFile("github.com/antha-lang/antha/api/v1/inventory.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 944 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x94, 0x95, 0xeb, 0x6e, 0xdb, 0x36,
	0x18, 0x86, 0x7d, 0xb6, 0xf5, 0xa5, 0xe9, 0x81, 0xd8, 0x41, 0x33, 0xd6, 0x36, 0xf3, 0x0e, 0x08,
	0x3a, 0x4c, 0x46, 0x9d, 0x62, 0x48, 0x87, 0x01, 0x43, 0x0e, 0xc5, 0x5c, 0x60, 0x5e, 0x3d, 0x35,
	0x5d, 0x7f, 0x0a, 0xb4, 0x45, 0xdb, 0x44, 0x44, 0x52, 0x90, 0x28, 0x37, 0xbe, 0x89, 0xed, 0xf2,
	0x76, 0x0b, 0xdb, 0xcf, 0xdd, 0xc1, 0xc0, 0x83, 0x6c, 0xa7, 0x8b, 0x52, 0xf9, 0x1f, 0x65, 0xbf,
	0xcf, 0x4b, 0xf2, 0xfb, 0x5e, 0x92, 0x30, 0x98, 0x53, 0xb9, 0xc8, 0x26, 0xde, 0x54, 0xb0, 0x3e,
	0xe6, 0x72, 0x81, 0xbf, 0x8b, 0x30, 0x9f, 0x9b, 0x61, 0x1f, 0xc7, 0xb4, 0xbf, 0x7c, 0xda, 0xa7,
	0x7c, 0x49, 0xb8, 0x14, 0xc9, 0xca, 0x8b, 0x13, 0x21, 0x05, 0xfa, 0x54, 0x24, 0x73, 0x4f, 0x2b,
	0x02, 0x25, 0x36, 0x43, 0x6f, 0xf9, 0xb4, 0xfb, 0xd9, 0x5c, 0x88, 0x79, 0x44, 0xfa, 0x5a, 0x36,
	0xc9, 0x66, 0x7d, 0xcc, 0x2d, 0xd3, 0x7d, 0xfc, 0xfe, 0x5f, 0x92, 0x32, 0x92, 0x4a, 0xcc, 0x62,
	0x2b, 0xf0, 0x4a, 0x2c, 0x64, 0x2a, 0x44, 0x12, 0x5a, 0xfd, 0x51, 0x09, 0x7d, 0x2c, 0xa2, 0x15,
	0x17, 0x8c, 0xe2, 0xc8, 0x42, 0xcf, 0x4a, 0x40, 0x8c, 0xe0, 0x34, 0x4b, 0x08, 0x23, 0x5c, 0x1a,
	0xaa, 0xf7, 0x47, 0x0b, 0xf6, 0x5f, 0xe6, 0x35, 0x78, 0x29, 0x09, 0x43, 0x77, 0xa1, 0x46, 0x43,
	0xb7, 0x7a, 0x50, 0x3d, 0x74, 0xfc, 0x1a, 0x0d, 0xd1, 0x18, 0x3a, 0x8c, 0x48, 0x1c, 0x62, 0x89,
	0xdd, 0xda, 0x41, 0xfd, 0x70, 0x6f, 0xf0, 0xcc, 0x2b, 0x28, 0x92, 0x77, 0xcd, 0xc9, 0x1b, 0x59,
	0xec, 0x05, 0x97, 0xc9, 0xca, 0x5f, 0xbb, 0xa0, 0xe7, 0x00, 0xd3, 0x84, 0x60, 0x49, 0xc2, 0x00,
	0x4b, 0xb7, 0x7e, 0x50, 0x3d, 0xdc, 0x1b, 0x74, 0x3d, 0x53, 0x44, 0x2f, 0x2f, 0xa2, 0x77, 0x91,
	0x17, 0xd1, 0x77, 0xac, 0xfa, 0x44, 0xa2, 0x17, 0x00, 0xb3, 0x44, 0xb0, 0x80, 0x4a, 0xc2, 0x52,
	0xb7, 0xa1, 0x97, 0xf3, 0x4d, 0xb9, 0xe5, 0xf8, 0x8e, 0x22, 0xd5, 0x28, 0x45, 0xcf, 0xa1, 0x25,
	0x69, 0x3c, 0x11, 0x57, 0x6e, 0x53, 0xcf, 0xfe, 0xb8, 0xd0, 0xe2, 0x42, 0xcb, 0x86, 0x15, 0xdf,
	0x02, 0xe8, 0x27, 0xe8, 0x48, 0x1a, 0xbf, 0xc3, 0xa9, 0x24, 0x6e, 0x4b, 0xc3, 0x5f, 0xdc, 0x06,
	0x6b, 0xe1, 0xb0, 0xe2, 0xaf, 0x21, 0xf4, 0x3d, 0x34, 0xe3, 0x08, 0x4b, 0xe2, 0xb6, 0x35, 0xfd,
	0xa8, 0x90, 0x1e, 0x2b, 0xd5, 0xb0, 0xe2, 0x1b, 0x39, 0xfa, 0x05, 0xf6, 0x43, 0x32, 0xbd, 0x0c,
	0x62, 0x91, 0x52, 0x49, 0x05, 0x77, 0x3b, 0x9a, 0xff, 0xba, 0x90, 0x3f, 0x27, 0xd3, 0xcb, 0xb1,
	0x15, 0x0f, 0x2b, 0xfe, 0x9d, 0x70, 0xeb, 0x1b, 0x9d, 0x82, 0x33, 0x15, 0x2c, 0x16, 0x9c, 0x70,
	0xe9, 0x3a, 0xda, 0xa9, 0x57, 0xe8, 0x74, 0x96, 0x2b, 0x87, 0x15, 0x7f, 0x83, 0xa1, 0x33, 0x00,
	0xbd, 0xb4, 0x40, 0xae, 0x62, 0xe2, 0xc2, 0x07, 0x4c, 0xf4, 0x76, 0x2e, 0x56, 0xb1, 0xda, 0x92,
	0x13, 0xe7, 0x1f, 0xe8, 0x08, 0x1a, 0x5c, 0x70, 0xe2, 0xee, 0x69, 0xfc, 0x61, 0x21, 0xfe, 0xab,
	0xe0, 0x8a, 0xd4, 0xe2, 0xee, 0x6f, 0xb0, 0x7f, 0x2d, 0x5c, 0xe8, 0x3e, 0xd4, 0x2f, 0xc9, 0xca,
	0xa6, 0x56, 0x0d, 0xd1, 0x13, 0x68, 0x2e, 0x71, 0x94, 0x11, 0xb7, 0xa6, 0x8d, 0x3f, 0xfa, 0x5f,
	0xbe, 0x4e, 0xf8, 0xca, 0x37, 0x92, 0x1f, 0x6a, 0xc7, 0xd5, 0xd3, 0x16, 0x34, 0x54, 0xa8, 0x7a,
	0x2d, 0x68, 0xa8, 0xa9, 0x7a, 0x9f, 0x43, 0xcb, 0xf4, 0x1e, 0x21, 0x68, 0xe8, 0x0d, 0x1a, 0x73,
	0x3d, 0xee, 0x3d, 0x82, 0x4e, 0xde, 0xdc, 0x1b, 0xff, 0x7f, 0x02, 0x77, 0xb6, 0xcb, 0x8f, 0xba,
	0xd0, 0x59, 0xf7, 0xcd, 0xe8, 0xd6, 0xdf, 0xbd, 0x31, 0x34, 0x75, 0x6d, 0x6e, 0x32, 0x42, 0x47,
	0xd0, 0x7c, 0x47, 0xa2, 0x28, 0xb5, 0x47, 0xaf, 0xb8, 0x3e, 0x6f, 0x49, 0x14, 0xf9, 0x46, 0xdb,
	0xfb, 0xb3, 0x0a, 0x0d, 0xf5, 0x8d, 0x4e, 0xde, 0x9b, 0xf6, 0xb6, 0xb8, 0xbc, 0x4a, 0x42, 0xca,
	0x71, 0x74, 0xa6, 0xee, 0xa1, 0xcd, 0xea, 0xd0, 0xf9, 0x76, 0x50, 0x4c, 0x2d, 0x4b, 0x1f, 0xb8,
	0x35, 0xd8, 0xfb, 0xb7, 0x06, 0xce, 0x3a, 0x45, 0x37, 0x6e, 0x14, 0x41, 0x83, 0x63, 0x66, 0xda,
	0xe5, 0xf8, 0x7a, 0x8c, 0x7e, 0x84, 0xd6, 0x52, 0x44, 0x19, 0x23, 0xf6, 0x92, 0xf8, 0xaa, 0x70,
	0xe2, 0xd1, 0xe6, 0x62, 0xf3, 0x2d, 0xa3, 0x22, 0xbe, 0xa4, 0xe9, 0x54, 0x6d, 0x64, 0xe5, 0x36,
	0x76, 0x30, 0xd8, 0x60, 0xe8, 0x18, 0x1a, 0x0c, 0xa7, 0xa9, 0xbd, 0x26, 0xca, 0xe1, 0x9a, 0x50,
	0x6b, 0xc7, 0x4c, 0x64, 0x5c, 0xda, 0x5b, 0xa2, 0xe4, 0xda, 0x0d, 0x83, 0x4e, 0x01, 0xd6, 0xc5,
	0x4b, 0xdd, 0xb6, 0xee, 0x7d, 0x89, 0xf3, 0xe9, 0x6f, 0x51, 0xbd, 0xbf, 0x5a, 0xe0, 0xac, 0x0f,
	0x1d, 0x3a, 0x86, 0x7a, 0x48, 0x99, 0x4d, 0x41, 0x71, 0x07, 0xc7, 0x8b, 0x55, 0x4a, 0xa7, 0x79,
	0x0c, 0x14, 0xa2, 0x42, 0xa4, 0x62, 0x15, 0x28, 0xbc, 0xb6, 0x13, 0xde, 0x56, 0xdc, 0x39, 0x65,
	0xaa, 0x15, 0x3c, 0x63, 0x81, 0x49, 0x72, 0x7d, 0xa7, 0x20, 0xf2, 0x8c, 0xa9, 0x28, 0xa7, 0xea,
	0xb6, 0x61, 0xf8, 0x2a, 0xb0, 0x81, 0xd8, 0xa9, 0x9f, 0x0c, 0x5f, 0xfd, 0x6e, 0x32, 0xf1, 0x10,
	0x40, 0xef, 0x25, 0x5d, 0xe0, 0x98, 0xe8, 0xae, 0x3a, 0xbe, 0xa3, 0x7e, 0x79, 0xad, 0x7e, 0x40,
	0x6f, 0xe0, 0x63, 0xe3, 0x1f, 0x64, 0x51, 0x20, 0x45, 0x80, 0x13, 0x82, 0x03, 0xc6, 0x06, 0xb6,
	0x87, 0x5f, 0x16, 0xef, 0x7b, 0xfd, 0x1a, 0xfb, 0x0f, 0x8c, 0xc3, 0x9b, 0xe8, 0x42, 0x9c, 0x24,
	0x04, 0x8f, 0xd8, 0x00, 0xbd, 0x85, 0x4f, 0xae, 0xd9, 0x2e, 0x08, 0x9d, 0x2f, 0x64, 0xc0, 0x98,
	0x7d, 0x03, 0x4a, 0xf9, 0xa2, 0x8d, 0xef, 0x50, 0xf3, 0x23, 0x86, 0x46, 0x70, 0x2f, 0x21, 0x29,
	0x0d, 0x33, 0x1c, 0xe5, 0x85, 0xe9, 0xec, 0x50, 0x98, 0xbb, 0x39, 0x6c, 0xab, 0xf3, 0x33, 0xec,
	0xe9, 0xea, 0x88, 0xd9, 0x2c, 0x25, 0xf9, 0xb3, 0x50, 0xb6, 0xd9, 0xba, 0xb0, 0xaf, 0x34, 0xb9,
	0x31, 0x4a, 0xe8, 0x9c, 0x72, 0xfb, 0x34, 0xec, 0x66, 0xa4, 0x49, 0x75, 0x2b, 0x30, 0x3e, 0x4b,
	0xf4, 0xeb, 0xe0, 0xf8, 0x7a, 0x8c, 0x5e, 0x03, 0xd2, 0xe6, 0x13, 0x21, 0xa5, 0x60, 0xb6, 0x98,
	0xee, 0xdf, 0xed, 0x1d, 0x36, 0x7e, 0x5f, 0x19, 0x9c, 0x6a, 0xde, 0xd4, 0x12, 0x7d, 0x0b, 0x0f,
	0xb6, 0x4d, 0x4d, 0x3e, 0xfe, 0x69, 0xeb, 0x69, 0xef, 0x6d, 0xd4, 0x3a, 0x26, 0x93, 0x96, 0x7e,
	0x44, 0x8e, 0xfe, 0x0b, 0x00, 0x00, 0xff, 0xff, 0x94, 0x48, 0x88, 0x76, 0x61, 0x0a, 0x00, 0x00,
}