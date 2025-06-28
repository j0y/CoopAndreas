package server

import (
	"os"

	"github.com/rs/zerolog"

	"coopandreas-server/internal/entities"
	"coopandreas-server/internal/network"
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
	case types.RESPAWN_PLAYER:
		return s.handleRespawnPlayer(client, packet.Data)
	case types.PLAYER_SET_HOST:
		return s.handlePlayerSetHost(client, packet.Data)
	case types.VEHICLE_SPAWN:
		return s.handleVehicleSpawn(client, packet.Data)
	case types.VEHICLE_REMOVE:
		return s.handleVehicleRemove(client, packet.Data)
	case types.VEHICLE_IDLE_UPDATE:
		return s.handleVehicleIdleUpdate(client, packet.Data)
	case types.VEHICLE_DRIVER_UPDATE:
		return s.handleVehicleDriverUpdate(client, packet.Data)
	case types.VEHICLE_ENTER:
		return s.handleVehicleEnter(client, packet.Data)
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

// SetNetworkServer sets the network server reference (for sending packets)
func (s *Server) SetNetworkServer(netServer *network.Server) {
	s.networkServer = netServer
}
