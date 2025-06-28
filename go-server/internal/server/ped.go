package server

import (
	"fmt"

	"coopandreas-server/internal/entities"
	"coopandreas-server/internal/network"
	"coopandreas-server/internal/packets"
	"coopandreas-server/internal/types"
)

// handlePedSpawn handles PED_SPAWN packets
func (s *Server) handlePedSpawn(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.PedSpawnPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal PedSpawn packet: %w", err)
	}

	// Validate the spawn request (matching C++ logic)
	if err := s.pedManager.ValidatePedSpawn(packet.ModelID, packet.SpecialModelName); err != nil {
		s.logger.Warn().
			Str("player", player.Name).
			Err(err).
			Msg("Player tried to spawn invalid ped")
		return nil // Don't return error, just ignore invalid request
	}

	// Assign new ped ID
	packet.PedID = int32(s.pedManager.GetFreeID())

	// Create ped entity
	ped := &entities.Ped{
		ID:               types.PedID(packet.PedID),
		Syncer:           player,
		ModelID:          packet.ModelID,
		PedType:          packet.PedType,
		Position:         packet.Position,
		CreatedBy:        packet.CreatedBy,
		SpecialModelName: packet.SpecialModelName,
	}

	// Add to ped manager
	s.pedManager.Add(ped)

	// Broadcast to all clients
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal ped spawn packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.PED_SPAWN,
		Data: packetData,
		Flag: network.PacketFlagReliable,
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast ped spawn: %w", err)
	}

	// Send confirmation back to spawner (matching C++ logic)
	confirmPacket := packets.NewPedConfirmPacket(packet.TempID, packet.PedID)

	confirmData, err := confirmPacket.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal ped confirm packet: %w", err)
	}

	confirmNetworkPacket := &network.NetworkPacket{
		ID:   types.PED_CONFIRM,
		Data: confirmData,
		Flag: network.PacketFlagReliable,
	}

	if err := s.networkServer.SendPacket(client, confirmNetworkPacket); err != nil {
		return fmt.Errorf("failed to send ped confirm: %w", err)
	}

	s.logger.Info().
		Str("player", player.Name).
		Int32("pedID", packet.PedID).
		Int16("modelID", packet.ModelID).
		Msg("Player spawned ped")

	return nil
}

// handlePedRemove handles PED_REMOVE packets
func (s *Server) handlePedRemove(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.PedRemovePacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal PedRemove packet: %w", err)
	}

	// Get the ped
	ped := s.pedManager.GetPed(types.PedID(packet.PedID))
	if ped == nil {
		return nil // Ped doesn't exist, ignore
	}

	// Validate ownership (anti-cheat check)
	if ped.Syncer != player {
		s.logger.Warn().
			Str("player", player.Name).
			Int32("pedID", packet.PedID).
			Msg("Player tried to delete someone else's pedestrian - possible hack or bug")
		return nil // Don't return error, just ignore
	}

	// Remove from manager
	s.pedManager.Remove(types.PedID(packet.PedID))

	// Broadcast removal to all clients
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal ped remove packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.PED_REMOVE,
		Data: packetData,
		Flag: network.PacketFlagReliable,
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast ped remove: %w", err)
	}

	s.logger.Info().
		Str("player", player.Name).
		Int32("pedID", packet.PedID).
		Msg("Player removed ped")

	return nil
}

// handlePedOnFoot handles PED_ONFOOT packets
func (s *Server) handlePedOnFoot(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Debug: Check packet size (expected: 59 bytes based on C++ struct)
	packets.DebugPacketSize("PED_ONFOOT", data, 59)

	// Parse the packet
	var packet packets.PedOnFootPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal PedOnFoot packet: %w", err)
	}

	// Get the ped
	ped := s.pedManager.GetPed(types.PedID(packet.PedID))
	if ped == nil {
		return nil // Ped doesn't exist, ignore
	}

	// Validate ownership (anti-cheat check)
	if ped.Syncer != player {
		s.logger.Warn().
			Str("player", player.Name).
			Int32("pedID", packet.PedID).
			Msg("Player tried to sync someone else's pedestrian on foot - possible hack or bug")
		return nil // Don't return error, just ignore
	}

	// Update ped position
	ped.Position = packet.Position

	// Broadcast to all other clients (unreliable for position updates)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal ped onfoot packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.PED_ONFOOT,
		Data: packetData,
		Flag: 0, // Unreliable for frequent position updates
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast ped onfoot: %w", err)
	}

	return nil
}

// handlePedDriverUpdate handles PED_DRIVER_UPDATE packets
func (s *Server) handlePedDriverUpdate(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.GetClientID())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Debug: Check packet size
	packets.DebugPacketSize("PED_DRIVER_UPDATE", data, 118)

	// Parse the packet
	var packet packets.PedDriverUpdatePacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal PedDriverUpdate packet: %w", err)
	}

	// Get the ped
	ped := s.pedManager.GetPed(packet.PedID)
	if ped == nil {
		return nil // Ped doesn't exist, ignore
	}

	// Validate ownership (anti-cheat check) - matching C++ logic
	if ped.Syncer != player {
		s.logger.Warn().
			Str("player", player.Name).
			Int32("pedID", int32(packet.PedID)).
			Msg("Player tried to sync (driver) someone else's pedestrian - possible hack or bug")
		return nil // Don't return error, just ignore
	}

	// Update ped position (matching C++ logic)
	ped.Position = packet.Position

	// Update vehicle if it exists (matching C++ logic)
	vehicle := s.vehicleManager.GetVehicle(types.VehicleID(packet.VehicleID))
	if vehicle != nil {
		vehicle.Position = packet.Position
		vehicle.FullRotation = packet.Rotation // Use FullRotation for 3D rotation
	}

	// Broadcast to all other clients (unreliable for frequent updates)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal ped driver update packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.PED_DRIVER_UPDATE,
		Data: packetData,
		Flag: 0, // Unreliable for frequent position updates (matching C++ logic)
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast ped driver update: %w", err)
	}

	s.logger.Debug().
		Str("player", player.Name).
		Int32("pedID", int32(packet.PedID)).
		Int32("vehicleID", packet.VehicleID).
		Msg("Player sent ped driver update")

	return nil
}

// sendExistingPedsTo sends information about all existing peds to a new player
func (s *Server) sendExistingPedsTo(client *network.Client) {
	allPeds := s.pedManager.GetAllPeds()

	for _, ped := range allPeds {
		pedSpawnPacket := packets.PedSpawnPacket{
			PedID:     int32(ped.ID),
			TempID:    0, // Existing peds don't need temp ID
			ModelID:   ped.ModelID,
			PedType:   ped.PedType,
			Position:  ped.Position,
			CreatedBy: ped.CreatedBy,
		}

		// Copy special model name
		copy(pedSpawnPacket.SpecialModelName[:], ped.SpecialModelName[:])

		packetData, err := pedSpawnPacket.Marshal()
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to marshal existing ped packet")
			continue
		}

		networkPacket := &network.NetworkPacket{
			ID:   types.PED_SPAWN,
			Data: packetData,
			Flag: network.PacketFlagReliable,
		}

		if err := s.networkServer.SendPacket(client, networkPacket); err != nil {
			s.logger.Error().Err(err).Msg("Failed to send existing ped packet")
		}

		s.logger.Debug().
			Int32("pedID", int32(ped.ID)).
			Uint32("clientID", client.ID).
			Msg("Sent existing ped info to new player")
	}
}
