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
}

// Ped represents a pedestrian/NPC in the game
type Ped struct {
	ID                types.PedID
	Syncer           *Player       // Player who controls this ped
	ModelID          int16
	PedType          uint8
	Position         types.Vector3
	CreatedBy        uint8
	SpecialModelName [8]byte
}

// PlayerManager manages all connected players
type PlayerManager struct {
	players map[string]*Player // key: peer address
	mutex   sync.RWMutex
}

// NewPlayerManager creates a new player manager
func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		players: make(map[string]*Player),
	}
}

// AddPlayer adds a new player
func (pm *PlayerManager) AddPlayer(peerAddr string, player *Player) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	player.PeerAddr = peerAddr
	pm.players[peerAddr] = player
}

// GetPlayer retrieves a player by peer address
func (pm *PlayerManager) GetPlayer(peerAddr string) *Player {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.players[peerAddr]
}

// RemovePlayer removes a player
func (pm *PlayerManager) RemovePlayer(peerAddr string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	delete(pm.players, peerAddr)
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
