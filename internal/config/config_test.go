package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	content := `
server:
  host: "127.0.0.1"
  port: 9090
  readTimeout: 30
  writeTimeout: 30
  adminToken: "my-secret-token"

database:
  dsn: "user:pass@tcp(localhost:3306)/testdb"
  maxOpenConns: 10
  maxIdleConns: 5
  connMaxLifetime: 120

notifier:
  pollInterval: 5

log:
  level: "debug"
  format: "text"

node:
  id: "node-001"
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(content)
	require.NoError(t, err)
	tmpfile.Close()

	cfg, err := Load(tmpfile.Name())
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, 30, cfg.Server.ReadTimeout)
	assert.Equal(t, 30, cfg.Server.WriteTimeout)
	assert.Equal(t, "my-secret-token", cfg.Server.AdminToken)

	assert.Equal(t, "user:pass@tcp(localhost:3306)/testdb", cfg.Database.DSN)
	assert.Equal(t, 10, cfg.Database.MaxOpenConns)
	assert.Equal(t, 5, cfg.Database.MaxIdleConns)
	assert.Equal(t, 120, cfg.Database.ConnMaxLifetime)

	assert.Equal(t, 5, cfg.Notifier.PollInterval)
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "text", cfg.Log.Format)
	assert.Equal(t, "node-001", cfg.Node.ID)
}

func TestLoad_Defaults(t *testing.T) {
	content := `
server:
  host: "0.0.0.0"
  port: 8080

database:
  dsn: "root:password@tcp(127.0.0.1:3306)/go_diamond"
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(content)
	require.NoError(t, err)
	tmpfile.Close()

	cfg, err := Load(tmpfile.Name())
	require.NoError(t, err)

	assert.Equal(t, 60, cfg.Server.ReadTimeout)
	assert.Equal(t, 60, cfg.Server.WriteTimeout)
	assert.Equal(t, 2, cfg.Notifier.PollInterval)
	assert.Equal(t, 20, cfg.Database.MaxOpenConns)
	assert.Equal(t, 10, cfg.Database.MaxIdleConns)
	assert.Equal(t, 300, cfg.Database.ConnMaxLifetime)
}

func TestLoad_AutoNodeID(t *testing.T) {
	content := `
server:
  host: "0.0.0.0"
  port: 8080

database:
  dsn: "root:password@tcp(127.0.0.1:3306)/go_diamond"
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(content)
	require.NoError(t, err)
	tmpfile.Close()

	cfg, err := Load(tmpfile.Name())
	require.NoError(t, err)

	assert.NotEmpty(t, cfg.Node.ID)
	assert.Contains(t, cfg.Node.ID, "node-")
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	assert.Error(t, err)
}

func TestServerConfig_Addr(t *testing.T) {
	sc := &ServerConfig{Host: "127.0.0.1", Port: 8080}
	assert.Equal(t, "127.0.0.1:8080", sc.Addr())

	sc = &ServerConfig{Host: "0.0.0.0", Port: 9090}
	assert.Equal(t, "0.0.0.0:9090", sc.Addr())
}

func TestServerConfig_Addr_DefaultPort(t *testing.T) {
	sc := &ServerConfig{Host: "127.0.0.1", Port: 0}
	assert.Equal(t, "127.0.0.1:8080", sc.Addr())
}

func TestFormatPort(t *testing.T) {
	assert.Equal(t, "8080", formatPort(8080))
	assert.Equal(t, "80", formatPort(80))
	assert.Equal(t, "9090", formatPort(9090))
	assert.Equal(t, "8080", formatPort(0))
}

func TestGenerateNodeID(t *testing.T) {
	id1 := generateNodeID()
	id2 := generateNodeID()

	assert.Contains(t, id1, "node-")
	assert.Len(t, id1, 17) // "node-" (5) + 12 chars
	assert.Equal(t, id1, id2) // deterministic
}

func TestRandomString(t *testing.T) {
	s1 := randomString(12)
	s2 := randomString(12)

	assert.Len(t, s1, 12)
	assert.Equal(t, s1, s2) // deterministic
}

func TestApplyDefaults(t *testing.T) {
	cfg := &Config{}
	applyDefaults(cfg)

	assert.Equal(t, 60, cfg.Server.ReadTimeout)
	assert.Equal(t, 60, cfg.Server.WriteTimeout)
	assert.Equal(t, 2, cfg.Notifier.PollInterval)
	assert.Equal(t, 20, cfg.Database.MaxOpenConns)
	assert.Equal(t, 10, cfg.Database.MaxIdleConns)
	assert.Equal(t, 300, cfg.Database.ConnMaxLifetime)
}

func TestApplyDefaults_DoesNotOverwrite(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			ReadTimeout: 120,
			Port:        9090,
		},
	}
	applyDefaults(cfg)

	assert.Equal(t, 120, cfg.Server.ReadTimeout)
	assert.Equal(t, 9090, cfg.Server.Port)
}