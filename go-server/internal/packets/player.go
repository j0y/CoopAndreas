package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"coopandreas-server/internal/types"
)

// CCompressedControllerState represents compressed controller state from C++
// This structure must match the C++ CCompressedControllerState exactly for binary compatibility
// C++ size: 10 bytes (tightly packed), Go size: 12 bytes (with padding)
// Manual marshaling is used in PlayerKeySyncPacket to achieve C++ binary compatibility
type CCompressedControllerState struct {
	LeftStickX int16 // move/steer left (-128)/right (+128)
	LeftStickY int16 // move back(+128)/forwards(-128)

	// Button and flag states packed into a 32-bit field
	// This matches the C++ union structure with bitfields
	Compressed uint32 // All button states as single value

	// Disable flags packed into a 16-bit field
	// This matches the C++ union structure with bitfields
	DisableFlags uint16 // All disable flags as single value
}

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

// PlayerOnFootPacket represents a player's on-foot movement and state
// This matches the C++ CPackets::PlayerOnFoot structure exactly
type PlayerOnFootPacket struct {
	ID            types.PlayerID // Player ID
	Position      types.Vector3  // Player position
	Velocity      types.Vector3  // Player velocity/movement speed
	Rotation      float32        // Player facing angle
	Health        uint8          // Player health (0-100)
	Armour        uint8          // Player armour (0-100)
	Weapon        uint8          // Current weapon ID
	Ammo          uint16         // Current ammo count
	Ducking       uint8          // 1 if ducking, 0 if not (bool as uint8 for C++ compatibility)
	HasJetpack    uint8          // 1 if has jetpack, 0 if not (bool as uint8 for C++ compatibility)
	FightingStyle int8           // Fighting style (4-16)
}

// Marshal serializes the packet to binary format
func (p *PlayerOnFootPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *PlayerOnFootPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// PlayerKeySyncPacket represents a packet for synchronizing player controller state
// This matches the C++ CPackets::PlayerKeySync structure exactly
// Uses manual marshaling/unmarshaling to match C++ struct packing (14 bytes total)
type PlayerKeySyncPacket struct {
	ID       types.PlayerID             // Player ID (matches C++ int playerid)
	NewState CCompressedControllerState // Compressed controller state
}

// Marshal serializes the packet to binary format (manual to match C++ struct packing)
func (p *PlayerKeySyncPacket) Marshal() ([]byte, error) {
	buf := make([]byte, 14) // C++ struct size: 4 bytes playerid + 10 bytes CCompressedControllerState

	// Write PlayerID (4 bytes)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(p.ID))

	// Write CCompressedControllerState (10 bytes, tightly packed)
	// LeftStickX (2 bytes)
	binary.LittleEndian.PutUint16(buf[4:6], uint16(p.NewState.LeftStickX))
	// LeftStickY (2 bytes)
	binary.LittleEndian.PutUint16(buf[6:8], uint16(p.NewState.LeftStickY))
	// Compressed (4 bytes)
	binary.LittleEndian.PutUint32(buf[8:12], p.NewState.Compressed)
	// DisableFlags (2 bytes)
	binary.LittleEndian.PutUint16(buf[12:14], p.NewState.DisableFlags)

	return buf, nil
}

// Unmarshal deserializes binary data to packet (manual to match C++ struct packing)
func (p *PlayerKeySyncPacket) Unmarshal(data []byte) error {
	if len(data) < 14 {
		return fmt.Errorf("PlayerKeySyncPacket: insufficient data, got %d bytes, expected 14", len(data))
	}

	// Read PlayerID (4 bytes)
	p.ID = types.PlayerID(binary.LittleEndian.Uint32(data[0:4]))

	// Read CCompressedControllerState (10 bytes, tightly packed)
	// LeftStickX (2 bytes)
	p.NewState.LeftStickX = int16(binary.LittleEndian.Uint16(data[4:6]))
	// LeftStickY (2 bytes)
	p.NewState.LeftStickY = int16(binary.LittleEndian.Uint16(data[6:8]))
	// Compressed (4 bytes)
	p.NewState.Compressed = binary.LittleEndian.Uint32(data[8:12])
	// DisableFlags (2 bytes)
	p.NewState.DisableFlags = binary.LittleEndian.Uint16(data[12:14])

	return nil
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

// NewPlayerOnFootPacket creates a new player on-foot packet
func NewPlayerOnFootPacket(playerID types.PlayerID, position, velocity types.Vector3, rotation float32, health, armour, weapon uint8, ammo uint16, ducking, hasJetpack bool, fightingStyle int8) *PlayerOnFootPacket {
	packet := &PlayerOnFootPacket{
		ID:            playerID,
		Position:      position,
		Velocity:      velocity,
		Rotation:      rotation,
		Health:        health,
		Armour:        armour,
		Weapon:        weapon,
		Ammo:          ammo,
		FightingStyle: fightingStyle,
	}

	// Convert booleans to uint8 for C++ compatibility
	if ducking {
		packet.Ducking = 1
	} else {
		packet.Ducking = 0
	}

	if hasJetpack {
		packet.HasJetpack = 1
	} else {
		packet.HasJetpack = 0
	}

	return packet
}

// NewPlayerKeySyncPacket creates a new player key synchronization packet
func NewPlayerKeySyncPacket(playerID types.PlayerID, newState CCompressedControllerState) *PlayerKeySyncPacket {
	return &PlayerKeySyncPacket{
		ID:       playerID,
		NewState: newState,
	}
}
