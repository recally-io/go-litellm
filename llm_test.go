package litellm

import (
	"os"
	"testing"

	"github.com/recally-io/go-litellm/llms"
)

func TestSetApiKeyFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		presetKey   string
		envKey      string
		envValue    string
		expectError bool
		expectedKey string
	}{
		{
			name:        "existing api key should not be overwritten",
			presetKey:   "preset-key",
			envKey:      "TEST_API_KEY",
			envValue:    "env-key",
			expectError: false,
			expectedKey: "preset-key",
		},
		{
			name:        "should set api key from environment",
			presetKey:   "",
			envKey:      "TEST_API_KEY",
			envValue:    "env-key",
			expectError: false,
			expectedKey: "env-key",
		},
		{
			name:        "should error when env variable is empty",
			presetKey:   "",
			envKey:      "TEST_API_KEY",
			envValue:    "",
			expectError: true,
			expectedKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			} else {
				os.Unsetenv(tt.envKey)
			}

			cfg := &llms.Config{APIKey: tt.presetKey}
			err := setApiKeyFromEnv(cfg, tt.envKey)

			// Check error
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check API key value
			if cfg.APIKey != tt.expectedKey {
				t.Errorf("expected API key %q but got %q", tt.expectedKey, cfg.APIKey)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		provider    ProviderName
		envKey      string
		envValue    string
		expectError bool
	}{
		{
			name:        "should create OpenAI client with valid API key",
			provider:    ProviderNameOpenAI,
			envKey:      "OPENAI_API_KEY",
			envValue:    "test-key",
			expectError: false,
		},
		{
			name:        "should fail creating OpenAI client without API key",
			provider:    ProviderNameOpenAI,
			envKey:      "OPENAI_API_KEY",
			envValue:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			} else {
				os.Unsetenv(tt.envKey)
			}

			// Create client
			client, err := New(tt.provider)

			// Check error
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check client
			if !tt.expectError && client == nil {
				t.Error("expected client but got nil")
			}
		})
	}
}
