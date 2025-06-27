package server

import (
	"fmt"
	"strings"

	"coopandreas-server/internal/entities"
	"coopandreas-server/internal/network"
	"coopandreas-server/internal/packets"
	"coopandreas-server/internal/types"
)

// handlePlayerGetName handles PLAYER_GET_NAME packets
func (s *Server) handlePlayerGetName(client *network.Client, data []byte) error {
	// Parse the player get name packet
	var packet packets.PlayerGetNamePacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal PlayerGetName packet: %w", err)
	}

	// Get the player name as a string
	playerName := packet.GetNameString()

	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		s.logger.Warn().
			Str("client", client.Addr.String()).
			Str("name", playerName).
			Msg("Received name from unknown player")
		return fmt.Errorf("no player found for client %s", client.Addr)
	}

	// Update player name
	oldName := player.Name
	player.Name = playerName

	s.logger.Info().
		Str("client", client.Addr.String()).
		Str("oldName", oldName).
		Str("newName", playerName).
		Int32("playerID", int32(player.ID)).
		Msg("Player name updated")

	// If this player wasn't already known to others (new connection), send chat message
	// In C++ this checks m_bHasBeenConnectedBeforeMe, but for simplicity we'll log it differently
	if oldName != playerName && strings.HasPrefix(oldName, "Player_") {
		s.logger.Info().
			Str("name", playerName).
			Int32("playerID", int32(player.ID)).
			Msg("Player introduced themselves")
	}

	// TODO: Trigger GameWeatherTime like in C++ (CPacketHandler::GameWeatherTime__Trigger)
	// This would send current weather/time state to the newly named player

	return nil
}

// handlePlayerOnFoot handles PLAYER_ONFOOT packets
func (s *Server) handlePlayerOnFoot(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.Addr)
	}

	// Parse the packet
	var packet packets.PlayerOnFootPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal PlayerOnFoot packet: %w", err)
	}

	// Set the player ID (clients send 0, server assigns the actual ID)
	packet.ID = player.ID

	// Validate weapon (matching C++ logic: 0-18 or 22-46 are valid)
	isValidWeapon := packet.Weapon <= 18 || (packet.Weapon >= 22 && packet.Weapon <= 46)
	if !isValidWeapon {
		packet.Weapon = 0
		packet.Ammo = 0
	}

	// Validate fighting style (matching C++ logic: 4-16 are valid)
	if packet.FightingStyle < 4 || packet.FightingStyle > 16 {
		packet.FightingStyle = 4
	}

	// Validate velocity to prevent speed hacking (matching C++ logic: max 10.0 per axis)
	if packet.Velocity.X > 10.0 || packet.Velocity.Y > 10.0 || packet.Velocity.Z > 10.0 {
		packet.Velocity = types.Vector3{X: 0.0, Y: 0.0, Z: 0.0}
	}

	// TODO: Remove player from vehicle if they're in one (when vehicle system is implemented)
	// In C++: if (player->m_nVehicleId >= 0) { player->RemoveFromVehicle(); }

	// Broadcast to all other clients (unreliable packet, high frequency)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal player onfoot packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.PLAYER_ONFOOT,
		Data: packetData,
		Flag: 0, // Unreliable for frequent position updates (matching C++ implementation)
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast player onfoot: %w", err)
	}

	s.logger.Debug().
		Str("player", player.Name).
		Float32("x", packet.Position.X).
		Float32("y", packet.Position.Y).
		Float32("z", packet.Position.Z).
		Uint8("health", packet.Health).
		Uint8("weapon", packet.Weapon).
		Msg("Player onfoot update")

	return nil
}

