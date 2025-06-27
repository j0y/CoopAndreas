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
)

// Server represents the game server
type Server struct {
	playerManager *entities.PlayerManager
	pedManager    *entities.PedManager
	networkServer *network.Server
	logger        zerolog.Logger
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
		playerManager: entities.NewPlayerManager(),
		pedManager:    entities.NewPedManager(),
		logger:        logger,
	}
}

// HandlePacket implements network.PacketHandler interface
func (s *Server) HandlePacket(client *network.Client, packet *network.NetworkPacket) error {
	switch packet.ID {
	case types.PED_SPAWN:
		return s.handlePedSpawn(client, packet.Data)
	case types.PED_REMOVE:
		return s.handlePedRemove(client, packet.Data)
	case types.PED_ONFOOT:
		return s.handlePedOnFoot(client, packet.Data)
	default:
		s.logger.Debug().
			Uint16("packetID", uint16(packet.ID)).
			Str("client", client.Addr.String()).
			Msg("Unhandled packet type")
		return nil
	}
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
		Syncer:          player,
		ModelID:         packet.ModelID,
		PedType:         packet.PedType,
		Position:        packet.Position,
		CreatedBy:       packet.CreatedBy,
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

// SetNetworkServer sets the network server reference (for sending packets)
func (s *Server) SetNetworkServer(netServer *network.Server) {
	s.networkServer = netServer
}

// HandlePlayerConnect handles when a player connects
func (s *Server) HandlePlayerConnect(client *network.Client) {
	// Create a new player (this would normally come from a handshake packet)
	player := &entities.Player{
		ID:   types.PlayerID(len(s.playerManager.GetAllPlayers()) + 1),
		Name: fmt.Sprintf("Player_%s", strings.Replace(client.Addr.String(), ":", "_", -1)),
	}

	s.playerManager.AddPlayer(client.Addr.String(), player)
	s.logger.Info().
		Str("player", player.Name).
		Str("address", client.Addr.String()).
		Msg("Player connected")
}

// HandlePlayerDisconnect handles when a player disconnects
func (s *Server) HandlePlayerDisconnect(client *network.Client) {
	player := s.playerManager.GetPlayer(client.Addr.String())
	if player == nil {
		return
	}

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

	s.playerManager.RemovePlayer(client.Addr.String())
	s.logger.Info().
		Str("player", player.Name).
		Int("pedsRemoved", len(removedPeds)).
		Msg("Player disconnected")
}
