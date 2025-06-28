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

// VehicleIdleUpdatePacket represents a vehicle idle state update
// This matches the C++ CPackets::VehicleIdleUpdate structure exactly
type VehicleIdleUpdatePacket struct {
	VehicleID      int32         // Vehicle ID (matches C++ int vehicleid)
	Position       types.Vector3 // Vehicle position (matches C++ CVector pos)
	Rotation       types.Vector3 // Vehicle rotation (matches C++ CVector rot)
	Roll           types.Vector3 // Vehicle roll (matches C++ CVector roll)
	Velocity       types.Vector3 // Vehicle velocity (matches C++ CVector velocity)
	TurnSpeed      types.Vector3 // Vehicle turn speed (matches C++ CVector turnSpeed)
	Color1         uint8         // Primary color (matches C++ unsigned char color1)
	Color2         uint8         // Secondary color (matches C++ unsigned char color2)
	Health         float32       // Vehicle health (matches C++ float health)
	Paintjob       int8          // Paintjob ID (matches C++ char paintjob)
	PlaneGearState float32       // Plane landing gear state (matches C++ float planeGearState)
	Locked         uint8         // Door lock state (matches C++ unsigned char locked)
}

// Marshal serializes the packet to binary format
func (p *VehicleIdleUpdatePacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *VehicleIdleUpdatePacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// VehicleDriverUpdatePacket represents a vehicle driver update
// This matches the C++ CPackets::VehicleDriverUpdate structure exactly
type VehicleDriverUpdatePacket struct {
	PlayerID           types.PlayerID // Player ID driving the vehicle (matches C++ int playerid)
	VehicleID          int32          // Vehicle ID (matches C++ int vehicleid)
	Position           types.Vector3  // Vehicle position (matches C++ CVector pos)
	Rotation           types.Vector3  // Vehicle rotation (matches C++ CVector rot)
	Roll               types.Vector3  // Vehicle roll (matches C++ CVector roll)
	Velocity           types.Vector3  // Vehicle velocity (matches C++ CVector velocity)
	PlayerHealth       uint8          // Player health (matches C++ unsigned char playerHealth)
	PlayerArmour       uint8          // Player armour (matches C++ unsigned char playerArmour)
	Weapon             uint8          // Player weapon (matches C++ unsigned char weapon)
	Ammo               uint16         // Player ammo (matches C++ unsigned short ammo)
	Color1             uint8          // Primary color (matches C++ unsigned char color1)
	Color2             uint8          // Secondary color (matches C++ unsigned char color2)
	Health             float32        // Vehicle health (matches C++ float health)
	Paintjob           int8           // Paintjob ID (matches C++ char paintjob)
	BikeLean           float32        // Bike lean angle (matches C++ float bikeLean)
	MiscComponentAngle uint16         // Misc component angle like hydra thrusters (matches C++ unsigned short miscComponentAngle)
	PlaneGearState     float32        // Plane landing gear state (matches C++ float planeGearState)
	Locked             uint8          // Door lock state (matches C++ unsigned char locked)
}

// Marshal serializes the packet to binary format
func (p *VehicleDriverUpdatePacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *VehicleDriverUpdatePacket) Unmarshal(data []byte) error {
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

// NewVehicleIdleUpdatePacket creates a new vehicle idle update packet
func NewVehicleIdleUpdatePacket(vehicleID int32, pos, rot, roll, velocity, turnSpeed types.Vector3, color1, color2 uint8, health float32, paintjob int8, planeGearState float32, locked uint8) *VehicleIdleUpdatePacket {
	return &VehicleIdleUpdatePacket{
		VehicleID:      vehicleID,
		Position:       pos,
		Rotation:       rot,
		Roll:           roll,
		Velocity:       velocity,
		TurnSpeed:      turnSpeed,
		Color1:         color1,
		Color2:         color2,
		Health:         health,
		Paintjob:       paintjob,
		PlaneGearState: planeGearState,
		Locked:         locked,
	}
}

// NewVehicleDriverUpdatePacket creates a new vehicle driver update packet
func NewVehicleDriverUpdatePacket(playerID types.PlayerID, vehicleID int32, pos, rot, roll, velocity types.Vector3, playerHealth, playerArmour, weapon uint8, ammo uint16, color1, color2 uint8, health float32, paintjob int8, bikeLean float32, miscComponentAngle uint16, planeGearState float32, locked uint8) *VehicleDriverUpdatePacket {
	return &VehicleDriverUpdatePacket{
		PlayerID:           playerID,
		VehicleID:          vehicleID,
		Position:           pos,
		Rotation:           rot,
		Roll:               roll,
		Velocity:           velocity,
		PlayerHealth:       playerHealth,
		PlayerArmour:       playerArmour,
		Weapon:             weapon,
		Ammo:               ammo,
		Color1:             color1,
		Color2:             color2,
		Health:             health,
		Paintjob:           paintjob,
		BikeLean:           bikeLean,
		MiscComponentAngle: miscComponentAngle,
		PlaneGearState:     planeGearState,
		Locked:             locked,
	}
}
