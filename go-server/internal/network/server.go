package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

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
	Addr         *net.UDPAddr
	LastSeen     time.Time
	Reliable     chan []byte // Channel for reliable packets
	Unreliable   chan []byte // Channel for unreliable packets
	Connected    bool
	mutex        sync.RWMutex
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

// UpdateLastSeen updates the client's last seen timestamp
func (c *Client) UpdateLastSeen() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.LastSeen = time.Now()
}

// IsConnected checks if the client is still connected
func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.Connected && time.Since(c.LastSeen) < ServerTimeout
}

// Disconnect marks the client as disconnected
func (c *Client) Disconnect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Connected = false
	close(c.Reliable)
	close(c.Unreliable)
}

// NetworkPacket represents a network packet
type NetworkPacket struct {
	ID   types.PacketID
	Data []byte
	Flag PacketFlag
}

// Marshal serializes the packet for network transmission
func (p *NetworkPacket) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	
	// Write packet ID (2 bytes)
	if err := binary.Write(buf, binary.LittleEndian, uint16(p.ID)); err != nil {
		return nil, err
	}
	
	// Write packet data
	if _, err := buf.Write(p.Data); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// UnmarshalPacket deserializes a network packet
func UnmarshalPacket(data []byte) (*NetworkPacket, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("packet too short: %d bytes", len(data))
	}
	
	buf := bytes.NewReader(data)
	
	// Read packet ID (2 bytes)
	var id uint16
	if err := binary.Read(buf, binary.LittleEndian, &id); err != nil {
		return nil, err
	}
	
	// Read remaining data
	packetData := make([]byte, len(data)-2)
	if _, err := buf.Read(packetData); err != nil {
		return nil, err
	}
	
	return &NetworkPacket{
		ID:   types.PacketID(id),
		Data: packetData,
	}, nil
}

// Server represents the UDP server
type Server struct {
	conn          *net.UDPConn
	clients       map[string]*Client
	clientsMutex  sync.RWMutex
	packetHandler PacketHandler
	running       bool
	runningMutex  sync.RWMutex
	logger        zerolog.Logger
}

// PacketHandler interface for handling different packet types
type PacketHandler interface {
	HandlePacket(client *Client, packet *NetworkPacket) error
}

// NewServer creates a new UDP server
func NewServer(port int, handler PacketHandler) *Server {
	// Configure zerolog for pretty console output
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}).
		With().
		Timestamp().
		Str("component", "network").
		Logger()
	
	return &Server{
		clients:       make(map[string]*Client),
		packetHandler: handler,
		logger:        logger,
	}
}

// Start starts the UDP server
func (s *Server) Start() error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", 6767))
	if err != nil {
		return fmt.Errorf("failed to resolve address: %w", err)
	}
	
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %w", err)
	}
	
	s.conn = conn
	s.setRunning(true)
	
	s.logger.Info().Str("address", addr.String()).Msg("Server listening")
	
	// Start cleanup goroutine
	go s.cleanupLoop()
	
	// Main packet receiving loop
	buffer := make([]byte, MaxPacketSize)
	for s.isRunning() {
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Timeout is expected for graceful shutdown
			}
			if s.isRunning() {
				s.logger.Error().Err(err).Msg("Error reading UDP packet")
			}
			continue
		}
		
		// Handle packet in goroutine
		go s.handlePacket(clientAddr, buffer[:n])
	}
	
	return nil
}

// Stop stops the server
func (s *Server) Stop() {
	s.setRunning(false)
	
	if s.conn != nil {
		s.conn.Close()
	}
	
	// Disconnect all clients
	s.clientsMutex.Lock()
	for _, client := range s.clients {
		client.Disconnect()
	}
	s.clients = make(map[string]*Client)
	s.clientsMutex.Unlock()
}

// handlePacket processes incoming packets
func (s *Server) handlePacket(clientAddr *net.UDPAddr, data []byte) {
	client := s.getOrCreateClient(clientAddr)
	client.UpdateLastSeen()
	
	packet, err := UnmarshalPacket(data)
	if err != nil {
		s.logger.Error().Err(err).Str("client", clientAddr.String()).Msg("Failed to unmarshal packet")
		return
	}
	
	s.logger.Debug().
		Uint16("packetID", uint16(packet.ID)).
		Str("client", clientAddr.String()).
		Msg("Received packet")
	
	// Handle the packet
	if err := s.packetHandler.HandlePacket(client, packet); err != nil {
		s.logger.Error().
			Err(err).
			Uint16("packetID", uint16(packet.ID)).
			Str("client", clientAddr.String()).
			Msg("Error handling packet")
	}
}

// getOrCreateClient gets or creates a client for the given address
func (s *Server) getOrCreateClient(addr *net.UDPAddr) *Client {
	addrStr := addr.String()
	
	s.clientsMutex.RLock()
	if client, exists := s.clients[addrStr]; exists {
		s.clientsMutex.RUnlock()
		return client
	}
	s.clientsMutex.RUnlock()
	
	// Create new client
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	
	// Double-check in case another goroutine created it
	if client, exists := s.clients[addrStr]; exists {
		return client
	}
	
	client := NewClient(addr)
	s.clients[addrStr] = client
	s.logger.Info().Str("client", addr.String()).Msg("New client connected")
	
	return client
}

// SendPacket sends a packet to a specific client
func (s *Server) SendPacket(client *Client, packet *NetworkPacket) error {
	data, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal packet: %w", err)
	}
	
	_, err = s.conn.WriteToUDP(data, client.Addr)
	if err != nil {
		return fmt.Errorf("failed to send packet: %w", err)
	}
	
	s.logger.Debug().
		Uint16("packetID", uint16(packet.ID)).
		Str("client", client.Addr.String()).
		Msg("Sent packet")
	return nil
}

// SendPacketToAll sends a packet to all connected clients except the excluded one
func (s *Server) SendPacketToAll(packet *NetworkPacket, exclude *Client) error {
	data, err := packet.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal packet: %w", err)
	}
	
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	
	for _, client := range s.clients {
		if exclude != nil && client == exclude {
			continue
		}
		
		if !client.IsConnected() {
			continue
		}
		
		if _, err := s.conn.WriteToUDP(data, client.Addr); err != nil {
			s.logger.Error().Err(err).Str("client", client.Addr.String()).Msg("Failed to send packet")
		}
	}
	
	s.logger.Debug().
		Uint16("packetID", uint16(packet.ID)).
		Int("clientCount", len(s.clients)).
		Msg("Broadcast packet to all clients")
	return nil
}

// cleanupLoop removes disconnected clients
func (s *Server) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for s.isRunning() {
		select {
		case <-ticker.C:
			s.cleanupClients()
		}
	}
}

// cleanupClients removes disconnected clients
func (s *Server) cleanupClients() {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	
	for addr, client := range s.clients {
		if !client.IsConnected() {
			client.Disconnect()
			delete(s.clients, addr)
			s.logger.Info().Str("client", addr).Msg("Cleaned up disconnected client")
		}
	}
}

// isRunning checks if the server is running
func (s *Server) isRunning() bool {
	s.runningMutex.RLock()
	defer s.runningMutex.RUnlock()
	return s.running
}

// setRunning sets the running state
func (s *Server) setRunning(running bool) {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()
	s.running = running
}

// GetClients returns all connected clients
func (s *Server) GetClients() []*Client {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	
	clients := make([]*Client, 0, len(s.clients))
	for _, client := range s.clients {
		if client.IsConnected() {
			clients = append(clients, client)
		}
	}
	return clients
}
