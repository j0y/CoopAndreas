package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"

	"coopandreas-server/internal/entities"
	"coopandreas-server/internal/network"
	"coopandreas-server/internal/packets"
	"coopandreas-server/internal/types"
	"coopandreas-server/internal/version"
)

// Server represents the game server
type Server struct {
	playerManager  *entities.PlayerManager
	pedManager     *entities.PedManager
	vehicleManager *entities.VehicleManager
	networkServer  *network.Server
	versionManager *version.Manager
	logger         zerolog.Logger
}

// New creates a new game server instance
func New() *Server {
	// Configure zerolog for game server
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}).
		With().
		Timestamp().
		Str("component", "game").
		Logger()

	return &Server{
		playerManager:  entities.NewPlayerManager(),
		pedManager:     entities.NewPedManager(),
		vehicleManager: entities.NewVehicleManager(),
		versionManager: version.NewManager(),
		logger:         logger,
	}
}

// HandlePacket implements network.GameServer interface
func (s *Server) HandlePacket(client *network.Client, packet *network.NetworkPacket) error {
	switch packet.ID {
	case types.CHECK_VERSION:
		return s.handleCheckVersion(client, packet.Data)
	case types.PLAYER_GET_NAME:
		return s.handlePlayerGetName(client, packet.Data)
	case types.PLAYER_ONFOOT:
		return s.handlePlayerOnFoot(client, packet.Data)
	case types.PED_SPAWN:
		return s.handlePedSpawn(client, packet.Data)
	case types.PED_REMOVE:
		return s.handlePedRemove(client, packet.Data)
	case types.PED_ONFOOT:
		return s.handlePedOnFoot(client, packet.Data)
	case types.PLAYER_KEY_SYNC:
		return s.handlePlayerKeySync(client, packet.Data)
	case types.VEHICLE_SPAWN:
		return s.handleVehicleSpawn(client, packet.Data)
	case types.MASS_PACKET_SEQUENCE:
		return s.handleMassPacketSequence(client, packet.Data)
	default:
		s.logger.Debug().
			Uint16("packetID", uint16(packet.ID)).
			Str("client", client.Addr.String()).
			Msg("Unhandled packet type")
		return nil
	}
}

// handleCheckVersion handles CHECK_VERSION packets
func (s *Server) handleCheckVersion(client *network.Client, data []byte) error {
	// Parse the version check packet
	var packet packets.CheckVersionPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal CheckVersion packet: %w", err)
	}

	// Extract client version string (null-terminated)
	clientVersionStr := string(packet.ClientVersion[:])
	if nullIndex := strings.IndexByte(clientVersionStr, 0); nullIndex != -1 {
		clientVersionStr = clientVersionStr[:nullIndex]
	}
	clientVersionStr = strings.TrimSpace(clientVersionStr)

	s.logger.Info().
		Str("client", client.Addr.String()).
		Str("clientVersion", clientVersionStr).
		Uint32("protocolVersion", packet.ProtocolVersion).
		Msg("Version check request received")

	// Validate client version
	isCompatible, message, err := s.versionManager.ValidateClientVersion(clientVersionStr)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("client", client.Addr.String()).
			Str("clientVersion", clientVersionStr).
			Msg("Failed to validate client version")

		// Send incompatible response due to validation error
		isCompatible = false
		message = "Version validation failed"
	}

	// Create response packet
	var responseMessage string
	if isCompatible {
		responseMessage = "Welcome to CoopAndreas Server!"
		s.logger.Info().
			Str("client", client.Addr.String()).
			Str("clientVersion", clientVersionStr).
			Msg("Client version accepted")
	} else {
		responseMessage = message
		s.logger.Warn().
			Str("client", client.Addr.String()).
			Str("clientVersion", clientVersionStr).
			Str("reason", message).
			Msg("Client version rejected")
	}

	// Send version check response
	response := packets.NewCheckVersionResponse(isCompatible, responseMessage)
	responseData, err := response.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal version response: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.CHECK_VERSION, // Respond with same packet ID
		Data: responseData,
		Flag: network.PacketFlagReliable,
	}

	if err := s.networkServer.SendPacket(client, networkPacket); err != nil {
		return fmt.Errorf("failed to send version response: %w", err)
	}

	// If version is incompatible, we might want to disconnect the client
	// For now, we'll let them stay connected but log the incompatibility
	if !isCompatible {
		s.logger.Info().
			Str("client", client.Addr.String()).
			Msg("Client with incompatible version remains connected (warning only)")
	}

	return nil
}

