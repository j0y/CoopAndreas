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

// Marshal serializes the packet to binary format with C++ struct packing
func (p *GameWeatherTimePacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write each field manually to ensure exact C++ struct layout
	if err := binary.Write(buf, binary.LittleEndian, p.NewWeather); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.OldWeather); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.ForcedWeather); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.CurrentMonth); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.CurrentDay); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.CurrentHour); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.CurrentMinute); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.GameTickCount); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes binary data to packet with C++ struct packing
func (p *GameWeatherTimePacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)

	// Read each field manually to ensure exact C++ struct layout
	if err := binary.Read(buf, binary.LittleEndian, &p.NewWeather); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.OldWeather); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.ForcedWeather); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.CurrentMonth); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.CurrentDay); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.CurrentHour); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.CurrentMinute); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.GameTickCount); err != nil {
		return err
	}

	return nil
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

// Marshal serializes the packet to binary format with C++ struct packing
func (p *PlayerStatsPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write PlayerID manually
	if err := binary.Write(buf, binary.LittleEndian, p.PlayerID); err != nil {
		return nil, err
	}

	// Write Stats array manually
	if err := binary.Write(buf, binary.LittleEndian, p.Stats); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes binary data to packet with C++ struct packing
func (p *PlayerStatsPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)

	// Read PlayerID manually
	if err := binary.Read(buf, binary.LittleEndian, &p.PlayerID); err != nil {
		return err
	}

	// Read Stats array manually
	if err := binary.Read(buf, binary.LittleEndian, &p.Stats); err != nil {
		return err
	}

	return nil
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

// Marshal serializes the packet to binary format with C++ struct packing
func (p *RebuildPlayerPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write PlayerID manually
	if err := binary.Write(buf, binary.LittleEndian, p.PlayerID); err != nil {
		return nil, err
	}

	// Write ModelKeys array manually
	if err := binary.Write(buf, binary.LittleEndian, p.ModelKeys); err != nil {
		return nil, err
	}

	// Write TextureKeys array manually
	if err := binary.Write(buf, binary.LittleEndian, p.TextureKeys); err != nil {
		return nil, err
	}

	// Write FatStat manually
	if err := binary.Write(buf, binary.LittleEndian, p.FatStat); err != nil {
		return nil, err
	}

	// Write MuscleStat manually
	if err := binary.Write(buf, binary.LittleEndian, p.MuscleStat); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes binary data to packet with C++ struct packing
func (p *RebuildPlayerPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)

	// Read PlayerID manually
	if err := binary.Read(buf, binary.LittleEndian, &p.PlayerID); err != nil {
		return err
	}

	// Read ModelKeys array manually
	if err := binary.Read(buf, binary.LittleEndian, &p.ModelKeys); err != nil {
		return err
	}

	// Read TextureKeys array manually
	if err := binary.Read(buf, binary.LittleEndian, &p.TextureKeys); err != nil {
		return err
	}

	// Read FatStat manually
	if err := binary.Read(buf, binary.LittleEndian, &p.FatStat); err != nil {
		return err
	}

	// Read MuscleStat manually
	if err := binary.Read(buf, binary.LittleEndian, &p.MuscleStat); err != nil {
		return err
	}

	return nil
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

// PlayMissionAudioPacket represents mission audio playback synchronization data
// This matches the C++ CPackets::PlayMissionAudio structure exactly
type PlayMissionAudioPacket struct {
	SlotID  uint8 // Audio slot ID (matches C++ uint8_t slotid)
	AudioID int32 // Audio ID to play (matches C++ int audioid)
}

// Marshal serializes the packet to binary format with C++ struct packing
func (p *PlayMissionAudioPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write each field manually to ensure exact C++ struct layout
	if err := binary.Write(buf, binary.LittleEndian, p.SlotID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.AudioID); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes binary data to packet with C++ struct packing
func (p *PlayMissionAudioPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)

	// Read each field manually to ensure exact C++ struct layout
	if err := binary.Read(buf, binary.LittleEndian, &p.SlotID); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.AudioID); err != nil {
		return err
	}

	return nil
}

// NewPlayMissionAudioPacket creates a new play mission audio packet
func NewPlayMissionAudioPacket(slotID uint8, audioID int32) *PlayMissionAudioPacket {
	return &PlayMissionAudioPacket{
		SlotID:  slotID,
		AudioID: audioID,
	}
}

// AddExplosionPacket represents an explosion event
// This matches the C++ CPackets::AddExplosion structure exactly
type AddExplosionPacket struct {
	Type        uint8         // Explosion type (eExplosionType enum)
	Position    types.Vector3 // Explosion position
	Time        int32         // Time parameter
	UsesSound   bool          // Whether explosion uses sound
	CameraShake float32       // Camera shake intensity
	IsVisible   bool          // Whether explosion is visible
}

// Marshal serializes the packet to binary format with C++ struct packing
func (p *AddExplosionPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write each field manually to ensure exact C++ struct layout
	if err := binary.Write(buf, binary.LittleEndian, p.Type); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Position); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Time); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.UsesSound); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.CameraShake); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.IsVisible); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes binary data to packet with C++ struct packing
