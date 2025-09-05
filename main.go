package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"tempest-homekit-go/pkg/config"
	"tempest-homekit-go/pkg/service"
)

func main() {
	cfg := config.LoadConfig()

	err := service.StartService(cfg)
	if err != nil {
		log.Fatalf("Service failed: %v", err)
	}

	// Wait for interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down...")
}
