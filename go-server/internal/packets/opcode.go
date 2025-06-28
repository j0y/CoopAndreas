package packets

import (
	"encoding/binary"
	"fmt"
)

// OpcodeSyncPacket represents an opcode synchronization packet
// This matches the C++ OpcodeSyncHeader structure and variable-length payload
type OpcodeSyncPacket struct {
	Opcode           uint16             // Script opcode being synchronized
	IntParamCount    uint8              // Number of integer parameters (4 bits in C++, using full byte for simplicity)
	StringParamCount uint8              // Number of string parameters (4 bits in C++, using full byte for simplicity)
	IntParams        []int32            // Integer parameters (up to 10 in C++)
	StringParams     []OpcodeSyncString // String parameters (up to 10 in C++)
}

// OpcodeSyncString represents a string parameter in opcode sync
type OpcodeSyncString struct {
	Length uint8  // String length (1 byte)
	Data   []byte // String data
}

// Marshal serializes the packet to binary format
// Manual marshaling to match C++ struct packing exactly
func (p *OpcodeSyncPacket) Marshal() ([]byte, error) {
	// Calculate total size
	totalSize := 4                    // 2 bytes opcode + 1 byte intParamCount + 1 byte stringParamCount
	totalSize += len(p.IntParams) * 4 // 4 bytes per int32
	for _, str := range p.StringParams {
		totalSize += 1 + len(str.Data) // 1 byte length + string data
	}

	buf := make([]byte, totalSize)
	offset := 0

	// Write header - matching C++ OpcodeSyncHeader layout
	binary.LittleEndian.PutUint16(buf[offset:offset+2], p.Opcode)
	offset += 2

	// Pack the two 4-bit counts into single byte (matching C++ bitfield layout)
	// intParamCount in lower 4 bits, stringParamCount in upper 4 bits
	packedCounts := (p.StringParamCount&0x0F)<<4 | (p.IntParamCount & 0x0F)
	buf[offset] = packedCounts
	offset++

	// Write integer parameters
	for _, param := range p.IntParams {
		binary.LittleEndian.PutUint32(buf[offset:offset+4], uint32(param))
		offset += 4
	}

	// Write string parameters
	for _, str := range p.StringParams {
		buf[offset] = str.Length
		offset++
		copy(buf[offset:offset+len(str.Data)], str.Data)
		offset += len(str.Data)
	}

	return buf, nil
}

// Unmarshal deserializes binary data to packet
// Manual unmarshaling to match C++ struct layout
func (p *OpcodeSyncPacket) Unmarshal(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("OpcodeSyncPacket: insufficient data, got %d bytes, expected at least 3", len(data))
	}

	offset := 0

	// Read header
	p.Opcode = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	// Unpack the two 4-bit counts from single byte
	packedCounts := data[offset]
	p.IntParamCount = packedCounts & 0x0F           // Lower 4 bits
	p.StringParamCount = (packedCounts >> 4) & 0x0F // Upper 4 bits
	offset++

	// Validate we have enough data for integer parameters
	intParamsSize := int(p.IntParamCount) * 4
	if len(data) < offset+intParamsSize {
		return fmt.Errorf("OpcodeSyncPacket: insufficient data for %d int params", p.IntParamCount)
	}

	// Read integer parameters
	p.IntParams = make([]int32, p.IntParamCount)
	for i := 0; i < int(p.IntParamCount); i++ {
		p.IntParams[i] = int32(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4
	}

	// Read string parameters
	p.StringParams = make([]OpcodeSyncString, p.StringParamCount)
	for i := 0; i < int(p.StringParamCount); i++ {
		if offset >= len(data) {
			return fmt.Errorf("OpcodeSyncPacket: insufficient data for string param %d", i)
		}

		// Read string length
		strLen := data[offset]
		offset++

		// Validate we have enough data for the string
		if offset+int(strLen) > len(data) {
			return fmt.Errorf("OpcodeSyncPacket: insufficient data for string param %d (length %d)", i, strLen)
		}

		// Read string data
		p.StringParams[i] = OpcodeSyncString{
			Length: strLen,
			Data:   make([]byte, strLen),
		}
		copy(p.StringParams[i].Data, data[offset:offset+int(strLen)])
		offset += int(strLen)
	}

	return nil
}

// NewOpcodeSyncPacket creates a new opcode sync packet
func NewOpcodeSyncPacket(opcode uint16, intParams []int32, stringParams []OpcodeSyncString) *OpcodeSyncPacket {
	// Limit parameters to match C++ implementation (max 10 each)
	if len(intParams) > 10 {
		intParams = intParams[:10]
	}
	if len(stringParams) > 10 {
		stringParams = stringParams[:10]
	}

	return &OpcodeSyncPacket{
		Opcode:           opcode,
		IntParamCount:    uint8(len(intParams)),
		StringParamCount: uint8(len(stringParams)),
		IntParams:        intParams,
		StringParams:     stringParams,
	}
}

// NewOpcodeSyncString creates a new opcode sync string parameter
func NewOpcodeSyncString(data string) OpcodeSyncString {
	dataBytes := []byte(data)
	// Limit to 255 bytes (uint8 length)
	if len(dataBytes) > 255 {
		dataBytes = dataBytes[:255]
	}

	return OpcodeSyncString{
		Length: uint8(len(dataBytes)),
		Data:   dataBytes,
	}
}
