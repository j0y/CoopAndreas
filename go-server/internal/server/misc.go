package server

import (
	"fmt"

	"coopandreas-server/internal/network"
	"coopandreas-server/internal/packets"
	"coopandreas-server/internal/types"
)

// handleMassPacketSequence handles MASS_PACKET_SEQUENCE packets
// Based on C++ server logic: simply rebroadcasts the entire mass packet to all other clients
// This matches the C++ implementation in CNetwork::HandlePacketReceive
func (s *Server) handleMassPacketSequence(client *network.Client, data []byte) error {
	// In C++: CNetwork::SendPacketRawToAll(event.packet->data, event.packet->dataLength, (ENetPacketFlag)event.packet->flags, event.peer);
	// The server simply rebroadcasts the entire mass packet sequence to all other clients
	// without parsing or processing the individual packets within it

	//s.logger.Debug().
	//	Str("client", client.Addr.String()).
	//	Int("dataSize", len(data)).
	//	Msg("Received mass packet sequence, rebroadcasting to all clients")

	// Create network packet for rebroadcast
	networkPacket := &network.NetworkPacket{
		ID:   types.MASS_PACKET_SEQUENCE,
		Data: data,
		Flag: 0, // Use same flags as original packet (for now, use unreliable)
	}

	// Rebroadcast to all other clients (excluding sender)
	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to rebroadcast mass packet sequence: %w", err)
	}

	return nil
}

// handleGameWeatherTime handles GAME_WEATHER_TIME packets
// This matches the C++ CPlayerPackets::GameWeatherTime::Handle() functionality
// Only the host player can send weather/time updates to synchronize all clients
func (s *Server) handleGameWeatherTime(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Check if player is host (matching C++ logic: if (!CPlayerManager::GetPlayer(peer)->m_bIsHost) return;)
	if !player.IsHost {
		s.logger.Warn().
			Str("player", player.Name).
			Int32("playerID", int32(player.ID)).
			Msg("Non-host player attempted to send weather/time update")
		return nil // Ignore non-host weather updates
	}

	// Debug: Check packet size (expected: 11 bytes based on C++ struct)
	packets.DebugPacketSize("GAME_WEATHER_TIME", data, 11)

	// Parse the packet
	var packet packets.GameWeatherTimePacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal GameWeatherTime packet: %w", err)
	}

	s.logger.Debug().
		Str("hostPlayer", player.Name).
		Uint8("newWeather", packet.NewWeather).
		Uint8("currentHour", packet.CurrentHour).
		Uint8("currentMinute", packet.CurrentMinute).
		Msg("Host updated weather/time")

	// Store the current weather/time state for new client synchronization
	s.currentWeatherTime = &packet

	// Marshal the packet for rebroadcast
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal GameWeatherTime packet: %w", err)
	}

	// Broadcast to all clients except the sender (matching C++ logic)
	networkPacket := &network.NetworkPacket{
		ID:   types.GAME_WEATHER_TIME,
		Data: packetData,
		Flag: network.PacketFlagReliable, // Reliable for important game state changes
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast GameWeatherTime: %w", err)
	}

	return nil
}

// sendCurrentWeatherTimeTo sends the current weather/time state to a specific client
// This matches the C++ logic where GameWeatherTime__Trigger is called after player connects
func (s *Server) sendCurrentWeatherTimeTo(client *network.Client) error {
	// Only send if we have current weather/time state
	if s.currentWeatherTime == nil {
		s.logger.Debug().
			Uint32("clientID", client.ID).
			Msg("No current weather/time state to send to new client")
		return nil
	}

	// Marshal the current weather/time packet
	packetData, err := s.currentWeatherTime.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal current weather/time: %w", err)
	}

	// Send to the specific client (reliable packet)
	networkPacket := &network.NetworkPacket{
		ID:   types.GAME_WEATHER_TIME,
		Data: packetData,
		Flag: network.PacketFlagReliable,
	}

	if err := s.networkServer.SendPacket(client, networkPacket); err != nil {
		return fmt.Errorf("failed to send weather/time to client: %w", err)
	}

	s.logger.Debug().
		Uint32("clientID", client.ID).
		Uint8("newWeather", s.currentWeatherTime.NewWeather).
		Uint8("currentHour", s.currentWeatherTime.CurrentHour).
		Uint8("currentMinute", s.currentWeatherTime.CurrentMinute).
		Msg("Sent current weather/time to client")

	return nil
}

