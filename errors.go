package polyllm

import (
	"errors"
)

// Common error types
var (
	// ErrProviderNotFound is returned when a provider is not found
	ErrProviderNotFound = errors.New("provider not found")

	// ErrModelNotFound is returned when a model is not found
	ErrModelNotFound = errors.New("model not found")

	// ErrInvalidConfiguration is returned when the configuration is invalid
	ErrInvalidConfiguration = errors.New("invalid configuration")

	// ErrAPIKeyNotSet is returned when an API key is not set
	ErrAPIKeyNotSet = errors.New("API key not set")

	// ErrRequestFailed is returned when a request to a provider fails
	ErrRequestFailed = errors.New("request failed")

	// ErrUnsupportedOperation is returned when an operation is not supported
	ErrUnsupportedOperation = errors.New("unsupported operation")
)
