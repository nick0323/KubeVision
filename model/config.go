package model

import (
	"fmt"
	"time"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server" json:"server"`
	Kubernetes KubernetesConfig `mapstructure:"kubernetes" json:"kubernetes"`
	JWT        JWTConfig        `mapstructure:"jwt" json:"jwt"`
	Log        LogConfig        `mapstructure:"log" json:"log"`
	Auth       AuthConfig       `mapstructure:"auth" json:"auth"`
	Cache      CacheConfig      `mapstructure:"cache" json:"cache"`
}

type ServerConfig struct {
	Port             string   `mapstructure:"port" json:"port"`
	Host             string   `mapstructure:"host" json:"host"`
	AllowedOrigin    []string `mapstructure:"allowedOrigin" json:"allowedOrigin"`
	MaxWsConnections int      `mapstructure:"maxWsConnections" json:"maxWsConnections"`
}

type KubernetesConfig struct {
	Kubeconfig string        `mapstructure:"kubeconfig" json:"kubeconfig"`
	Timeout    time.Duration `mapstructure:"timeout" json:"timeout"`
	QPS        float32       `mapstructure:"qps" json:"qps"`
	Burst      int           `mapstructure:"burst" json:"burst"`
	Insecure   bool          `mapstructure:"insecure" json:"insecure"`
	CAFile     string        `mapstructure:"caFile" json:"caFile"`
	CertFile   string        `mapstructure:"certFile" json:"certFile"`
	KeyFile    string        `mapstructure:"keyFile" json:"keyFile"`
	Token      string        `mapstructure:"token" json:"-"`
	APIServer  string        `mapstructure:"apiServer" json:"apiServer"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret" json:"-"`
	Expiration time.Duration `mapstructure:"expiration" json:"expiration"`
}

type LogConfig struct {
	Level  string `mapstructure:"level" json:"level"`
	Format string `mapstructure:"format" json:"format"`
}

type AuthConfig struct {
	Username        string        `mapstructure:"username" json:"username"`
	Password        string        `mapstructure:"password" json:"-"`
	MaxLoginFail    int           `mapstructure:"maxLoginFail" json:"max_login_fail"`
	LockDuration    time.Duration `mapstructure:"lockDuration" json:"lock_duration"`
	SessionTimeout  time.Duration `mapstructure:"sessionTimeout" json:"session_timeout"`
	EnableRateLimit bool          `mapstructure:"enableRateLimit" json:"enable_rate_limit"`
	RateLimit       int           `mapstructure:"rateLimit" json:"rate_limit"`
	BcryptCost      int           `mapstructure:"bcryptCost" json:"bcrypt_cost"`
}

type CacheConfig struct {
	Enabled         bool          `mapstructure:"enabled" json:"enabled"`
	Type            string        `mapstructure:"type" json:"type"`
	TTL             time.Duration `mapstructure:"ttl" json:"ttl"`
	MaxSize         int           `mapstructure:"maxSize" json:"max_size"`
	CleanupInterval time.Duration `mapstructure:"cleanupInterval" json:"cleanup_interval"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:             "8080",
			Host:             "0.0.0.0",
			AllowedOrigin:    []string{"http://localhost:3000", "http://localhost:8080"},
			MaxWsConnections: 100,
		},
		Kubernetes: KubernetesConfig{
			Timeout:  30 * time.Second,
			QPS:      100,
			Burst:    200,
			Insecure: false,
		},
		JWT: JWTConfig{
			Secret:     "",
			Expiration: 24 * time.Hour,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		Auth: AuthConfig{
			Username:        "admin",
			Password:        "",
			MaxLoginFail:    5,
			LockDuration:    10 * time.Minute,
			SessionTimeout:  24 * time.Hour,
			EnableRateLimit: true,
			RateLimit:       100,
			BcryptCost:      12,
		},
		Cache: CacheConfig{
			Enabled:         true,
			Type:            "memory",
			TTL:             5 * time.Minute,
			MaxSize:         1000,
			CleanupInterval: 10 * time.Minute,
		},
	}
}

func (c *Config) Validate() error {
	var errs []error

	if err := c.Server.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("server config: %w", err))
	}
	if err := c.Log.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("log config: %w", err))
	}
	if err := c.Cache.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("cache config: %w", err))
	}
	if err := c.Kubernetes.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("Kubernetes config: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed: %d errors", len(errs))
	}
	return nil
}

func (c *ServerConfig) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("server port cannot be empty")
	}
	if c.Host == "" {
		return fmt.Errorf("server host cannot be empty")
	}
	return nil
}

func (c *LogConfig) Validate() error {
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Level] {
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	validFormats := map[string]bool{"json": true, "console": true}
	if !validFormats[c.Format] {
		return fmt.Errorf("invalid log format: %s", c.Format)
	}
	return nil
}

func (c *CacheConfig) Validate() error {
	if c.Enabled {
		if c.TTL <= 0 {
			return fmt.Errorf("cache TTL must be greater than 0")
		}
		if c.MaxSize <= 0 {
			return fmt.Errorf("cache max size must be greater than 0")
		}
	}
	return nil
}

func (c *KubernetesConfig) Validate() error {
	if c.QPS <= 0 {
		return fmt.Errorf("Kubernetes QPS must be greater than 0")
	}
	if c.Burst <= 0 {
		return fmt.Errorf("Kubernetes Burst must be greater than 0")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("Kubernetes timeout must be greater than 0")
	}
	return nil
}

func (c *Config) GetServerAddress() string {
	return c.Server.Host + ":" + c.Server.Port
}

func (c *Config) IsDevelopment() bool {
	return c.Log.Level == "debug"
}
