package diamond

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSnapshotManager_SaveLoad(t *testing.T) {
	home, _ := os.UserHomeDir()
	baseDir := filepath.Join(home, ".go-diamond", "test-snapshot")

	mgr := &SnapshotManager{baseDir: baseDir}

	ns, group, dataID := "default", "DEFAULT_GROUP", "test.json"
	content := `{"key":"value"}`

	err := mgr.Save(ns, group, dataID, content)
	assert.NoError(t, err)

	loaded, err := mgr.Load(ns, group, dataID)
	assert.NoError(t, err)
	assert.Equal(t, content, loaded)

	os.RemoveAll(baseDir)
}

func TestSnapshotManager_Delete(t *testing.T) {
	home, _ := os.UserHomeDir()
	baseDir := filepath.Join(home, ".go-diamond", "test-snapshot-delete")

	mgr := &SnapshotManager{baseDir: baseDir}

	ns, group, dataID := "default", "DEFAULT_GROUP", "test.json"
	content := `{"key":"value"}`

	err := mgr.Save(ns, group, dataID, content)
	assert.NoError(t, err)

	err = mgr.Delete(ns, group, dataID)
	assert.NoError(t, err)

	_, err = mgr.Load(ns, group, dataID)
	assert.Error(t, err)

	os.RemoveAll(baseDir)
}

func TestConfigCache_GetSet(t *testing.T) {
	cache := NewConfigCache()

	cache.Set("key1", "value1")

	v, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", v)

	_, ok = cache.Get("nonexistent")
	assert.False(t, ok)
}

func TestCacheKey(t *testing.T) {
	key := cacheKey("DEFAULT_GROUP", "database.json")
	assert.Equal(t, "DEFAULT_GROUP/database.json", key)
}

func TestNewClient(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080", "default")
	assert.NotNil(t, client)
	assert.Equal(t, "http://127.0.0.1:8080", client.serverAddr)
	assert.Equal(t, "default", client.namespace)
}

func TestClient_AddListener(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080", "default")

	client.AddListener("group1", "data1", func(content string) {})
	client.AddListener("group1", "data2", func(content string) {})

	client.mu.RLock()
	defer client.mu.RUnlock()
	assert.Len(t, client.listeners, 2)
}

func TestClient_AddListener_SameKey(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080", "default")

	var count int
	client.AddListener("group1", "data1", func(content string) { count++ })
	client.AddListener("group1", "data1", func(content string) { count++ })

	client.mu.RLock()
	defer client.mu.RUnlock()
	assert.Len(t, client.listeners, 1)
	assert.Len(t, client.listeners["group1/data1"], 2)
}

func TestNewConfigCache(t *testing.T) {
	cache := NewConfigCache()
	assert.NotNil(t, cache)
}

func TestConfigCache_Get_NotFound(t *testing.T) {
	cache := NewConfigCache()

	_, ok := cache.Get("nonexistent")
	assert.False(t, ok)
}

func TestConfigCache_Set_Overwrite(t *testing.T) {
	cache := NewConfigCache()

	cache.Set("key1", "value1")
	cache.Set("key1", "value2")

	val, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value2", val)
}

func TestNewSnapshotManager(t *testing.T) {
	manager := NewSnapshotManager()
	assert.NotNil(t, manager)
	assert.NotEmpty(t, manager.baseDir)
}

func TestSnapshotManager_snapshotPath(t *testing.T) {
	manager := NewSnapshotManager()

	path := manager.snapshotPath("default", "DEFAULT_GROUP", "app.json")
	assert.Contains(t, path, "default")
	assert.Contains(t, path, "DEFAULT_GROUP")
	assert.Contains(t, path, "app.json")
}

func TestSnapshotManager_snapshotPath_WithSlash(t *testing.T) {
	manager := NewSnapshotManager()

	path := manager.snapshotPath("default", "DEFAULT_GROUP", "path/to/app.json")
	assert.NotContains(t, path, "path/to")
	assert.Contains(t, path, "path_to_app.json")
}

func TestMd5Hash(t *testing.T) {
	hash := md5Hash("test content")
	assert.Len(t, hash, 32)
}

func TestMd5Hash_DifferentContent(t *testing.T) {
	hash1 := md5Hash("content1")
	hash2 := md5Hash("content2")
	assert.NotEqual(t, hash1, hash2)
}

func TestMd5Hash_EmptyString(t *testing.T) {
	hash := md5Hash("")
	assert.Len(t, hash, 32)
}

func TestWithHTTPClient(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080", "default")

	customClient := &http.Client{Timeout: 100 * time.Second}
	opt := WithHTTPClient(customClient)
	opt(client)

	assert.Equal(t, customClient, client.httpClient)
}

func TestClient_Stop(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080", "default")
	client.Stop()
}

func TestClient_Start(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080", "default")
	ctx := context.Background()
	client.Start(ctx)
	client.Stop()
}

func TestClient_StartStop(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080", "default")
	ctx := context.Background()
	client.Start(ctx)
	client.Stop()
}