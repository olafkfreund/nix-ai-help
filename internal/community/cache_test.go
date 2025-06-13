package community

import (
	"os"
	"testing"
	"time"
)

func TestCacheManager_SetAndGet_Current(t *testing.T) {
	cacheDir := t.TempDir()
	cm := NewCacheManager(cacheDir)
	cm.SetMaxAge(1 * time.Hour)
	key := "test-key"
	data := map[string]string{"foo": "bar"}
	if err := cm.Set(key, data, "test"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	var result map[string]string
	found, err := cm.Get(key, &result)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !found || result["foo"] != "bar" {
		t.Error("Cache did not return expected data")
	}
}

func TestCacheManager_DeleteAndClear_Current(t *testing.T) {
	cacheDir := t.TempDir()
	cm := NewCacheManager(cacheDir)
	key := "delete-key"
	_ = cm.Set(key, "value", "test")
	if err := cm.Delete(key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	var result string
	found, _ := cm.Get(key, &result)
	if found {
		t.Error("Expected cache entry to be deleted")
	}
	// Test Clear
	_ = cm.Set("another-key", "value", "test")
	if err := cm.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	files, _ := os.ReadDir(cacheDir)
	if len(files) != 0 {
		t.Error("Expected cache directory to be empty after Clear")
	}
}

func TestCacheManager_CleanExpired_Current(t *testing.T) {
	cacheDir := t.TempDir()
	cm := NewCacheManager(cacheDir)
	cm.SetMaxAge(1 * time.Millisecond)
	key := "expire-key"
	_ = cm.Set(key, "value", "test")
	time.Sleep(10 * time.Millisecond)
	if err := cm.CleanExpired(); err != nil {
		t.Fatalf("CleanExpired failed: %v", err)
	}
	var result string
	found, _ := cm.Get(key, &result)
	if found {
		t.Error("Expected cache entry to be expired and removed")
	}
}
