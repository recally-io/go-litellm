package polyllm

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/recally-io/polyllm/llms"
)

// ModelCache represents the cache of models
type ModelCache struct {
	Models    []llms.Model `json:"models"`
	Timestamp time.Time    `json:"timestamp"`
}

// Global variables for model cache
var (
	modelCacheMutex sync.RWMutex
	cachePath       string
)

func init() {
	// Create the cache directory if it does not exist
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		slog.Error("Error getting user cache directory", "err", err)
		cachePath = filepath.Join(os.TempDir(), "polyllm", "models")
	} else {
		cachePath = filepath.Join(userCacheDir, "polyllm", "models")
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		slog.Error("Error creating cache directory", "err", err)
	}
}

// Load loads the model cache from the file
func loadModelCache(providerName string) (ModelCache, error) {
	modelCacheMutex.RLock()
	defer modelCacheMutex.RUnlock()
	modelCachePath := filepath.Join(cachePath, providerName+".json")

	var cache ModelCache
	slog.Debug("loading model cache", "path", modelCachePath, "provider", providerName)

	// Check if cache file exists
	if _, err := os.Stat(modelCachePath); os.IsNotExist(err) {
		return ModelCache{
			Models:    make([]llms.Model, 0),
			Timestamp: time.Time{},
		}, nil
	}

	// Read cache file
	data, err := os.ReadFile(modelCachePath)
	if err != nil {
		return cache, err
	}

	// Unmarshal cache
	err = json.Unmarshal(data, &cache)
	if err != nil {
		return ModelCache{
			Models:    make([]llms.Model, 0),
			Timestamp: time.Time{},
		}, err
	}

	return cache, nil
}

// Save saves the model cache to the file
func saveModelCache(providerName string, cache ModelCache) error {
	modelCacheMutex.Lock()
	defer modelCacheMutex.Unlock()
	modelCachePath := filepath.Join(cachePath, providerName+".json")

	// Marshal cache
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	// Ensure the directory exists
	dir := filepath.Dir(modelCachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for model cache: %w", err)
	}

	// Write cache file
	return os.WriteFile(modelCachePath, data, 0644)
}

// IsValid checks if the cache is valid (less than 1 hour old)
func isModelCacheValid(cache ModelCache) bool {
	return !cache.Timestamp.IsZero() && time.Since(cache.Timestamp) < time.Hour
}
