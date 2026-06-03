package atk

// DeviceDefinition represents the identification parameters for a supported ATK peripheral.
type DeviceDefinition struct {
	Name       string   `json:"name"`
	VendorID   uint16   `json:"vendor_id"`
	ProductID  uint16   `json:"product_id"`
	UsagePages []uint16 `json:"usage_pages"` // Usage pages that receive raw communication (e.g. 0xFF02, 0xFF04)
	ReportID   byte     `json:"report_id"`   // Report ID used for packet transmission
}

// Matches checks if the provided usage page is supported by this device definition.
func (d DeviceDefinition) Matches(usagePage uint16) bool {
	if len(d.UsagePages) == 0 {
		return true
	}
	for _, page := range d.UsagePages {
		if page == usagePage {
			return true
		}
	}
	return false
}

// defaultRegistry stores the list of supported devices.
var defaultRegistry = []DeviceDefinition{
	{
		Name:       "ATK A9 Plus",
		VendorID:   0x373b,
		ProductID:  0x10c9,
		UsagePages: []uint16{0xFF02, 0xFF04},
		ReportID:   DefaultReportID,
	},
}

// RegisteredDevices returns a copy of the default list of supported devices.
func RegisteredDevices() []DeviceDefinition {
	devices := make([]DeviceDefinition, len(defaultRegistry))
	copy(devices, defaultRegistry)
	return devices
}

// FindDefinition searches the registry for a definition matching the VendorID and ProductID.
func FindDefinition(vendorID, productID uint16) (DeviceDefinition, bool) {
	for _, def := range defaultRegistry {
		if def.VendorID == vendorID && def.ProductID == productID {
			return def, true
		}
	}
	return DeviceDefinition{}, false
}

// RegisterDevice adds a new device definition to the registry.
// This enables consumers of the library to dynamically register custom devices at runtime.
func RegisterDevice(def DeviceDefinition) {
	defaultRegistry = append(defaultRegistry, def)
}
