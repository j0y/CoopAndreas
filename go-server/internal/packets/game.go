package packets

import (
	"bytes"
	"encoding/binary"
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