func (p *AddExplosionPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)

	// Read each field manually to ensure exact C++ struct layout
	if err := binary.Read(buf, binary.LittleEndian, &p.Type); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Position); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Time); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.UsesSound); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.CameraShake); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.IsVisible); err != nil {
		return err
	}

	return nil
}

// NewAddExplosionPacket creates a new add explosion packet
func NewAddExplosionPacket(explosionType uint8, position types.Vector3, time int32, usesSound bool, cameraShake float32, isVisible bool) *AddExplosionPacket {
	return &AddExplosionPacket{
		Type:        explosionType,
		Position:    position,
		Time:        time,
		UsesSound:   usesSound,
		CameraShake: cameraShake,
		IsVisible:   isVisible,
	}
}

// StartCutscenePacket represents a start cutscene request
// This matches the C++ CPackets::StartCutscene structure exactly
type StartCutscenePacket struct {
	Name     [8]byte // Cutscene name (8 chars, null-terminated)
	CurrArea uint8   // Current area/interior ID
}

// Marshal serializes the packet to binary format
func (p *StartCutscenePacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *StartCutscenePacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// GetCutsceneName returns the cutscene name as a string
func (p *StartCutscenePacket) GetCutsceneName() string {
	// Find the null terminator and return the string up to that point
	for i, b := range p.Name {
		if b == 0 {
			return string(p.Name[:i])
		}
	}
	return string(p.Name[:])
}

// SetCutsceneName sets the cutscene name (truncates if longer than 7 chars)
func (p *StartCutscenePacket) SetCutsceneName(name string) {
	// Clear the array first
	for i := range p.Name {
		p.Name[i] = 0
	}

	// Copy the name (max 7 chars to leave room for null terminator)
	maxLen := len(p.Name) - 1
	if len(name) < maxLen {
		maxLen = len(name)
	}
	copy(p.Name[:], name[:maxLen])
}

// SkipCutscenePacket represents a skip cutscene request
// This matches the C++ CPackets::SkipCutscene structure exactly
type SkipCutscenePacket struct {
	PlayerID int32 // Player ID who requested the skip
	Votes    int32 // Voting system (currently unused, as per C++ comment)
}

// Marshal serializes the packet to binary format
func (p *SkipCutscenePacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *SkipCutscenePacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, p)
}

// NewSkipCutscenePacket creates a new skip cutscene packet
func NewSkipCutscenePacket(playerID int32, votes int32) *SkipCutscenePacket {
	return &SkipCutscenePacket{
		PlayerID: playerID,
		Votes:    votes,
	}
}

// OnMissionFlagSyncPacket represents mission flag synchronization
// This matches the C++ CPackets::OnMissionFlagSync structure exactly
// Note: C++ uses bitfield (uint8_t bOnMission : 1), but we use full uint8 for simplicity
type OnMissionFlagSyncPacket struct {
	OnMission uint8 // 1 if player is on mission, 0 if not (matches C++ bitfield behavior)
}

// Marshal serializes the packet to binary format
func (p *OnMissionFlagSyncPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p.OnMission)
	return buf.Bytes(), err
}

// Unmarshal deserializes binary data to packet
func (p *OnMissionFlagSyncPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, &p.OnMission)
}

// IsOnMission returns true if player is on a mission
func (p *OnMissionFlagSyncPacket) IsOnMission() bool {
	return p.OnMission != 0
}

// SetOnMission sets the mission flag
func (p *OnMissionFlagSyncPacket) SetOnMission(onMission bool) {
	if onMission {
		p.OnMission = 1
	} else {
		p.OnMission = 0
	}
}

