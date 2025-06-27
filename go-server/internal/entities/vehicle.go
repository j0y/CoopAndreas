package entities

import (
	"fmt"
	"sync"

	"coopandreas-server/internal/types"
)

// Vehicle represents a network vehicle entity
type Vehicle struct {
	ID             types.VehicleID // Unique vehicle ID
	ModelID        uint16          // Vehicle model ID (400-611 for SA)
	Position       types.Vector3   // Current position
	Rotation       float32         // Current rotation angle
	PrimaryColor   uint8           // Primary color
	SecondaryColor uint8           // Secondary color
	Health         float32         // Vehicle health (0.0 - 1000.0)
	Syncer         *Player         // Player responsible for syncing this vehicle
	CreatedBy      uint8           // Who created this vehicle
	Active         bool            // Whether the vehicle is active
}

// VehicleManager manages all network vehicles
type VehicleManager struct {
	vehicles map[types.VehicleID]*Vehicle
	nextID   types.VehicleID
	mutex    sync.RWMutex
}

// NewVehicleManager creates a new vehicle manager
func NewVehicleManager() *VehicleManager {
	return &VehicleManager{
		vehicles: make(map[types.VehicleID]*Vehicle),
		nextID:   1, // Start from 1, 0 is typically invalid
	}
}

// GetFreeID gets the next available vehicle ID
func (vm *VehicleManager) GetFreeID() types.VehicleID {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	// Find a free ID (matching C++ CVehicleManager::GetFreeID logic)
	for {
		id := vm.nextID
		vm.nextID++

		// Wrap around if we hit a reasonable limit
		if vm.nextID > 1000 {
			vm.nextID = 1
		}

		if _, exists := vm.vehicles[id]; !exists {
			return id
		}
	}
}

// AddVehicle adds a new vehicle to the manager
func (vm *VehicleManager) AddVehicle(vehicle *Vehicle) error {
	if vehicle == nil {
		return fmt.Errorf("vehicle cannot be nil")
	}

	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	if _, exists := vm.vehicles[vehicle.ID]; exists {
		return fmt.Errorf("vehicle with ID %d already exists", vehicle.ID)
	}

	vehicle.Active = true
	vm.vehicles[vehicle.ID] = vehicle
	return nil
}

// GetVehicle retrieves a vehicle by ID
func (vm *VehicleManager) GetVehicle(id types.VehicleID) *Vehicle {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()

	return vm.vehicles[id]
}

// RemoveVehicle removes a vehicle from the manager
func (vm *VehicleManager) RemoveVehicle(id types.VehicleID) *Vehicle {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	vehicle := vm.vehicles[id]
	if vehicle != nil {
		vehicle.Active = false
		delete(vm.vehicles, id)
	}
	return vehicle
}

// GetAllVehicles returns a slice of all vehicles
func (vm *VehicleManager) GetAllVehicles() []*Vehicle {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()

	vehicles := make([]*Vehicle, 0, len(vm.vehicles))
	for _, vehicle := range vm.vehicles {
		vehicles = append(vehicles, vehicle)
	}
	return vehicles
}

// GetVehicleCount returns the number of active vehicles
func (vm *VehicleManager) GetVehicleCount() int {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()

	return len(vm.vehicles)
}

// GetVehiclesBySyncer returns all vehicles synced by a specific player
func (vm *VehicleManager) GetVehiclesBySyncer(syncer *Player) []*Vehicle {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()

	var vehicles []*Vehicle
	for _, vehicle := range vm.vehicles {
		if vehicle.Syncer != nil && vehicle.Syncer.ID == syncer.ID {
			vehicles = append(vehicles, vehicle)
		}
	}
	return vehicles
}

// RemoveVehiclesBySyncer removes all vehicles synced by a specific player
func (vm *VehicleManager) RemoveVehiclesBySyncer(syncer *Player) []*Vehicle {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	var removedVehicles []*Vehicle
	for id, vehicle := range vm.vehicles {
		if vehicle.Syncer != nil && vehicle.Syncer.ID == syncer.ID {
			vehicle.Active = false
			removedVehicles = append(removedVehicles, vehicle)
			delete(vm.vehicles, id)
		}
	}
	return removedVehicles
}

// NewVehicle creates a new vehicle entity
func NewVehicle(id types.VehicleID, modelID uint16, pos types.Vector3, rot float32) *Vehicle {
	return &Vehicle{
		ID:             id,
		ModelID:        modelID,
		Position:       pos,
		Rotation:       rot,
		PrimaryColor:   0,
		SecondaryColor: 0,
		Health:         1000.0, // Full health
		Active:         false,  // Will be set to true when added to manager
	}
}

// IsValidVehicleModel checks if a model ID is a valid vehicle model
func IsValidVehicleModel(modelID uint16) bool {
	// In GTA San Andreas, vehicle models are from 400 to 611
	return modelID >= 400 && modelID <= 611
}
