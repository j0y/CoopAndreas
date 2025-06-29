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
		return fmt.Errorf("no player found for client %s", client.GetClientID())
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

		// Check if vehicle has damage data (matching C++ logic)
		modifiedDamageStatus := false
		for _, damage := range vehicle.DamageManager {
			if damage != 0 {
				modifiedDamageStatus = true
				break
			}
		}

		// Send damage data if vehicle has been damaged (matching C++ logic)
		if modifiedDamageStatus {
			vehicleDamagePacket := packets.VehicleDamagePacket{
				VehicleID:     int32(vehicle.ID),
				DamageManager: vehicle.DamageManager,
			}

			damageData, err := vehicleDamagePacket.Marshal()
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to marshal vehicle damage packet")
				continue
			}

			damageNetworkPacket := &network.NetworkPacket{
				ID:   types.VEHICLE_DAMAGE,
				Data: damageData,
				Flag: network.PacketFlagReliable,
			}

			if err := s.networkServer.SendPacket(client, damageNetworkPacket); err != nil {
				s.logger.Error().Err(err).Msg("Failed to send vehicle damage packet")
			}

			s.logger.Debug().
				Int32("vehicleID", int32(vehicle.ID)).
				Uint32("clientID", client.ID).
				Msg("Sent vehicle damage data to new player")
		}

		// Send all vehicle components to new player (matching C++ logic)
		for _, componentID := range vehicle.Components {
			componentPacket := packets.VehicleComponentAddPacket{
				VehicleID:   int32(vehicle.ID),
				ComponentID: componentID,
			}

			componentData, err := componentPacket.Marshal()
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to marshal vehicle component packet")
				continue
			}

			componentNetworkPacket := &network.NetworkPacket{
				ID:   types.VEHICLE_COMPONENT_ADD,
				Data: componentData,
				Flag: network.PacketFlagReliable,
			}

			if err := s.networkServer.SendPacket(client, componentNetworkPacket); err != nil {
				s.logger.Error().Err(err).Msg("Failed to send vehicle component packet")
			}

			s.logger.Debug().
				Int32("vehicleID", int32(vehicle.ID)).
				Int32("componentID", componentID).
				Uint32("clientID", client.ID).
				Msg("Sent vehicle component to new player")
		}

		s.logger.Debug().
			Int32("vehicleID", int32(vehicle.ID)).
			Uint32("clientID", client.ID).
			Msg("Sent existing vehicle info to new player")
	}
}

// handleVehicleRemove handles VEHICLE_REMOVE packets
func (s *Server) handleVehicleRemove(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.VehicleRemovePacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal VehicleRemove packet: %w", err)
	}

	s.logger.Debug().
		Int32("vehicleID", packet.VehicleID).
		Str("player", player.Name).
		Msg("Received vehicle remove request")

	// Get the vehicle
	vehicle := s.vehicleManager.GetVehicle(types.VehicleID(packet.VehicleID))
	if vehicle == nil {
		s.logger.Warn().
			Int32("vehicleID", packet.VehicleID).
			Str("player", player.Name).
			Msg("Vehicle not found for removal")
		return nil // Not an error - vehicle might already be removed
	}

	// Check if the player is authorized to remove this vehicle
	// Only the syncer (creator) can remove the vehicle
	if vehicle.Syncer != player {
		s.logger.Warn().
			Int32("vehicleID", packet.VehicleID).
			Str("player", player.Name).
			Str("syncer", vehicle.Syncer.Name).
			Msg("Player attempted to remove vehicle they don't own")
		return nil // Ignore unauthorized removal attempts
	}

	// Remove from vehicle manager
	s.vehicleManager.RemoveVehicle(types.VehicleID(packet.VehicleID))

	// Broadcast removal to all other clients (reliable packet)
	// This matches the C++ logic: CNetwork::SendPacketToAll with ENET_PACKET_FLAG_RELIABLE
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle remove packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.VEHICLE_REMOVE,
		Data: packetData,
		Flag: network.PacketFlagReliable, // Reliable for important removal events
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast vehicle remove: %w", err)
	}

	s.logger.Info().
		Int32("vehicleID", packet.VehicleID).
		Str("player", player.Name).
		Uint16("modelID", vehicle.ModelID).
		Msg("Vehicle removed")

	return nil
}

