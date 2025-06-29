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
// NOTE: Client sends incomplete data - only 12 bytes instead of expected 14 bytes
// This suggests the client might not send the disableFlags field (last 2 bytes)
type PlayerKeySyncPacket struct {
	ID       types.PlayerID // Player ID (4 bytes) - set by server, client may send garbage
	NewState struct {
		LeftStickX int16  // 2 bytes
		LeftStickY int16  // 2 bytes
		Compressed uint32 // 4 bytes
		// DisableFlags field appears to be missing in client packets (2 bytes missing)
		// Total: 4 + 2 + 2 + 4 = 12 bytes (matches received data)
	}
}

// Marshal serializes the packet to binary format (manual to match C++ struct packing)
func (p *PlayerKeySyncPacket) Marshal() ([]byte, error) {
	buf := make([]byte, 12) // Actual client packet size: 4 bytes playerid + 8 bytes partial controller state

	// Write PlayerID (4 bytes)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(p.ID))

	// Write partial controller state (8 bytes, missing DisableFlags)
	// LeftStickX (2 bytes)
	binary.LittleEndian.PutUint16(buf[4:6], uint16(p.NewState.LeftStickX))
	// LeftStickY (2 bytes)
	binary.LittleEndian.PutUint16(buf[6:8], uint16(p.NewState.LeftStickY))
	// Compressed (4 bytes)
	binary.LittleEndian.PutUint32(buf[8:12], p.NewState.Compressed)

	return buf, nil
}

// Unmarshal deserializes binary data to packet (manual to match C++ struct packing)
func (p *PlayerKeySyncPacket) Unmarshal(data []byte) error {
	if len(data) < 12 {
		return fmt.Errorf("PlayerKeySyncPacket: insufficient data, got %d bytes, expected 12", len(data))
	}

	// Read PlayerID (4 bytes)
	p.ID = types.PlayerID(binary.LittleEndian.Uint32(data[0:4]))

	// Read partial controller state (8 bytes, missing DisableFlags)
	// LeftStickX (2 bytes)
	p.NewState.LeftStickX = int16(binary.LittleEndian.Uint16(data[4:6]))
	// LeftStickY (2 bytes)
	p.NewState.LeftStickY = int16(binary.LittleEndian.Uint16(data[6:8]))
	// Compressed (4 bytes)
	p.NewState.Compressed = binary.LittleEndian.Uint32(data[8:12])

	return nil
}

// PlayerSetHostPacket represents a notification that a player is now the host
// This is sent from server to all clients to notify about host changes
type PlayerSetHostPacket struct {
	ID types.PlayerID // Player ID that is now the host (matches C++ int playerid)
}