// handlePlayerKeySync handles PLAYER_KEY_SYNC packets
func (s *Server) handlePlayerKeySync(client *network.Client, data []byte) error {
	// Debug packet size for troubleshooting
	s.logger.Debug().
		Int("dataSize", len(data)).
		Str("dataHex", fmt.Sprintf("%x", data)).
		Msg("Received PLAYER_KEY_SYNC packet")

	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.Addr)
	}

	// Parse the packet
	var packet packets.PlayerKeySyncPacket
	if err := packet.Unmarshal(data); err != nil {
		s.logger.Error().
			Err(err).
			Int("dataSize", len(data)).
			Str("dataHex", fmt.Sprintf("%x", data)).
			Str("client", client.Addr.String()).
			Msg("Failed to unmarshal PLAYER_KEY_SYNC packet")
		return fmt.Errorf("failed to unmarshal PlayerKeySync packet: %w", err)
	}

	// Set the player ID (clients send 0, server assigns the actual ID)
	// This matches the C++ logic: packet->playerid = CPlayerManager::GetPlayer(peer)->m_iPlayerId;
	packet.ID = player.ID

	// Broadcast to all other clients (unreliable packet, high frequency like onfoot)
	// This matches the C++ logic: CNetwork::SendPacketToAll(CPacketsID::PLAYER_KEY_SYNC, packet, sizeof * packet, (ENetPacketFlag)0, peer);
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal player key sync packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.PLAYER_KEY_SYNC,
		Data: packetData,
		Flag: 0, // Unreliable for frequent key updates (matching C++ implementation)
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast player key sync: %w", err)
	}

	s.logger.Debug().
		Str("player", player.Name).
		Uint32("compressed", packet.NewState.Compressed).
		Int16("leftX", packet.NewState.LeftStickX).
		Int16("leftY", packet.NewState.LeftStickY).
		Msg("Player key sync update")

	return nil
}

// generatePlayerID generates a unique player ID (similar to C++ GetFreeID())
func (s *Server) generatePlayerID() types.PlayerID {
	allPlayers := s.playerManager.GetAllPlayers()
	return types.PlayerID(len(allPlayers) + 1)
}

// broadcastPlayerConnectedPacket sends a PlayerConnected packet to all players except the specified client
func (s *Server) broadcastPlayerConnectedPacket(packet *packets.PlayerConnectedPacket, excludeClient *network.Client) {
	packetData, err := packet.Marshal()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to marshal player connected packet")
		return
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.PLAYER_CONNECTED,
		Data: packetData,
		Flag: network.PacketFlagReliable,
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, excludeClient); err != nil {
		s.logger.Error().Err(err).Msg("Failed to broadcast player connected packet")
	}

	s.logger.Info().
		Int32("playerID", int32(packet.ID)).
		Bool("isAlreadyConnected", packet.IsAlreadyConnected == 1).
		Msg("Broadcasted PlayerConnected packet to existing players")
}

// sendExistingPlayersTo sends information about all existing players to a new player
func (s *Server) sendExistingPlayersTo(client *network.Client, newPlayer *entities.Player) {
	allPlayers := s.playerManager.GetAllPlayers()

	for _, existingPlayer := range allPlayers {
		if existingPlayer.ID == newPlayer.ID {
			continue // Skip the new player itself
		}

		// Send PlayerConnected packet with isAlreadyConnected = true
		playerPacket := packets.NewPlayerConnectedPacket(existingPlayer.ID, true)
		packetData, err := playerPacket.Marshal()
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to marshal existing player packet")
			continue
		}

		networkPacket := &network.NetworkPacket{
			ID:   types.PLAYER_CONNECTED,
			Data: packetData,
			Flag: network.PacketFlagReliable,
		}

		if err := s.networkServer.SendPacket(client, networkPacket); err != nil {
			s.logger.Error().Err(err).Msg("Failed to send existing player packet")
		}

		s.logger.Debug().
			Int32("existingPlayerID", int32(existingPlayer.ID)).
			Str("newPlayer", newPlayer.Name).
			Msg("Sent existing player info to new player")
	}
}

// sendHandshakeTo sends a handshake packet to a new player
func (s *Server) sendHandshakeTo(client *network.Client, handshakePacket *packets.PlayerHandshakePacket) {
	packetData, err := handshakePacket.Marshal()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to marshal handshake packet")
		return
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.PLAYER_HANDSHAKE,
		Data: packetData,
		Flag: network.PacketFlagReliable,
	}

	if err := s.networkServer.SendPacket(client, networkPacket); err != nil {
		s.logger.Error().Err(err).Msg("Failed to send handshake packet")
		return
	}

	s.logger.Info().
		Int32("playerID", int32(handshakePacket.YourID)).
		Str("client", client.Addr.String()).
		Msg("Sent handshake packet to new player")
}