// handleVehicleIdleUpdate handles VEHICLE_IDLE_UPDATE packets
func (s *Server) handleVehicleIdleUpdate(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.VehicleIdleUpdatePacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal VehicleIdleUpdate packet: %w", err)
	}

	s.logger.Debug().
		Int32("vehicleID", packet.VehicleID).
		Str("player", player.Name).
		Float32("x", packet.Position.X).
		Float32("y", packet.Position.Y).
		Float32("z", packet.Position.Z).
		Float32("health", packet.Health).
		Msg("Received vehicle idle update")

	// Get the vehicle
	vehicle := s.vehicleManager.GetVehicle(types.VehicleID(packet.VehicleID))
	if vehicle == nil {
		s.logger.Warn().
			Int32("vehicleID", packet.VehicleID).
			Str("player", player.Name).
			Msg("Vehicle not found for idle update")
		return nil // Not an error - vehicle might have been removed
	}

	// Check if the player is authorized to update this vehicle
	// Only the syncer (controller) can send idle updates
	if vehicle.Syncer != player {
		s.logger.Warn().
			Int32("vehicleID", packet.VehicleID).
			Str("player", player.Name).
			Str("syncer", vehicle.Syncer.Name).
			Msg("Player attempted to send idle update for vehicle they don't control")
		return nil // Ignore unauthorized updates
	}

	// Update the vehicle's state (matching C++ logic)
	vehicle.Position = packet.Position
	vehicle.FullRotation = packet.Rotation // Use FullRotation for 3D rotation
	vehicle.Roll = packet.Roll
	vehicle.Velocity = packet.Velocity
	vehicle.TurnSpeed = packet.TurnSpeed
	vehicle.PrimaryColor = packet.Color1
	vehicle.SecondaryColor = packet.Color2
	vehicle.Health = packet.Health
	vehicle.Paintjob = packet.Paintjob
	vehicle.PlaneGearState = packet.PlaneGearState
	vehicle.Locked = packet.Locked

	// Broadcast the update to all other clients (unreliable packet for frequent updates)
	// This matches the C++ logic: CNetwork::SendPacketToAll with ENetPacketFlag 0 (unreliable)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle idle update packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.VEHICLE_IDLE_UPDATE,
		Data: packetData,
		Flag: 0, // Unreliable for frequent position/state updates
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast vehicle idle update: %w", err)
	}

	return nil
}

