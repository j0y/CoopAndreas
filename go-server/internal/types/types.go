package types

// PacketID represents different packet types
type PacketID uint16

const (
	// Packet IDs matching the C++ implementation
	CHECK_VERSION PacketID = iota
	PLAYER_CONNECTED
	PLAYER_DISCONNECTED
	PLAYER_ONFOOT
	PLAYER_BULLET_SHOT
	PLAYER_HANDSHAKE
	PLAYER_PLACE_WAYPOINT
	PLAYER_GET_NAME
	VEHICLE_SPAWN
	PLAYER_SET_HOST
	ADD_EXPLOSION
	VEHICLE_REMOVE
	VEHICLE_IDLE_UPDATE
	VEHICLE_DRIVER_UPDATE
	VEHICLE_ENTER
	VEHICLE_EXIT
	VEHICLE_DAMAGE
	VEHICLE_COMPONENT_ADD
	VEHICLE_COMPONENT_REMOVE
	VEHICLE_PASSENGER_UPDATE
	PLAYER_CHAT_MESSAGE
	PED_SPAWN
	PED_REMOVE
	PED_ONFOOT
	PED_DRIVER_UPDATE
	PED_PASSENGER_UPDATE
	PED_SHOT_SYNC
	PED_ADD_TASK
	PED_CONFIRM
	MASS_PACKET_SEQUENCE = 255
)

// Vector3 represents a 3D position/vector
type Vector3 struct {
	X, Y, Z float32
}

// PlayerID represents a unique player identifier
type PlayerID int32

// PedID represents a unique pedestrian identifier  
type PedID int32

// VehicleID represents a unique vehicle identifier
type VehicleID int32
