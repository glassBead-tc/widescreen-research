package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/spawn-mcp/coordinator/pkg/drone"
)

func main() {
	log.Println("Starting Drone...")
	drone, err := drone.NewResearcherDrone()
	if err != nil {
		log.Fatal(err)
	}
	defer drone.Close()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		drone.Serve()
	}()
	<-sigChan
	log.Println("Drone stopped")
}