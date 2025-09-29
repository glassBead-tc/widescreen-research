// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/glassBead-tc/widescreen-research/pkg/coordinator"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("Starting Coordinator...")
	server := coordinator.NewServer()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		server.Serve()
	}()
	<-sigChan
	log.Println("Coordinator stopped")
}
