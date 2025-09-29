package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/glassBead-tc/widescreen-research/pkg/coordinator"
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
