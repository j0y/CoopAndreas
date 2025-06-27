package server

import (
	"fmt"

	"coopandreas-server/internal/network"
	"coopandreas-server/internal/types"
)

// handleMassPacketSequence handles MASS_PACKET_SEQUENCE packets
// Based on C++ server logic: simply rebroadcast the entire mass packet to all other clients
// This matches the C++ implementation in CNetwork::HandlePacketReceive
func (s *Server) handleMassPacketSequence(client *network.Client, data []byte) error {
	// In C++: CNetwork::SendPacketRawToAll(event.packet->data, event.packet->dataLength, (ENetPacketFlag)event.packet->flags, event.peer);
	// The server simply rebroadcasts the entire mass packet sequence to all other clients
	// without parsing or processing the individual packets within it

	s.logger.Debug().
		Str("client", client.Addr.String()).
		Int("dataSize", len(data)).
		Msg("Received mass packet sequence, rebroadcasting to all clients")

	// Create network packet for rebroadcast
	networkPacket := &network.NetworkPacket{
		ID:   types.MASS_PACKET_SEQUENCE,
		Data: data,
		Flag: 0, // Use same flags as original packet (for now, use unreliable)
	}

	// Rebroadcast to all other clients (excluding sender)
	if err := s.networkServer.SendPacketToAll(networkPacket, client); err != nil {
		return fmt.Errorf("failed to rebroadcast mass packet sequence: %w", err)
	}

	return nil
}
