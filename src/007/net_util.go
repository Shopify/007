package main

import (
	"os"
)

// Returns the host as reported by the kernel
func GetLocalHostname() string {
	if hostname, err := os.Hostname(); err != nil {
		Log.WithField("error", err).Errorf("Error getting hostname.")
		return ""
	} else {
		return hostname
	}
}
