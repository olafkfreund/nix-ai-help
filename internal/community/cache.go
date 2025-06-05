package community

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// CacheManager handles caching for community data to improve performance
type CacheManager struct {
	cacheDir string
	maxAge   time.Duration
	logger   *logger.Logger
}

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Key       string      `json:"key"`
	Type      string      `json:"type"`
}

// NewCacheManager creates a new cache manager instance
func NewCacheManager(cacheDir string) *CacheManager {
	if cacheDir == "" {
		homeDir, _ := os.UserHomeDir()
		cacheDir = filepath.Join(homeDir, ".cache", "nixai", "community")
	}

	return &CacheManager{
		cacheDir: cacheDir,
		maxAge:   1 * time.Hour, // Default cache duration
		logger:   logger.NewLoggerWithLevel("info"),
	}
}

// SetMaxAge sets the maximum age for cache entries
func (cm *CacheManager) SetMaxAge(duration time.Duration) {
	cm.maxAge = duration
}

// ensureCacheDir creates the cache directory if it doesn't exist
func (cm *CacheManager) ensureCacheDir() error {
	if err := os.MkdirAll(cm.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	return nil
}

// getCacheFilePath returns the file path for a cache key
func (cm *CacheManager) getCacheFilePath(key string) string {
	// Sanitize the key for use as filename
	safeKey := sanitizeFileName(key)
	return filepath.Join(cm.cacheDir, safeKey+".json")
}

// Get retrieves data from cache if it exists and is not expired
func (cm *CacheManager) Get(key string, result interface{}) (bool, error) {
	filePath := cm.getCacheFilePath(key)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false, nil
	}

	// Read cache file
	data, err := os.ReadFile(filePath)
	if err != nil {
		cm.logger.Error("Failed to read cache file: " + filePath + " - " + err.Error())
		return false, nil
	}

	// Parse cache entry
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		cm.logger.Error("Failed to parse cache entry: " + filePath + " - " + err.Error())
		return false, nil
	}

	// Check if expired
	if time.Since(entry.Timestamp) > cm.maxAge {
		// Clean up expired entry
		_ = os.Remove(filePath)
		return false, nil
	}

	// Unmarshal the cached data into result
	entryData, err := json.Marshal(entry.Data)
	if err != nil {
		return false, fmt.Errorf("failed to marshal cache data: %w", err)
	}

	if err := json.Unmarshal(entryData, result); err != nil {
		return false, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	cm.logger.Debug("Cache hit for key: " + key + " (age: " + time.Since(entry.Timestamp).String() + ")")
	return true, nil
}

// Set stores data in cache with the given key
func (cm *CacheManager) Set(key string, data interface{}, cacheType string) error {
	if err := cm.ensureCacheDir(); err != nil {
		return err
	}

	filePath := cm.getCacheFilePath(key)

	entry := CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		Key:       key,
		Type:      cacheType,
	}

	// Marshal cache entry
	entryData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, entryData, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	cm.logger.Debug("Cache set for key: " + key + " (type: " + cacheType + ")")
	return nil
}

// Delete removes a specific cache entry
func (cm *CacheManager) Delete(key string) error {
	filePath := cm.getCacheFilePath(key)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache entry: %w", err)
	}
	return nil
}

// Clear removes all cache entries
func (cm *CacheManager) Clear() error {
	if err := os.RemoveAll(cm.cacheDir); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	return nil
}

// CleanExpired removes all expired cache entries
func (cm *CacheManager) CleanExpired() error {
	entries, err := os.ReadDir(cm.cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Cache directory doesn't exist, nothing to clean
		}
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	cleaned := 0
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			filePath := filepath.Join(cm.cacheDir, entry.Name())

			// Read and check if expired
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			var cacheEntry CacheEntry
			if err := json.Unmarshal(data, &cacheEntry); err != nil {
				continue
			}

			if time.Since(cacheEntry.Timestamp) > cm.maxAge {
				if err := os.Remove(filePath); err == nil {
					cleaned++
				}
			}
		}
	}

	if cleaned > 0 {
		cm.logger.Info(fmt.Sprintf("Cleaned %d expired cache entries", cleaned))
	}

	return nil
}

// GetStats returns cache statistics
func (cm *CacheManager) GetStats() (CacheStats, error) {
	stats := CacheStats{}

	entries, err := os.ReadDir(cm.cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return stats, nil // Cache directory doesn't exist
		}
		return stats, fmt.Errorf("failed to read cache directory: %w", err)
	}

	var totalSize int64
	expired := 0

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			filePath := filepath.Join(cm.cacheDir, entry.Name())

			// Get file size
			if info, err := entry.Info(); err == nil {
				totalSize += info.Size()
			}

			// Check if expired
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			var cacheEntry CacheEntry
			if err := json.Unmarshal(data, &cacheEntry); err != nil {
				continue
			}

			if time.Since(cacheEntry.Timestamp) > cm.maxAge {
				expired++
			}

			stats.Entries++
		}
	}

	stats.TotalSize = totalSize
	stats.ExpiredEntries = expired

	return stats, nil
}

// CacheStats represents cache statistics
type CacheStats struct {
	Entries        int   `json:"entries"`
	ExpiredEntries int   `json:"expired_entries"`
	TotalSize      int64 `json:"total_size"`
}

// GetCacheKey generates a cache key from multiple parts
func GetCacheKey(parts ...string) string {
	key := ""
	for i, part := range parts {
		if i > 0 {
			key += "_"
		}
		key += sanitizeFileName(part)
	}
	return key
}

// sanitizeFileName removes/replaces characters that are invalid in filenames
func sanitizeFileName(name string) string {
	// Replace invalid characters with underscores
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " ", "."}
	safe := name

	for _, char := range invalid {
		safe = strings.ReplaceAll(safe, char, "_")
	}

	// Limit length
	if len(safe) > 50 {
		safe = safe[:50]
	}

	return safe
}
