package server

import (
	"os"

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
	// Global game state
	currentWeatherTime *packets.GameWeatherTimePacket // Current weather/time state, nil if not set
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
	case types.PLAYER_PLACE_WAYPOINT:
		return s.handlePlayerPlaceWaypoint(client, packet.Data)
	case types.PLAYER_ONFOOT:
		return s.handlePlayerOnFoot(client, packet.Data)
	case types.PLAYER_BULLET_SHOT:
		return s.handlePlayerBulletShot(client, packet.Data)
	case types.PED_SPAWN:
		return s.handlePedSpawn(client, packet.Data)
	case types.PED_REMOVE:
		return s.handlePedRemove(client, packet.Data)
	case types.PED_ONFOOT:
		return s.handlePedOnFoot(client, packet.Data)
	case types.PED_ADD_TASK:
		return s.handlePedAddTask(client, packet.Data)
	case types.PED_REMOVE_TASK:
		return s.handlePedRemoveTask(client, packet.Data)
	case types.PED_SHOT_SYNC:
		return s.handlePedShotSync(client, packet.Data)
	case types.PED_PASSENGER_UPDATE:
		return s.handlePedPassengerUpdate(client, packet.Data)
	case types.PED_DRIVER_UPDATE:
		return s.handlePedDriverUpdate(client, packet.Data)
	case types.PLAYER_KEY_SYNC:
		return s.handlePlayerKeySync(client, packet.Data)
	case types.PLAYER_AIM_SYNC:
		return s.handlePlayerAimSync(client, packet.Data)
	case types.RESPAWN_PLAYER:
		return s.handleRespawnPlayer(client, packet.Data)
	case types.PLAYER_SET_HOST:
		return s.handlePlayerSetHost(client, packet.Data)
	case types.ADD_EXPLOSION:
		return s.handleAddExplosion(client, packet.Data)
	case types.PLAYER_STATS:
		return s.handlePlayerStats(client, packet.Data)
	case types.REBUILD_PLAYER:
		return s.handleRebuildPlayer(client, packet.Data)
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
	case types.VEHICLE_EXIT:
		return s.handleVehicleExit(client, packet.Data)
	case types.VEHICLE_DAMAGE:
		return s.handleVehicleDamage(client, packet.Data)
	case types.VEHICLE_COMPONENT_ADD:
		return s.handleVehicleComponentAdd(client, packet.Data)
	case types.VEHICLE_COMPONENT_REMOVE:
		return s.handleVehicleComponentRemove(client, packet.Data)
	case types.VEHICLE_PASSENGER_UPDATE:
		return s.handleVehiclePassengerUpdate(client, packet.Data)
	case types.ASSIGN_VEHICLE:
		return s.handleAssignVehicle(client, packet.Data)
	case types.PLAYER_CHAT_MESSAGE:
		return s.handlePlayerChatMessage(client, packet.Data)
	case types.GAME_WEATHER_TIME:
		return s.handleGameWeatherTime(client, packet.Data)
	case types.OPCODE_SYNC:
		return s.handleOpCodeSync(client, packet.Data)
	case types.PLAY_MISSION_AUDIO:
		return s.handlePlayMissionAudio(client, packet.Data)
	case types.START_CUTSCENE:
		return s.handleStartCutscene(client, packet.Data)
	case types.SKIP_CUTSCENE:
		return s.handleSkipCutscene(client, packet.Data)
	case types.MASS_PACKET_SEQUENCE:
		return s.handleMassPacketSequence(client, packet.Data)
	default:
		s.logger.Debug().
			Uint16("packetID", uint16(packet.ID)).
			Uint32("clientID", client.ID).
			Msg("Unhandled packet type")
		return nil
	}
}

// SetNetworkServer sets the network server reference (for sending packets)
func (s *Server) SetNetworkServer(netServer *network.Server) {
	s.networkServer = netServer
}

// getClientByPlayer finds the network client associated with a specific player
// This matches the C++ approach where they have direct access to player->m_pPeer
func (s *Server) getClientByPlayer(player *entities.Player) *network.Client {
	if player == nil {
		return nil
	}

	// Get the client ID associated with this player
	clientID := s.playerManager.GetClientIDByPlayer(player)
	if clientID == "" {
		return nil
	}

	// Get the client from the network server
	return s.networkServer.GetClient(clientID)
}
