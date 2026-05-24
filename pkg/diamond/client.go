package diamond

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Client struct {
	serverAddr string
	namespace  string
	httpClient *http.Client
	snapshot   *SnapshotManager
	cache      *ConfigCache
	listeners  map[string][]func(content string)
	mu         sync.RWMutex
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

func NewClient(serverAddr, namespace string, opts ...Option) *Client {
	c := &Client{
		serverAddr: serverAddr,
		namespace:  namespace,
		httpClient: &http.Client{
			Timeout: 65 * time.Second,
		},
		snapshot:  NewSnapshotManager(),
		cache:    NewConfigCache(),
		listeners: make(map[string][]func(content string)),
		stopCh:   make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) GetConfig(group, dataID string) (string, error) {
	cacheKey := cacheKey(group, dataID)

	if content, ok := c.cache.Get(cacheKey); ok {
		return content, nil
	}

	url := fmt.Sprintf("%s/api/v1/configs/%s/%s/%s", c.serverAddr, c.namespace, group, dataID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return c.snapshot.Load(c.namespace, group, dataID)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("config not found")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.snapshot.Load(c.namespace, group, dataID)
	}

	c.cache.Set(cacheKey, string(body))
	c.snapshot.Save(c.namespace, group, dataID, string(body))

	return string(body), nil
}

func (c *Client) AddListener(group, dataID string, fn func(content string)) {
	key := cacheKey(group, dataID)
	c.mu.Lock()
	c.listeners[key] = append(c.listeners[key], fn)
	c.mu.Unlock()
}

func (c *Client) Start(ctx context.Context) {
	c.wg.Add(1)
	go c.longPolling(ctx)
}

func (c *Client) Stop() {
	close(c.stopCh)
	c.wg.Wait()
}

func (c *Client) longPolling(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.pollOnce()
		}
	}
}

func (c *Client) pollOnce() {
	c.mu.RLock()
	listeners := c.listeners
	c.mu.RUnlock()

	for key := range listeners {
		parts := strings.SplitN(key, "/", 2)
		if len(parts) != 2 {
			continue
		}
		group, dataID := parts[0], parts[1]

		content, err := c.GetConfig(group, dataID)
		if err != nil {
			continue
		}

		c.mu.RLock()
		cbs := c.listeners[key]
		c.mu.RUnlock()

		for _, cb := range cbs {
			go cb(content)
		}
	}
}

func cacheKey(group, dataID string) string {
	return group + "/" + dataID
}

type ConfigCache struct {
	data sync.Map
}

func NewConfigCache() *ConfigCache {
	return &ConfigCache{}
}

func (c *ConfigCache) Get(key string) (string, bool) {
	v, ok := c.data.Load(key)
	if !ok {
		return "", false
	}
	return v.(string), true
}

func (c *ConfigCache) Set(key, value string) {
	c.data.Store(key, value)
}

type SnapshotManager struct {
	baseDir string
}

func NewSnapshotManager() *SnapshotManager {
	home, _ := os.UserHomeDir()
	return &SnapshotManager{
		baseDir: filepath.Join(home, ".go-diamond", "snapshot"),
	}
}

func (m *SnapshotManager) Save(namespace, group, dataID, content string) error {
	path := m.snapshotPath(namespace, group, dataID)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func (m *SnapshotManager) Load(namespace, group, dataID string) (string, error) {
	path := m.snapshotPath(namespace, group, dataID)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (m *SnapshotManager) Delete(namespace, group, dataID string) error {
	path := m.snapshotPath(namespace, group, dataID)
	return os.Remove(path)
}

func (m *SnapshotManager) snapshotPath(namespace, group, dataID string) string {
	safeDataID := strings.ReplaceAll(dataID, "/", "_")
	return filepath.Join(m.baseDir, namespace, group, safeDataID+".cache")
}

type Option func(*Client)

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

func md5Hash(content string) string {
	h := md5.Sum([]byte(content))
	return hex.EncodeToString(h[:])
}