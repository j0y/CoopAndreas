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
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		s.logger.Warn().
			Uint32("clientID", client.ID).
			Str("name", playerName).
			Msg("Received name from unknown player")
		return fmt.Errorf("no player found for client %s", client.Addr)
	}

	// Update player name
	oldName := player.Name
	player.Name = playerName

	s.logger.Info().
		Uint32("clientID", client.ID).
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

	// Note: In C++ client code, PlayerGetName__Handle calls GameWeatherTime__Trigger()
	// However, this triggers the CLIENT to send weather data to the SERVER (if client is host)
	// The server does NOT send weather data to clients during name updates
	// Weather synchronization happens during connection, not during name updates

	return nil
}

// handlePlayerOnFoot handles PLAYER_ONFOOT packets
func (s *Server) handlePlayerOnFoot(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
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

	//s.logger.Debug().
	//	Str("player", player.Name).
	//	Float32("x", packet.Position.X).
	//	Float32("y", packet.Position.Y).
	//	Float32("z", packet.Position.Z).
	//	Uint8("health", packet.Health).
	//	Uint8("weapon", packet.Weapon).
	//	Msg("Player onfoot update")

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
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.PlayerKeySyncPacket
	if err := packet.Unmarshal(data); err != nil {
		s.logger.Error().
			Err(err).
			Int("dataSize", len(data)).
			Str("dataHex", fmt.Sprintf("%x", data)).
			Uint32("clientID", client.ID).
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

		// Send PLAYER_GET_NAME packet for existing player (matching C++ logic)
		// This tells the new player what the existing player's name is
		getNamePacket := packets.PlayerGetNamePacket{
			PlayerID: existingPlayer.ID,
		}
		getNamePacket.SetNameString(existingPlayer.Name)

		getNameData, err := getNamePacket.Marshal()
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to marshal player get name packet")
		} else {
			getNameNetworkPacket := &network.NetworkPacket{
				ID:   types.PLAYER_GET_NAME,
				Data: getNameData,
				Flag: network.PacketFlagReliable,
			}

			if err := s.networkServer.SendPacket(client, getNameNetworkPacket); err != nil {
				s.logger.Error().Err(err).Msg("Failed to send player get name packet")
			} else {
				s.logger.Debug().
					Int32("existingPlayerID", int32(existingPlayer.ID)).
					Str("existingPlayerName", existingPlayer.Name).
					Str("newPlayer", newPlayer.Name).
					Msg("Sent existing player name to new player")
			}
		}

		// Send player stats if they have been modified (matching C++ logic)
		if existingPlayer.StatsModified {
			statsPacket := packets.NewPlayerStatsPacket(existingPlayer.ID, existingPlayer.Stats)
			statsData, err := statsPacket.Marshal()
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to marshal player stats packet")
				continue
			}

			statsNetworkPacket := &network.NetworkPacket{
				ID:   types.PLAYER_STATS,
				Data: statsData,
				Flag: network.PacketFlagReliable,
			}

			if err := s.networkServer.SendPacket(client, statsNetworkPacket); err != nil {
				s.logger.Error().Err(err).Msg("Failed to send player stats packet")
			} else {
				s.logger.Debug().
					Int32("existingPlayerID", int32(existingPlayer.ID)).
					Str("newPlayer", newPlayer.Name).
					Msg("Sent existing player stats to new player")
			}
		}

		// Send player clothes/appearance if they have been modified (matching C++ logic)
		if existingPlayer.ClothesModified {
			rebuildPacket := packets.NewRebuildPlayerPacket(
				existingPlayer.ID,
				existingPlayer.ModelKeys,
				existingPlayer.TextureKeys,
				existingPlayer.FatStat,
				existingPlayer.MuscleStat,
			)
			rebuildData, err := rebuildPacket.Marshal()
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to marshal rebuild player packet")
				continue
			}

			rebuildNetworkPacket := &network.NetworkPacket{
				ID:   types.REBUILD_PLAYER,
				Data: rebuildData,
				Flag: network.PacketFlagReliable,
			}

			if err := s.networkServer.SendPacket(client, rebuildNetworkPacket); err != nil {
				s.logger.Error().Err(err).Msg("Failed to send rebuild player packet")
			} else {
				s.logger.Debug().
					Int32("existingPlayerID", int32(existingPlayer.ID)).
					Str("newPlayer", newPlayer.Name).
					Msg("Sent existing player clothes to new player")
			}
		}

		// Send player waypoint if they have one (matching C++ logic)
		if existingPlayer.WaypointModified {
			waypointPacket := packets.NewPlayerPlaceWaypointPacket(
				existingPlayer.ID,
				true, // place = true since waypoint exists
				existingPlayer.WaypointPosition,
			)
			waypointData, err := waypointPacket.Marshal()
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to marshal waypoint packet")
				continue
			}

			waypointNetworkPacket := &network.NetworkPacket{
				ID:   types.PLAYER_PLACE_WAYPOINT,
				Data: waypointData,
				Flag: network.PacketFlagReliable,
			}

			if err := s.networkServer.SendPacket(client, waypointNetworkPacket); err != nil {
				s.logger.Error().Err(err).Msg("Failed to send waypoint packet")
			} else {
				s.logger.Debug().
					Int32("existingPlayerID", int32(existingPlayer.ID)).
					Str("newPlayer", newPlayer.Name).
					Msg("Sent existing player waypoint to new player")
			}
		}
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
		Uint32("clientID", client.ID).
		Msg("Sent handshake packet to new player")
}

