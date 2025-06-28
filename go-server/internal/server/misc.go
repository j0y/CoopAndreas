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
			Str("client", client.Addr.String()).
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
		Str("client", client.Addr.String()).
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
