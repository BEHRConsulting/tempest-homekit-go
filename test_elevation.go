//go:build manual
// +build manual

package main

import (
	"fmt"

	"tempest-homekit-go/pkg/config"
)

func main() {
	fmt.Println("Testing elevation parsing...")

	// Test the default configuration
	cfg := config.LoadConfig()
	fmt.Printf("Default elevation: %.2f meters (%.1f feet)\n", cfg.Elevation, cfg.Elevation*3.28084)

	// You can test with command line args like: go run test_elevation.go --elevation 1200ft
}