// handlePlayerSetHost handles PLAYER_SET_HOST packets
// Note: This packet is typically sent FROM server TO clients, not the other way around
// But we implement the handler for completeness and potential client-to-server scenarios
func (s *Server) handlePlayerSetHost(client *network.Client, data []byte) error {
	// Parse the packet
	var packet packets.PlayerSetHostPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal PlayerSetHost packet: %w", err)
	}

	s.logger.Info().
		Int32("hostPlayerID", int32(packet.ID)).
		Uint32("clientID", client.ID).
		Msg("Received PlayerSetHost packet (unusual - typically server-to-client)")

	// Note: In the C++ implementation, this packet is typically sent FROM server TO clients
	// If we receive it from a client, we could either ignore it or treat it as a request
	// For now, we'll log it but not change host status based on client requests

	return nil
}

// assignHostToFirstPlayer assigns host status to the first player and notifies all clients
// This matches the C++ CPlayerManager::AssignHostToFirstPlayer() functionality
func (s *Server) assignHostToFirstPlayer() error {
	newHost := s.playerManager.AssignHostToFirstPlayer()
	if newHost == nil {
		s.logger.Debug().Msg("No players to assign as host")
		return nil
	}

	s.logger.Info().
		Int32("hostPlayerID", int32(newHost.ID)).
		Str("hostName", newHost.Name).
		Msg("Assigned host to first player")

	// Send PlayerSetHost packet to all clients
	packet := packets.NewPlayerSetHostPacket(newHost.ID)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal PlayerSetHost packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.PLAYER_SET_HOST,
		Data: packetData,
		Flag: network.PacketFlagReliable, // Reliable for important host changes
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, nil); err != nil {
		return fmt.Errorf("failed to broadcast PlayerSetHost: %w", err)
	}

	// When a new host is assigned, send current weather/time to all clients
	// This ensures all clients have synchronized weather after a host change
	if s.currentWeatherTime != nil {
		if err := s.broadcastCurrentWeatherTime(); err != nil {
			s.logger.Error().Err(err).Msg("Failed to broadcast weather/time after host assignment")
		}
	}

	return nil
}

// handleRespawnPlayer handles RESPAWN_PLAYER packets
// This matches the C++ CPlayerPackets::RespawnPlayer::Handle() functionality
func (s *Server) handleRespawnPlayer(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet (though it only contains player ID that we'll override)
	var packet packets.RespawnPlayerPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal RespawnPlayer packet: %w", err)
	}

	s.logger.Info().
		Int32("playerID", int32(player.ID)).
		Str("playerName", player.Name).
		Msg("Player respawned")

	// Set the correct player ID (matching C++ logic: packet->playerid = CPlayerManager::GetPlayer(peer)->m_iPlayerId;)
	packet.PlayerID = player.ID

	// Marshal the updated packet
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal RespawnPlayer packet: %w", err)
	}

	// Broadcast to all clients (matching C++ logic: SendPacketToAll with ENET_PACKET_FLAG_RELIABLE)
	networkPacket := &network.NetworkPacket{
		ID:   types.RESPAWN_PLAYER,
		Data: packetData,
		Flag: network.PacketFlagReliable, // Reliable for important respawn notifications
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, nil); err != nil {
		return fmt.Errorf("failed to broadcast RespawnPlayer: %w", err)
	}

	return nil
}

// handlePlayerStats handles PLAYER_STATS packets
// This matches the C++ CPlayerPackets::PlayerStats::Handle() functionality
func (s *Server) handlePlayerStats(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.PlayerStatsPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal PlayerStats packet: %w", err)
	}

	// Set the player ID (security measure - don't trust client, matching C++ logic)
	packet.PlayerID = player.ID

	s.logger.Debug().
		Int32("playerID", int32(packet.PlayerID)).
		Str("player", player.Name).
		Float32("stat0", packet.Stats[0]).
		Float32("stat1", packet.Stats[1]).
		Msg("Received player stats update")

	// Store the stats in the player entity (matching C++ logic)
	player.Stats = packet.Stats
	player.StatsModified = true

	// Broadcast the stats packet to all other clients (reliable packet)
	// This matches the C++ logic: CNetwork::SendPacketToAll(CPacketsID::PLAYER_STATS, packet, sizeof * packet, ENET_PACKET_FLAG_RELIABLE, peer);
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal player stats packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.PLAYER_STATS,
		Data: packetData,
		Flag: network.PacketFlagReliable, // Reliable for important player state changes
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast player stats: %w", err)
	}

	return nil
}

