package server

import (
	"fmt"

	"coopandreas-server/internal/entities"
	"coopandreas-server/internal/network"
	"coopandreas-server/internal/packets"
	"coopandreas-server/internal/types"
)

// HandlePlayerConnect handles when a player connects
func (s *Server) HandlePlayerConnect(client *network.Client) {
	// Generate a unique player ID (similar to C++ GetFreeID())
	playerID := s.generatePlayerID()

	// Create a new player with default name
	player := &entities.Player{
		ID:       playerID,
		Name:     fmt.Sprintf("Player_%d", int(playerID)),
		PeerAddr: client.Addr.String(),
		IsHost:   false,
	}

	s.playerManager.AddPlayer(client.Addr.String(), player)
	s.logger.Info().
		Str("player", player.Name).
		Str("address", client.Addr.String()).
		Int32("playerID", int32(playerID)).
		Msg("Player connected")

	// Send PlayerConnected packet to all existing players (excluding new player)
	playerConnectedPacket := packets.NewPlayerConnectedPacket(player.ID, false)
	s.broadcastPlayerConnectedPacket(playerConnectedPacket, client)

	// Send existing players to the new player (with isAlreadyConnected = true)
	s.sendExistingPlayersTo(client, player)

	// Send existing vehicles to the new player
	s.sendExistingVehiclesTo(client)

	// Send existing peds to the new player
	s.sendExistingPedsTo(client)

	// Send handshake packet to the new player
	handshakePacket := packets.NewPlayerHandshakePacket(player.ID)
	s.sendHandshakeTo(client, handshakePacket)

	// Assign host if no current host exists (first player becomes host)
	if s.playerManager.GetHost() == nil {
		if err := s.assignHostToFirstPlayer(); err != nil {
			s.logger.Error().Err(err).Msg("Failed to assign host to first player")
		}
	}
}

// HandlePlayerDisconnect handles when a player disconnects
// This method:
// 1. Removes all peds owned by the disconnecting player
// 2. Broadcasts PED_REMOVE packets for cleanup
// 3. Removes the player from the player manager
// 4. Broadcasts PLAYER_DISCONNECTED packet to all remaining clients (matching C++ behavior)
func (s *Server) HandlePlayerDisconnect(client *network.Client) {
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return
	}

	// Store player info for broadcasting and host checking
	disconnectedPlayerID := player.ID
	wasHost := player.IsHost

	// Remove all peds owned by this player
	removedPeds := s.pedManager.RemoveAllHostedBy(player)

	// Broadcast ped removals
	for _, pedID := range removedPeds {
		removePacket := packets.PedRemovePacket{
			PedID: int32(pedID),
		}

		if data, err := removePacket.Marshal(); err == nil {
			networkPacket := &network.NetworkPacket{
				ID:   types.PED_REMOVE,
				Data: data,
				Flag: network.PacketFlagReliable,
			}
			s.networkServer.SendPacketToAll(networkPacket, nil)
		}
	}

	// Remove all vehicles owned by this player
	removedVehicles := s.vehicleManager.RemoveVehiclesBySyncer(player)

	// Broadcast vehicle removals
	for _, vehicle := range removedVehicles {
		removePacket := packets.NewVehicleRemovePacket(int32(vehicle.ID))

		if data, err := removePacket.Marshal(); err == nil {
			networkPacket := &network.NetworkPacket{
				ID:   types.VEHICLE_REMOVE,
				Data: data,
				Flag: network.PacketFlagReliable,
			}
			s.networkServer.SendPacketToAll(networkPacket, nil)
		}
	}

	// Remove player from manager
	s.playerManager.RemovePlayer(client.Addr.String())

	// Create and broadcast PLAYER_DISCONNECTED packet to all remaining clients
	disconnectPacket := packets.NewPlayerDisconnectedPacket(disconnectedPlayerID, 0) // Normal disconnect (no specific reason code)

	if data, err := disconnectPacket.Marshal(); err == nil {
		networkPacket := &network.NetworkPacket{
			ID:   types.PLAYER_DISCONNECTED,
			Data: data,
			Flag: 0, // Use unreliable for disconnect notifications (matching C++ implementation)
		}
		s.networkServer.SendPacketToAll(networkPacket, nil) // Send to all clients (not excluding anyone since the client is already disconnected)
	} else {
		s.logger.Error().Err(err).Msg("Failed to marshal player disconnect packet")
	}

	s.logger.Info().
		Str("player", player.Name).
		Int32("playerID", int32(disconnectedPlayerID)).
		Int("pedsRemoved", len(removedPeds)).
		Int("vehiclesRemoved", len(removedVehicles)).
		Bool("wasHost", wasHost).
		Msg("Player disconnected")

	// If the disconnected player was the host, assign host to another player
	if wasHost {
		if err := s.assignHostToFirstPlayer(); err != nil {
			s.logger.Error().Err(err).Msg("Failed to reassign host after host disconnection")
		}
	}
}
