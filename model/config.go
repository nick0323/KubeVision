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
	Port          string   `mapstructure:"port" json:"port"`
	Host          string   `mapstructure:"host" json:"host"`
	AllowedOrigin []string `mapstructure:"allowedOrigin" json:"allowedOrigin"`
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
	Token      string        `mapstructure:"token" json:"-"` // 不输出到 JSON
	APIServer  string        `mapstructure:"apiServer" json:"apiServer"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string        `mapstructure:"secret" json:"-"` // 不输出到 JSON
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
	Password        string        `mapstructure:"password" json:"-"` // 不输出到 JSON
	MaxLoginFail    int           `mapstructure:"maxLoginFail" json:"max_login_fail"`
	LockDuration    time.Duration `mapstructure:"lockDuration" json:"lock_duration"`
	SessionTimeout  time.Duration `mapstructure:"sessionTimeout" json:"session_timeout"`
	EnableRateLimit bool          `mapstructure:"enableRateLimit" json:"enable_rate_limit"`
	RateLimit       int           `mapstructure:"rateLimit" json:"rate_limit"`
	BcryptCost      int           `mapstructure:"bcryptCost" json:"bcrypt_cost"` // bcrypt 成本因子
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled         bool          `mapstructure:"enabled" json:"enabled"`
	Type            string        `mapstructure:"type" json:"type"`
	TTL             time.Duration `mapstructure:"ttl" json:"ttl"`
	MaxSize         int           `mapstructure:"maxSize" json:"max_size"`
	CleanupInterval time.Duration `mapstructure:"cleanupInterval" json:"cleanup_interval"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:          "8080",
			Host:          "0.0.0.0",
			AllowedOrigin: []string{"http://localhost:3000", "http://localhost:8080"},
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
			Password:        "", // 首次启动时应生成或要求设置
			MaxLoginFail:    5,
			LockDuration:    10 * time.Minute,
			SessionTimeout:  24 * time.Hour,
			EnableRateLimit: true,
			RateLimit:       100,
			BcryptCost:      12, // 推荐值
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

// Validate 验证配置
func (c *Config) Validate() error {
	var errs []error

	// 验证服务器配置
	if err := c.Server.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("服务器配置：%w", err))
	}

	// 验证 JWT 配置
	if err := c.JWT.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("JWT 配置：%w", err))
	}

	// 验证认证配置
	if err := c.Auth.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("认证配置：%w", err))
	}

	// 验证日志配置
	if err := c.Log.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("日志配置：%w", err))
	}

	// 验证缓存配置
	if err := c.Cache.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("缓存配置：%w", err))
	}

	// 验证 Kubernetes 配置
	if err := c.Kubernetes.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("Kubernetes 配置：%w", err))
	}

	if len(errs) > 0 {
		errMsg := fmt.Sprintf("配置验证失败：%d 个错误\n", len(errs))
		for i, err := range errs {
			errMsg += fmt.Sprintf("  %d. %v\n", i+1, err)
		}
		return fmt.Errorf("%s", errMsg)
	}
	return nil
}

// Validate 验证服务器配置
func (c *ServerConfig) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("服务器端口不能为空")
	}
	if c.Host == "" {
		return fmt.Errorf("服务器主机不能为空")
	}
	return nil
}

// Validate 验证 JWT 配置
func (c *JWTConfig) Validate() error {
	// Secret 必须配置
	if c.Secret == "" {
		return fmt.Errorf("JWT secret 未配置，请设置环境变量 K8SVISION_JWT_SECRET")
	}
	// 检查长度
	if len(c.Secret) < 16 {
		return fmt.Errorf("JWT 密钥长度至少 16 位字符，当前长度：%d", len(c.Secret))
	}
	if c.Expiration <= 0 {
		return fmt.Errorf("JWT 过期时间必须大于 0")
	}
	return nil
}

// Validate 验证认证配置
func (c *AuthConfig) Validate() error {
	if c.Username == "" {
		return fmt.Errorf("认证用户名不能为空，请设置环境变量 K8SVISION_AUTH_USERNAME")
	}
	// Password 必须配置
	if c.Password == "" {
		return fmt.Errorf("认证密码未配置，请设置环境变量 K8SVISION_AUTH_PASSWORD")
	}
	if c.MaxLoginFail <= 0 {
		return fmt.Errorf("最大登录失败次数必须大于 0")
	}
	if c.LockDuration <= 0 {
		return fmt.Errorf("锁定时间必须大于 0")
	}
	// BcryptCost 为 0 时使用默认值 12
	if c.BcryptCost == 0 {
		c.BcryptCost = 12
	}
	return nil
}