// handleVehicleDriverUpdate handles VEHICLE_DRIVER_UPDATE packets
func (s *Server) handleVehicleDriverUpdate(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.VehicleDriverUpdatePacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal VehicleDriverUpdate packet: %w", err)
	}

	// Set the player ID (clients send 0, server assigns the actual ID)
	packet.PlayerID = player.ID

	s.logger.Debug().
		Int32("vehicleID", packet.VehicleID).
		Int32("playerID", int32(packet.PlayerID)).
		Str("player", player.Name).
		Float32("x", packet.Position.X).
		Float32("y", packet.Position.Y).
		Float32("z", packet.Position.Z).
		Float32("health", packet.Health).
		Uint8("playerHealth", packet.PlayerHealth).
		Msg("Received vehicle driver update")

	// Get the vehicle
	vehicle := s.vehicleManager.GetVehicle(types.VehicleID(packet.VehicleID))
	if vehicle == nil {
		s.logger.Warn().
			Int32("vehicleID", packet.VehicleID).
			Str("player", player.Name).
			Msg("Vehicle not found for driver update")
		return nil // Not an error - vehicle might have been removed
	}

	// Check if the player is authorized to drive this vehicle
	// Only the syncer can send driver updates, or we can assign them as driver
	if vehicle.Syncer != player {
		s.logger.Warn().
			Int32("vehicleID", packet.VehicleID).
			Str("player", player.Name).
			Str("syncer", vehicle.Syncer.Name).
			Msg("Player attempted to send driver update for vehicle they don't control")
		return nil // Ignore unauthorized updates
	}

	// Validate weapon (matching C++ logic: 0-18 or 22-46 are valid)
	isValidWeapon := packet.Weapon <= 18 || (packet.Weapon >= 22 && packet.Weapon <= 46)
	if !isValidWeapon {
		packet.Weapon = 0
		packet.Ammo = 0
	}

	// Validate player health and armour
	if packet.PlayerHealth > 100 {
		packet.PlayerHealth = 100
	}
	if packet.PlayerArmour > 100 {
		packet.PlayerArmour = 100
	}

	// Update the vehicle's state (matching C++ logic)
	vehicle.Position = packet.Position
	vehicle.FullRotation = packet.Rotation
	vehicle.Roll = packet.Roll
	vehicle.Velocity = packet.Velocity
	vehicle.PrimaryColor = packet.Color1
	vehicle.SecondaryColor = packet.Color2
	vehicle.Health = packet.Health
	vehicle.Paintjob = packet.Paintjob
	vehicle.BikeLean = packet.BikeLean
	vehicle.MiscComponentAngle = packet.MiscComponentAngle
	vehicle.PlaneGearState = packet.PlaneGearState
	vehicle.Locked = packet.Locked
	vehicle.Driver = player // Set this player as the driver

	// Broadcast the update to all other clients (unreliable packet for frequent updates)
	// This matches the C++ logic: CNetwork::SendPacketToAll with ENetPacketFlag 0 (unreliable)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle driver update packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.VEHICLE_DRIVER_UPDATE,
		Data: packetData,
		Flag: 0, // Unreliable for frequent position/state updates
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast vehicle driver update: %w", err)
	}

	return nil
}

// handleVehicleEnter handles VEHICLE_ENTER packets
func (s *Server) handleVehicleEnter(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.VehicleEnterPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal VehicleEnter packet: %w", err)
	}

	// Set the player ID (clients send 0, server assigns the actual ID)
	packet.PlayerID = player.ID

	s.logger.Debug().
		Int32("vehicleID", packet.VehicleID).
		Int32("playerID", int32(packet.PlayerID)).
		Str("player", player.Name).
		Uint8("seatID", packet.SeatID).
		Bool("force", packet.Force).
		Bool("passenger", packet.Passenger).
		Msg("Received vehicle enter request")

	// Get the vehicle
	vehicle := s.vehicleManager.GetVehicle(types.VehicleID(packet.VehicleID))
	if vehicle == nil {
		s.logger.Warn().
			Int32("vehicleID", packet.VehicleID).
			Str("player", player.Name).
			Msg("Vehicle not found for enter request")
		return nil // Not an error - vehicle might have been removed
	}

	// Validate seat ID (0 = driver, 1-3 = passengers)
	if packet.SeatID > 3 {
		s.logger.Warn().
			Int32("vehicleID", packet.VehicleID).
			Str("player", player.Name).
			Uint8("seatID", packet.SeatID).
			Msg("Invalid seat ID for vehicle enter")
		return nil
	}

	// Check if player is trying to enter as driver (seat 0)
	if packet.SeatID == 0 && !packet.Passenger {
		// Player wants to drive - this matches C++ VehicleEnter logic
		if vehicle.GetOccupant(0) != nil && vehicle.GetOccupant(0) != player {
			s.logger.Warn().
				Int32("vehicleID", packet.VehicleID).
				Str("player", player.Name).
				Str("currentDriver", vehicle.GetOccupant(0).Name).
				Msg("Vehicle already has a driver")
			return nil // Vehicle already has a driver
		}

		// Set occupant and reassign syncer (matching C++ logic)
		vehicle.SetOccupant(0, player)
		s.reassignVehicleSyncer(vehicle, player)
	} else {
		// Player wants to enter as passenger (matching C++ logic for passenger seats)
		// Check if seat is already occupied
		if vehicle.GetOccupant(int(packet.SeatID)) != nil {
			s.logger.Warn().
				Int32("vehicleID", packet.VehicleID).
				Str("player", player.Name).
				Uint8("seatID", packet.SeatID).
				Msg("Vehicle seat already occupied")
			return nil
		}

		// Set passenger and check if we need to reassign syncer
		vehicle.SetOccupant(int(packet.SeatID), player)

		// If there's no driver, reassign syncer to a passenger (matching C++ logic)
		if vehicle.GetOccupant(0) == nil {
			// Find any passenger to become syncer (matching C++ logic)
			for i := 1; i < 8; i++ {
				if occupant := vehicle.GetOccupant(i); occupant != nil {
					s.reassignVehicleSyncer(vehicle, occupant)
					break
				}
			}
		}
	}

	s.logger.Info().
		Int32("vehicleID", packet.VehicleID).
		Str("player", player.Name).
		Uint8("seatID", packet.SeatID).
		Bool("isDriver", packet.SeatID == 0 && !packet.Passenger).
		Bool("force", packet.Force).
		Msg("Player entering vehicle")

	// Broadcast the enter request to all other clients (reliable packet)
	// This matches the C++ logic: CNetwork::SendPacketToAll with ENET_PACKET_FLAG_RELIABLE
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle enter packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.VEHICLE_ENTER,
		Data: packetData,
		Flag: network.PacketFlagReliable, // Reliable for important vehicle state changes
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast vehicle enter: %w", err)
	}

	return nil
}

