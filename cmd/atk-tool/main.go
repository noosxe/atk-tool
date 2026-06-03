package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/noosxe/atk-tool"
	"github.com/spf13/cobra"
)

var (
	jsonFlag   bool
	deviceFlag string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "atk-tool",
		Short: "ATK Peripheral Utility",
		Long:  `A command-line tool designed to interface with and query ATK peripherals.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return atk.Init()
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			_ = atk.Exit()
		},
	}

	// Global flag (available to all commands)
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output in JSON format")

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all connected and supported ATK devices",
		Run: func(cmd *cobra.Command, args []string) {
			handleList(jsonFlag)
		},
	}

	var statusCmd = &cobra.Command{
		Use:     "status",
		Aliases: []string{"battery"},
		Short:   "Query and print battery status/voltage of a device",
		Run: func(cmd *cobra.Command, args []string) {
			handleStatus(deviceFlag, jsonFlag)
		},
	}

	// Local flag for status command
	statusCmd.Flags().StringVar(&deviceFlag, "device", "", "Target specific device path")

	rootCmd.AddCommand(listCmd, statusCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func handleList(jsonOut bool) {
	devices, err := atk.Enumerate()
	if err != nil {
		handleError(err, jsonOut)
		os.Exit(1)
	}

	if jsonOut {
		if devices == nil {
			devices = []*atk.DeviceInfo{}
		}
		b, _ := json.MarshalIndent(devices, "", "  ")
		fmt.Println(string(b))
		return
	}

	if len(devices) == 0 {
		fmt.Println("No supported ATK devices found.")
		return
	}

	fmt.Printf("Found %d supported ATK device(s):\n\n", len(devices))
	fmt.Printf("%-20s %-25s %-12s %-12s %s\n", "Model", "Product Name", "Vendor ID", "Product ID", "Path")
	fmt.Println("------------------------------------------------------------------------------------------------------------------")
	for _, dev := range devices {
		fmt.Printf("%-20s %-25s 0x%04X       0x%04X       %s\n",
			dev.ModelName,
			dev.ProductName,
			dev.VendorID,
			dev.ProductID,
			dev.Path,
		)
	}
}

func handleStatus(devicePath string, jsonOut bool) {
	devices, err := atk.Enumerate()
	if err != nil {
		handleError(err, jsonOut)
		os.Exit(1)
	}

	if len(devices) == 0 {
		if jsonOut {
			printJSONError(fmt.Errorf("no connected ATK devices found"))
		} else {
			fmt.Fprintf(os.Stderr, "Error: No connected ATK devices found.\n")
		}
		os.Exit(1)
	}

	var target *atk.DeviceInfo
	if devicePath != "" {
		for _, dev := range devices {
			if dev.Path == devicePath {
				target = dev
				break
			}
		}
		if target == nil {
			if jsonOut {
				printJSONError(fmt.Errorf("device path %s not found", devicePath))
			} else {
				fmt.Fprintf(os.Stderr, "Error: Device path %s not found.\n", devicePath)
			}
			os.Exit(1)
		}
	} else {
		target = devices[0]
	}

	dev, err := atk.Open(target)
	if err != nil {
		handleError(err, jsonOut)
		os.Exit(1)
	}
	defer dev.Close()

	batt, err := dev.QueryBattery()
	if err != nil {
		handleError(err, jsonOut)
		os.Exit(1)
	}

	if jsonOut {
		type statusResponse struct {
			Device  *atk.DeviceInfo  `json:"device"`
			Battery *atk.BatteryInfo `json:"battery"`
		}
		res := statusResponse{
			Device:  target,
			Battery: batt,
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(b))
		return
	}

	fmt.Printf("🔋 %s Status:\n", target.ModelName)
	fmt.Printf("   Device Path: %s\n", target.Path)
	fmt.Printf("   Battery:     %d%%\n", batt.Percentage)
	fmt.Printf("   Voltage:     %.3f V\n", batt.Voltage)
}

func printJSONError(err error) {
	errMap := map[string]string{"error": err.Error()}
	b, _ := json.MarshalIndent(errMap, "", "  ")
	fmt.Println(string(b))
}

func handleError(err error, jsonOut bool) {
	if jsonOut {
		printJSONError(err)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
