.PHONY: build clean run

# Default target
all: build

# Build the CLI
build:
	go build -o polyllm-cli ./cmd/polyllm-cli

# Clean build artifacts
clean:
	rm -f polyllm-cli

# Run the CLI with models command
run:
	./polyllm-cli models

# Install the CLI to $GOPATH/bin
install:
	go install ./cmd/polyllm-cli
