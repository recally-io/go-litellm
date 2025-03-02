package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/recally-io/polyllm"
	"github.com/recally-io/polyllm/internal/cli"
)

func init() {
	// Create log file in the current directory
	logFileName := "polyllm.log"
	// Open the log file with create, append, and write permissions
	logFile, err := os.OpenFile(filepath.Join(".", logFileName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// If we can't open the log file, just log to stderr
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v, using stderr instead\n", err)
		return
	}

	// Create a text handler for structured logging
	logHandler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	// Set as the default logger
	slog.SetDefault(slog.New(logHandler))

}

// printUsage prints the usage information
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  polyllm-cli models                  - List all available models")
	fmt.Println("  polyllm-cli -m \"<model>\" -c \"<config-file>\" \"<prompt>\" - Chat with a model")
	fmt.Println("\nExamples:")
	fmt.Println("  polyllm-cli models")
	fmt.Println("  polyllm-cli -m \"gpt-4\" -c \"config.json\" \"Tell me a joke\"")
	fmt.Println("  polyllm-cli -m \"deepseek/deepseek-chat\" -c \"config.json\" \"What is the meaning of life?\"")
}

func main() {
	// Define command line flags
	modelFlag := flag.String("m", "", "Model to use for chat")
	configFlag := flag.String("c", "", "Path to the config file")
	flag.Parse()

	// Get remaining arguments
	args := flag.Args()

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

	service := cli.NewLLMService(polyllm.NewFromConfig(config))

	// Check if the command is "models"
	if len(args) > 0 {
		switch args[0] {
		case "models":
			service.ListModels()
			return
		case "help":
			printUsage()
			return
		case "tools":
			service.ListMCPTools()
			return
		}
	}

	// Check if we need to chat with a model
	if *modelFlag != "" {
		// Join remaining arguments as the prompt
		prompt := strings.Join(args, " ")
		if prompt == "" {
			fmt.Println("Error: No prompt provided")
			os.Exit(1)
		}

		// Chat with the model
		service.ChatCompletion(*modelFlag, prompt)
		return
	}

	// If no valid command is provided, print usage
	printUsage()
}
