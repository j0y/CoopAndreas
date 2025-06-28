package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"coopandreas-server/internal/types"
)

// GameWeatherTimePacket represents weather and time synchronization data
// This matches the C++ CPackets::GameWeatherTime structure exactly
type GameWeatherTimePacket struct {
	NewWeather    uint8  // Current weather type
	OldWeather    uint8  // Previous weather type
	ForcedWeather uint8  // Forced weather type
	CurrentMonth  uint8  // Current month (1-12)
	CurrentDay    uint8  // Current day (1-31)
	CurrentHour   uint8  // Current hour (0-23)
	CurrentMinute uint8  // Current minute (0-59)
	GameTickCount uint32 // Game milliseconds per minute
}

// Marshal serializes the packet to binary format
func (p *GameWeatherTimePacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *GameWeatherTimePacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// NewGameWeatherTimePacket creates a new game weather/time packet
func NewGameWeatherTimePacket(newWeather, oldWeather, forcedWeather, month, day, hour, minute uint8, gameTickCount uint32) *GameWeatherTimePacket {
	return &GameWeatherTimePacket{
		NewWeather:    newWeather,
		OldWeather:    oldWeather,
		ForcedWeather: forcedWeather,
		CurrentMonth:  month,
		CurrentDay:    day,
		CurrentHour:   hour,
		CurrentMinute: minute,
		GameTickCount: gameTickCount,
	}
}

// PlayerStatsPacket represents player statistics synchronization data
// This matches the C++ CPackets::PlayerStats structure exactly
type PlayerStatsPacket struct {
	PlayerID types.PlayerID // Player ID (matches C++ int playerid)
	Stats    [14]float32    // Player statistics array (matches C++ float stats[14])
}

// Marshal serializes the packet to binary format
func (p *PlayerStatsPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *PlayerStatsPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// NewPlayerStatsPacket creates a new player stats packet
func NewPlayerStatsPacket(playerID types.PlayerID, stats [14]float32) *PlayerStatsPacket {
	return &PlayerStatsPacket{
		PlayerID: playerID,
		Stats:    stats,
	}
}

// RebuildPlayerPacket represents player appearance/clothes rebuilding data
// This matches the C++ CPackets::RebuildPlayer structure exactly
type RebuildPlayerPacket struct {
	PlayerID    types.PlayerID // Player ID (matches C++ int playerid)
	ModelKeys   [10]uint32     // Model keys for clothes/appearance (matches C++ unsigned int m_anModelKeys[10])
	TextureKeys [18]uint32     // Texture keys for clothes/appearance (matches C++ unsigned int m_anTextureKeys[18])
	FatStat     float32        // Fat statistic (matches C++ float m_fFatStat)
	MuscleStat  float32        // Muscle statistic (matches C++ float m_fMuscleStat)
}

// Marshal serializes the packet to binary format
func (p *RebuildPlayerPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *RebuildPlayerPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// NewRebuildPlayerPacket creates a new rebuild player packet
func NewRebuildPlayerPacket(playerID types.PlayerID, modelKeys [10]uint32, textureKeys [18]uint32, fatStat, muscleStat float32) *RebuildPlayerPacket {
	return &RebuildPlayerPacket{
		PlayerID:    playerID,
		ModelKeys:   modelKeys,
		TextureKeys: textureKeys,
		FatStat:     fatStat,
		MuscleStat:  muscleStat,
	}
}

// PlayerPlaceWaypointPacket represents player waypoint placement/removal data
// This matches the C++ CPackets::PlayerPlaceWaypoint structure exactly
type PlayerPlaceWaypointPacket struct {
	PlayerID types.PlayerID // Player ID (matches C++ int playerid)
	Place    bool           // Whether to place (true) or remove (false) waypoint (matches C++ bool place)
	Position types.Vector3  // Waypoint position (matches C++ CVector position)
}

// Marshal serializes the packet to binary format
// Manual marshaling to match C++ struct layout (bool is 1 byte in C++)
func (p *PlayerPlaceWaypointPacket) Marshal() ([]byte, error) {
	buf := make([]byte, 17) // 4 bytes playerid + 1 byte bool + 12 bytes Vector3

	// Write PlayerID (4 bytes)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(p.PlayerID))

	// Write Place as byte (1 byte, matching C++ bool)
	if p.Place {
		buf[4] = 1
	} else {
		buf[4] = 0
	}

	// Write Position (12 bytes - 3 float32s)
	binary.LittleEndian.PutUint32(buf[5:9], math.Float32bits(p.Position.X))
	binary.LittleEndian.PutUint32(buf[9:13], math.Float32bits(p.Position.Y))
	binary.LittleEndian.PutUint32(buf[13:17], math.Float32bits(p.Position.Z))

	return buf, nil
}

// Unmarshal deserializes binary data to packet
func (p *PlayerPlaceWaypointPacket) Unmarshal(data []byte) error {
	if len(data) < 17 { // 4 bytes PlayerID + 1 byte Place + 12 bytes Position
		return fmt.Errorf("PlayerPlaceWaypointPacket: insufficient data, got %d bytes, expected 17", len(data))
	}

	// Read PlayerID (4 bytes)
	p.PlayerID = types.PlayerID(binary.LittleEndian.Uint32(data[0:4]))

	// Read Place (1 byte)
	p.Place = data[4] != 0

	// Read Position (12 bytes)
	p.Position.X = math.Float32frombits(binary.LittleEndian.Uint32(data[5:9]))
	p.Position.Y = math.Float32frombits(binary.LittleEndian.Uint32(data[9:13]))
	p.Position.Z = math.Float32frombits(binary.LittleEndian.Uint32(data[13:17]))

	return nil
}

// NewPlayerPlaceWaypointPacket creates a new player place waypoint packet
func NewPlayerPlaceWaypointPacket(playerID types.PlayerID, place bool, position types.Vector3) *PlayerPlaceWaypointPacket {
	return &PlayerPlaceWaypointPacket{
		PlayerID: playerID,
		Place:    place,
		Position: position,
	}
}