// broadcastCurrentWeatherTime broadcasts the current weather/time state to all clients
// This is used when a new host is assigned to synchronize all clients
func (s *Server) broadcastCurrentWeatherTime() error {
	// Only broadcast if we have current weather/time state
	if s.currentWeatherTime == nil {
		s.logger.Debug().Msg("No current weather/time state to broadcast")
		return nil
	}

	// Marshal the current weather/time packet
	packetData, err := s.currentWeatherTime.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal current weather/time: %w", err)
	}

	// Debug: Check outgoing packet size
	s.logger.Debug().
		Int("packetSize", len(packetData)).
		Uint8("newWeather", s.currentWeatherTime.NewWeather).
		Uint8("oldWeather", s.currentWeatherTime.OldWeather).
		Uint8("forcedWeather", s.currentWeatherTime.ForcedWeather).
		Uint8("currentMonth", s.currentWeatherTime.CurrentMonth).
		Uint8("currentDay", s.currentWeatherTime.CurrentDay).
		Uint8("currentHour", s.currentWeatherTime.CurrentHour).
		Uint8("currentMinute", s.currentWeatherTime.CurrentMinute).
		Uint32("gameTickCount", s.currentWeatherTime.GameTickCount).
		Msg("Broadcasting weather/time packet")

	// Broadcast to all clients (reliable packet)
	networkPacket := &network.NetworkPacket{
		ID:   types.GAME_WEATHER_TIME,
		Data: packetData,
		Flag: network.PacketFlagReliable,
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, nil); err != nil {
		return fmt.Errorf("failed to broadcast weather/time: %w", err)
	}

	s.logger.Debug().
		Uint8("newWeather", s.currentWeatherTime.NewWeather).
		Uint8("currentHour", s.currentWeatherTime.CurrentHour).
		Uint8("currentMinute", s.currentWeatherTime.CurrentMinute).
		Msg("Broadcasted current weather/time to all clients")

	return nil
}

// handleOpCodeSync handles OPCODE_SYNC packets
// This matches the C++ CPlayerPackets::OpCodeSync::Handle() functionality
// Only the host player can send opcode synchronization packets
func (s *Server) handleOpCodeSync(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Check if player is host (matching C++ logic: if (!CPlayerManager::GetPlayer(peer)->m_bIsHost) return;)
	if !player.IsHost {
		s.logger.Warn().
			Str("player", player.Name).
			Int32("playerID", int32(player.ID)).
			Msg("Non-host player attempted to send opcode sync")
		return nil // Ignore non-host opcode sync
	}

	// Parse the packet to validate structure (optional but good for debugging)
	var packet packets.OpcodeSyncPacket
	if err := packet.Unmarshal(data); err != nil {
		s.logger.Warn().
			Err(err).
			Str("hostPlayer", player.Name).
			Msg("Failed to parse opcode sync packet")
		// Still rebroadcast the raw data even if parsing fails, to match C++ behavior
	} else {
		s.logger.Debug().
			Str("hostPlayer", player.Name).
			Uint16("opcode", packet.Opcode).
			Uint8("intParams", packet.IntParamCount).
			Uint8("stringParams", packet.StringParamCount).
			Msg("Host sent opcode sync")
	}

	// Rebroadcast raw data to all clients except the sender (matching C++ logic)
	// This matches: CNetwork::SendPacketToAll(CPacketsID::OPCODE_SYNC, data, size, ENET_PACKET_FLAG_RELIABLE, peer);
	networkPacket := &network.NetworkPacket{
		ID:   types.OPCODE_SYNC,
		Data: data,                       // Send raw data as received
		Flag: network.PacketFlagReliable, // Reliable for script synchronization
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast opcode sync: %w", err)
	}

	return nil
}

