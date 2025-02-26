package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/recally-io/polyllm/llms"
)

// ModelCache represents the cache of models
type ModelCache struct {
	Models    map[string][]llms.Model `json:"models"`
	Timestamp time.Time               `json:"timestamp"`
}

// Global variables for model cache
var (
	modelCacheMutex = &sync.RWMutex{}
	modelCachePath  = filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "polyllm", "models.json")
)

// Load loads the model cache from the file
func Load() (ModelCache, error) {
	modelCacheMutex.RLock()
	defer modelCacheMutex.RUnlock()
	
	var cache ModelCache
	
	// Check if cache file exists
	if _, err := os.Stat(modelCachePath); os.IsNotExist(err) {
		return ModelCache{
			Models:    make(map[string][]llms.Model),
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
			Models:    make(map[string][]llms.Model),
			Timestamp: time.Time{},
		}, err
	}
	
	return cache, nil
}

// Save saves the model cache to the file
func Save(cache ModelCache) error {
	modelCacheMutex.Lock()
	defer modelCacheMutex.Unlock()
	
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
func IsValid(cache ModelCache) bool {
	return !cache.Timestamp.IsZero() && time.Since(cache.Timestamp) < time.Hour
}
