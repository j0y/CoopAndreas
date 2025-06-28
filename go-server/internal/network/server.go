package network

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/codecat/go-enet"
	"github.com/rs/zerolog"

	"coopandreas-server/internal/types"
)

const (
	MaxPacketSize = 1024
	ServerTimeout = 30 * time.Second
)

// PacketFlag represents ENet packet flags
type PacketFlag uint8

const (
	PacketFlagReliable PacketFlag = 1 << 0
	// Add other flags as needed
)

// Client represents a connected client
type Client struct {
	Addr       *net.UDPAddr
	LastSeen   time.Time
	Reliable   chan []byte // Channel for reliable packets
	Unreliable chan []byte // Channel for unreliable packets
	Connected  bool
	mutex      sync.RWMutex
	// ENet support (required for ENet clients)
	ENetPeer interface{} // Holds enet.Peer for ENet clients
}

// NewClient creates a new client instance
func NewClient(addr *net.UDPAddr) *Client {
	return &Client{
		Addr:       addr,
		LastSeen:   time.Now(),
		Reliable:   make(chan []byte, 100),
		Unreliable: make(chan []byte, 100),
		Connected:  true,
	}
}

// UpdateLastSeen updates the client's last seen time
func (c *Client) UpdateLastSeen() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.LastSeen = time.Now()
}

// IsTimedOut returns true if the client has timed out
func (c *Client) IsTimedOut() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return time.Since(c.LastSeen) > ServerTimeout
}

// NetworkPacket represents a network packet
type NetworkPacket struct {
	ID   types.PacketID
	Data []byte
	Flag PacketFlag
}

// UnmarshalPacket unmarshals a packet from raw bytes (C++ format)
func UnmarshalPacket(data []byte) (*NetworkPacket, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("packet too short: %d bytes", len(data))
	}

	// Read packet ID (2 bytes, little endian)
	id := uint16(data[0]) | (uint16(data[1]) << 8)

	// Read remaining data
	packetData := make([]byte, len(data)-2)
	copy(packetData, data[2:])

	return &NetworkPacket{
		ID:   types.PacketID(id),
		Data: packetData,
	}, nil
}

// GameServer represents the game server that handles packets
type GameServer interface {
	HandlePacket(client *Client, packet *NetworkPacket) error
	HandlePlayerConnect(client *Client)
	HandlePlayerDisconnect(client *Client)
}

// Server represents the ENet-compatible server (replaces old UDP server)
type Server struct {
	host         enet.Host
	clients      map[uint32]*Client
	clientsMutex sync.RWMutex
	gameServer   GameServer
	running      bool
	runningMutex sync.RWMutex
	logger       zerolog.Logger
	nextClientID uint32
}

// NewServer creates a new ENet-compatible server
func NewServer(port int, gameServer GameServer) *Server {
	// Configure zerolog for pretty console output
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}).
		With().
		Timestamp().
		Str("component", "enet").
		Logger()

	return &Server{
		clients:      make(map[uint32]*Client),
		gameServer:   gameServer,
		logger:       logger,
		nextClientID: 1,
	}
}

// Start starts the ENet server
func (s *Server) Start() error {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	// Initialize ENet
	enet.Initialize()

	// Create listen address
	addr := enet.NewListenAddress(uint16(6767))

	// Create ENet host
	host, err := enet.NewHost(addr, 32, 2, 0, 0) // 32 peers, 2 channels, no bandwidth limits
	if err != nil {
		enet.Deinitialize()
		return fmt.Errorf("failed to create ENet host: %w", err)
	}

	s.host = host
	s.running = true

	s.logger.Info().Int("port", 6767).Msg("ENet server started - compatible with C++ clients")

	// Start the main server loop
	go s.run()

	return nil
}

// Stop stops the ENet server
func (s *Server) Stop() {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()

	if !s.running {
		return
	}

	s.running = false

	if s.host != nil {
		s.host.Destroy()
	}
	enet.Deinitialize()

	s.logger.Info().Msg("ENet server stopped")
}

// IsRunning returns true if the server is running
func (s *Server) IsRunning() bool {
	s.runningMutex.RLock()
	defer s.runningMutex.RUnlock()
	return s.running
}

// run is the main server loop
func (s *Server) run() {
	for s.IsRunning() {
		event := s.host.Service(100) // 100ms timeout
		if event == nil {
			continue
		}

		switch event.GetType() {
		case enet.EventConnect:
			s.handleConnect(event)
		case enet.EventDisconnect:
			s.handleDisconnect(event)
		case enet.EventReceive:
			s.handleReceive(event)
		}
	}
}

// handleConnect handles new peer connections
func (s *Server) handleConnect(event enet.Event) {
	peer := event.GetPeer()
	clientID := s.nextClientID
	s.nextClientID++

	// Get actual client address from ENet peer
	enetAddr := peer.GetAddress()
	addrStr := enetAddr.String()

	// Parse the address string to create a UDP address
	// ENet address format is typically "IP:PORT"
	udpAddr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		// Fallback to placeholder if parsing fails
		s.logger.Warn().
			Str("enetAddr", addrStr).
			Err(err).
			Msg("Failed to parse ENet address, using placeholder")
		udpAddr = &net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 0,
		}
	}

	client := NewClient(udpAddr)
	client.ENetPeer = peer // Store the ENet peer for sending packets

	s.clientsMutex.Lock()
	s.clients[clientID] = client
	s.clientsMutex.Unlock()

	// Store client ID in peer data (as bytes)
	clientIDBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(clientIDBytes, clientID)
	peer.SetData(clientIDBytes)

	s.logger.Info().
		Uint32("clientID", clientID).
		Str("clientAddr", client.Addr.String()).
		Str("enetAddr", addrStr).
		Msg("ENet client connected")

	// Call the connection handler
	s.gameServer.HandlePlayerConnect(client)
}

