package server

import (
	"fmt"
	"strings"

	"coopandreas-server/internal/network"
	"coopandreas-server/internal/packets"
	"coopandreas-server/internal/types"
)

// handleCheckVersion handles CHECK_VERSION packets
func (s *Server) handleCheckVersion(client *network.Client, data []byte) error {
	// Parse the version check packet
	var packet packets.CheckVersionPacket
	if err := packet.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal CheckVersion packet: %w", err)
	}

	// Extract client version string (null-terminated)
	clientVersionStr := string(packet.ClientVersion[:])
	if nullIndex := strings.IndexByte(clientVersionStr, 0); nullIndex != -1 {
		clientVersionStr = clientVersionStr[:nullIndex]
	}
	clientVersionStr = strings.TrimSpace(clientVersionStr)

	s.logger.Info().
		Uint32("clientID", client.ID).
		Str("clientVersion", clientVersionStr).
		Uint32("protocolVersion", packet.ProtocolVersion).
		Msg("Version check request received")

	// Validate client version
	isCompatible, message, err := s.versionManager.ValidateClientVersion(clientVersionStr)
	if err != nil {
		s.logger.Error().
			Err(err).
			Uint32("clientID", client.ID).
			Str("clientVersion", clientVersionStr).
			Msg("Failed to validate client version")

		// Send incompatible response due to validation error
		isCompatible = false
		message = "Version validation failed"
	}

	// Create response packet
	var responseMessage string
	if isCompatible {
		responseMessage = "Welcome to CoopAndreas Server!"
		s.logger.Info().
			Uint32("clientID", client.ID).
			Str("clientVersion", clientVersionStr).
			Msg("Client version accepted")
	} else {
		responseMessage = message
		s.logger.Warn().
			Uint32("clientID", client.ID).
			Str("clientVersion", clientVersionStr).
			Str("reason", message).
			Msg("Client version rejected")
	}

	// Send version check response
	response := packets.NewCheckVersionResponse(isCompatible, responseMessage)
	responseData, err := response.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal version response: %w", err)
	}

	networkPacket := &network.NetworkPacket{
		ID:   types.CHECK_VERSION, // Respond with same packet ID
		Data: responseData,
		Flag: network.PacketFlagReliable,
	}

	if err := s.networkServer.SendPacket(client, networkPacket); err != nil {
		return fmt.Errorf("failed to send version response: %w", err)
	}

	// If version is incompatible, we might want to disconnect the client
	// For now, we'll let them stay connected but log the incompatibility
	if !isCompatible {
		s.logger.Info().
			Uint32("clientID", client.ID).
			Msg("Client with incompatible version remains connected (warning only)")
	}

	return nil
}
