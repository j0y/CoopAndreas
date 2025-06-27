package packets

import (
	"bytes"
	"encoding/binary"

	"coopandreas-server/internal/types"
)

// PedSpawnPacket matches the C++ PedSpawn struct
type PedSpawnPacket struct {
	PedID             int32           // 4 bytes
	TempID            uint8           // 1 byte
	ModelID           int16           // 2 bytes
	PedType           uint8           // 1 byte
	Position          types.Vector3   // 12 bytes (3 * float32)
	CreatedBy         uint8           // 1 byte
	SpecialModelName  [8]byte         // 8 bytes
}

// Marshal serializes the packet to binary format
func (p *PedSpawnPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *PedSpawnPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// PedRemovePacket matches the C++ PedRemove struct
type PedRemovePacket struct {
	PedID int32 // 4 bytes
}

// Marshal serializes the packet to binary format
func (p *PedRemovePacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *PedRemovePacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// PedOnFootPacket matches the C++ PedOnFoot struct
type PedOnFootPacket struct {
	PedID           int32         // 4 bytes
	Position        types.Vector3 // 12 bytes
	Velocity        types.Vector3 // 12 bytes
	Health          uint8         // 1 byte
	Armour          uint8         // 1 byte
	Weapon          uint8         // 1 byte
	WeaponState     uint8         // 1 byte
	Ammo            uint16        // 2 bytes
	AimingRotation  float32       // 4 bytes
	CurrentRotation float32       // 4 bytes
	LookDirection   int32         // 4 bytes
	// Bitfield struct - represented as single byte in Go
	MoveStateAndFlags uint8         // 1 byte (moveState:3, ducked:1, aiming:1)
	FightingStyle     uint8         // 1 byte
	WeaponAim         types.Vector3 // 12 bytes
}

// Marshal serializes the packet to binary format
func (p *PedOnFootPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *PedOnFootPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// PedConfirmPacket matches the C++ PedConfirm struct
type PedConfirmPacket struct {
	TempID uint8 // 1 byte
	PedID  int32 // 4 bytes
}

// Marshal serializes the packet to binary format
func (p *PedConfirmPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *PedConfirmPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}