// handlePlayMissionAudio handles PLAY_MISSION_AUDIO packets
// This matches the C++ CPlayerPackets::PlayMissionAudio::Handle() functionality
// Only the host player can send mission audio playback packets
func (s *Server) handlePlayMissionAudio(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Check if player is host (matching C++ logic: if (!player->m_bIsHost) return;)
	if !player.IsHost {
		s.logger.Warn().
			Str("player", player.Name).
			Int32("playerID", int32(player.ID)).
			Msg("Non-host player attempted to send mission audio")
		return nil // Ignore non-host mission audio
	}

	// Debug: Check packet size (expected: 5 bytes based on C++ struct)
	packets.DebugPacketSize("PLAY_MISSION_AUDIO", data, 5)

	// Parse the packet
	var packet packets.PlayMissionAudioPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal PlayMissionAudio packet: %w", err)
	}

	// Log the mission audio playback
	s.logger.Debug().
		Str("hostPlayer", player.Name).
		Uint8("slotID", packet.SlotID).
		Int32("audioID", packet.AudioID).
		Msg("Host triggered mission audio playback")

	// Broadcast to all other clients (reliable packet, matching C++ ENET_PACKET_FLAG_RELIABLE)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal PlayMissionAudio packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.PLAY_MISSION_AUDIO,
		Data: packetData,
		Flag: network.PacketFlagReliable, // Reliable for mission audio synchronization
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast mission audio: %w", err)
	}

	s.logger.Info().
		Str("hostPlayer", player.Name).
		Uint8("slotID", packet.SlotID).
		Int32("audioID", packet.AudioID).
		Msg("Broadcasted mission audio to all clients")

	return nil
}

// handleAddExplosion handles ADD_EXPLOSION packets
// This matches the C++ CPlayerPackets::AddExplosion::Handle() functionality
// Simply rebroadcasts the explosion to all other clients
func (s *Server) handleAddExplosion(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.AddExplosionPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal AddExplosion packet: %w", err)
	}

	s.logger.Debug().
		Str("player", player.Name).
		Uint8("explosionType", packet.Type).
		Float32("x", packet.Position.X).
		Float32("y", packet.Position.Y).
		Float32("z", packet.Position.Z).
		Int32("time", packet.Time).
		Bool("usesSound", packet.UsesSound).
		Float32("cameraShake", packet.CameraShake).
		Bool("isVisible", packet.IsVisible).
		Msg("Player created explosion")

	// Broadcast to all other clients (reliable packet)
	// This matches the C++ logic: SendPacketToAll with ENET_PACKET_FLAG_RELIABLE
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal AddExplosion packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.ADD_EXPLOSION,
		Data: packetData,
		Flag: network.PacketFlagReliable, // Reliable for important explosion events
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast explosion: %w", err)
	}

	return nil
}

// handleStartCutscene handles START_CUTSCENE packets
// This matches the C++ CPlayerPackets::StartCutscene::Handle() functionality
// Only the host player can start cutscenes, and it gets broadcast to all other clients
func (s *Server) handleStartCutscene(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Check if player is host (matching C++ logic: if (player->m_bIsHost))
	if !player.IsHost {
		s.logger.Warn().
			Str("player", player.Name).
			Str("client", client.GetClientID()).
			Msg("Non-host player attempted to start cutscene")
		return nil // Ignore non-host cutscene requests
	}

	// Parse the packet
	var packet packets.StartCutscenePacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal StartCutscene packet: %w", err)
	}

	s.logger.Info().
		Str("player", player.Name).
		Str("cutsceneName", packet.GetCutsceneName()).
		Uint8("currArea", packet.CurrArea).
		Msg("Host starting cutscene")

	// Broadcast to all other clients (reliable packet, matching C++ logic)
	// This matches C++ logic: CNetwork::SendPacketToAll with ENET_PACKET_FLAG_RELIABLE
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal start cutscene packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.START_CUTSCENE,
		Data: packetData,
		Flag: 1, // Reliable for cutscene synchronization (matching C++ implementation)
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast start cutscene: %w", err)
	}

	return nil
}