// handlePedSpawn handles PED_SPAWN packets
func (s *Server) handlePedSpawn(client *network.Client, data []byte) error {
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.Addr)
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

	// Send confirmation back to spawner
	confirmPacket := packets.PedConfirmPacket{
		TempID: packet.TempID,
		PedID:  packet.PedID,
	}

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
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.Addr)
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
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.Addr)
	}

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

// handleMassPacketSequence handles MASS_PACKET_SEQUENCE packets
// Based on C++ server logic: simply rebroadcast the entire mass packet to all other clients
// This matches the C++ implementation in CNetwork::HandlePacketReceive
func (s *Server) handleMassPacketSequence(client *network.Client, data []byte) error {
	// In C++: CNetwork::SendPacketRawToAll(event.packet->data, event.packet->dataLength, (ENetPacketFlag)event.packet->flags, event.peer);
	// The server simply rebroadcasts the entire mass packet sequence to all other clients
	// without parsing or processing the individual packets within it

	s.logger.Debug().
		Str("client", client.Addr.String()).
		Int("dataSize", len(data)).
		Msg("Received mass packet sequence, rebroadcasting to all clients")

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

// SetNetworkServer sets the network server reference (for sending packets)
func (s *Server) SetNetworkServer(netServer *network.Server) {
	s.networkServer = netServer
}

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

	// Store player ID for broadcasting
	disconnectedPlayerID := player.ID

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
		Msg("Player disconnected")
}

// broadcastPlayerConnectedPacket sends a PlayerConnected packet to all players except the specified client
func (s *Server) broadcastPlayerConnectedPacket(packet *packets.PlayerConnectedPacket, excludeClient *network.Client) {
	// For now, we'll log the intent since we don't have a SendPacketToAll method yet
	// TODO: Implement proper broadcasting to all clients except excludeClient
	s.logger.Info().
		Int32("playerID", int32(packet.ID)).
		Bool("isAlreadyConnected", packet.IsAlreadyConnected == 1).
		Msg("Broadcasting PlayerConnected packet to existing players")
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

// sendExistingVehiclesTo sends information about all existing vehicles to a new player
func (s *Server) sendExistingVehiclesTo(client *network.Client) {
	// TODO: Implement when vehicle management is added
	s.logger.Debug().
		Str("client", client.Addr.String()).
		Msg("Sending existing vehicles to new player (placeholder)")
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
			Str("client", client.Addr.String()).
			Msg("Sent existing ped info to new player")
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

// generatePlayerID generates a unique player ID (similar to C++ GetFreeID())
func (s *Server) generatePlayerID() types.PlayerID {
	allPlayers := s.playerManager.GetAllPlayers()
	return types.PlayerID(len(allPlayers) + 1)
}

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
// This method processes player movement and state updates from clients
// Based on C++ CPlayerPackets::PlayerOnFoot::Handle implementation
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
	isValidWeapon := (packet.Weapon >= 0 && packet.Weapon <= 18) || (packet.Weapon >= 22 && packet.Weapon <= 46)
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
	// Get player associated with this client
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return fmt.Errorf("no player found for client %s", client.Addr)
	}

	// Parse the packet
	var packet packets.PlayerKeySyncPacket
	if err := packet.Unmarshal(data); err != nil {
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
