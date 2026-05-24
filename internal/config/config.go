package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Notifier NotifierConfig `mapstructure:"notifier"`
	Log      LogConfig      `mapstructure:"log"`
	Node     NodeConfig     `mapstructure:"node"`
}

type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"readTimeout"`
	WriteTimeout int    `mapstructure:"writeTimeout"`
	AdminToken   string `mapstructure:"adminToken"`
}

type DatabaseConfig struct {
	DSN            string `mapstructure:"dsn"`
	MaxOpenConns   int    `mapstructure:"maxOpenConns"`
	MaxIdleConns   int    `mapstructure:"maxIdleConns"`
	ConnMaxLifetime int    `mapstructure:"connMaxLifetime"`
}

type NotifierConfig struct {
	PollInterval int `mapstructure:"pollInterval"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type NodeConfig struct {
	ID string `mapstructure:"id"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	applyDefaults(&cfg)

	if cfg.Node.ID == "" {
		cfg.Node.ID = generateNodeID()
	}

	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 60
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 60
	}
	if cfg.Notifier.PollInterval == 0 {
		cfg.Notifier.PollInterval = 2
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 20
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = 300
	}
}

func (s *ServerConfig) Addr() string {
	return s.Host + ":" + formatPort(s.Port)
}

func formatPort(port int) string {
	if port == 0 {
		return "8080"
	}
	b := make([]byte, 0, 6)
	for port > 0 {
		b = append([]byte{byte('0' + port%10)}, b...)
		port /= 10
	}
	if len(b) == 0 {
		return "8080"
	}
	return string(b)
}

func generateNodeID() string {
	return "node-" + randomString(12)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}