// handleSkipCutscene handles SKIP_CUTSCENE packets
// This matches the C++ CPlayerPackets::SkipCutscene::Handle() functionality
// Any player can request to skip a cutscene, and it gets broadcast to all other clients
func (s *Server) handleSkipCutscene(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.SkipCutscenePacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal SkipCutscene packet: %w", err)
	}

	// Set the correct player ID (security measure - don't trust client)
	packet.PlayerID = int32(player.ID)

	s.logger.Info().
		Str("player", player.Name).
		Int32("playerID", packet.PlayerID).
		Int32("votes", packet.Votes).
		Msg("Player requesting cutscene skip")

	// Broadcast to all other clients (reliable packet, matching C++ logic)
	// This matches C++ logic: CNetwork::SendPacketToAll with ENET_PACKET_FLAG_RELIABLE
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal skip cutscene packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.SKIP_CUTSCENE,
		Data: packetData,
		Flag: 1, // Reliable for cutscene synchronization (matching C++ implementation)
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast skip cutscene: %w", err)
	}

	return nil
}

// handleOnMissionFlagSync handles ON_MISSION_FLAG_SYNC packets
// This matches the C++ CPlayerPackets::OnMissionFlagSync::Handle() functionality
// Only the host player can sync mission flags, and it gets broadcast to all other clients
func (s *Server) handleOnMissionFlagSync(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Check if player is host (matching C++ logic: if (player->m_bIsHost))
	if !player.IsHost {
		s.logger.Warn().
			Str("player", player.Name).
			Str("client", client.GetClientID()).
			Msg("Non-host player attempted to sync mission flag")
		return nil // Ignore non-host mission flag sync requests
	}

	// Parse the packet
	var packet packets.OnMissionFlagSyncPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal OnMissionFlagSync packet: %w", err)
	}

	s.logger.Info().
		Str("player", player.Name).
		Bool("onMission", packet.IsOnMission()).
		Msg("Host syncing mission flag")

	// Broadcast to all other clients (reliable packet, matching C++ logic)
	// This matches C++ logic: CNetwork::SendPacketToAll with ENET_PACKET_FLAG_RELIABLE
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal mission flag sync packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.ON_MISSION_FLAG_SYNC,
		Data: packetData,
		Flag: 1, // Reliable for mission synchronization (matching C++ implementation)
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast mission flag sync: %w", err)
	}

	return nil
}

// handleUpdateEntityBlip handles UPDATE_ENTITY_BLIP packets
// This matches the C++ CPlayerPackets::UpdateEntityBlip::Handle() functionality
// Only the host player can update entity blips, and it gets sent to the specific target player
func (s *Server) handleUpdateEntityBlip(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Check if player is host (matching C++ logic: if (player->m_bIsHost))
	if !player.IsHost {
		s.logger.Warn().
			Str("player", player.Name).
			Str("client", client.GetClientID()).
			Msg("Non-host player attempted to update entity blip")
		return nil // Ignore non-host entity blip updates
	}

	// Parse the packet
	var packet packets.UpdateEntityBlipPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal UpdateEntityBlip packet: %w", err)
	}

	s.logger.Info().
		Str("player", player.Name).
		Int32("targetPlayerID", packet.PlayerID).
		Uint8("entityType", uint8(packet.EntityType)).
		Int32("entityID", packet.EntityID).
		Bool("isFriendly", packet.IsFriendly).
		Uint8("color", packet.Color).
		Msg("Host updating entity blip")

	// Find the target player and get their client (matching C++ logic: if (auto targetPlayer = CPlayerManager::GetPlayer(packet->playerid)))
	targetClient := s.getClientByPlayerID(types.PlayerID(packet.PlayerID))
	if targetClient == nil {
		s.logger.Warn().
			Int32("targetPlayerID", packet.PlayerID).
			Msg("Target player or client not found for entity blip update")
		return nil // Target player doesn't exist or client not found
	}

	// Send packet to target client (reliable packet, matching C++ logic)
	// This matches C++ logic: CNetwork::SendPacket(targetPlayer->m_pPeer, ...)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal entity blip update packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.UPDATE_ENTITY_BLIP,
		Data: packetData,
		Flag: 1, // Reliable for blip synchronization (matching C++ implementation)
	}

	if err := s.networkServer.SendPacket(targetClient, networkPacket); err != nil {
		return fmt.Errorf("failed to send entity blip update to target player: %w", err)
	}

	return nil
}

