package atk

import (
	"fmt"
	"time"

	"github.com/sstallion/go-hid"
)

// BatteryInfo holds the parsed battery percentage and voltage.
type BatteryInfo struct {
	Percentage uint8   `json:"percentage"`
	Voltage    float32 `json:"voltage"`
}

// DeviceInfo represents details of a discovered ATK peripheral.
type DeviceInfo struct {
	Path        string `json:"path"`
	VendorID    uint16 `json:"vendor_id"`
	ProductID   uint16 `json:"product_id"`
	Interface   int    `json:"interface"`
	UsagePage   uint16 `json:"usage_page"`
	Usage       uint16 `json:"usage"`
	ProductName string `json:"product_name"` // Retreived from device
	ModelName   string `json:"model_name"`   // Matched from registry
	ReportID    byte   `json:"report_id"`    // Matched from registry
}

// Device represents an open connection to an ATK device.
type Device struct {
	info *DeviceInfo
	dev  *hid.Device
}

// Info returns metadata about the connected device.
func (d *Device) Info() *DeviceInfo {
	return d.info
}

// Close terminates the HID connection to the device.
func (d *Device) Close() error {
	if d.dev != nil {
		return d.dev.Close()
	}
	return nil
}

// QueryBattery queries the device for battery percentage and voltage,
// then parses the response.
func (d *Device) QueryBattery() (*BatteryInfo, error) {
	// Construct raw payload frame (16 bytes)
	payload := FinalizePayload(CmdQueryBattery, [14]byte{})

	// Prepend Report ID to create the transfer packet
	packet := make([]byte, 1+len(payload))
	packet[0] = d.info.ReportID
	copy(packet[1:], payload)

	// Write packet to device
	_, err := d.dev.Write(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to write query packet: %w", err)
	}

	// Read response (expected layout: header/opcode, metadata, data bytes)
	inBuf := make([]byte, 64)
	n, err := d.dev.ReadWithTimeout(inBuf, 200*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("read timeout or error: %w", err)
	}
	if n < 10 {
		return nil, fmt.Errorf("response frame too short (got %d bytes, expected >= 10)", n)
	}

	// Parse out battery percentage (index 6)
	batteryPercent := inBuf[6]

	// Parse out battery voltage in millivolts (indices 8 & 9)
	millivolts := (uint16(inBuf[8]) << 8) | uint16(inBuf[9])
	voltage := float32(millivolts) / 1000.0

	return &BatteryInfo{
		Percentage: batteryPercent,
		Voltage:    voltage,
	}, nil
}
