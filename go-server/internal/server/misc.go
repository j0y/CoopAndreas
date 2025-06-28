package server

import (
	"fmt"

	"coopandreas-server/internal/network"
	"coopandreas-server/internal/packets"
	"coopandreas-server/internal/types"
)

// handleMassPacketSequence handles MASS_PACKET_SEQUENCE packets
// Based on C++ server logic: simply rebroadcast the entire mass packet to all other clients
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
