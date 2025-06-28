package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"

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

// VehicleEnterPacket represents a vehicle enter request/notification
// This matches the C++ CPackets::VehicleEnter structure exactly
// Note: C++ uses bitfields which need manual marshaling to match binary layout
type VehicleEnterPacket struct {
	PlayerID  types.PlayerID // Player ID entering the vehicle (matches C++ int playerid)
	VehicleID int32          // Vehicle ID (matches C++ int vehicleid)
	SeatID    uint8          // Seat ID (0=driver, 1-3=passengers) - matches C++ unsigned char seatid : 3
	Force     bool           // Force enter flag - matches C++ unsigned char force : 1
	Passenger bool           // Is passenger flag - matches C++ unsigned char passenger : 1
}

// Marshal serializes the packet to binary format
// Manual marshaling to match C++ bitfield layout
func (p *VehicleEnterPacket) Marshal() ([]byte, error) {
	buf := make([]byte, 9) // 4 bytes playerid + 4 bytes vehicleid + 1 byte bitfield

	// Write PlayerID (4 bytes)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(p.PlayerID))

	// Write VehicleID (4 bytes)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(p.VehicleID))

	// Pack bitfields into single byte (matches C++ bitfield layout)
	var bitfield uint8
	bitfield |= p.SeatID & 0x07 // 3 bits for seatid (0-7)
	if p.Force {
		bitfield |= 0x08 // bit 3 for force
	}
	if p.Passenger {
		bitfield |= 0x10 // bit 4 for passenger
	}
	buf[8] = bitfield

	return buf, nil
}

// Unmarshal deserializes binary data to packet
// Manual unmarshaling to match C++ bitfield layout
func (p *VehicleEnterPacket) Unmarshal(data []byte) error {
	if len(data) < 9 {
		return fmt.Errorf("VehicleEnterPacket: insufficient data, got %d bytes, expected 9", len(data))
	}

	// Read PlayerID (4 bytes)
	p.PlayerID = types.PlayerID(binary.LittleEndian.Uint32(data[0:4]))

	// Read VehicleID (4 bytes)
	p.VehicleID = int32(binary.LittleEndian.Uint32(data[4:8]))

	// Unpack bitfields from single byte
	bitfield := data[8]
	p.SeatID = bitfield & 0x07           // Extract bits 0-2 for seatid
	p.Force = (bitfield & 0x08) != 0     // Extract bit 3 for force
	p.Passenger = (bitfield & 0x10) != 0 // Extract bit 4 for passenger

	return nil
}

// VehicleExitPacket represents a player exiting a vehicle
// This matches the C++ CVehiclePackets::VehicleExit structure exactly
type VehicleExitPacket struct {
	PlayerID types.PlayerID // Player ID exiting the vehicle (matches C++ int playerid)
	Force    bool           // Whether exit is forced (matches C++ bool force)
}

// Marshal serializes the packet to binary format
func (p *VehicleExitPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write PlayerID (4 bytes)
	if err := binary.Write(buf, binary.LittleEndian, p.PlayerID); err != nil {
		return nil, err
	}

	// Write Force as byte (1 byte, matching C++ bool)
	var forceByte uint8
	if p.Force {
		forceByte = 1
	}
	if err := binary.Write(buf, binary.LittleEndian, forceByte); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes binary data to packet
func (p *VehicleExitPacket) Unmarshal(data []byte) error {
	if len(data) < 5 { // 4 bytes PlayerID + 1 byte Force
		return fmt.Errorf("VehicleExitPacket: insufficient data, got %d bytes, expected 5", len(data))
	}

	buf := bytes.NewReader(data)

	// Read PlayerID (4 bytes)
	if err := binary.Read(buf, binary.LittleEndian, &p.PlayerID); err != nil {
		return err
	}

	// Read Force as byte (1 byte)
	var forceByte uint8
	if err := binary.Read(buf, binary.LittleEndian, &forceByte); err != nil {
		return err
	}
	p.Force = forceByte != 0

	return nil
}

// VehiclePassengerUpdate represents a passenger update in a vehicle
// This matches the C++ CVehiclePackets::VehiclePassengerUpdate structure exactly
type VehiclePassengerUpdatePacket struct {
	PlayerID     int32  // Player ID (set by server)
	VehicleID    int32  // Vehicle ID
	PlayerHealth uint8  // Player health (0-255)
	PlayerArmour uint8  // Player armour (0-255)
	Weapon       uint8  // Current weapon ID
	Ammo         uint16 // Current weapon ammo
	Driveby      uint8  // Whether player is in driveby mode (0/1)
	SeatID       uint8  // Seat ID (passenger seat number)
}

// Marshal serializes the packet to binary format
func (p *VehiclePassengerUpdatePacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *VehiclePassengerUpdatePacket) Unmarshal(data []byte) error {
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

// NewVehicleEnterPacket creates a new vehicle enter packet
func NewVehicleEnterPacket(playerID types.PlayerID, vehicleID int32, seatID uint8, force, passenger bool) *VehicleEnterPacket {
	return &VehicleEnterPacket{
		PlayerID:  playerID,
		VehicleID: vehicleID,
		SeatID:    seatID,
		Force:     force,
		Passenger: passenger,
	}
}

// NewVehicleExitPacket creates a new vehicle exit packet
func NewVehicleExitPacket(playerID types.PlayerID, force bool) *VehicleExitPacket {
	return &VehicleExitPacket{
		PlayerID: playerID,
		Force:    force,
	}
}

// NewVehiclePassengerUpdatePacket creates a new vehicle passenger update packet
func NewVehiclePassengerUpdatePacket(playerID types.PlayerID, vehicleID int32, health, armour, weapon uint8, ammo uint16, driveby, seatID uint8) *VehiclePassengerUpdatePacket {
	return &VehiclePassengerUpdatePacket{
		PlayerID:     int32(playerID),
		VehicleID:    vehicleID,
		PlayerHealth: health,
		PlayerArmour: armour,
		Weapon:       weapon,
		Ammo:         ammo,
		Driveby:      driveby,
		SeatID:       seatID,
	}
}