// handleVehicleExit handles VEHICLE_EXIT packets
func (s *Server) handleVehicleExit(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.VehicleExitPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal VehicleExit packet: %w", err)
	}

	// Set the player ID (security measure - don't trust client, matching C++ logic)
	packet.PlayerID = player.ID

	s.logger.Debug().
		Int32("playerID", int32(packet.PlayerID)).
		Str("player", player.Name).
		Bool("force", packet.Force).
		Msg("Received vehicle exit request")

	// Find the vehicle the player is currently in and their seat
	// We need to iterate through vehicles to find which one this player is in
	// This matches the C++ logic where they check player occupancy
	var currentVehicle *entities.Vehicle
	var playerSeatID int = -1
	for _, vehicle := range s.vehicleManager.GetAllVehicles() {
		for seatID := 0; seatID < 8; seatID++ {
			if occupant := vehicle.GetOccupant(seatID); occupant != nil && occupant.ID == player.ID {
				currentVehicle = vehicle
				playerSeatID = seatID
				break
			}
		}
		if currentVehicle != nil {
			break
		}
	}

	if currentVehicle == nil {
		s.logger.Debug().
			Str("player", player.Name).
			Msg("Player not in any vehicle - ignoring exit request")
		return nil // Player is not in a vehicle, ignore the request
	}

	s.logger.Info().
		Int32("vehicleID", int32(currentVehicle.ID)).
		Str("player", player.Name).
		Int("seatID", playerSeatID).
		Bool("wasDriver", playerSeatID == 0).
		Bool("force", packet.Force).
		Msg("Player exiting vehicle")

	// Remove player from vehicle occupancy (matching C++ logic)
	currentVehicle.RemoveOccupant(playerSeatID)

	// Handle syncer reassignment if driver left (matching C++ CVehicle exit logic)
	if playerSeatID == 0 { // Driver exited
		// Find a passenger to become the new syncer (matching C++ logic)
		var newSyncer *entities.Player
		for i := 1; i < 8; i++ {
			if occupant := currentVehicle.GetOccupant(i); occupant != nil {
				newSyncer = occupant
				break
			}
		}

		if newSyncer != nil {
			s.reassignVehicleSyncer(currentVehicle, newSyncer)
		} else {
			// No passengers left, keep current syncer (they may still sync the empty vehicle)
			s.logger.Debug().
				Int32("vehicleID", int32(currentVehicle.ID)).
				Msg("Driver exited but no passengers to reassign syncer")
		}
	}

	// Broadcast the exit packet to all other clients (reliable packet)
	// This matches the C++ logic: CNetwork::SendPacketToAll with ENET_PACKET_FLAG_RELIABLE
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle exit packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.VEHICLE_EXIT,
		Data: packetData,
		Flag: network.PacketFlagReliable, // Reliable for important vehicle state changes
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast vehicle exit: %w", err)
	}

	return nil
}

