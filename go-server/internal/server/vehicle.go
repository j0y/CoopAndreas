package server

import (
	"fmt"

	"coopandreas-server/internal/entities"
	"coopandreas-server/internal/network"
	"coopandreas-server/internal/packets"
	"coopandreas-server/internal/types"
)

// handleVehicleSpawn handles VEHICLE_SPAWN packets
func (s *Server) handleVehicleSpawn(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.Addr)
	}

	// Parse the packet
	var packet packets.VehicleSpawnPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal VehicleSpawn packet: %w", err)
	}

	// Validate model ID (matching C++ logic: 400-611 are valid vehicle models)
	if packet.ModelID < 400 || packet.ModelID > 611 {
		s.logger.Warn().
			Uint16("modelID", packet.ModelID).
			Str("player", player.Name).
			Msg("Invalid vehicle model ID")
		return nil // Ignore invalid model IDs
	}

	// Assign a server-side vehicle ID (matching C++ logic)
	vehicleID := s.vehicleManager.GetFreeID()
	packet.VehicleID = int32(vehicleID)

	// Create the vehicle entity
	vehicle := entities.NewVehicle(vehicleID, packet.ModelID, packet.Position, packet.Rotation)
	vehicle.Syncer = player
	vehicle.PrimaryColor = packet.Color1
	vehicle.SecondaryColor = packet.Color2
	vehicle.CreatedBy = packet.CreatedBy

	// Add to vehicle manager
	if err := s.vehicleManager.AddVehicle(vehicle); err != nil {
		return fmt.Errorf("failed to add vehicle to manager: %w", err)
	}

	// Broadcast spawn packet to all other clients (reliable, matching C++ implementation)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle spawn packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.VEHICLE_SPAWN,
		Data: packetData,
		Flag: 1, // Reliable for vehicle spawn (matching C++ ENET_PACKET_FLAG_RELIABLE)
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast vehicle spawn: %w", err)
	}

	// Send vehicle confirmation back to the spawning client (matching C++ logic)
	confirmPacket := packets.NewVehicleConfirmPacket(packet.TempID, packet.VehicleID)
	confirmData, err := confirmPacket.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle confirm packet: %w", err)
	}

	confirmNetworkPacket := &network.NetworkPacket{
		ID:   types.VEHICLE_CONFIRM,
		Data: confirmData,
		Flag: 1, // Reliable
	}

	if err := s.networkServer.SendPacket(client, confirmNetworkPacket); err != nil {
		return fmt.Errorf("failed to send vehicle confirm: %w", err)
	}

	s.logger.Info().
		Str("player", player.Name).
		Int32("vehicleID", packet.VehicleID).
		Uint8("tempID", packet.TempID).
		Uint16("modelID", packet.ModelID).
		Float32("x", packet.Position.X).
		Float32("y", packet.Position.Y).
		Float32("z", packet.Position.Z).
		Msg("Vehicle spawned")

	return nil
}

// sendExistingVehiclesTo sends information about all existing vehicles to a new player
func (s *Server) sendExistingVehiclesTo(client *network.Client) {
	allVehicles := s.vehicleManager.GetAllVehicles()

	for _, vehicle := range allVehicles {
		vehicleSpawnPacket := packets.VehicleSpawnPacket{
			VehicleID: int32(vehicle.ID),
			TempID:    0, // Existing vehicles don't need temp ID
			ModelID:   vehicle.ModelID,
			Position:  vehicle.Position,
			Rotation:  vehicle.Rotation,
			Color1:    vehicle.PrimaryColor,
			Color2:    vehicle.SecondaryColor,
			CreatedBy: vehicle.CreatedBy,
		}

		packetData, err := vehicleSpawnPacket.Marshal()
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to marshal existing vehicle packet")
			continue
		}

		networkPacket := &network.NetworkPacket{
			ID:   types.VEHICLE_SPAWN,
			Data: packetData,
			Flag: network.PacketFlagReliable,
		}

		if err := s.networkServer.SendPacket(client, networkPacket); err != nil {
			s.logger.Error().Err(err).Msg("Failed to send existing vehicle packet")
		}

		s.logger.Debug().
			Int32("vehicleID", int32(vehicle.ID)).
			Str("client", client.Addr.String()).
			Msg("Sent existing vehicle info to new player")
	}
}
