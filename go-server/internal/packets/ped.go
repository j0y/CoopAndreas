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

// PedDriverUpdatePacket represents a ped driving a vehicle update
// This matches the C++ CPackets::PedDriverUpdate structure exactly
type PedDriverUpdatePacket struct {
	PedID             types.PedID   // Ped ID (matches C++ int pedid)
	VehicleID         int32         // Vehicle ID (matches C++ int vehicleid)
	Position          types.Vector3 // Vehicle position (matches C++ CVector pos)
	Rotation          types.Vector3 // Vehicle rotation (matches C++ CVector rot)
	Roll              types.Vector3 // Vehicle roll (matches C++ CVector roll)
	Velocity          types.Vector3 // Vehicle velocity (matches C++ CVector velocity)
	TurnSpeed         types.Vector3 // Vehicle turn speed (matches C++ CVector turnSpeed)
	PedHealth         uint8         // Ped health (matches C++ unsigned char pedHealth)
	PedArmour         uint8         // Ped armour (matches C++ unsigned char pedArmour)
	Weapon            uint8         // Ped weapon (matches C++ unsigned char weapon)
	Ammo              uint16        // Ped ammo (matches C++ unsigned short ammo)
	Color1            uint8         // Vehicle primary color (matches C++ unsigned char color1)
	Color2            uint8         // Vehicle secondary color (matches C++ unsigned char color2)
	Health            float32       // Vehicle health (matches C++ float health)
	Paintjob          int8          // Vehicle paintjob (matches C++ char paintjob)
	BikeLean          float32       // Bike lean angle (matches C++ float bikeLean)
	PlaneGearState    float32       // Plane gear state / control pedaling (matches C++ union field)
	Locked            uint8         // Door lock state (matches C++ unsigned char locked)
	GasPedal          float32       // Gas pedal position (matches C++ float gasPedal)
	BreakPedal        float32       // Brake pedal position (matches C++ float breakPedal)
	DrivingStyle      uint8         // Driving style (matches C++ uint8_t drivingStyle)
	CarMission        uint8         // Car mission type (matches C++ uint8_t carMission)
	CruiseSpeed       int8          // Cruise speed (matches C++ int8_t cruiseSpeed)
	CtrlFlags         uint8         // Control flags (matches C++ uint8_t ctrlFlags)
	MovementFlags     uint8         // Movement flags (matches C++ uint8_t movementFlags)
	TargetVehicleID   int32         // Target vehicle ID (matches C++ int targetVehicleId)
	DestinationCoords types.Vector3 // Destination coordinates (matches C++ CVector destinationCoors)
}

// Marshal serializes the packet to binary format with C++ struct packing
func (p *PedDriverUpdatePacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write each field manually to ensure exact C++ struct layout
	if err := binary.Write(buf, binary.LittleEndian, p.PedID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.VehicleID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Position); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Rotation); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Roll); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Velocity); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.TurnSpeed); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.PedHealth); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.PedArmour); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Weapon); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Ammo); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Color1); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Color2); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Health); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Paintjob); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.BikeLean); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.PlaneGearState); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Locked); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.GasPedal); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.BreakPedal); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.DrivingStyle); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.CarMission); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.CruiseSpeed); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.CtrlFlags); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.MovementFlags); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.TargetVehicleID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.DestinationCoords); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes binary data to packet with C++ struct packing
func (p *PedDriverUpdatePacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)

	// Read each field manually to ensure exact C++ struct layout
	if err := binary.Read(buf, binary.LittleEndian, &p.PedID); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.VehicleID); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Position); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Rotation); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Roll); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Velocity); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.TurnSpeed); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.PedHealth); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.PedArmour); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Weapon); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Ammo); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Color1); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Color2); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Health); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Paintjob); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.BikeLean); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.PlaneGearState); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Locked); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.GasPedal); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.BreakPedal); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.DrivingStyle); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.CarMission); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.CruiseSpeed); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.CtrlFlags); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.MovementFlags); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.TargetVehicleID); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.DestinationCoords); err != nil {
		return err
	}

	return nil
}

// NewPedDriverUpdatePacket creates a new ped driver update packet
func NewPedDriverUpdatePacket(pedID types.PedID, vehicleID int32, pos, rot, roll, velocity, turnSpeed types.Vector3, pedHealth, pedArmour, weapon uint8, ammo uint16, color1, color2 uint8, health float32, paintjob int8, bikeLean, planeGearState float32, locked uint8, gasPedal, breakPedal float32, drivingStyle, carMission uint8, cruiseSpeed int8, ctrlFlags, movementFlags uint8, targetVehicleID int32, destinationCoords types.Vector3) *PedDriverUpdatePacket {
	return &PedDriverUpdatePacket{
		PedID:             pedID,
		VehicleID:         vehicleID,
		Position:          pos,
		Rotation:          rot,
		Roll:              roll,
		Velocity:          velocity,
		TurnSpeed:         turnSpeed,
		PedHealth:         pedHealth,
		PedArmour:         pedArmour,
		Weapon:            weapon,
		Ammo:              ammo,
		Color1:            color1,
		Color2:            color2,
		Health:            health,
		Paintjob:          paintjob,
		BikeLean:          bikeLean,
		PlaneGearState:    planeGearState,
		Locked:            locked,
		GasPedal:          gasPedal,
		BreakPedal:        breakPedal,
		DrivingStyle:      drivingStyle,
		CarMission:        carMission,
		CruiseSpeed:       cruiseSpeed,
		CtrlFlags:         ctrlFlags,
		MovementFlags:     movementFlags,
		TargetVehicleID:   targetVehicleID,
		DestinationCoords: destinationCoords,
	}
}