// handleVehiclePassengerUpdate handles VEHICLE_PASSENGER_UPDATE packets
// This matches the C++ CVehiclePackets::VehiclePassengerUpdate::Handle() functionality
func (s *Server) handleVehiclePassengerUpdate(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.GetClientID())
	}

	// Parse the packet
	var packet packets.VehiclePassengerUpdatePacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal VehiclePassengerUpdate packet: %w", err)
	}

	// Set the player ID (security measure - don't trust client, matching C++ logic)
	packet.PlayerID = int32(player.ID)

	// Validate weapon (matching C++ logic: 0-18 or 22-46 are valid)
	isValidWeapon := packet.Weapon <= 18 || (packet.Weapon >= 22 && packet.Weapon <= 46)
	if !isValidWeapon {
		packet.Weapon = 0
		packet.Ammo = 0
	}

	// Get the vehicle entity
	vehicle := s.vehicleManager.GetVehicle(types.VehicleID(packet.VehicleID))
	if vehicle != nil {
		// Set the player as occupant of the specified seat (seat ID + 1 for driver/passenger indexing)
		// This matches the C++ logic: vehicle->SetOccupant(packet->seatid + 1, player)
		vehicle.SetOccupant(int(packet.SeatID)+1, player)

		// If there's no driver (seat 0), try to reassign syncer to a passenger
		// This matches the C++ logic for reassigning syncer when driver leaves
		if vehicle.GetOccupant(0) == nil {
			for i := 1; i < 8; i++ { // Check passenger seats
				if occupant := vehicle.GetOccupant(i); occupant != nil {
					vehicle.Syncer = occupant
					s.logger.Debug().
						Int32("vehicleID", int32(vehicle.ID)).
						Int32("newSyncerID", int32(occupant.ID)).
						Str("newSyncerName", occupant.Name).
						Msg("Reassigned vehicle syncer to passenger")
					break
				}
			}
		}
	}

	s.logger.Debug().
		Str("player", player.Name).
		Int32("vehicleID", packet.VehicleID).
		Uint8("seatID", packet.SeatID).
		Uint8("health", packet.PlayerHealth).
		Uint8("weapon", packet.Weapon).
		Uint8("driveby", packet.Driveby).
		Msg("Vehicle passenger update")

	// Broadcast to all other clients (unreliable packet, frequent updates)
	// This matches the C++ logic: SendPacketToAll with (ENetPacketFlag)0 (unreliable)
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle passenger update packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.VEHICLE_PASSENGER_UPDATE,
		Data: packetData,
		Flag: 0, // Unreliable for frequent passenger updates (matching C++ implementation)
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast vehicle passenger update: %w", err)
	}

	return nil
}

