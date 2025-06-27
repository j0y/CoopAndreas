package packets

import (
	"bytes"
	"encoding/binary"

	"coopandreas-server/internal/types"
)

// CheckVersionPacket represents a version check request from client
// Note: In the C++ implementation, this is noted as "reserved but not used, see enet_host_connect"
// However, we'll implement it for explicit version validation
type CheckVersionPacket struct {
	ClientVersion [32]byte // Client version string (null-terminated)
	ProtocolVersion uint32  // Protocol version number
}

// Marshal serializes the packet to binary format
func (p *CheckVersionPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *CheckVersionPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// CheckVersionResponsePacket represents the server's version check response
type CheckVersionResponsePacket struct {
	ServerVersion   [32]byte // Server version string (null-terminated)
	ProtocolVersion uint32   // Server protocol version
	IsCompatible    uint8    // 1 if compatible, 0 if not
	Message         [128]byte // Optional message (e.g., update required)
}

// Marshal serializes the packet to binary format
func (p *CheckVersionResponsePacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *CheckVersionResponsePacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// NewCheckVersionResponse creates a new version response packet
func NewCheckVersionResponse(isCompatible bool, message string) *CheckVersionResponsePacket {
	response := &CheckVersionResponsePacket{
		ProtocolVersion: types.ProtocolVersion,
	}
	
	// Copy server version
	copy(response.ServerVersion[:], types.ServerVersion)
	
	// Set compatibility
	if isCompatible {
		response.IsCompatible = 1
	} else {
		response.IsCompatible = 0
	}
	
	// Copy message
	if len(message) > 0 {
		copy(response.Message[:], message)
	}
	
	return response
}
