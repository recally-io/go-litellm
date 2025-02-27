package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/recally-io/polyllm"
	"github.com/recally-io/polyllm/internal/cli"
)

// printUsage prints the usage information
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  polyllm-cli models                  - List all available models")
	fmt.Println("  polyllm-cli -m \"<model>\" \"<prompt>\" - Chat with a model")
	fmt.Println("\nExamples:")
	fmt.Println("  polyllm-cli models")
	fmt.Println("  polyllm-cli -m \"gpt-4\" \"Tell me a joke\"")
	fmt.Println("  polyllm-cli -m \"deepseek/deepseek-chat\" \"What is the meaning of life?\"")
}

func main() {
	// Define command line flags
	modelFlag := flag.String("m", "", "Model to use for chat")
	flag.Parse()

	// Get remaining arguments
	args := flag.Args()

	service := cli.NewLLMService(polyllm.New())

	// Check if the command is "models"
	if len(args) > 0 && args[0] == "models" {
		service.ListModels()
		return
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
