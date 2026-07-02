package main

import (
	"github.com/sgdevelopers29-afk/GoKit-Lite/logger"
)

func main() {
	// Basic logging
	logger.Info("Starting the application server")

	// Formatted logging
	port := 8080
	logger.Infof("Server listening on port %d", port)

	logger.Debugf("Debugging enabled for session %s", "12345")
	logger.Warnf("Memory usage is getting high: %dMB", 512)
	logger.Errorf("Failed to connect to database: %v", "timeout")
}
