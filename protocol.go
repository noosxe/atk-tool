package atk

const (
	// DefaultReportID is the query Report ID used by ATK devices.
	DefaultReportID = 0x08

	// FrameSize is the size of the payload frame (excluding the Report ID).
	FrameSize = 16

	// CmdQueryBattery is the subcommand opcode for querying battery information.
	CmdQueryBattery = 0x04
)

// FinalizePayload constructs a 16-byte payload starting with the opcode,
// followed by the body, and ending with a checksum at the last byte.
// The checksum is computed such that the sum of all 16 bytes equals 0x4D.
func FinalizePayload(op byte, body [14]byte) []byte {
	payload := make([]byte, FrameSize)
	payload[0] = op
	copy(payload[1:15], body[:])

	var sum byte
	for i := 0; i < 15; i++ {
		sum += payload[i]
	}
	payload[15] = 0x4D - sum
	return payload
}