// Marshal serializes the packet to binary format
func (p *PlayerSetHostPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *PlayerSetHostPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// RespawnPlayerPacket represents a notification that a player has respawned
// This matches the C++ CPackets::RespawnPlayer structure exactly
type RespawnPlayerPacket struct {
	PlayerID types.PlayerID // Player ID that respawned (matches C++ int playerid)
}

// Marshal serializes the packet to binary format
func (p *RespawnPlayerPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *RespawnPlayerPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// PlayerBulletShotPacket represents a bullet shot event from a player
// This matches the C++ CPlayerPackets::PlayerBulletShot structure exactly
// The collision point is represented as 44 bytes of padding to match server implementation
type PlayerBulletShotPacket struct {
	PlayerID       int32                   // Player ID (set by server)
	TargetID       int32                   // Target entity ID (-1 if no target)
	StartPos       types.Vector3           // Bullet start position
	EndPos         types.Vector3           // Bullet end position
	ColPoint       [44]uint8               // Collision point data (44 bytes padding, matching C++ server)
	IncrementalHit int32                   // Incremental hit value
	EntityType     types.NetworkEntityType // Type of entity hit (if any)
}

// Marshal serializes the packet to binary format
func (p *PlayerBulletShotPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *PlayerBulletShotPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// PlayerChatMessagePacket represents a chat message from a player
// This matches the C++ CPackets::PlayerChatMessage structure exactly
// Note: Uses UTF-16 encoding (wchar_t) with 129 wide characters (128 + null terminator)
type PlayerChatMessagePacket struct {
	PlayerID types.PlayerID // Player ID (matches C++ int playerid)
	Message  [129]uint16    // UTF-16 message (matches C++ wchar_t message[128+1])
}

// Marshal serializes the packet to binary format
func (p *PlayerChatMessagePacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *PlayerChatMessagePacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// SetMessage sets the message from a UTF-8 string, converting to UTF-16
func (p *PlayerChatMessagePacket) SetMessage(message string) {
	// Clear the message array
	for i := range p.Message {
		p.Message[i] = 0
	}

	// Convert UTF-8 string to UTF-16 runes and store
	runes := []rune(message)
	maxLen := len(p.Message) - 1 // Reserve space for null terminator

	for i, r := range runes {
		if i >= maxLen {
			break // Truncate if message is too long
		}
		p.Message[i] = uint16(r)
	}
	// Null terminator is already set by clearing the array
}

// GetMessage extracts the UTF-8 string from the UTF-16 message
func (p *PlayerChatMessagePacket) GetMessage() string {
	// Find the null terminator
	length := 0
	for i, char := range p.Message {
		if char == 0 {
			length = i
			break
		}
	}

	// Convert UTF-16 to UTF-8
	runes := make([]rune, length)
	for i := 0; i < length; i++ {
		runes[i] = rune(p.Message[i])
	}

	return string(runes)
}

// PlayerAimSyncPacket represents a packet for synchronizing player aiming data
// This matches the C++ CPackets::PlayerAimSync structure exactly
type PlayerAimSyncPacket struct {
	PlayerID    types.PlayerID // Player ID (matches C++ int playerid)
	CameraMode  uint8          // Camera mode (matches C++ unsigned char cameraMode)
	CameraFov   float32        // Camera field of view (matches C++ float cameraFov)
	Front       types.Vector3  // Camera front vector (matches C++ CVector front)
	Source      types.Vector3  // Camera source vector (matches C++ CVector source)
	Up          types.Vector3  // Camera up vector (matches C++ CVector up)
	MoveHeading float32        // Movement heading (matches C++ float moveHeading)
	AimY        float32        // Aim Y angle (matches C++ float aimY)
	AimZ        float32        // Aim Z angle (matches C++ float aimZ)
}

// Marshal serializes the packet to binary format
func (p *PlayerAimSyncPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *PlayerAimSyncPacket) Unmarshal(data []byte) error {
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
func NewPlayerKeySyncPacket(playerID types.PlayerID, leftStickX, leftStickY int16, compressed uint32) *PlayerKeySyncPacket {
	return &PlayerKeySyncPacket{
		ID: playerID,
		NewState: struct {
			LeftStickX int16
			LeftStickY int16
			Compressed uint32
		}{
			LeftStickX: leftStickX,
			LeftStickY: leftStickY,
			Compressed: compressed,
		},
	}
}

// NewPlayerSetHostPacket creates a new player set host packet
func NewPlayerSetHostPacket(playerID types.PlayerID) *PlayerSetHostPacket {
	return &PlayerSetHostPacket{
		ID: playerID,
	}
}

// NewRespawnPlayerPacket creates a new player respawn notification
func NewRespawnPlayerPacket(playerID types.PlayerID) *RespawnPlayerPacket {
	return &RespawnPlayerPacket{
		PlayerID: playerID,
	}
}

// NewPlayerBulletShotPacket creates a new player bullet shot packet
func NewPlayerBulletShotPacket(playerID types.PlayerID, targetID int32, startPos, endPos types.Vector3, entityType types.NetworkEntityType, incrementalHit int32) *PlayerBulletShotPacket {
	return &PlayerBulletShotPacket{
		PlayerID:       int32(playerID),
		TargetID:       targetID,
		StartPos:       startPos,
		EndPos:         endPos,
		ColPoint:       [44]uint8{}, // Zero-initialized collision point data
		IncrementalHit: incrementalHit,
		EntityType:     entityType,
	}
}

// NewPlayerChatMessagePacket creates a new chat message packet
func NewPlayerChatMessagePacket(playerID types.PlayerID, message string) *PlayerChatMessagePacket {
	packet := &PlayerChatMessagePacket{
		PlayerID: playerID,
	}
	packet.SetMessage(message)
	return packet
}

// NewPlayerAimSyncPacket creates a new player aim sync packet
func NewPlayerAimSyncPacket(playerID types.PlayerID, cameraMode uint8, cameraFov float32,
	front, source, up types.Vector3, moveHeading, aimY, aimZ float32) *PlayerAimSyncPacket {
	return &PlayerAimSyncPacket{
		PlayerID:    playerID,
		CameraMode:  cameraMode,
		CameraFov:   cameraFov,
		Front:       front,
		Source:      source,
		Up:          up,
		MoveHeading: moveHeading,
		AimY:        aimY,
		AimZ:        aimZ,
	}
}
