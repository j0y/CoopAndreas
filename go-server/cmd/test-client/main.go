package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"coopandreas-server/internal/packets"
	"coopandreas-server/internal/types"
)

func main() {
	// Configure logger for test client
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	log.Info().Msg("CoopAndreas Go Server Test Client")

	// Connect to server
	conn, err := net.Dial("udp", "localhost:6767")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect")
		return
	}
	defer conn.Close()

	log.Info().Msg("Connected to server")

	// Test CheckVersion packet first
	testCheckVersion(conn)

	// Wait a bit
	time.Sleep(1 * time.Second)

	// Test PlayerGetName packet
	testPlayerGetName(conn)

	// Wait a bit
	time.Sleep(1 * time.Second)

	// Test PedSpawn packet
	testPedSpawn(conn)

	// Wait a bit
	time.Sleep(2 * time.Second)

	// Test PedRemove packet
	testPedRemove(conn)

	// Wait a bit
	time.Sleep(1 * time.Second)

	// Test PlayerGetName packet
	testPlayerGetName(conn)

	log.Info().Msg("Test completed")
}

func testPedSpawn(conn net.Conn) {
	log.Info().Msg("Testing PedSpawn packet")

	// Create a PedSpawn packet
	pedSpawn := packets.PedSpawnPacket{
		PedID:     0, // Will be assigned by server
		TempID:    1,
		ModelID:   290, // Special model requiring name validation
		PedType:   4,
		Position:  types.Vector3{X: 100.0, Y: 200.0, Z: 10.0},
		CreatedBy: 1,
	}

	// Set special model name (should be valid)
	copy(pedSpawn.SpecialModelName[:], "ANDRE")

	// Marshal packet data
	packetData, err := pedSpawn.Marshal()
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal PedSpawn")
		return
	}

	// Create network packet
	networkPacket := createNetworkPacket(types.PED_SPAWN, packetData)

	// Send packet
	if _, err := conn.Write(networkPacket); err != nil {
		log.Error().Err(err).Msg("Failed to send PedSpawn")
		return
	}

	log.Info().Msg("PedSpawn packet sent")
}

func testPedRemove(conn net.Conn) {
	log.Info().Msg("Testing PedRemove packet")

	// Create a PedRemove packet (assuming ped ID 1 was created)
	pedRemove := packets.PedRemovePacket{
		PedID: 1,
	}

	// Marshal packet data
	packetData, err := pedRemove.Marshal()
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal PedRemove")
		return
	}

	// Create network packet
	networkPacket := createNetworkPacket(types.PED_REMOVE, packetData)

	// Send packet
	if _, err := conn.Write(networkPacket); err != nil {
		log.Error().Err(err).Msg("Failed to send PedRemove")
		return
	}

	log.Info().Msg("PedRemove packet sent")
}

func testCheckVersion(conn net.Conn) {
	log.Info().Msg("Testing CheckVersion packet")

	// Create a CheckVersion packet
	versionCheck := packets.CheckVersionPacket{
		ProtocolVersion: types.ProtocolVersion,
	}

	// Set client version
	copy(versionCheck.ClientVersion[:], types.ServerVersion) // Use server version as client version for testing

	// Marshal packet data
	packetData, err := versionCheck.Marshal()
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal CheckVersion")
		return
	}

	// Create network packet
	networkPacket := createNetworkPacket(types.CHECK_VERSION, packetData)

	// Send packet
	if _, err := conn.Write(networkPacket); err != nil {
		log.Error().Err(err).Msg("Failed to send CheckVersion")
		return
	}

	log.Info().Msg("CheckVersion packet sent")
}

func testPlayerGetName(conn net.Conn) {
	log.Info().Msg("Testing PlayerGetName packet")

	// Create a PlayerGetName packet
	playerGetName := packets.NewPlayerGetNamePacket(1, "TestPlayer")

	// Marshal packet data
	packetData, err := playerGetName.Marshal()
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal PlayerGetName")
		return
	}

	// Create network packet
	networkPacket := createNetworkPacket(types.PLAYER_GET_NAME, packetData)

	// Send packet
	if _, err := conn.Write(networkPacket); err != nil {
		log.Error().Err(err).Msg("Failed to send PlayerGetName")
		return
	}

	log.Info().Msg("PlayerGetName packet sent")
}

func createNetworkPacket(id types.PacketID, data []byte) []byte {
	buf := new(bytes.Buffer)

	// Write packet ID (2 bytes)
	binary.Write(buf, binary.LittleEndian, uint16(id))

	// Write packet data
	buf.Write(data)

	return buf.Bytes()
}
