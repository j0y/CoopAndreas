package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"coopandreas-server/internal/types"
)

// DebugPacketSize logs the size of packet data for debugging struct packing issues
func DebugPacketSize(packetName string, data []byte, expectedSize int) {
	actualSize := len(data)
	if actualSize != expectedSize {
		fmt.Printf("WARNING: %s packet size mismatch - received: %d bytes, expected: %d bytes\n",
			packetName, actualSize, expectedSize)
	}
}

// PedSpawnPacket matches the C++ PedSpawn struct
type PedSpawnPacket struct {
	PedID            int32         // 4 bytes
	TempID           uint8         // 1 byte
	ModelID          int16         // 2 bytes
	PedType          uint8         // 1 byte
	Position         types.Vector3 // 12 bytes (3 * float32)
	CreatedBy        uint8         // 1 byte
	SpecialModelName [8]byte       // 8 bytes
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

// PedOnFootPacket matches the C++ PedOnFoot struct exactly
// Note: Based on actual client packets, weaponState appears to be omitted (59 bytes vs 60 bytes expected)
type PedOnFootPacket struct {
	PedID    int32         // 4 bytes - int pedid
	Position types.Vector3 // 12 bytes - CVector pos
	Velocity types.Vector3 // 12 bytes - CVector velocity
	Health   uint8         // 1 byte - unsigned char health
	Armour   uint8         // 1 byte - unsigned char armour
	Weapon   uint8         // 1 byte - unsigned char weapon
	// weaponState field appears to be missing in client packets (59 bytes vs 60 expected)
	Ammo            uint16  // 2 bytes - unsigned short ammo
	AimingRotation  float32 // 4 bytes - float aimingRotation
	CurrentRotation float32 // 4 bytes - float currentRotation
	LookDirection   int32   // 4 bytes - int lookDirection
	// Bitfield struct - represented as single byte in Go
	// struct { unsigned char moveState:3; ducked:1; aiming:1; }
	MoveStateAndFlags uint8         // 1 byte (moveState:3, ducked:1, aiming:1)
	FightingStyle     uint8         // 1 byte - unsigned char fightingStyle
	WeaponAim         types.Vector3 // 12 bytes - CVector weaponAim
}

// Marshal serializes the packet to binary format with C++ struct packing
func (p *PedOnFootPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write each field manually to ensure exact C++ struct layout
	if err := binary.Write(buf, binary.LittleEndian, p.PedID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Position); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Velocity); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Health); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Armour); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Weapon); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Ammo); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.AimingRotation); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.CurrentRotation); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.LookDirection); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.MoveStateAndFlags); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.FightingStyle); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.WeaponAim); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes binary data to packet with C++ struct packing
func (p *PedOnFootPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)

	// Read each field manually to ensure exact C++ struct layout
	if err := binary.Read(buf, binary.LittleEndian, &p.PedID); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Position); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Velocity); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Health); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Armour); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Weapon); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Ammo); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.AimingRotation); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.CurrentRotation); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.LookDirection); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.MoveStateAndFlags); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.FightingStyle); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.WeaponAim); err != nil {
		return err
	}

	return nil
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

// NewPedConfirmPacket creates a new ped confirmation packet
func NewPedConfirmPacket(tempID uint8, pedID int32) *PedConfirmPacket {
	return &PedConfirmPacket{
		TempID: tempID,
		PedID:  pedID,
	}
}