// Validate 验证日志配置
func (c *LogConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Level] {
		return fmt.Errorf("无效的日志级别：%s", c.Level)
	}

	validFormats := map[string]bool{
		"json":    true,
		"console": true,
	}
	if !validFormats[c.Format] {
		return fmt.Errorf("无效的日志格式：%s", c.Format)
	}
	return nil
}

// Validate 验证缓存配置
func (c *CacheConfig) Validate() error {
	if c.Enabled {
		if c.TTL <= 0 {
			return fmt.Errorf("缓存 TTL 必须大于 0")
		}
		if c.MaxSize <= 0 {
			return fmt.Errorf("缓存最大大小必须大于 0")
		}
	}
	return nil
}

// Validate 验证 Kubernetes 配置
func (c *KubernetesConfig) Validate() error {
	if c.QPS <= 0 {
		return fmt.Errorf("Kubernetes QPS 必须大于 0")
	}
	if c.Burst <= 0 {
		return fmt.Errorf("Kubernetes Burst 必须大于 0")
	}
	if c.Timeout <= 0 {
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

// DeepCopy 创建配置的深拷贝
func (c *Config) DeepCopy() *Config {
	if c == nil {
		return nil
	}
	return &Config{
		Server:     c.Server.DeepCopy(),
		Kubernetes: c.Kubernetes.DeepCopy(),
		JWT:        c.JWT.DeepCopy(),
		Log:        c.Log,
		Auth:       c.Auth.DeepCopy(),
		Cache:      c.Cache.DeepCopy(),
	}
}

// DeepCopy 创建服务器配置的深拷贝
func (c *ServerConfig) DeepCopy() ServerConfig {
	if c == nil {
		return ServerConfig{}
	}

	// 深拷贝切片字段
	var allowedOriginCopy []string
	if c.AllowedOrigin != nil {
		allowedOriginCopy = make([]string, len(c.AllowedOrigin))
		copy(allowedOriginCopy, c.AllowedOrigin)
	}

	return ServerConfig{
		Port:          c.Port,
		Host:          c.Host,
		AllowedOrigin: allowedOriginCopy,
	}
}

// DeepCopy 创建 Kubernetes 配置的深拷贝
func (c *KubernetesConfig) DeepCopy() KubernetesConfig {
	if c == nil {
		return KubernetesConfig{}
	}

	// KubernetesConfig 只包含基本类型和字符串，直接拷贝即可
	return KubernetesConfig{
		Kubeconfig: c.Kubeconfig,
		Timeout:    c.Timeout,
		QPS:        c.QPS,
		Burst:      c.Burst,
		Insecure:   c.Insecure,
		CAFile:     c.CAFile,
		CertFile:   c.CertFile,
		KeyFile:    c.KeyFile,
		Token:      c.Token,
		APIServer:  c.APIServer,
	}
}

// DeepCopy 创建 JWT 配置的深拷贝
func (c *JWTConfig) DeepCopy() JWTConfig {
	if c == nil {
		return JWTConfig{}
	}

	return JWTConfig{
		Secret:     c.Secret,
		Expiration: c.Expiration,
	}
}

// DeepCopy 创建认证配置的深拷贝
func (c *AuthConfig) DeepCopy() AuthConfig {
	if c == nil {
		return AuthConfig{}
	}

	return AuthConfig{
		Username:        c.Username,
		Password:        c.Password,
		MaxLoginFail:    c.MaxLoginFail,
		LockDuration:    c.LockDuration,
		SessionTimeout:  c.SessionTimeout,
		EnableRateLimit: c.EnableRateLimit,
		RateLimit:       c.RateLimit,
		BcryptCost:      c.BcryptCost,
	}
}

// DeepCopy 创建缓存配置的深拷贝
func (c *CacheConfig) DeepCopy() CacheConfig {
	if c == nil {
		return CacheConfig{}
	}

	return CacheConfig{
		Enabled:         c.Enabled,
		Type:            c.Type,
		TTL:             c.TTL,
		MaxSize:         c.MaxSize,
		CleanupInterval: c.CleanupInterval,
	}
}

// SafeString 返回配置的安全字符串表示（不暴露敏感信息）
func (c *Config) SafeString() string {
	return fmt.Sprintf("Config{Server:%s Auth:%s Log:%s}",
		c.GetServerAddress(),
		c.Auth.Username,
		c.Log.Level)
}

// MaskPassword 返回掩码后的密码（用于日志）
func (c *AuthConfig) MaskPassword() string {
	if c.Password == "" {
		return "(empty)"
	}
	if len(c.Password) <= 4 {
		return "****"
	}
	return c.Password[:2] + "****" + c.Password[len(c.Password)-2:]
}