// NewOnMissionFlagSyncPacket creates a new mission flag sync packet
func NewOnMissionFlagSyncPacket(onMission bool) *OnMissionFlagSyncPacket {
	packet := &OnMissionFlagSyncPacket{}
	packet.SetOnMission(onMission)
	return packet
}

// UpdateEntityBlipPacket represents entity blip update data
// This matches the C++ CPackets::UpdateEntityBlip structure exactly
type UpdateEntityBlipPacket struct {
	PlayerID   int32                   // Target player ID (matches C++ int playerid)
	EntityType types.NetworkEntityType // Entity type (matches C++ eNetworkEntityType)
	EntityID   int32                   // Entity ID (matches C++ int entityId)
	IsFriendly bool                    // Whether blip is friendly (matches C++ bool isFriendly)
	Color      uint8                   // Blip color (matches C++ uint8_t color)
	Display    uint8                   // Display type (matches C++ uint8_t display)
	Scale      uint8                   // Blip scale (matches C++ uint8_t scale)
}

// Marshal serializes the packet to binary format with C++ struct packing
func (p *UpdateEntityBlipPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write each field manually to ensure exact C++ struct layout
	if err := binary.Write(buf, binary.LittleEndian, p.PlayerID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.EntityType); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.EntityID); err != nil {
		return nil, err
	}
	// Convert bool to uint8 for C++ compatibility
	var isFriendlyByte uint8
	if p.IsFriendly {
		isFriendlyByte = 1
	}
	if err := binary.Write(buf, binary.LittleEndian, isFriendlyByte); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Color); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Display); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Scale); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes binary data to packet with C++ struct packing
func (p *UpdateEntityBlipPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)

	// Read each field manually to ensure exact C++ struct layout
	if err := binary.Read(buf, binary.LittleEndian, &p.PlayerID); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.EntityType); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.EntityID); err != nil {
		return err
	}
	// Read bool as uint8 for C++ compatibility
	var isFriendlyByte uint8
	if err := binary.Read(buf, binary.LittleEndian, &isFriendlyByte); err != nil {
		return err
	}
	p.IsFriendly = isFriendlyByte != 0
	if err := binary.Read(buf, binary.LittleEndian, &p.Color); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Display); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Scale); err != nil {
		return err
	}

	return nil
}

// NewUpdateEntityBlipPacket creates a new entity blip update packet
func NewUpdateEntityBlipPacket(playerID int32, entityType types.NetworkEntityType, entityID int32,
	isFriendly bool, color, display, scale uint8) *UpdateEntityBlipPacket {
	return &UpdateEntityBlipPacket{
		PlayerID:   playerID,
		EntityType: entityType,
		EntityID:   entityID,
		IsFriendly: isFriendly,
		Color:      color,
		Display:    display,
		Scale:      scale,
	}
}

// RemoveEntityBlipPacket represents entity blip removal data
// This matches the C++ CPackets::RemoveEntityBlip structure exactly
type RemoveEntityBlipPacket struct {
	PlayerID   int32                   // Target player ID (matches C++ int playerid)
	EntityType types.NetworkEntityType // Entity type (matches C++ eNetworkEntityType)
	EntityID   int32                   // Entity ID (matches C++ int entityId)
}

// Marshal serializes the packet to binary format with C++ struct packing
func (p *RemoveEntityBlipPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write each field manually to ensure exact C++ struct layout
	if err := binary.Write(buf, binary.LittleEndian, p.PlayerID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.EntityType); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.EntityID); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes binary data to packet with C++ struct packing
func (p *RemoveEntityBlipPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)

	// Read each field manually to ensure exact C++ struct layout
	if err := binary.Read(buf, binary.LittleEndian, &p.PlayerID); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.EntityType); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.EntityID); err != nil {
		return err
	}

	return nil
}

// NewRemoveEntityBlipPacket creates a new entity blip removal packet
func NewRemoveEntityBlipPacket(playerID int32, entityType types.NetworkEntityType, entityID int32) *RemoveEntityBlipPacket {
	return &RemoveEntityBlipPacket{
		PlayerID:   playerID,
		EntityType: entityType,
		EntityID:   entityID,
	}
}