// handleRemoveEntityBlip handles REMOVE_ENTITY_BLIP packets
// This matches the C++ CPlayerPackets::RemoveEntityBlip::Handle() functionality
// Only the host player can remove entity blips, and it gets sent to the specific target player
func (s *Server) handleRemoveEntityBlip(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Check if player is host (matching C++ logic: if (player->m_bIsHost))
	if !player.IsHost {
		s.logger.Warn().
			Str("player", player.Name).
			Str("client", client.GetClientID()).
			Msg("Non-host player attempted to remove entity blip")
		return nil // Ignore non-host entity blip removals
	}

	// Parse the packet
	var packet packets.RemoveEntityBlipPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal RemoveEntityBlip packet: %w", err)
	}

	s.logger.Info().
		Str("player", player.Name).
		Int32("targetPlayerID", packet.PlayerID).
		Uint8("entityType", uint8(packet.EntityType)).
		Int32("entityID", packet.EntityID).
		Msg("Host removing entity blip")

	// Find the target player and get their client (matching C++ logic: if (auto targetPlayer = CPlayerManager::GetPlayer(packet->playerid)))
	targetClient := s.getClientByPlayerID(types.PlayerID(packet.PlayerID))
	if targetClient == nil {
		s.logger.Warn().
			Int32("targetPlayerID", packet.PlayerID).
			Msg("Target player or client not found for entity blip removal")
		return nil // Target player doesn't exist or client not found
	}

	// Send packet to target client (reliable packet, matching C++ logic)
	// This matches C++ logic: CNetwork::SendPacket(targetPlayer->m_pPeer, ...)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal entity blip removal packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.REMOVE_ENTITY_BLIP,
		Data: packetData,
		Flag: 1, // Reliable for blip synchronization (matching C++ implementation)
	}

	if err := s.networkServer.SendPacket(targetClient, networkPacket); err != nil {
		return fmt.Errorf("failed to send entity blip removal to target player: %w", err)
	}

	return nil
}

// handleAddMessageGXT handles ADD_MESSAGE_GXT packets
// This matches the C++ CPlayerPackets::AddMessageGXT::Handle() functionality
// Only the host player can add GXT messages, and it gets sent to the specific target player
func (s *Server) handleAddMessageGXT(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Check if player is host (matching C++ logic: if (player->m_bIsHost))
	if !player.IsHost {
		s.logger.Warn().
			Str("player", player.Name).
			Str("client", client.GetClientID()).
			Msg("Non-host player attempted to add GXT message")
		return nil // Ignore non-host GXT message additions
	}

	// Parse the packet
	var packet packets.AddMessageGXTPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal AddMessageGXT packet: %w", err)
	}

	s.logger.Info().
		Str("player", player.Name).
		Int32("targetPlayerID", packet.PlayerID).
		Uint8("type", packet.Type).
		Str("gxt", packet.GetGXTString()).
		Uint32("time", packet.Time).
		Uint8("flag", packet.Flag).
		Msg("Host adding GXT message")

	// Find the target player and get their client (matching C++ logic: if (auto targetPlayer = CPlayerManager::GetPlayer(packet->playerid)))
	targetClient := s.getClientByPlayerID(types.PlayerID(packet.PlayerID))
	if targetClient == nil {
		s.logger.Warn().
			Int32("targetPlayerID", packet.PlayerID).
			Msg("Target player or client not found for GXT message addition")
		return nil // Target player doesn't exist or client not found
	}

	// Send packet to target client (reliable packet, matching C++ logic)
	// This matches C++ logic: CNetwork::SendPacket(targetPlayer->m_pPeer, ...)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal GXT message addition packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.ADD_MESSAGE_GXT,
		Data: packetData,
		Flag: 1, // Reliable for message synchronization (matching C++ implementation)
	}

	if err := s.networkServer.SendPacket(targetClient, networkPacket); err != nil {
		return fmt.Errorf("failed to send GXT message addition to target player: %w", err)
	}

	return nil
}

