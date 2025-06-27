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
