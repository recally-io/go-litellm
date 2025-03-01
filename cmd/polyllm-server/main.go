package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/recally-io/polyllm"
	"github.com/recally-io/polyllm/internal/server"
)

func main() {
	// Define command line flags
	configFlag := flag.String("c", "", "Path to the config file")
	flag.Parse()

	config := polyllm.Config{}
	if *configFlag != "" {
		// Set the config file if provided
		cfg, err := polyllm.LoadConfig(*configFlag)
		if err != nil {
			fmt.Printf("Error loading config file: %v\n", err)
			os.Exit(1)
		}
		config = cfg
	}

	server.StartServer(config)
}
