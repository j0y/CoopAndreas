package entities

import (
	"fmt"
	"strings"
	"sync"

	"coopandreas-server/internal/types"
)

// AllowedSpecialActors matches the C++ ms_aszAllowedSpecialActors array
var allowedSpecialActors = [52]string{
	"ANDRE", "BBTHIN", "BB", "CAT", "CESAR", "COPGRL1", "COPGRL2", "COPGRL3",
	"CLAUDE", "CROGRL1", "CROGRL2", "CROGRL3", "DWAYNE", "EMMET", "FORELLI", "GANGRL1",
	"GANGRL2", "GANGRL3", "GUNGRL1", "GUNGRL2", "GUNGRL3", "HERN", "JANITOR", "JETHRO",
	"JIZZY", "KENDL", "MACCER", "MADDOGG", "MECGRL1", "MECGRL2", "MECGRL3", "NURGRL1",
	"NURGRL2", "NURGRL3", "OGLOC", "PAUL", "PULASKI", "ROSE", "RYDER1", "RYDER2",
	"RYDER3", "SINDACO", "SMOKE", "SMOKEV", "SUZIE", "SWEET", "TBONE", "TENPEN",
	"TORINO", "TRUTH", "WUZIMU", "ZERO",
}

// PedManager manages all pedestrians in the game
type PedManager struct {
	peds     map[types.PedID]*Ped
	mutex    sync.RWMutex
	nextID   types.PedID
	idMutex  sync.Mutex
}

// NewPedManager creates a new ped manager
func NewPedManager() *PedManager {
	return &PedManager{
		peds:   make(map[types.PedID]*Ped),
		nextID: 1,
	}
}

// Add adds a new ped
func (pm *PedManager) Add(ped *Ped) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.peds[ped.ID] = ped
}

// Remove removes a ped
func (pm *PedManager) Remove(pedID types.PedID) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	delete(pm.peds, pedID)
}

// GetPed retrieves a ped by ID
func (pm *PedManager) GetPed(pedID types.PedID) *Ped {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.peds[pedID]
}

// GetFreeID returns the next available ped ID
func (pm *PedManager) GetFreeID() types.PedID {
	pm.idMutex.Lock()
	defer pm.idMutex.Unlock()
	
	id := pm.nextID
	pm.nextID++
	return id
}

// RemoveAllHostedBy removes all peds created by a specific player
func (pm *PedManager) RemoveAllHostedBy(player *Player) []types.PedID {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	var removedIDs []types.PedID
	for id, ped := range pm.peds {
		if ped.Syncer == player {
			delete(pm.peds, id)
			removedIDs = append(removedIDs, id)
		}
	}
	return removedIDs
}

// IsValidModelID checks if the model ID is within valid range
func (pm *PedManager) IsValidModelID(modelID int16) bool {
	return modelID >= 1 && modelID <= 311
}

// IsSpecialModel checks if the model ID is a special model (290-299)
func (pm *PedManager) IsSpecialModel(modelID int16) bool {
	return modelID >= 290 && modelID <= 299
}

// IsValidSpecialModelName checks if the special model name is allowed
func (pm *PedManager) IsValidSpecialModelName(modelName [8]byte) bool {
	// Convert byte array to string, handling null termination
	nameStr := string(modelName[:])
	if nullIndex := strings.IndexByte(nameStr, 0); nullIndex != -1 {
		nameStr = nameStr[:nullIndex]
	}
	nameStr = strings.ToUpper(nameStr)
	
	// Check against allowed special actors (case-insensitive like _strnicmp)
	for _, allowedName := range allowedSpecialActors {
		if strings.HasPrefix(nameStr, allowedName) {
			return true
		}
	}
	return false
}

// ValidatePedSpawn validates a ped spawn request
func (pm *PedManager) ValidatePedSpawn(modelID int16, specialModelName [8]byte) error {
	if !pm.IsValidModelID(modelID) {
		return fmt.Errorf("invalid model ID: %d (must be 1-311)", modelID)
	}
	
	if pm.IsSpecialModel(modelID) {
		if !pm.IsValidSpecialModelName(specialModelName) {
			nameStr := string(specialModelName[:])
			if nullIndex := strings.IndexByte(nameStr, 0); nullIndex != -1 {
				nameStr = nameStr[:nullIndex]
			}
			return fmt.Errorf("invalid special model name: %s", nameStr)
		}
	}
	
	return nil
}

// GetAllPeds returns all peds (for debugging/admin purposes)
func (pm *PedManager) GetAllPeds() []*Ped {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	peds := make([]*Ped, 0, len(pm.peds))
	for _, ped := range pm.peds {
		peds = append(peds, ped)
	}
	return peds
}
