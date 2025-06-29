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
	players   map[string]*Player // key: unique client ID (not peer address)
	joinOrder []*Player          // tracks players in join order (like C++ vector)
	mutex     sync.RWMutex
}

// NewPlayerManager creates a new player manager
func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		players:   make(map[string]*Player),
		joinOrder: make([]*Player, 0),
	}
}

// AddPlayer adds a new player
func (pm *PlayerManager) AddPlayer(clientID string, player *Player) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.players[clientID] = player
	pm.joinOrder = append(pm.joinOrder, player) // Track join order like C++ vector
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

	player := pm.players[clientID]
	delete(pm.players, clientID)

	// Also remove from join order slice
	if player != nil {
		for i, p := range pm.joinOrder {
			if p == player {
				// Remove from slice preserving order
				pm.joinOrder = append(pm.joinOrder[:i], pm.joinOrder[i+1:]...)
				break
			}
		}
	}
}

// GetAllPlayers returns all connected players in join order
func (pm *PlayerManager) GetAllPlayers() []*Player {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	// Return a copy of the join order slice to maintain ordering
	players := make([]*Player, len(pm.joinOrder))
	copy(players, pm.joinOrder)
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
// This matches the C++ CPlayerManager::AssignHostToFirstPlayer() logic exactly
func (pm *PlayerManager) AssignHostToFirstPlayer() *Player {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if len(pm.joinOrder) <= 0 {
		return nil
	}

	// Get the first player that joined (matching C++ m_pPlayers.front())
	firstPlayer := pm.joinOrder[0]

	// Get current host
	currentHost := pm.getHostUnsafe()

	// If the first player is already the host, no need to change anything
	if firstPlayer == currentHost {
		return firstPlayer
	}

	// Remove host status from current host (if any)
	if currentHost != nil {
		currentHost.IsHost = false
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

// GetClientIDByPlayer returns the client ID associated with a player
func (pm *PlayerManager) GetClientIDByPlayer(player *Player) string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	for clientID, managedPlayer := range pm.players {
		if managedPlayer.ID == player.ID {
			return clientID
		}
	}
	return ""
}

// GetPlayerByID returns a player by their Player ID
func (pm *PlayerManager) GetPlayerByID(playerID types.PlayerID) *Player {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	for _, player := range pm.players {
		if player.ID == playerID {
			return player
		}
	}
	return nil
}
