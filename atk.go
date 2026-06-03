package atk

import "github.com/sstallion/go-hid"

// Init initializes the underlying HID library.
// It must be called before searching or opening devices.
func Init() error {
	return hid.Init()
}

// Exit cleans up the underlying HID library.
// It should be called when the application is exiting.
func Exit() error {
	return hid.Exit()
}