// handleVehicleDamage handles VEHICLE_DAMAGE packets
func (s *Server) handleVehicleDamage(client *network.Client, data []byte) error {
	// Parse the packet
	var packet packets.VehicleDamagePacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal VehicleDamage packet: %w", err)
	}

	// Get the vehicle
	vehicle := s.vehicleManager.GetVehicle(types.VehicleID(packet.VehicleID))
	if vehicle == nil {
		s.logger.Warn().
			Int32("vehicleID", packet.VehicleID).
			Str("client", client.GetClientID()).
			Msg("Vehicle damage packet for non-existent vehicle")
		return nil // Ignore damage packets for non-existent vehicles
	}

	// Store the damage manager data (matching C++ logic)
	copy(vehicle.DamageManager[:], packet.DamageManager[:])

	s.logger.Debug().
		Int32("vehicleID", packet.VehicleID).
		Str("client", client.GetClientID()).
		Msg("Vehicle damage updated")

	// Broadcast to all other clients (reliable packet)
	// This matches C++ logic: SendPacketToAll with ENET_PACKET_FLAG_RELIABLE
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle damage packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.VEHICLE_DAMAGE,
		Data: packetData,
		Flag: 1, // Reliable for damage updates (matching C++ implementation)
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast vehicle damage: %w", err)
	}

	return nil
}

// handleVehicleComponentAdd handles VEHICLE_COMPONENT_ADD packets
func (s *Server) handleVehicleComponentAdd(client *network.Client, data []byte) error {
	// Parse the packet
	var packet packets.VehicleComponentAddPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal VehicleComponentAdd packet: %w", err)
	}

	// Get the vehicle
	vehicle := s.vehicleManager.GetVehicle(types.VehicleID(packet.VehicleID))
	if vehicle == nil {
		s.logger.Warn().
			Int32("vehicleID", packet.VehicleID).
			Str("client", client.GetClientID()).
			Msg("Vehicle component add packet for non-existent vehicle")
		return nil // Ignore packets for non-existent vehicles
	}

	// Add the component to the vehicle (matching C++ logic)
	vehicle.AddComponent(packet.ComponentID)

	s.logger.Debug().
		Int32("vehicleID", packet.VehicleID).
		Int32("componentID", packet.ComponentID).
		Str("client", client.GetClientID()).
		Msg("Vehicle component added")

	// Broadcast to all other clients (reliable packet)
	// This matches C++ logic: SendPacketToAll with ENET_PACKET_FLAG_RELIABLE
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle component add packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.VEHICLE_COMPONENT_ADD,
		Data: packetData,
		Flag: 1, // Reliable for component updates (matching C++ implementation)
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast vehicle component add: %w", err)
	}

	return nil
}

// handleVehicleComponentRemove handles VEHICLE_COMPONENT_REMOVE packets
func (s *Server) handleVehicleComponentRemove(client *network.Client, data []byte) error {
	// Parse the packet
	var packet packets.VehicleComponentRemovePacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal VehicleComponentRemove packet: %w", err)
	}

	// Get the vehicle
	vehicle := s.vehicleManager.GetVehicle(types.VehicleID(packet.VehicleID))
	if vehicle == nil {
		s.logger.Warn().
			Int32("vehicleID", packet.VehicleID).
			Str("client", client.GetClientID()).
			Msg("Vehicle component remove packet for non-existent vehicle")
		return nil // Ignore packets for non-existent vehicles
	}

	// Remove the component from the vehicle (matching C++ logic)
	vehicle.RemoveComponent(packet.ComponentID)

	s.logger.Debug().
		Int32("vehicleID", packet.VehicleID).
		Int32("componentID", packet.ComponentID).
		Str("client", client.GetClientID()).
		Msg("Vehicle component removed")

	// Broadcast to all other clients (reliable packet)
	// This matches C++ logic: SendPacketToAll with ENET_PACKET_FLAG_RELIABLE
	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle component remove packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.VEHICLE_COMPONENT_REMOVE,
		Data: packetData,
		Flag: 1, // Reliable for component updates (matching C++ implementation)
	}

	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to broadcast vehicle component remove: %w", err)
	}

	return nil
}

