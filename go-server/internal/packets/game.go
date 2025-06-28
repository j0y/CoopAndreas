package packets

import (
	"bytes"
	"encoding/binary"

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
