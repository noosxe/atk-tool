package atk

import (
	"fmt"

	"github.com/sstallion/go-hid"
)

// Enumerate scans the USB bus for registered and supported ATK devices.
func Enumerate() ([]*DeviceInfo, error) {
	var found []*DeviceInfo
	seenPaths := make(map[string]bool)

	for _, def := range RegisteredDevices() {
		err := hid.Enumerate(def.VendorID, def.ProductID, func(info *hid.DeviceInfo) error {
			if def.Matches(info.UsagePage) {
				if seenPaths[info.Path] {
					return nil
				}
				seenPaths[info.Path] = true

				productName := info.ProductStr
				if productName == "" {
					productName = "ATK Peripheral"
				}

				found = append(found, &DeviceInfo{
					Path:        info.Path,
					VendorID:    info.VendorID,
					ProductID:   info.ProductID,
					Interface:   info.InterfaceNbr,
					UsagePage:   info.UsagePage,
					Usage:       info.Usage,
					ProductName: productName,
					ModelName:   def.Name,
					ReportID:    def.ReportID,
				})
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to enumerate device %s (VID: 0x%04x, PID: 0x%04x): %w", def.Name, def.VendorID, def.ProductID, err)
		}
	}

	return found, nil
}

// Open establishes a connection to a specific ATK device using its DeviceInfo path.
func Open(info *DeviceInfo) (*Device, error) {
	dev, err := hid.OpenPath(info.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open HID path %s: %w", info.Path, err)
	}

	return &Device{
		info: info,
		dev:  dev,
	}, nil
}
