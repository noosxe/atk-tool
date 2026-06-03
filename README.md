# ATK Peripheral Tool

[![Go Report Card](https://goreportcard.com/badge/github.com/noosxe/atk-tool)](https://goreportcard.com/report/github.com/noosxe/atk-tool)
[![Go Reference](https://pkg.go.dev/badge/github.com/noosxe/atk-tool.svg)](https://pkg.go.dev/github.com/noosxe/atk-tool)

`atk-tool` is a modular Golang library and command-line utility for querying and interfacing with **ATK gaming peripherals** (focusing on mouse telemetry like battery status and voltage).

It is structured both as a reusable Go library that can be integrated into other projects and as a standalone CLI tool.

---

## Features

- **Telemetry Queries:** Fetch battery charge percentage and current operating voltage.
- **Deduplication:** Merges multiple HID interface/endpoint paths into single logical devices.
- **JSON Outputs:** Native structured JSON output formats for seamless scripting and dashboard integration.
- **Highly Modular:** Centralized device registry for easy support expansion of additional models.
- **Subcommand Aliases:** Fast, intuitive CLI commands powered by Cobra (e.g. `atk-tool battery`).

---

## Linux Setup (udev rules)

Because this tool interacts with low-level USB HID interfaces (`hidraw`), your Linux user needs read/write permissions for the target device files. Create a udev rule to run the tool without `sudo`:

1. Create a file `/etc/udev/rules.d/99-atk.rules` with the following content:
   ```udev
   # ATK / Compx USB Peripherals & Receiver Dongles
   SUBSYSTEM=="hidraw", ATTRS{idVendor}=="373b", MODE="0666"
   ```
2. Reload and trigger the rule configuration:
   ```bash
   sudo udevadm control --reload-rules && sudo udevadm trigger
   ```

---

## Command Line Interface (CLI)

### Installation

To install the latest version of the CLI utility directly into your `$GOPATH/bin`:

```bash
go install github.com/noosxe/atk-tool/cmd/atk-tool@latest
```

### Usage

Run the utility without any arguments (or with `--help`) to view the standard help:

```bash
atk-tool
```

#### 1. List connected ATK devices
```bash
atk-tool list
```

#### 2. Get battery telemetry
```bash
atk-tool status
# OR using the alias:
atk-tool battery
```

#### 3. Output in JSON format
```bash
atk-tool status --json
```

#### 4. Target a specific device path (useful when multiple dongles are plugged in)
```bash
atk-tool status --device /dev/hidraw5
```

---

## Library Usage

You can import `github.com/noosxe/atk-tool` as a dependency in your own Go projects to scan and query ATK peripherals.

### Installation

```bash
go get github.com/noosxe/atk-tool
```

### Code Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/noosxe/atk-tool"
)

func main() {
	// 1. Initialize the underlying HID library
	if err := atk.Init(); err != nil {
		log.Fatalf("Failed to init library: %v", err)
	}
	defer atk.Exit()

	// 2. Scan for supported ATK devices
	devices, err := atk.Enumerate()
	if err != nil {
		log.Fatalf("Failed to scan devices: %v", err)
	}

	if len(devices) == 0 {
		fmt.Println("No supported ATK peripherals found.")
		return
	}

	// 3. Select the first discovered peripheral
	target := devices[0]
	fmt.Printf("Connecting to %s (%s) at %s...\n", target.ModelName, target.ProductName, target.Path)

	// 4. Open a connection
	dev, err := atk.Open(target)
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	defer dev.Close()

	// 5. Query telemetry data
	status, err := dev.QueryBattery()
	if err != nil {
		log.Fatalf("Failed to query battery status: %v", err)
	}

	fmt.Printf("Telemetry:\n")
	fmt.Printf("  Battery level: %d%%\n", status.Percentage)
	fmt.Printf("  Voltage:       %.3f V\n", status.Voltage)
}
```

---

## Extending Supported Devices

Adding support for new models is straightforward. You can register custom models at runtime before calling scanning functions:

```go
import "github.com/noosxe/atk-tool"

func init() {
	atk.RegisterDevice(atk.DeviceDefinition{
		Name:       "ATK F1 Extreme",
		VendorID:   0x2bdf,                      // Example Vendor ID
		ProductID:  0x0a0e,                      // Example Product ID
		UsagePages: []uint16{0xFF02, 0xFF04},    // raw communication usage pages
		ReportID:   0x08,                        // Command report ID
	})
}
```
