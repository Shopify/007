package main

import (
	"net"
)

func collectInterfaces() {
	// Get a list of all interfaces.
	ifaces, err := net.Interfaces()
	if err != nil {
		Log.WithField("error", err).Fatalf("Error getting the list of interfaces.")
	}

	Log.WithField("count", len(ifaces)).Info("Found interfaces.")

	IfaceList.ClearAndAppend(ifaces)
}
