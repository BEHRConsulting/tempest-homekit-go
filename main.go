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

	log.Printf("Starting service with config: WebPort=%s, LogLevel=%s", cfg.WebPort, cfg.LogLevel)
	err := service.StartService(cfg)
	if err != nil {
		log.Fatalf("Service failed: %v", err)
	}

	log.Println("Service started successfully, waiting for interrupt signal...")

	// Wait for interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("Received signal %v, shutting down...", sig)
}
