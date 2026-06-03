package atk

import (
	"bytes"
	"testing"
)

func TestFinalizePayload(t *testing.T) {
	tests := []struct {
		name     string
		op       byte
		body     [14]byte
		expected []byte // expected 16-byte payload
	}{
		{
			name: "Query Battery opcode 0x04 with empty body",
			op:   0x04,
			body: [14]byte{},
			expected: []byte{
				0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x49,
			},
		},
		{
			name: "Arbitrary opcode and body values",
			op:   0x12,
			body: [14]byte{0x01, 0x02, 0x03},
			expected: []byte{
				0x12, 0x01, 0x02, 0x03, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x35, // 0x4D - (0x12 + 0x01 + 0x02 + 0x03) = 0x4D - 0x18 = 0x35
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FinalizePayload(tt.op, tt.body)

			if len(result) != FrameSize {
				t.Fatalf("expected payload size %d, got %d", FrameSize, len(result))
			}

			// Validate checksum logic: sum of all 16 bytes must equal 0x4D
			var sum byte
			for _, b := range result {
				sum += b
			}

			// Since checksum byte (index 15) is (0x4D - sum_of_first_15), the sum of all 16 bytes:
			// sum_all = sum_of_first_15 + (0x4D - sum_of_first_15) = 0x4D.
			if sum != 0x4D {
				t.Errorf("expected sum of bytes to be 0x4D, got 0x%02X", sum)
			}

			if !bytes.Equal(result, tt.expected) {
				t.Errorf("expected payload %v, got %v", tt.expected, result)
			}
		})
	}
}
