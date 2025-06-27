package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"coopandreas-server/internal/types"
)

// MassPacketSequence represents a MASS_PACKET_SEQUENCE packet
// This packet contains multiple packets bundled together for efficiency
// Format: [packet_count:1][packet_id:2][packet_data:variable]...
type MassPacketSequence struct {
	PacketCount uint8             // Number of packets in the sequence
	Packets     []MassPacketEntry // Individual packets
}

// MassPacketEntry represents a single packet within a mass packet sequence
type MassPacketEntry struct {
	ID   types.PacketID // Packet ID
	Data []byte         // Packet data
}

// MassPacketBuilder helps build mass packet sequences
// Equivalent to CMassPacketBuilder in C++
type MassPacketBuilder struct {
	packets []MassPacketEntry
}

// NewMassPacketBuilder creates a new mass packet builder
func NewMassPacketBuilder() *MassPacketBuilder {
	return &MassPacketBuilder{
		packets: make([]MassPacketEntry, 0),
	}
}

// AddPacket adds a packet to the mass packet sequence
func (b *MassPacketBuilder) AddPacket(id types.PacketID, data []byte) {
	// Copy data to avoid issues with shared memory
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	b.packets = append(b.packets, MassPacketEntry{
		ID:   id,
		Data: dataCopy,
	})
}

// Build creates the final mass packet sequence
func (b *MassPacketBuilder) Build() (*MassPacketSequence, error) {
	if len(b.packets) == 0 {
		return nil, fmt.Errorf("no packets to build")
	}

	if len(b.packets) > 255 {
		return nil, fmt.Errorf("too many packets: %d (max 255)", len(b.packets))
	}

	return &MassPacketSequence{
		PacketCount: uint8(len(b.packets)),
		Packets:     b.packets,
	}, nil
}

// Clear removes all packets from the builder
func (b *MassPacketBuilder) Clear() {
	b.packets = b.packets[:0]
}

// IsEmpty returns true if no packets have been added
func (b *MassPacketBuilder) IsEmpty() bool {
	return len(b.packets) == 0
}

// Count returns the number of packets in the builder
func (b *MassPacketBuilder) Count() int {
	return len(b.packets)
}

// Marshal serializes the mass packet sequence to binary format
// Binary format: [packet_count:1][packet_id:2][packet_data:variable]...
func (m *MassPacketSequence) Marshal() ([]byte, error) {
	if m.PacketCount == 0 {
		return nil, fmt.Errorf("empty mass packet sequence")
	}

	buf := new(bytes.Buffer)

	// Write packet count
	if err := binary.Write(buf, binary.LittleEndian, m.PacketCount); err != nil {
		return nil, fmt.Errorf("failed to write packet count: %w", err)
	}

	// Write each packet
	for _, packet := range m.Packets {
		// Write packet ID
		if err := binary.Write(buf, binary.LittleEndian, packet.ID); err != nil {
			return nil, fmt.Errorf("failed to write packet ID: %w", err)
		}

		// Write packet data
		if _, err := buf.Write(packet.Data); err != nil {
			return nil, fmt.Errorf("failed to write packet data: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes binary data to mass packet sequence
func (m *MassPacketSequence) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return fmt.Errorf("invalid mass packet data: too short")
	}

	buf := bytes.NewReader(data)

	// Read packet count
	if err := binary.Read(buf, binary.LittleEndian, &m.PacketCount); err != nil {
		return fmt.Errorf("failed to read packet count: %w", err)
	}

	if m.PacketCount == 0 {
		return fmt.Errorf("invalid packet count: 0")
	}

	m.Packets = make([]MassPacketEntry, 0, m.PacketCount)

	// Read each packet
	for i := uint8(0); i < m.PacketCount; i++ {
		var packetID types.PacketID

		// Read packet ID
		if err := binary.Read(buf, binary.LittleEndian, &packetID); err != nil {
			return fmt.Errorf("failed to read packet ID for packet %d: %w", i, err)
		}

		// For proper implementation, we'd need packet size mapping like C++ CPackets::GetPacketSize()
		// For now, we'll read all remaining data for the last packet or use a simplified approach
		// This is a limitation that should be addressed when implementing full packet size support

		var packetData []byte
		if i == m.PacketCount-1 {
			// Last packet - read all remaining data
			remainingData := make([]byte, buf.Len())
			n, err := buf.Read(remainingData)
			if err != nil && err.Error() != "EOF" {
				return fmt.Errorf("failed to read packet data for packet %d: %w", i, err)
			}
			packetData = remainingData[:n]
		} else {
			// For non-last packets, we'd need to know exact packet sizes
			// This would require implementing getPacketSize() function
			// For now, return an error for multi-packet sequences
			return fmt.Errorf("multi-packet mass sequences not fully supported yet (packet %d of %d)", i+1, m.PacketCount)
		}

		m.Packets = append(m.Packets, MassPacketEntry{
			ID:   packetID,
			Data: packetData,
		})
	}

	return nil
}

// NewMassPacketSequence creates a new mass packet sequence from a builder
func NewMassPacketSequence(builder *MassPacketBuilder) (*MassPacketSequence, error) {
	return builder.Build()
}

// CreateMassPacket is a utility function to create a mass packet from multiple individual packets
func CreateMassPacket(packets ...struct {
	ID   types.PacketID
	Data []byte
}) (*MassPacketSequence, error) {
	builder := NewMassPacketBuilder()

	for _, pkt := range packets {
		builder.AddPacket(pkt.ID, pkt.Data)
	}

	return builder.Build()
}

// SendMassPacket is a utility to build and marshal a mass packet in one call
func SendMassPacket(packets ...struct {
	ID   types.PacketID
	Data []byte
}) ([]byte, error) {
	massPacket, err := CreateMassPacket(packets...)
	if err != nil {
		return nil, err
	}

	return massPacket.Marshal()
}
