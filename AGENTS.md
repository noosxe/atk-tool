# Developer & Agent Guide: ATK Peripheral Tool

This document is designed to help LLM coding agents and developers quickly understand the repository structure, code flow, protocols, and APIs of the `atk-tool` project.

---

## 📂 Repository Layout

The project structure is minimal and structured logically:

- [README.md](file:///home/mechsoull/Projects/atk-tool/README.md): High-level user documentation, udev rules setup, CLI usage examples, and library quick-start.
- [docs/design.md](file:///home/mechsoull/Projects/atk-tool/docs/design.md): Deep-dive design document outlining library, CLI, and registry design principles.
- [docs/protocol.md](file:///home/mechsoull/Projects/atk-tool/docs/protocol.md): Low-level hardware communication packet structures, opcodes, checksum formulas, and parsing specs.
- [atk.go](file:///home/mechsoull/Projects/atk-tool/atk.go): Wrapper for initializing and exiting the underlying HID library.
- [discovery.go](file:///home/mechsoull/Projects/atk-tool/discovery.go): Code for scanning/enumerating connected USB HID devices and opening connections.
- [device.go](file:///home/mechsoull/Projects/atk-tool/device.go): Models for device details, battery telemetry data, and methods to write/read raw HID packets.
- [registry.go](file:///home/mechsoull/Projects/atk-tool/registry.go): Central registry definition for supported ATK devices and match logic.
- [protocol.go](file:///home/mechsoull/Projects/atk-tool/protocol.go): Constants and payload builders (e.g., checksum finalization).
- [protocol_test.go](file:///home/mechsoull/Projects/atk-tool/protocol_test.go): Verification of protocol payload checksum building logic.
- [cmd/atk-tool/main.go](file:///home/mechsoull/Projects/atk-tool/cmd/atk-tool/main.go): The entry point for the CLI tool (built using Cobra).
- [go.mod](file:///home/mechsoull/Projects/atk-tool/go.mod): Go module definition (requires Go 1.26.3+ and `go-hid`).

---

## ⚙️ Dependencies & Prerequisites

- **Go Version:** `1.26.3` or later.
- **External Packages:**
  - `github.com/sstallion/go-hid` (a CGo wrapper around `hidapi`).
  - `github.com/spf13/cobra` (for the CLI application).
- **System Requirements (Linux):**
  - Dev packages: `libudev-dev` and `pkg-config`.
  - Permissions: Access to `/dev/hidraw*` files is required. A udev rule is typically set up in `/etc/udev/rules.d/99-atk.rules` targeting Vendor ID `373b` with mode `0666`.

---

## ⚡ Main Logic & Flow

### 1. Initialization and Cleanup
Before performing any operations on the HID subsystem, the underlying C library must be initialized:
- Call [Init](file:///home/mechsoull/Projects/atk-tool/atk.go#L7) to setup.
- Call [Exit](file:///home/mechsoull/Projects/atk-tool/atk.go#L13) to clean up resources upon program termination.

### 2. Device Registration & Matching
All supported peripherals are declared using the [DeviceDefinition](file:///home/mechsoull/Projects/atk-tool/registry.go#L4) struct in [registry.go](file:///home/mechsoull/Projects/atk-tool/registry.go):
```go
type DeviceDefinition struct {
	Name       string   `json:"name"`
	VendorID   uint16   `json:"vendor_id"`
	ProductID  uint16   `json:"product_id"`
	UsagePages []uint16 `json:"usage_pages"`
	ReportID   byte     `json:"report_id"`
}
```
- **Matching Criteria:** For a scanned USB device to be matched, its `VendorID` and `ProductID` must match a definition, and the device's current HID `UsagePage` must match one of the allowed usage pages in `UsagePages` (typically `0xFF02` or `0xFF04` for raw manufacturer interface communications).
- **Extensibility:** You can register custom devices at runtime using [RegisterDevice](file:///home/mechsoull/Projects/atk-tool/registry.go#L62) or add them natively to the `defaultRegistry` slice.

### 3. Enumeration & Discovery
Calling [Enumerate](file:///home/mechsoull/Projects/atk-tool/discovery.go#L10) will:
1. Iterate over the registered device definitions.
2. Search matching USB HID nodes on the system.
3. De-duplicate identical paths to avoid redundant query handles.
4. Return a slice of pointers to [DeviceInfo](file:///home/mechsoull/Projects/atk-tool/device.go#L17) structs.

### 4. Connection & Telemetry Query
- Use [Open](file:///home/mechsoull/Projects/atk-tool/discovery.go#L50) to initialize a [Device](file:///home/mechsoull/Projects/atk-tool/device.go#L30) with an active connection handle.
- Call [QueryBattery](file:///home/mechsoull/Projects/atk-tool/device.go#L50) to write a request packet to the mouse, wait for its response, and parse it into a [BatteryInfo](file:///home/mechsoull/Projects/atk-tool/device.go#L11) struct.

---

## 🖧 Hardware Protocol Specifications

Communication uses a 16-byte raw payload frame, prepended with a 1-byte Report ID.

### Write Payload Construction
A 16-byte command payload is generated via [FinalizePayload](file:///home/mechsoull/Projects/atk-tool/protocol.go#L17):
1. **Opcode (Byte 0):** Specifies the command. For battery queries, this is `0x04` ([CmdQueryBattery](file:///home/mechsoull/Projects/atk-tool/protocol.go#L11)).
2. **Body (Bytes 1–14):** Arbitrary data bytes, padded with zeros if empty.
3. **Checksum (Byte 15):** The final byte acts as a checksum. It is computed as:
   $$\text{Checksum} = 0\text{x}4\text{D} - \sum_{i=0}^{14} \text{Payload}[i]$$
   This guarantees that the 8-bit sum of all 16 bytes of the payload equals `0x4D`.

The full packet sent to the device includes the **Report ID** (usually `0x08`) at the very beginning (Byte 0), followed by the 16-byte payload (total size = 17 bytes).

### Read Response Parsing
After sending the write packet, the tool reads a 64-byte buffer from the device with a 200ms timeout:
- A valid response must be at least 10 bytes long.
- **Battery Percentage:** Extracted from byte index 6 (`inBuf[6]`).
- **Battery Voltage:** Extracted from byte indices 8 and 9 as a 16-bit big-endian integer:
  $$\text{Voltage (Volts)} = \frac{(\text{inBuf}[8] \ll 8) \mid \text{inBuf}[9]}{1000.0}$$

---

## 💻 CLI Commands

The command-line tool `atk-tool` supports the following:
- **`atk-tool list`**: Enumerate and print all connected supported ATK devices in a table or JSON format.
- **`atk-tool status` / `atk-tool battery`**: Query battery telemetry.
- **Global Flags:**
  - `--json`: Format the output as JSON for easier parsing and integration.
- **Local Flags (status command only):**
  - `--device <path>`: Query a specific `/dev/hidraw*` path directly.

---

## 🛠️ Verification & Development Commands

Always run these commands in the workspace root to ensure code health and validation:

```bash
# Build the CLI tool
go build -v ./...

# Run the test suite
go test -v -race ./...
```

Ensure native libraries (`libudev-dev` and `pkg-config`) are installed locally before executing standard Go toolchain commands.
