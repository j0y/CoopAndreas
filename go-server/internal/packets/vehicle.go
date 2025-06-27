package packets

import (
	"bytes"
	"encoding/binary"

	"coopandreas-server/internal/types"
)

// VehicleSpawnPacket represents a vehicle spawn request/notification
// This matches the C++ CVehiclePackets::VehicleSpawn structure exactly
type VehicleSpawnPacket struct {
	VehicleID int32         // Vehicle ID assigned by server
	TempID    uint8         // Temporary ID from client
	ModelID   uint16        // Vehicle model ID (400-611)
	Position  types.Vector3 // Spawn position
	Rotation  float32       // Rotation angle
	Color1    uint8         // Primary color
	Color2    uint8         // Secondary color
	CreatedBy uint8         // Who created the vehicle
}

// Marshal serializes the packet to binary format
func (p *VehicleSpawnPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *VehicleSpawnPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// VehicleConfirmPacket represents a vehicle ID confirmation to the client
// This is sent back to the client that spawned the vehicle
type VehicleConfirmPacket struct {
	TempID    uint8 // Original temporary ID from client
	VehicleID int32 // Server-assigned vehicle ID
}

// Marshal serializes the packet to binary format
func (p *VehicleConfirmPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *VehicleConfirmPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// VehicleRemovePacket represents a vehicle removal notification
type VehicleRemovePacket struct {
	VehicleID int32 // Vehicle ID to remove
}

// Marshal serializes the packet to binary format
func (p *VehicleRemovePacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *VehicleRemovePacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// NewVehicleSpawnPacket creates a new vehicle spawn packet
func NewVehicleSpawnPacket(vehicleID int32, tempID uint8, modelID uint16, pos types.Vector3, rot float32, color1, color2, createdBy uint8) *VehicleSpawnPacket {
	return &VehicleSpawnPacket{
		VehicleID: vehicleID,
		TempID:    tempID,
		ModelID:   modelID,
		Position:  pos,
		Rotation:  rot,
		Color1:    color1,
		Color2:    color2,
		CreatedBy: createdBy,
	}
}

// NewVehicleConfirmPacket creates a new vehicle confirm packet
func NewVehicleConfirmPacket(tempID uint8, vehicleID int32) *VehicleConfirmPacket {
	return &VehicleConfirmPacket{
		TempID:    tempID,
		VehicleID: vehicleID,
	}
}

// NewVehicleRemovePacket creates a new vehicle remove packet
func NewVehicleRemovePacket(vehicleID int32) *VehicleRemovePacket {
	return &VehicleRemovePacket{
		VehicleID: vehicleID,
	}
}
