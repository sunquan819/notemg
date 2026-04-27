package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Data     DataConfig     `yaml:"data"`
	Auth     AuthConfig     `yaml:"auth"`
	Editor   EditorConfig   `yaml:"editor"`
	Security SecurityConfig `yaml:"security"`
	Plugins  PluginsConfig  `yaml:"plugins"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	BaseURL string `yaml:"base_url"`
}

type DataConfig struct {
	Dir            string `yaml:"dir"`
	DBName         string `yaml:"db_name"`
	NotesDir       string `yaml:"notes_dir"`
	AttachmentsDir string `yaml:"attachments_dir"`
	IndexDir       string `yaml:"index_dir"`
}

type AuthConfig struct {
	JWTSecret       string        `yaml:"jwt_secret"`
	TokenExpire     time.Duration `yaml:"token_expire"`
	RefreshExpire   time.Duration `yaml:"refresh_expire"`
	BcryptCost      int           `yaml:"bcrypt_cost"`
	MaxLoginAttempts int          `yaml:"max_login_attempts"`
	LockDuration    time.Duration `yaml:"lock_duration"`
}

type EditorConfig struct {
	AutosaveInterval int    `yaml:"autosave_interval"`
	DefaultMode      string `yaml:"default_mode"`
	TabSize          int    `yaml:"tab_size"`
	Theme            string `yaml:"theme"`
}

type SecurityConfig struct {
	EnableHTTPS bool   `yaml:"enable_https"`
	CertFile    string `yaml:"cert_file"`
	KeyFile     string `yaml:"key_file"`
	CSRFEnabled bool   `yaml:"csrf_enabled"`
	RateLimit   int    `yaml:"rate_limit"`
}

type PluginsConfig struct {
	Enabled []string `yaml:"enabled"`
	Dir     string   `yaml:"dir"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host:    "0.0.0.0",
			Port:    8080,
			BaseURL: "",
		},
		Data: DataConfig{
			Dir:            "./data",
			DBName:         "notemg.db",
			NotesDir:       "notes",
			AttachmentsDir: "attachments",
			IndexDir:       "index",
		},
		Auth: AuthConfig{
			JWTSecret:        "change-me-to-a-random-secret",
			TokenExpire:      72 * time.Hour,
			RefreshExpire:    168 * time.Hour,
			BcryptCost:       12,
			MaxLoginAttempts: 5,
			LockDuration:     15 * time.Minute,
		},
		Editor: EditorConfig{
			AutosaveInterval: 1000,
			DefaultMode:      "ir",
			TabSize:          4,
			Theme:            "dark",
		},
		Security: SecurityConfig{
			CSRFEnabled: true,
			RateLimit:   60,
		},
		Plugins: PluginsConfig{
			Dir: "./plugins",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Auth.JWTSecret == "change-me-to-a-random-secret" {
		fmt.Println("WARNING: Using default JWT secret. Please change it in config.")
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Server.Port)
	}
	return nil
}

func (c *Config) DataDir(sub ...string) string {
	parts := append([]string{c.Data.Dir}, sub...)
	return filepath.Join(parts...)
}

func (c *Config) DBPath() string {
	return c.DataDir(c.Data.DBName)
}

func (c *Config) NotesPath() string {
	return c.DataDir(c.Data.NotesDir)
}

func (c *Config) AttachmentsPath() string {
	return c.DataDir(c.Data.AttachmentsDir)
}

func (c *Config) IndexPath() string {
	return c.DataDir(c.Data.IndexDir)
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func (c *Config) EnsureDataDirs() error {
	dirs := []string{
		c.Data.Dir,
		c.NotesPath(),
		c.AttachmentsPath(),
		c.IndexPath(),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create data directory %s: %w", dir, err)
		}
	}
	return nil
}
