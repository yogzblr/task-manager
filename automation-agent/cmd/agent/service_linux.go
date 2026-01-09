//go:build !windows
// +build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"
)

func initService() {
	// Systemd integration for Linux
	// In production, you'd use systemd notify support
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	
	go func() {
		sig := <-sigChan
		// Handle graceful shutdown
		_ = sig
	}()
}
