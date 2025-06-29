package main

import (
	"encoding/binary"
	"flag"
	"math"
	"os"
	"time"

	"github.com/codecat/go-enet"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"coopandreas-server/internal/packets"
	"coopandreas-server/internal/types"
)

type DummyClient struct {
	host         enet.Host
	peer         enet.Peer
	playerName   string
	position     types.Vector3
	angle        float32
	moveSpeed    float32
	running      bool
	connected    bool
	playerID     types.PlayerID
	lastUpdate   time.Time
	movementTime float64 // Track movement time for circular motion
}

func main() {
	// Configure logger for dummy client
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	// Define command line flags
	var (
		serverIP   = flag.String("ip", "localhost", "Server IP address")
		serverPort = flag.Int("port", 6767, "Server port")
		playerName = flag.String("name", "DummyPlayer", "Player name")
	)
	flag.Parse()

	log.Info().
		Str("ip", *serverIP).
		Int("port", *serverPort).
		Str("name", *playerName).
		Msg("CoopAndreas Dummy Client Starting")

	// Initialize ENet
	enet.Initialize()
	defer enet.Deinitialize()

	// Create ENet host for client
	host, err := enet.NewHost(nil, 1, 2, 0, 0) // No address (client), 1 peer, 2 channels, no bandwidth limits
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create ENet host")
		return
	}
	defer host.Destroy()

	// Connect to server
	addr := enet.NewAddress(*serverIP, uint16(*serverPort))
	peer, err := host.Connect(addr, 2, 0) // 2 channels, no data
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to server")
		return
	}

	client := &DummyClient{
		host:         host,
		peer:         peer,
		playerName:   *playerName,
		position:     types.Vector3{X: 1000.0, Y: 1500.0, Z: 3.0}, // Start at spawn point
		angle:        0.0,
		moveSpeed:    2.0, // Units per second
		running:      true,
		connected:    false,
		playerID:     0, // Will be set by server
		lastUpdate:   time.Now(),
		movementTime: 0.0,
	}

	log.Info().Msg("Connecting to server...")

	// Start the client simulation
	client.run()
}

func (c *DummyClient) run() {
	// Wait for connection to be established
	connectionTimeout := time.NewTimer(5 * time.Second)
	defer connectionTimeout.Stop()

	for !c.connected && c.running {
		select {
		case <-connectionTimeout.C:
			log.Error().Msg("Connection timeout")
			c.running = false
			return
		default:
			event := c.host.Service(100) // 100ms timeout
			if event != nil {
				switch event.GetType() {
				case enet.EventConnect:
					log.Info().Msg("Connected to server")
					c.connected = true
				case enet.EventDisconnect:
					log.Error().Msg("Disconnected from server during connection")
					c.running = false
					return
				case enet.EventReceive:
					// Handle any early packets
					packet := event.GetPacket()
					log.Debug().Msg("Received early packet")
					packet.Destroy()
				}
			}
		}
	}

	if !c.connected {
		log.Error().Msg("Failed to connect to server")
		return
	}

	// Step 1: Send version check
	if !c.sendVersionCheck() {
		return
	}

	// Wait a bit for server response
	time.Sleep(500 * time.Millisecond)

	// Step 2: Send player name
	if !c.sendPlayerName() {
		return
	}

	// Wait a bit for server to process
	time.Sleep(500 * time.Millisecond)

	log.Info().Msg("Starting position updates...")

	// Step 3: Start sending position updates in a loop
	ticker := time.NewTicker(2 * time.Second) // Every 2 seconds
	defer ticker.Stop()

	// Send initial position update
	c.updatePosition()
	if !c.sendPositionUpdate() {
		log.Error().Msg("Failed to send initial position update")
		return
	}

	for c.running {
		// Process ENet events with a small timeout
		event := c.host.Service(50) // 50ms timeout
		if event != nil {
			switch event.GetType() {
			case enet.EventDisconnect:
				log.Info().Msg("Disconnected from server")
				c.running = false
				return
			case enet.EventReceive:
				// Handle received packets (we're not processing server responses in this dummy client)
				packet := event.GetPacket()
				log.Debug().
					Int("size", len(packet.GetData())).
					Msg("Received packet from server")
				packet.Destroy()
			}
		}

		// Check if it's time to send position update (non-blocking)
		select {
		case <-ticker.C:
			log.Info().Msg("Sending scheduled position update")
			c.updatePosition()
			if !c.sendPositionUpdate() {
				log.Error().Msg("Failed to send position update, stopping client")
				return
			}
		default:
			// No position update needed right now, continue processing events
		}
	}

	log.Info().Msg("Client stopped")
}

