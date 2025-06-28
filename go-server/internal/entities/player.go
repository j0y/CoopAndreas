package entities

import (
	"sync"

	"coopandreas-server/internal/types"
)

// Player represents a connected client
type Player struct {
	ID       types.PlayerID
	Name     string
	PeerAddr string // IP:Port for identification
	IsHost   bool
	// Player statistics (matching C++ m_afStats[14])
	Stats         [14]float32 // Player stats (strength, stamina, etc.)
	StatsModified bool        // Whether stats have been modified (matching C++ m_ucSyncFlags.bStatsModified)
	// Player appearance/clothes (matching C++ clothes data)
	ModelKeys       [10]uint32 // Model keys for clothes/appearance (matches C++ m_anModelKeys[10])
	TextureKeys     [18]uint32 // Texture keys for clothes/appearance (matches C++ m_anTextureKeys[18])
	FatStat         float32    // Fat statistic (matches C++ m_fFatStat)
	MuscleStat      float32    // Muscle statistic (matches C++ m_fMuscleStat)
	ClothesModified bool       // Whether clothes have been modified (matching C++ m_ucSyncFlags.bClothesModified)
	// Player waypoint (matching C++ waypoint data)
	WaypointPosition types.Vector3 // Waypoint position (matches C++ m_vecWaypointPos)
	WaypointModified bool          // Whether waypoint has been modified (matching C++ m_ucSyncFlags.bWaypointModified)
}

// Ped represents a pedestrian/NPC in the game
type Ped struct {
	ID               types.PedID
	Syncer           *Player // Player who controls this ped
	ModelID          int16
	PedType          uint8
	Position         types.Vector3
	CreatedBy        uint8
	SpecialModelName [8]byte
}

// PlayerManager manages all connected players
type PlayerManager struct {
	players map[string]*Player // key: unique client ID (not peer address)
	mutex   sync.RWMutex
}

// NewPlayerManager creates a new player manager
func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		players: make(map[string]*Player),
	}
}

// AddPlayer adds a new player
func (pm *PlayerManager) AddPlayer(clientID string, player *Player) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.players[clientID] = player
}

// GetPlayer retrieves a player by client ID
func (pm *PlayerManager) GetPlayer(clientID string) *Player {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.players[clientID]
}

// RemovePlayer removes a player
func (pm *PlayerManager) RemovePlayer(clientID string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	delete(pm.players, clientID)
}

// GetAllPlayers returns all connected players
func (pm *PlayerManager) GetAllPlayers() []*Player {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	players := make([]*Player, 0, len(pm.players))
	for _, player := range pm.players {
		players = append(players, player)
	}
	return players
}

// GetHost returns the current host player
func (pm *PlayerManager) GetHost() *Player {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	for _, player := range pm.players {
		if player.IsHost {
			return player
		}
	}
	return nil
}

// AssignHostToFirstPlayer assigns host status to the first connected player
// This matches the C++ CPlayerManager::AssignHostToFirstPlayer() logic
func (pm *PlayerManager) AssignHostToFirstPlayer() *Player {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if len(pm.players) <= 0 {
		return nil
	}

	// Check if there's already a host
	currentHost := pm.getHostUnsafe()
	if currentHost != nil {
		// There's already a host, don't reassign
		return currentHost
	}

	// Find the first player (in Go maps are unordered, so we'll pick any player)
	var firstPlayer *Player
	for _, player := range pm.players {
		firstPlayer = player
		break
	}

	if firstPlayer == nil {
		return nil
	}

	// Assign host status to the first player
	firstPlayer.IsHost = true

	return firstPlayer
}

// getHostUnsafe returns the current host without locking (internal use only)
func (pm *PlayerManager) getHostUnsafe() *Player {
	for _, player := range pm.players {
		if player.IsHost {
			return player
		}
	}
	return nil
}
