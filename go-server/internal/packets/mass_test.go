package packets

import (
	"testing"

	"coopandreas-server/internal/types"
)

// TestMassPacketBuilder tests the mass packet builder functionality
func TestMassPacketBuilder(t *testing.T) {
	builder := NewMassPacketBuilder()

	// Test empty builder
	if !builder.IsEmpty() {
		t.Error("New builder should be empty")
	}

	if builder.Count() != 0 {
		t.Error("New builder should have count 0")
	}

	// Add some test packets
	testData1 := []byte{0x01, 0x02, 0x03}
	testData2 := []byte{0x04, 0x05, 0x06, 0x07}

	builder.AddPacket(types.PED_SPAWN, testData1)
	builder.AddPacket(types.PED_REMOVE, testData2)

	if builder.IsEmpty() {
		t.Error("Builder should not be empty after adding packets")
	}

	if builder.Count() != 2 {
		t.Errorf("Builder should have count 2, got %d", builder.Count())
	}

	// Build the mass packet
	massPacket, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build mass packet: %v", err)
	}

	if massPacket.PacketCount != 2 {
		t.Errorf("Mass packet should have 2 packets, got %d", massPacket.PacketCount)
	}

	if len(massPacket.Packets) != 2 {
		t.Errorf("Mass packet should have 2 packet entries, got %d", len(massPacket.Packets))
	}

	// Check packet data
	if massPacket.Packets[0].ID != types.PED_SPAWN {
		t.Errorf("First packet should be PED_SPAWN, got %d", massPacket.Packets[0].ID)
	}

	if massPacket.Packets[1].ID != types.PED_REMOVE {
		t.Errorf("Second packet should be PED_REMOVE, got %d", massPacket.Packets[1].ID)
	}
}

// TestMassPacketMarshalUnmarshal tests serialization and deserialization
func TestMassPacketMarshalUnmarshal(t *testing.T) {
	builder := NewMassPacketBuilder()

	// Add a test packet
	testData := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	builder.AddPacket(types.PED_SPAWN, testData)

	// Build mass packet
	original, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build mass packet: %v", err)
	}

	// Marshal to binary
	data, err := original.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal mass packet: %v", err)
	}

	// Unmarshal from binary
	unmarshaled := &MassPacketSequence{}
	err = unmarshaled.Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal mass packet: %v", err)
	}

	// Verify data
	if unmarshaled.PacketCount != original.PacketCount {
		t.Errorf("Packet count mismatch: expected %d, got %d", original.PacketCount, unmarshaled.PacketCount)
	}

	if len(unmarshaled.Packets) != len(original.Packets) {
		t.Errorf("Packet array length mismatch: expected %d, got %d", len(original.Packets), len(unmarshaled.Packets))
	}

	if len(unmarshaled.Packets) > 0 {
		if unmarshaled.Packets[0].ID != original.Packets[0].ID {
			t.Errorf("Packet ID mismatch: expected %d, got %d", original.Packets[0].ID, unmarshaled.Packets[0].ID)
		}

		if len(unmarshaled.Packets[0].Data) != len(original.Packets[0].Data) {
			t.Errorf("Packet data length mismatch: expected %d, got %d", len(original.Packets[0].Data), len(unmarshaled.Packets[0].Data))
		}

		for i, b := range unmarshaled.Packets[0].Data {
			if b != original.Packets[0].Data[i] {
				t.Errorf("Packet data mismatch at index %d: expected %02x, got %02x", i, original.Packets[0].Data[i], b)
			}
		}
	}
}

// TestCreateMassPacket tests the utility function
func TestCreateMassPacket(t *testing.T) {
	data1 := []byte{0x01, 0x02}
	data2 := []byte{0x03, 0x04}

	massPacket, err := CreateMassPacket(
		struct {
			ID   types.PacketID
			Data []byte
		}{types.PED_SPAWN, data1},
		struct {
			ID   types.PacketID
			Data []byte
		}{types.PED_REMOVE, data2},
	)

	if err != nil {
		t.Fatalf("Failed to create mass packet: %v", err)
	}

	if massPacket.PacketCount != 2 {
		t.Errorf("Expected 2 packets, got %d", massPacket.PacketCount)
	}
}

// TestSendMassPacket tests the utility function for marshaling
func TestSendMassPacket(t *testing.T) {
	data1 := []byte{0x01, 0x02}

	binaryData, err := SendMassPacket(
		struct {
			ID   types.PacketID
			Data []byte
		}{types.PED_SPAWN, data1},
	)

	if err != nil {
		t.Fatalf("Failed to send mass packet: %v", err)
	}

	if len(binaryData) == 0 {
		t.Error("Binary data should not be empty")
	}

	// Should start with packet count (1)
	if binaryData[0] != 1 {
		t.Errorf("First byte should be packet count 1, got %d", binaryData[0])
	}
}