func (c *DummyClient) sendVersionCheck() bool {
	log.Info().Msg("Sending version check")

	// Create a CheckVersion packet
	versionCheck := packets.CheckVersionPacket{
		ProtocolVersion: types.ProtocolVersion,
	}

	// Set client version
	copy(versionCheck.ClientVersion[:], types.ServerVersion)

	// Marshal packet data
	packetData, err := versionCheck.Marshal()
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal CheckVersion")
		return false
	}

	// Create and send network packet
	return c.sendPacket(types.CHECK_VERSION, packetData)
}

func (c *DummyClient) sendPlayerName() bool {
	log.Info().Str("name", c.playerName).Msg("Sending player name")

	// Create a PlayerGetName packet (this is how clients send their names)
	playerGetName := packets.NewPlayerGetNamePacket(c.playerID, c.playerName)

	// Marshal packet data
	packetData, err := playerGetName.Marshal()
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal PlayerGetName")
		return false
	}

	// Create and send network packet
	return c.sendPacket(types.PLAYER_GET_NAME, packetData)
}

func (c *DummyClient) updatePosition() {
	now := time.Now()
	deltaTime := now.Sub(c.lastUpdate).Seconds()
	c.lastUpdate = now

	// Update movement time for smooth motion
	c.movementTime += deltaTime

	// Calculate movement based on current angle and speed
	// Move forward at moveSpeed units per second
	moveDistance := float32(deltaTime) * c.moveSpeed

	// Calculate velocity based on current angle
	velocityX := moveDistance * float32(math.Cos(float64(c.angle)))
	velocityY := moveDistance * float32(math.Sin(float64(c.angle)))

	// Update position based on previous position + velocity
	c.position.X += velocityX
	c.position.Y += velocityY

	// Add some Z-axis bobbing for visual effect
	c.position.Z = 3.0 + float32(math.Sin(c.movementTime*2.0))*0.5 // Subtle bobbing

	// Gradually turn the player (complete rotation every ~30 seconds)
	angleChange := float32(deltaTime * 0.2) // Radians per second
	c.angle += angleChange

	// Keep angle in 0-2π range
	if c.angle > 2*math.Pi {
		c.angle -= 2 * math.Pi
	}
}

func (c *DummyClient) sendPositionUpdate() bool {
	// Create a PlayerOnFoot packet
	onFootPacket := packets.PlayerOnFootPacket{
		ID:            c.playerID,
		Position:      c.position,
		Velocity:      types.Vector3{X: 0.0, Y: 0.0, Z: 0.0}, // Standing still
		Rotation:      c.angle,
		Health:        100, // Full health
		Armour:        0,   // No armour
		Weapon:        0,   // No weapon
		Ammo:          0,   // No ammo
		Ducking:       0,   // Not ducking
		HasJetpack:    0,   // No jetpack
		FightingStyle: 4,   // Default fighting style
	}

	// Marshal packet data
	packetData, err := onFootPacket.Marshal()
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal PlayerOnFoot")
		return false
	}

	log.Debug().
		Float32("x", c.position.X).
		Float32("y", c.position.Y).
		Float32("z", c.position.Z).
		Float32("angle", c.angle).
		Msg("Sending position update")

	// Create and send network packet
	return c.sendPacket(types.PLAYER_ONFOOT, packetData)
}

func (c *DummyClient) sendPacket(id types.PacketID, data []byte) bool {
	if !c.connected {
		log.Error().Msg("Cannot send packet: not connected")
		return false
	}

	// Create network packet data (matching server format: 2-byte ID + data)
	packetData := make([]byte, 2+len(data))
	binary.LittleEndian.PutUint16(packetData[:2], uint16(id))
	copy(packetData[2:], data)

	// Create ENet packet (reliable for all packets in this dummy client)
	enetPacket, err := enet.NewPacket(packetData, enet.PacketFlagReliable)
	if err != nil {
		log.Error().Err(err).Uint16("packetID", uint16(id)).Msg("Failed to create ENet packet")
		return false
	}

	// Send packet on channel 0
	if err := c.peer.SendPacket(enetPacket, 0); err != nil {
		log.Error().Err(err).Uint16("packetID", uint16(id)).Msg("Failed to send ENet packet")
		return false
	}

	return true
}