// handleRemoveMessageGXT handles REMOVE_MESSAGE_GXT packets
// This matches the C++ CPlayerPackets::RemoveMessageGXT::Handle() functionality
// Only the host player can remove GXT messages, and it gets sent to the specific target player
func (s *Server) handleRemoveMessageGXT(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Check if player is host (matching C++ logic: if (player->m_bIsHost))
	if !player.IsHost {
		s.logger.Warn().
			Str("player", player.Name).
			Str("client", client.GetClientID()).
			Msg("Non-host player attempted to remove GXT message")
		return nil // Ignore non-host GXT message removals
	}

	// Parse the packet
	var packet packets.RemoveMessageGXTPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal RemoveMessageGXT packet: %w", err)
	}

	s.logger.Info().
		Str("player", player.Name).
		Int32("targetPlayerID", packet.PlayerID).
		Str("gxt", packet.GetGXTString()).
		Msg("Host removing GXT message")

	// Find the target player and get their client (matching C++ logic: if (auto targetPlayer = CPlayerManager::GetPlayer(packet->playerid)))
	targetClient := s.getClientByPlayerID(types.PlayerID(packet.PlayerID))
	if targetClient == nil {
		s.logger.Warn().
			Int32("targetPlayerID", packet.PlayerID).
			Msg("Target player or client not found for GXT message removal")
		return nil // Target player doesn't exist or client not found
	}

	// Send packet to target client (reliable packet, matching C++ logic)
	// This matches C++ logic: CNetwork::SendPacket(targetPlayer->m_pPeer, ...)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal GXT message removal packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.REMOVE_MESSAGE_GXT,
		Data: packetData,
		Flag: 1, // Reliable for message synchronization (matching C++ implementation)
	}

	if err := s.networkServer.SendPacket(targetClient, networkPacket); err != nil {
		return fmt.Errorf("failed to send GXT message removal to target player: %w", err)
	}

	return nil
}

// handleClearEntityBlips handles CLEAR_ENTITY_BLIPS packets
// This matches the C++ CPlayerPackets::ClearEntityBlips::Handle() functionality
// Only the host player can clear entity blips, and it gets sent to the specific target player
func (s *Server) handleClearEntityBlips(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Check if player is host (matching C++ logic: if (player->m_bIsHost))
	if !player.IsHost {
		s.logger.Warn().
			Str("player", player.Name).
			Str("client", client.GetClientID()).
			Msg("Non-host player attempted to clear entity blips")
		return nil // Ignore non-host entity blip clearing
	}

	// Parse the packet
	var packet packets.ClearEntityBlipsPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal ClearEntityBlips packet: %w", err)
	}

	s.logger.Info().
		Str("player", player.Name).
		Int32("targetPlayerID", packet.PlayerID).
		Msg("Host clearing entity blips")

	// Find the target player and get their client (matching C++ logic: if (auto targetPlayer = CPlayerManager::GetPlayer(packet->playerid)))
	targetClient := s.getClientByPlayerID(types.PlayerID(packet.PlayerID))
	if targetClient == nil {
		s.logger.Warn().
			Int32("targetPlayerID", packet.PlayerID).
			Msg("Target player or client not found for entity blip clearing")
		return nil // Target player doesn't exist or client not found
	}

	// Send packet to target client (reliable packet, matching C++ logic)
	// This matches C++ logic: CNetwork::SendPacket(targetPlayer->m_pPeer, ...)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal entity blip clearing packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.CLEAR_ENTITY_BLIPS,
		Data: packetData,
		Flag: 1, // Reliable for blip synchronization (matching C++ implementation)
	}

	if err := s.networkServer.SendPacket(targetClient, networkPacket); err != nil {
		return fmt.Errorf("failed to send entity blip clearing to target player: %w", err)
	}

	return nil
}