// AddMessageGXTPacket represents GXT message addition data
// This matches the C++ CPackets::AddMessageGXT structure exactly
type AddMessageGXTPacket struct {
	PlayerID int32   // Target player ID (matches C++ int playerid)
	Type     uint8   // Message type: 0=PRINT, 1=PRINT_BIG, 2=PRINT_NOW, 3=PRINT_HELP (matches C++ uint8_t type)
	Time     uint32  // Display time in milliseconds (matches C++ uint32_t time)
	Flag     uint8   // Additional flags (matches C++ uint8_t flag)
	GXT      [8]byte // GXT string identifier (matches C++ char gxt[8])
}

// Marshal serializes the packet to binary format with C++ struct packing
func (p *AddMessageGXTPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write each field manually to ensure exact C++ struct layout
	if err := binary.Write(buf, binary.LittleEndian, p.PlayerID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Type); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Time); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.Flag); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.GXT); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes binary data to packet with C++ struct packing
func (p *AddMessageGXTPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)

	// Read each field manually to ensure exact C++ struct layout
	if err := binary.Read(buf, binary.LittleEndian, &p.PlayerID); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Type); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Time); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.Flag); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.GXT); err != nil {
		return err
	}

	return nil
}

// GetGXTString returns the GXT string as a string
func (p *AddMessageGXTPacket) GetGXTString() string {
	// Find the null terminator and return the string up to that point
	for i, b := range p.GXT {
		if b == 0 {
			return string(p.GXT[:i])
		}
	}
	return string(p.GXT[:])
}

// SetGXTString sets the GXT string (truncates if longer than 7 chars)
func (p *AddMessageGXTPacket) SetGXTString(gxt string) {
	// Clear the array first
	for i := range p.GXT {
		p.GXT[i] = 0
	}

	// Copy the string (max 7 chars to leave room for null terminator)
	maxLen := len(p.GXT) - 1
	if len(gxt) < maxLen {
		maxLen = len(gxt)
	}
	copy(p.GXT[:], gxt[:maxLen])
}

// NewAddMessageGXTPacket creates a new GXT message addition packet
func NewAddMessageGXTPacket(playerID int32, msgType uint8, time uint32, flag uint8, gxt string) *AddMessageGXTPacket {
	packet := &AddMessageGXTPacket{
		PlayerID: playerID,
		Type:     msgType,
		Time:     time,
		Flag:     flag,
	}
	packet.SetGXTString(gxt)
	return packet
}

// RemoveMessageGXTPacket represents GXT message removal data
// This matches the C++ CPackets::RemoveMessageGXT structure exactly
type RemoveMessageGXTPacket struct {
	PlayerID int32   // Target player ID (matches C++ int playerid)
	GXT      [8]byte // GXT string identifier (matches C++ char gxt[8])
}

// Marshal serializes the packet to binary format with C++ struct packing
func (p *RemoveMessageGXTPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write each field manually to ensure exact C++ struct layout
	if err := binary.Write(buf, binary.LittleEndian, p.PlayerID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, p.GXT); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes binary data to packet with C++ struct packing
func (p *RemoveMessageGXTPacket) Unmarshal(data []byte) error {
	buf := bytes.NewReader(data)

	// Read each field manually to ensure exact C++ struct layout
	if err := binary.Read(buf, binary.LittleEndian, &p.PlayerID); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.LittleEndian, &p.GXT); err != nil {
		return err
	}

	return nil
}

// GetGXTString returns the GXT string as a string
func (p *RemoveMessageGXTPacket) GetGXTString() string {
	// Find the null terminator and return the string up to that point
	for i, b := range p.GXT {
		if b == 0 {
			return string(p.GXT[:i])
		}
	}
	return string(p.GXT[:])
}

// SetGXTString sets the GXT string (truncates if longer than 7 chars)
func (p *RemoveMessageGXTPacket) SetGXTString(gxt string) {
	// Clear the array first
	for i := range p.GXT {
		p.GXT[i] = 0
	}

	// Copy the string (max 7 chars to leave room for null terminator)
	maxLen := len(p.GXT) - 1
	if len(gxt) < maxLen {
		maxLen = len(gxt)
	}
	copy(p.GXT[:], gxt[:maxLen])
}

// NewRemoveMessageGXTPacket creates a new GXT message removal packet
func NewRemoveMessageGXTPacket(playerID int32, gxt string) *RemoveMessageGXTPacket {
	packet := &RemoveMessageGXTPacket{
		PlayerID: playerID,
	}
	packet.SetGXTString(gxt)
	return packet
}
