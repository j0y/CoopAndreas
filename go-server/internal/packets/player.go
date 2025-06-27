package packets

import (
	"bytes"
	"encoding/binary"

	"coopandreas-server/internal/types"
)

// PlayerConnectedPacket represents a notification that a player has connected
// This is sent from server to clients to notify about new player connections
type PlayerConnectedPacket struct {
	ID                 types.PlayerID // Player ID assigned by server
	IsAlreadyConnected uint8          // 1 if player was already connected, 0 for new connection
}

// Marshal serializes the packet to binary format
func (p *PlayerConnectedPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *PlayerConnectedPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// PlayerDisconnectedPacket represents a notification that a player has disconnected
type PlayerDisconnectedPacket struct {
	ID     types.PlayerID // Player ID that disconnected
	Reason uint8          // Disconnection reason
}

// Marshal serializes the packet to binary format
func (p *PlayerDisconnectedPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *PlayerDisconnectedPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// PlayerHandshakePacket represents a handshake sent to new players
type PlayerHandshakePacket struct {
	YourID types.PlayerID // The player ID assigned to this client
}

// Marshal serializes the packet to binary format
func (p *PlayerHandshakePacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *PlayerHandshakePacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// PlayerGetNamePacket represents a player's name update
// This is sent from client to server after receiving handshake
type PlayerGetNamePacket struct {
	PlayerID types.PlayerID // Player ID (not used in C++ but included for completeness)
	Name     [33]byte       // Player name (32 chars + null terminator)
}

// Marshal serializes the packet to binary format
func (p *PlayerGetNamePacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *PlayerGetNamePacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// GetNameString returns the name as a Go string (null-terminated)
func (p *PlayerGetNamePacket) GetNameString() string {
	// Find the null terminator and return substring
	for i, b := range p.Name {
		if b == 0 {
			return string(p.Name[:i])
		}
	}
	return string(p.Name[:])
}

// SetNameString sets the name from a Go string (adds null termination)
func (p *PlayerGetNamePacket) SetNameString(name string) {
	// Clear the name array
	for i := range p.Name {
		p.Name[i] = 0
	}

	// Copy the name (up to 32 characters)
	nameBytes := []byte(name)
	maxLen := len(p.Name) - 1 // Reserve space for null terminator
	if len(nameBytes) > maxLen {
		nameBytes = nameBytes[:maxLen]
	}

	copy(p.Name[:], nameBytes)
	// Null terminator is already there from clearing
}

// NewPlayerConnectedPacket creates a new player connected notification
func NewPlayerConnectedPacket(playerID types.PlayerID, isAlreadyConnected bool) *PlayerConnectedPacket {
	packet := &PlayerConnectedPacket{
		ID: playerID,
	}

	if isAlreadyConnected {
		packet.IsAlreadyConnected = 1
	} else {
		packet.IsAlreadyConnected = 0
	}

	return packet
}

// NewPlayerDisconnectedPacket creates a new player disconnected notification
func NewPlayerDisconnectedPacket(playerID types.PlayerID, reason uint8) *PlayerDisconnectedPacket {
	return &PlayerDisconnectedPacket{
		ID:     playerID,
		Reason: reason,
	}
}

// NewPlayerHandshakePacket creates a new handshake packet
func NewPlayerHandshakePacket(playerID types.PlayerID) *PlayerHandshakePacket {
	return &PlayerHandshakePacket{
		YourID: playerID,
	}
}

// NewPlayerGetNamePacket creates a new player get name packet
func NewPlayerGetNamePacket(playerID types.PlayerID, name string) *PlayerGetNamePacket {
	packet := &PlayerGetNamePacket{
		PlayerID: playerID,
	}
	packet.SetNameString(name)
	return packet
}