// handleDisconnect handles peer disconnections
func (s *Server) handleDisconnect(event enet.Event) {
	peer := event.GetPeer()
	clientIDBytes := peer.GetData()

	if len(clientIDBytes) < 4 {
		s.logger.Warn().Msg("Invalid client ID data on disconnect")
		return
	}

	clientID := binary.LittleEndian.Uint32(clientIDBytes)

	s.clientsMutex.Lock()
	client, exists := s.clients[clientID]
	if exists {
		delete(s.clients, clientID)
	}
	s.clientsMutex.Unlock()

	if exists {
		s.logger.Info().
			Uint32("clientID", clientID).
			Msg("ENet client disconnected")

		// Call the disconnection handler
		s.gameServer.HandlePlayerDisconnect(client)
	}

	// Clear peer data
	peer.SetData(nil)
}

// handleReceive handles received packets
func (s *Server) handleReceive(event enet.Event) {
	peer := event.GetPeer()
	clientIDBytes := peer.GetData()

	if len(clientIDBytes) < 4 {
		s.logger.Warn().Msg("Invalid client ID data on receive")
		return
	}

	clientID := binary.LittleEndian.Uint32(clientIDBytes)

	s.clientsMutex.RLock()
	client, exists := s.clients[clientID]
	s.clientsMutex.RUnlock()

	if !exists {
		s.logger.Warn().Uint32("clientID", clientID).Msg("Received packet from unknown client")
		return
	}

	client.UpdateLastSeen()

	packet := event.GetPacket()
	data := packet.GetData()

	gamePacket, err := UnmarshalPacket(data)
	if err != nil {
		s.logger.Error().Err(err).Uint32("clientID", clientID).Msg("Failed to unmarshal game packet")
		packet.Destroy()
		return
	}

	//s.logger.Debug().
	//	Uint16("packetID", uint16(gamePacket.ID)).
	//	Uint32("clientID", clientID).
	//	Msg("Received ENet packet")

	// Handle the packet
	if err := s.gameServer.HandlePacket(client, gamePacket); err != nil {
		s.logger.Error().
			Err(err).
			Uint16("packetID", uint16(gamePacket.ID)).
			Uint32("clientID", clientID).
			Msg("Error handling packet")
	}

	// Destroy the packet
	packet.Destroy()
}

// SendPacket sends a packet to a specific client (ENet version)
func (s *Server) SendPacket(client *Client, packet *NetworkPacket) error {
	if client.ENetPeer == nil {
		return fmt.Errorf("client does not have an ENet peer")
	}

	// Cast the interface{} to enet.Peer
	peer, ok := client.ENetPeer.(enet.Peer)
	if !ok {
		return fmt.Errorf("client ENetPeer is not an enet.Peer")
	}

	// Marshal the game packet (C++ format: 2-byte ID + data)
	data := make([]byte, 2+len(packet.Data))
	binary.LittleEndian.PutUint16(data[:2], uint16(packet.ID))
	copy(data[2:], packet.Data)

	// Create ENet packet
	var flags enet.PacketFlags
	if packet.Flag&PacketFlagReliable != 0 {
		flags = enet.PacketFlagReliable
	}

	enetPacket, err := enet.NewPacket(data, flags)
	if err != nil {
		return fmt.Errorf("failed to create ENet packet: %w", err)
	}

	// Send packet
	return peer.SendPacket(enetPacket, 0) // Channel 0
}

// BroadcastPacket sends a packet to all connected clients
func (s *Server) BroadcastPacket(packet *NetworkPacket) {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	for _, client := range s.clients {
		if err := s.SendPacket(client, packet); err != nil {
			s.logger.Error().
				Err(err).
				Str("client", client.Addr.String()).
				Msg("Failed to broadcast packet to client")
		}
	}
}

// BroadcastPacketExclude sends a packet to all connected clients except the excluded one
func (s *Server) BroadcastPacketExclude(packet *NetworkPacket, excludeClient *Client) {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	for _, client := range s.clients {
		if client.Addr.String() != excludeClient.Addr.String() {
			if err := s.SendPacket(client, packet); err != nil {
				s.logger.Error().
					Err(err).
					Str("client", client.Addr.String()).
					Msg("Failed to broadcast packet to client")
			}
		}
	}
}

// SendPacketToAll sends a packet to all clients, optionally excluding one
func (s *Server) SendPacketToAll(packet *NetworkPacket, excludeClient *Client) error {
	if excludeClient == nil {
		s.BroadcastPacket(packet)
	} else {
		s.BroadcastPacketExclude(packet, excludeClient)
	}
	return nil
}

// GetClientCount returns the number of connected clients
func (s *Server) GetClientCount() int {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	return len(s.clients)
}