// handleRebuildPlayer handles REBUILD_PLAYER packets
// This matches the C++ CPlayerPackets::RebuildPlayer::Handle() functionality
func (s *Server) handleRebuildPlayer(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.RebuildPlayerPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal RebuildPlayer packet: %w", err)
	}

	// Set the player ID (security measure - don't trust client, matching C++ logic)
	packet.PlayerID = player.ID

	s.logger.Debug().
		Int32("playerID", int32(packet.PlayerID)).
		Str("player", player.Name).
		Float32("fatStat", packet.FatStat).
		Float32("muscleStat", packet.MuscleStat).
		Msg("Received player rebuild request")

	// Store the clothes/appearance data in the player entity (matching C++ logic)
	player.ModelKeys = packet.ModelKeys
	player.TextureKeys = packet.TextureKeys
	player.FatStat = packet.FatStat
	player.MuscleStat = packet.MuscleStat
	player.ClothesModified = true

	// Broadcast the rebuild packet to all other clients (reliable packet)
	// This matches the C++ logic: CNetwork::SendPacketToAll(CPacketsID::REBUILD_PLAYER, packet, sizeof * packet, ENET_PACKET_FLAG_RELIABLE, peer);
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal rebuild player packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.REBUILD_PLAYER,
		Data: packetData,
		Flag: network.PacketFlagReliable, // Reliable for important player appearance changes
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast rebuild player: %w", err)
	}

	return nil
}

// handlePlayerPlaceWaypoint handles PLAYER_PLACE_WAYPOINT packets
// This matches the C++ CPlayerPackets::PlayerPlaceWaypoint::Handle() functionality
func (s *Server) handlePlayerPlaceWaypoint(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.PlayerPlaceWaypointPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal PlayerPlaceWaypoint packet: %w", err)
	}

	// Set the player ID (security measure - don't trust client, matching C++ logic)
	packet.PlayerID = player.ID

	// Validate and clamp position coordinates (matching C++ logic: std::clamp(-3000.0f, 3000.0f))
	const maxCoord = 3000.0
	const minCoord = -3000.0

	if packet.Position.X > maxCoord {
		packet.Position.X = maxCoord
	} else if packet.Position.X < minCoord {
		packet.Position.X = minCoord
	}

	if packet.Position.Y > maxCoord {
		packet.Position.Y = maxCoord
	} else if packet.Position.Y < minCoord {
		packet.Position.Y = minCoord
	}

	s.logger.Debug().
		Int32("playerID", int32(packet.PlayerID)).
		Str("player", player.Name).
		Bool("place", packet.Place).
		Float32("x", packet.Position.X).
		Float32("y", packet.Position.Y).
		Float32("z", packet.Position.Z).
		Msg("Received player waypoint request")

	// Store the waypoint data in the player entity (matching C++ logic)
	player.WaypointModified = packet.Place
	player.WaypointPosition = packet.Position

	// Broadcast the waypoint packet to all other clients (reliable packet)
	// This matches the C++ logic: CNetwork::SendPacketToAll(CPacketsID::PLAYER_PLACE_WAYPOINT, packet, sizeof * packet, ENET_PACKET_FLAG_RELIABLE, peer);
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal player place waypoint packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.PLAYER_PLACE_WAYPOINT,
		Data: packetData,
		Flag: network.PacketFlagReliable, // Reliable for important waypoint changes
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast player place waypoint: %w", err)
	}

	return nil
}

// handlePlayerBulletShot handles PLAYER_BULLET_SHOT packets
// This matches the C++ CPlayerPackets::PlayerBulletShot::Handle() functionality
func (s *Server) handlePlayerBulletShot(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.PlayerBulletShotPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal PlayerBulletShot packet: %w", err)
	}

	// Set the player ID (security measure - don't trust client, matching C++ logic)
	packet.PlayerID = int32(player.ID)

	s.logger.Debug().
		Str("player", player.Name).
		Int32("targetID", packet.TargetID).
		Uint8("entityType", uint8(packet.EntityType)).
		Float32("startX", packet.StartPos.X).
		Float32("startY", packet.StartPos.Y).
		Float32("startZ", packet.StartPos.Z).
		Float32("endX", packet.EndPos.X).
		Float32("endY", packet.EndPos.Y).
		Float32("endZ", packet.EndPos.Z).
		Int32("incrementalHit", packet.IncrementalHit).
		Msg("Player bullet shot")

	// Broadcast to all other clients (unreliable packet, high frequency)
	// This matches the C++ logic: SendPacketToAll with (ENetPacketFlag)0 (unreliable)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal player bullet shot packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.PLAYER_BULLET_SHOT,
		Data: packetData,
		Flag: 0, // Unreliable for frequent bullet shot events (matching C++ implementation)
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast player bullet shot: %w", err)
	}

	return nil
}
