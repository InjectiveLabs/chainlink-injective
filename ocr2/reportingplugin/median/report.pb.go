// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: report.proto

package median

import (
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type Report struct {
	ObservationsTimestamp int64                                    `protobuf:"varint,1,opt,name=observations_timestamp,json=observationsTimestamp,proto3" json:"observations_timestamp,omitempty"`
	Observers             []byte                                   `protobuf:"bytes,2,opt,name=observers,proto3" json:"observers,omitempty"`
	Observations          []github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,3,rep,name=observations,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"observations"`
}

func (m *Report) Reset()         { *m = Report{} }
func (m *Report) String() string { return proto.CompactTextString(m) }
func (*Report) ProtoMessage()    {}
func (*Report) Descriptor() ([]byte, []int) {
	return fileDescriptor_3eedb623aa6ca98c, []int{0}
}
func (m *Report) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Report) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Report.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Report) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Report.Merge(m, src)
}
func (m *Report) XXX_Size() int {
	return m.Size()
}
func (m *Report) XXX_DiscardUnknown() {
	xxx_messageInfo_Report.DiscardUnknown(m)
}

var xxx_messageInfo_Report proto.InternalMessageInfo

func (m *Report) GetObservationsTimestamp() int64 {
	if m != nil {
		return m.ObservationsTimestamp
	}
	return 0
}

func (m *Report) GetObservers() []byte {
	if m != nil {
		return m.Observers
	}
	return nil
}

func init() {
	proto.RegisterType((*Report)(nil), "chainlink.cosmos.reportingplugin.median.v1beta1.Report")
}

func init() { proto.RegisterFile("report.proto", fileDescriptor_3eedb623aa6ca98c) }

var fileDescriptor_3eedb623aa6ca98c = []byte{
	// 254 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x8f, 0xc1, 0x4a, 0xc3, 0x30,
	0x18, 0xc7, 0x1b, 0x0b, 0x83, 0x96, 0x9e, 0x8a, 0x4a, 0x11, 0xc9, 0x8a, 0x82, 0xf4, 0xe2, 0x57,
	0x86, 0xf8, 0x02, 0xc3, 0x27, 0x28, 0x9e, 0xbc, 0x48, 0xda, 0x85, 0x2c, 0x6c, 0xc9, 0x17, 0x92,
	0x6c, 0xe0, 0x5b, 0xf8, 0x2a, 0xbe, 0xc5, 0x8e, 0x3b, 0x8a, 0x87, 0x21, 0xed, 0x8b, 0x88, 0x8d,
	0xe2, 0x3c, 0x25, 0x7c, 0x3f, 0x7e, 0x7f, 0xf8, 0xa5, 0x99, 0xe5, 0x06, 0xad, 0x07, 0x63, 0xd1,
	0x63, 0x5e, 0x77, 0x4b, 0x26, 0xf5, 0x5a, 0xea, 0x15, 0x74, 0xe8, 0x14, 0x3a, 0x08, 0x58, 0x6a,
	0x61, 0xd6, 0x1b, 0x21, 0x35, 0x28, 0xbe, 0x90, 0x4c, 0xc3, 0x76, 0xd6, 0x72, 0xcf, 0x66, 0x17,
	0xa7, 0x02, 0x05, 0x8e, 0x6e, 0xfd, 0xfd, 0x0b, 0x33, 0x57, 0x6f, 0x24, 0x9d, 0x34, 0xa3, 0x98,
	0xdf, 0xa7, 0xe7, 0xd8, 0x3a, 0x6e, 0xb7, 0xcc, 0x4b, 0xd4, 0xee, 0xd9, 0x4b, 0xc5, 0x9d, 0x67,
	0xca, 0x14, 0xa4, 0x24, 0x55, 0xdc, 0x9c, 0x1d, 0xd3, 0xc7, 0x5f, 0x98, 0x5f, 0xa6, 0x49, 0x00,
	0xdc, 0xba, 0xe2, 0xa4, 0x24, 0x55, 0xd6, 0xfc, 0x1d, 0xf2, 0x26, 0xcd, 0x8e, 0xb5, 0x22, 0x2e,
	0xe3, 0x2a, 0x99, 0xc3, 0xee, 0x30, 0x8d, 0x3e, 0x0e, 0xd3, 0x1b, 0x21, 0xfd, 0x72, 0xd3, 0x42,
	0x87, 0xaa, 0x0e, 0x15, 0x3f, 0xcf, 0xad, 0x5b, 0xac, 0x6a, 0xff, 0x62, 0xb8, 0x83, 0x07, 0xde,
	0x35, 0xff, 0x36, 0xe6, 0xd7, 0xbb, 0x9e, 0x92, 0x7d, 0x4f, 0xc9, 0x67, 0x4f, 0xc9, 0xeb, 0x40,
	0xa3, 0xfd, 0x40, 0xa3, 0xf7, 0x81, 0x46, 0x4f, 0x09, 0x40, 0x1d, 0xb2, 0xdb, 0xc9, 0xd8, 0x77,
	0xf7, 0x15, 0x00, 0x00, 0xff, 0xff, 0x3d, 0xec, 0x74, 0x50, 0x36, 0x01, 0x00, 0x00,
}

func (m *Report) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Report) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Report) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Observations) > 0 {
		for iNdEx := len(m.Observations) - 1; iNdEx >= 0; iNdEx-- {
			{
				size := m.Observations[iNdEx].Size()
				i -= size
				if _, err := m.Observations[iNdEx].MarshalTo(dAtA[i:]); err != nil {
					return 0, err
				}
				i = encodeVarintReport(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.Observers) > 0 {
		i -= len(m.Observers)
		copy(dAtA[i:], m.Observers)
		i = encodeVarintReport(dAtA, i, uint64(len(m.Observers)))
		i--
		dAtA[i] = 0x12
	}
	if m.ObservationsTimestamp != 0 {
		i = encodeVarintReport(dAtA, i, uint64(m.ObservationsTimestamp))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintReport(dAtA []byte, offset int, v uint64) int {
	offset -= sovReport(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Report) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ObservationsTimestamp != 0 {
		n += 1 + sovReport(uint64(m.ObservationsTimestamp))
	}
	l = len(m.Observers)
	if l > 0 {
		n += 1 + l + sovReport(uint64(l))
	}
	if len(m.Observations) > 0 {
		for _, e := range m.Observations {
			l = e.Size()
			n += 1 + l + sovReport(uint64(l))
		}
	}
	return n
}

func sovReport(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozReport(x uint64) (n int) {
	return sovReport(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Report) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowReport
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Report: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Report: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObservationsTimestamp", wireType)
			}
			m.ObservationsTimestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReport
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ObservationsTimestamp |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Observers", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReport
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthReport
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthReport
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Observers = append(m.Observers[:0], dAtA[iNdEx:postIndex]...)
			if m.Observers == nil {
				m.Observers = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Observations", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReport
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthReport
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthReport
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			var v github_com_cosmos_cosmos_sdk_types.Dec
			m.Observations = append(m.Observations, v)
			if err := m.Observations[len(m.Observations)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipReport(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthReport
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipReport(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowReport
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowReport
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowReport
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthReport
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupReport
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthReport
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthReport        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowReport          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupReport = fmt.Errorf("proto: unexpected end of group")
)
