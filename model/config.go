package model

import (
	"fmt"
	"time"
)

// Config 应用配置
type Config struct {
	Server     ServerConfig     `mapstructure:"server" json:"server"`
	Kubernetes KubernetesConfig `mapstructure:"kubernetes" json:"kubernetes"`
	JWT        JWTConfig        `mapstructure:"jwt" json:"jwt"`
	Log        LogConfig        `mapstructure:"log" json:"log"`
	Auth       AuthConfig       `mapstructure:"auth" json:"auth"`
	Cache      CacheConfig      `mapstructure:"cache" json:"cache"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `mapstructure:"port" json:"port"`
	Host string `mapstructure:"host" json:"host"`
}

// KubernetesConfig Kubernetes 配置
type KubernetesConfig struct {
	Kubeconfig string        `mapstructure:"kubeconfig" json:"kubeconfig"`
	Timeout    time.Duration `mapstructure:"timeout" json:"timeout"`
	QPS        float32       `mapstructure:"qps" json:"qps"`
	Burst      int           `mapstructure:"burst" json:"burst"`
	Insecure   bool          `mapstructure:"insecure" json:"insecure"`
	CAFile     string        `mapstructure:"caFile" json:"caFile"`
	CertFile   string        `mapstructure:"certFile" json:"certFile"`
	KeyFile    string        `mapstructure:"keyFile" json:"keyFile"`
	Token      string        `mapstructure:"token" json:"token"`
	APIServer  string        `mapstructure:"apiServer" json:"apiServer"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string        `mapstructure:"secret" json:"secret"`
	Expiration time.Duration `mapstructure:"expiration" json:"expiration"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level" json:"level"`
	Format string `mapstructure:"format" json:"format"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Username        string        `mapstructure:"username" json:"username"`
	Password        string        `mapstructure:"password" json:"password"`
	MaxLoginFail    int           `mapstructure:"maxLoginFail" json:"maxLoginFail"`
	LockDuration    time.Duration `mapstructure:"lockDuration" json:"lockDuration"`
	SessionTimeout  time.Duration `mapstructure:"sessionTimeout" json:"sessionTimeout"`
	EnableRateLimit bool          `mapstructure:"enableRateLimit" json:"enableRateLimit"`
	RateLimit       int           `mapstructure:"rateLimit" json:"rateLimit"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled         bool          `mapstructure:"enabled" json:"enabled"`
	Type            string        `mapstructure:"type" json:"type"`
	TTL             time.Duration `mapstructure:"ttl" json:"ttl"`
	MaxSize         int           `mapstructure:"maxSize" json:"maxSize"`
	CleanupInterval time.Duration `mapstructure:"cleanupInterval" json:"cleanupInterval"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		Kubernetes: KubernetesConfig{
			Timeout:  30 * time.Second,
			QPS:      100,
			Burst:    200,
			Insecure: true,
		},
		JWT: JWTConfig{
			Secret:     "k8svision-default-jwt-secret-key-32-chars",
			Expiration: 24 * time.Hour,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		Auth: AuthConfig{
			Username:        "admin",
			Password:        "admin123!",
			MaxLoginFail:    5,
			LockDuration:    10 * time.Minute,
			SessionTimeout:  24 * time.Hour,
			EnableRateLimit: true,
			RateLimit:       100,
		},
		Cache: CacheConfig{
			Enabled: true,
			TTL:     5 * time.Minute,
			MaxSize: 1000,
		},
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证服务器配置
	if c.Server.Port == "" {
		return fmt.Errorf("服务器端口不能为空")
	}
	if c.Server.Host == "" {
		return fmt.Errorf("服务器主机不能为空")
	}

	// 验证 JWT 配置
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT 密钥不能为空，请设置环境变量 K8SVISION_JWT_SECRET")
	}
	if len(c.JWT.Secret) < 16 {
		return fmt.Errorf("JWT 密钥长度至少 16 位字符，当前长度：%d", len(c.JWT.Secret))
	}
	if c.JWT.Expiration <= 0 {
		return fmt.Errorf("JWT 过期时间必须大于 0")
	}

	// 验证认证配置
	if c.Auth.Username == "" {
		return fmt.Errorf("认证用户名不能为空，请设置环境变量 K8SVISION_AUTH_USERNAME")
	}
	if c.Auth.Password == "" {
		return fmt.Errorf("认证密码不能为空，请设置环境变量 K8SVISION_AUTH_PASSWORD")
	}
	if c.Auth.MaxLoginFail <= 0 {
		return fmt.Errorf("最大登录失败次数必须大于 0")
	}
	if c.Auth.LockDuration <= 0 {
		return fmt.Errorf("锁定时间必须大于 0")
	}

	// 验证日志配置
	switch c.Log.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("无效的日志级别：%s", c.Log.Level)
	}

	switch c.Log.Format {
	case "json", "console":
	default:
		return fmt.Errorf("无效的日志格式：%s", c.Log.Format)
	}

	// 验证缓存配置
	if c.Cache.Enabled {
		if c.Cache.TTL <= 0 {
			return fmt.Errorf("缓存 TTL 必须大于 0")
		}
		if c.Cache.MaxSize <= 0 {
			return fmt.Errorf("缓存最大大小必须大于 0")
		}
	}

	// 验证 Kubernetes 配置
	if c.Kubernetes.QPS <= 0 {
		return fmt.Errorf("Kubernetes QPS 必须大于 0")
	}
	if c.Kubernetes.Burst <= 0 {
		return fmt.Errorf("Kubernetes Burst 必须大于 0")
	}
	if c.Kubernetes.Timeout <= 0 {
		return fmt.Errorf("Kubernetes 超时时间必须大于 0")
	}

	return nil
}

// GetServerAddress 获取服务器地址
func (c *Config) GetServerAddress() string {
	return c.Server.Host + ":" + c.Server.Port
}

// IsDevelopment 判断是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.Log.Level == "debug"
}