// handleAssignVehicle handles ASSIGN_VEHICLE packets
// This is a client-side packet handler (rarely used on server)
// The client uses this to toggle vehicle syncing state
func (s *Server) handleAssignVehicle(client *network.Client, data []byte) error {
	// Parse the packet
	var packet packets.AssignVehiclePacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal AssignVehicle packet: %w", err)
	}

	// Find the vehicle
	vehicle := s.vehicleManager.GetVehicle(types.VehicleID(packet.VehicleID))
	if vehicle == nil {
		s.logger.Debug().
			Int32("vehicleID", packet.VehicleID).
			Str("client", client.GetClientID()).
			Msg("AssignVehicle packet for non-existent vehicle")
		return nil // Ignore packets for non-existent vehicles
	}

	s.logger.Debug().
		Int32("vehicleID", packet.VehicleID).
		Str("client", client.GetClientID()).
		Msg("Received AssignVehicle packet (client-side handling)")

	// Note: This packet is typically handled on the client side to toggle
	// vehicle syncing state. The server usually doesn't need to process this,
	// but we log it for debugging purposes.

	return nil
}

// sendAssignVehiclePacket sends an ASSIGN_VEHICLE packet to a specific client
// This matches the C++ server logic in CVehicle::ReassignSyncer
func (s *Server) sendAssignVehiclePacket(client *network.Client, vehicleID types.VehicleID) error {
	packet := packets.AssignVehiclePacket{
		VehicleID: int32(vehicleID),
	}

	packetData, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal assign vehicle packet: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.ASSIGN_VEHICLE,
		Data: packetData,
		Flag: 1, // Reliable for syncer assignment (matching C++ implementation)
	}

	if err := s.networkServer.SendPacket(client, networkPacket); err != nil {
		return fmt.Errorf("failed to send assign vehicle packet: %w", err)
	}

	s.logger.Debug().
		Int32("vehicleID", int32(vehicleID)).
		Str("client", client.GetClientID()).
		Msg("Sent AssignVehicle packet")

	return nil
}

// reassignVehicleSyncer reassigns a vehicle syncer and sends appropriate packets
// This matches the C++ CVehicle::ReassignSyncer logic exactly
func (s *Server) reassignVehicleSyncer(vehicle *entities.Vehicle, newSyncer *entities.Player) error {
	if vehicle.Syncer != newSyncer {
		oldSyncer := vehicle.Syncer

		// Send to the old vehicle syncer (matching C++ logic)
		if oldSyncer != nil {
			oldSyncerClient := s.getClientByPlayer(oldSyncer)
			if oldSyncerClient != nil {
				if err := s.sendAssignVehiclePacket(oldSyncerClient, vehicle.ID); err != nil {
					s.logger.Error().
						Err(err).
						Int32("vehicleID", int32(vehicle.ID)).
						Str("oldSyncer", oldSyncer.Name).
						Msg("Failed to send assign vehicle packet to old syncer")
				}
			}
		}

		// Send to the new vehicle syncer (matching C++ logic)
		if newSyncer != nil {
			newSyncerClient := s.getClientByPlayer(newSyncer)
			if newSyncerClient != nil {
				if err := s.sendAssignVehiclePacket(newSyncerClient, vehicle.ID); err != nil {
					s.logger.Error().
						Err(err).
						Int32("vehicleID", int32(vehicle.ID)).
						Str("newSyncer", newSyncer.Name).
						Msg("Failed to send assign vehicle packet to new syncer")
				}
			}
		}

		// Update the vehicle syncer (matching C++ logic: this->m_pSyncer = newSyncer)
		vehicle.ReassignSyncer(newSyncer)

		s.logger.Debug().
			Int32("vehicleID", int32(vehicle.ID)).
			Str("oldSyncer", func() string {
				if oldSyncer != nil {
					return oldSyncer.Name
				}
				return "none"
			}()).
			Str("newSyncer", func() string {
				if newSyncer != nil {
					return newSyncer.Name
				}
				return "none"
			}()).
			Msg("Vehicle syncer reassigned")
	}

	return nil
}
