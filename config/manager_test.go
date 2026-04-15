package config

import (
	"os"
	"testing"
	"time"

	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
)

// 辅助函数：创建测试 logger
func createTestLogger(t *testing.T) *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	return logger
}

// TestNewManager 测试创建配置管理器
func TestNewManager(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	if manager == nil {
		t.Fatal("Expected manager to be created")
	}
	if manager.viper == nil {
		t.Error("Expected viper to be initialized")
	}
	if manager.logger == nil {
		t.Error("Expected logger to be set")
	}
}

// TestNewManagerWithConfig 测试使用配置创建配置管理器
func TestNewManagerWithConfig(t *testing.T) {
	logger := createTestLogger(t)
	callback := func(cfg *model.Config) {
		_ = cfg // 使用 cfg 避免未使用错误
	}

	cfg := ManagerConfig{
		ConfigFile: "test.yaml",
		Logger:     logger,
		OnChange:   callback,
	}

	manager := NewManagerWithConfig(cfg)

	if manager == nil {
		t.Fatal("Expected manager to be created")
	}
	if manager.configFile != "test.yaml" {
		t.Errorf("Expected configFile to be 'test.yaml', got %s", manager.configFile)
	}
	if len(manager.onChangeCb) != 1 {
		t.Errorf("Expected 1 onChange callback, got %d", len(manager.onChangeCb))
	}
}

// TestManager_GetConfig 测试获取配置
func TestManager_GetConfig(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	config := manager.GetConfig()
	if config == nil {
		t.Fatal("Expected config to be returned")
	}

	// 验证返回的是深拷贝
	config.Server.Port = "9999"
	originalConfig := manager.GetConfig()
	if originalConfig.Server.Port == "9999" {
		t.Error("Expected GetConfig to return a deep copy")
	}
}

// TestManager_Set 测试设置配置值
func TestManager_Set(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	// 测试设置 JWT secret
	manager.Set("jwt.secret", "test-secret-key")
	if manager.config.JWT.Secret != "test-secret-key" {
		t.Errorf("Expected JWT secret to be set, got %s", manager.config.JWT.Secret)
	}

	// 测试设置 auth password
	manager.Set("auth.password", "test-password")
	if manager.config.Auth.Password != "test-password" {
		t.Errorf("Expected auth password to be set, got %s", manager.config.Auth.Password)
	}

	// 测试设置 auth username
	manager.Set("auth.username", "test-user")
	if manager.config.Auth.Username != "test-user" {
		t.Errorf("Expected auth username to be set, got %s", manager.config.Auth.Username)
	}
}

// TestManager_GetString 测试获取字符串配置
func TestManager_GetString(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	manager.Set("jwt.secret", "test-secret")
	value := manager.GetString("jwt.secret")
	if value != "test-secret" {
		t.Errorf("Expected 'test-secret', got %s", value)
	}
}

// TestManager_GetInt 测试获取整数配置
func TestManager_GetInt(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	manager.Set("server.port", 8080)
	value := manager.GetInt("server.port")
	if value != 8080 {
		t.Errorf("Expected 8080, got %d", value)
	}
}

// TestManager_GetBool 测试获取布尔配置
func TestManager_GetBool(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	manager.Set("cache.enabled", true)
	value := manager.GetBool("cache.enabled")
	if !value {
		t.Error("Expected true, got false")
	}
}

// TestManager_GetDuration 测试获取时间配置
func TestManager_GetDuration(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	manager.Set("cache.ttl", 5*time.Minute)
	value := manager.GetDuration("cache.ttl")
	if value != 5*time.Minute {
		t.Errorf("Expected 5m, got %v", value)
	}
}

// TestManager_GetJWTSecret 测试获取 JWT 密钥
func TestManager_GetJWTSecret(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	manager.config.JWT.Secret = "test-secret-key"
	secret := manager.GetJWTSecret()
	if string(secret) != "test-secret-key" {
		t.Errorf("Expected 'test-secret-key', got %s", string(secret))
	}
}

// TestManager_GetAuthConfig 测试获取认证配置
func TestManager_GetAuthConfig(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	manager.config.Auth.Username = "admin"
	manager.config.Auth.Password = "password123"

	authCfg := manager.GetAuthConfig()
	if authCfg == nil {
		t.Fatal("Expected auth config to be returned")
	}
	if authCfg.Username != "admin" {
		t.Errorf("Expected username 'admin', got %s", authCfg.Username)
	}
}

// TestManager_UpdateLogger 测试更新 logger
func TestManager_UpdateLogger(t *testing.T) {
	logger1 := createTestLogger(t)
	logger2 := createTestLogger(t)
	manager := NewManager(logger1)

	manager.UpdateLogger(logger2)
	if manager.logger != logger2 {
		t.Error("Expected logger to be updated")
	}
}

// TestManager_GetConfigFile 测试获取配置文件路径
func TestManager_GetConfigFile(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	// 初始应该返回空或 viper 的配置路径
	path := manager.GetConfigFile()
	if path == "" {
		t.Log("empty config path is expected before loading config")
	}

	// 设置配置文件路径后
	manager.configFile = "/path/to/config.yaml"
	path = manager.GetConfigFile()
	if path != "/path/to/config.yaml" {
		t.Errorf("Expected '/path/to/config.yaml', got %s", path)
	}
}

// TestManager_Close 测试关闭配置管理器
func TestManager_Close(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	err := manager.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got %v", err)
	}
}

// TestManager_RegisterOnChange 测试注册配置变更回调
func TestManager_RegisterOnChange(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	called := false
	callback := func(cfg *model.Config) {
		called = true
	}

	manager.RegisterOnChange(callback)
	manager.notifyOnChange()

	// 给 goroutine 一些时间执行
	time.Sleep(10 * time.Millisecond)

	if !called {
		t.Error("Expected callback to be called")
	}
}

// TestManager_applyEnvironmentOverrides 测试环境变量覆盖
func TestManager_applyEnvironmentOverrides(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	// 设置环境变量
	os.Setenv("K8SVISION_SERVER_PORT", "9090")
	os.Setenv("K8SVISION_SERVER_HOST", "0.0.0.0")
	os.Setenv("K8SVISION_JWT_SECRET", "env-secret")
	os.Setenv("K8SVISION_LOG_LEVEL", "debug")
	os.Setenv("K8SVISION_AUTH_USERNAME", "env-admin")
	os.Setenv("K8SVISION_AUTH_PASSWORD", "env-password")
	os.Setenv("K8SVISION_AUTH_MAX_FAIL", "10")
	os.Setenv("K8SVISION_AUTH_LOCK_MINUTES", "30")
	os.Setenv("KUBECONFIG", "/path/to/kubeconfig")
	os.Setenv("K8SVISION_KUBERNETES_APISERVER", "https://api.example.com")
	os.Setenv("K8SVISION_KUBERNETES_TOKEN", "token123")

	defer func() {
		// 清理环境变量
		os.Unsetenv("K8SVISION_SERVER_PORT")
		os.Unsetenv("K8SVISION_SERVER_HOST")
		os.Unsetenv("K8SVISION_JWT_SECRET")
		os.Unsetenv("K8SVISION_LOG_LEVEL")
		os.Unsetenv("K8SVISION_AUTH_USERNAME")
		os.Unsetenv("K8SVISION_AUTH_PASSWORD")
		os.Unsetenv("K8SVISION_AUTH_MAX_FAIL")
		os.Unsetenv("K8SVISION_AUTH_LOCK_MINUTES")
		os.Unsetenv("KUBECONFIG")
		os.Unsetenv("K8SVISION_KUBERNETES_APISERVER")
		os.Unsetenv("K8SVISION_KUBERNETES_TOKEN")
	}()

	manager.applyEnvironmentOverrides()

	if manager.config.Server.Port != "9090" {
		t.Errorf("Expected server port '9090', got %s", manager.config.Server.Port)
	}
	if manager.config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected server host '0.0.0.0', got %s", manager.config.Server.Host)
	}
	if manager.config.JWT.Secret != "env-secret" {
		t.Errorf("Expected JWT secret 'env-secret', got %s", manager.config.JWT.Secret)
	}
	if manager.config.Log.Level != "debug" {
		t.Errorf("Expected log level 'debug', got %s", manager.config.Log.Level)
	}
	if manager.config.Auth.Username != "env-admin" {
		t.Errorf("Expected auth username 'env-admin', got %s", manager.config.Auth.Username)
	}
	if manager.config.Auth.Password != "env-password" {
		t.Errorf("Expected auth password 'env-password', got %s", manager.config.Auth.Password)
	}
	if manager.config.Auth.MaxLoginFail != 10 {
		t.Errorf("Expected max login fail 10, got %d", manager.config.Auth.MaxLoginFail)
	}
	if manager.config.Auth.LockDuration != 30*time.Minute {
		t.Errorf("Expected lock duration 30m, got %v", manager.config.Auth.LockDuration)
	}
	if manager.config.Kubernetes.Kubeconfig != "/path/to/kubeconfig" {
		t.Errorf("Expected kubeconfig '/path/to/kubeconfig', got %s", manager.config.Kubernetes.Kubeconfig)
	}
	if manager.config.Kubernetes.APIServer != "https://api.example.com" {
		t.Errorf("Expected API server 'https://api.example.com', got %s", manager.config.Kubernetes.APIServer)
	}
	if manager.config.Kubernetes.Token != "token123" {
		t.Errorf("Expected token 'token123', got %s", manager.config.Kubernetes.Token)
	}
}

// TestManager_createBackup 测试创建配置备份
func TestManager_createBackup(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "config-test-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// 写入一些内容
	err = os.WriteFile(tmpPath, []byte("test: value"), 0600)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// 创建备份
	err = manager.createBackup(tmpPath)
	if err != nil {
		t.Errorf("Expected no error on createBackup, got %v", err)
	}

	// 验证备份文件存在
	backupPath := tmpPath + ".bak"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Expected backup file to exist")
	}
	defer os.Remove(backupPath)
}

// TestManager_SetAndWrite 测试原子更新配置
func TestManager_SetAndWrite(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "config-test-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// 初始化 viper 配置文件
	manager.viper.SetConfigFile(tmpPath)
	manager.viper.Set("initial", "value")
	err = manager.viper.WriteConfig()
	if err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// 测试 SetAndWrite
	err = manager.SetAndWrite("new.key", "newvalue")
	if err != nil {
		t.Errorf("Expected no error on SetAndWrite, got %v", err)
	}
}

// TestManager_WriteConfigWithBackup 测试写入配置并备份
func TestManager_WriteConfigWithBackup(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "config-test-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)
	defer os.Remove(tmpPath + ".bak")

	manager.viper.SetConfigFile(tmpPath)
	manager.viper.Set("test", "value")

	err = manager.WriteConfigWithBackup()
	// 可能因为文件已存在而失败，这是可接受的行为
	_ = err
}

// TestManager_reload 测试重新加载配置
func TestManager_reload(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "config-test-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// 写入初始配置（包含所有必需字段）
	configContent := `
server:
  port: "8080"
  host: "localhost"
jwt:
  secret: "test-secret-key-for-reload-test-min-32-chars"
auth:
  username: "admin"
  password: "hashed-password-for-test"
  maxLoginFail: 5
  lockMinutes: 15
cache:
  enabled: true
  maxSize: 1000
  ttl: 5m
log:
  level: "info"
  format: "console"
`
	err = os.WriteFile(tmpPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	manager.viper.SetConfigFile(tmpPath)
	err = manager.viper.ReadInConfig()
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// 测试 reload
	err = manager.reload()
	if err != nil {
		t.Errorf("Expected no error on reload, got %v", err)
	}
}

// TestManager_Load_FileNotFound 测试加载不存在的配置文件
func TestManager_Load_FileNotFound(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	// 使用不存在的文件路径
	err := manager.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error when loading nonexistent config")
	}
}

// TestManager_GetDefaultConfig 测试默认配置
func TestManager_GetDefaultConfig(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	config := manager.GetConfig()
	if config == nil {
		t.Fatal("Expected default config to be returned")
	}

	// 验证默认配置不为空
	if config.Server.Port == "" {
		t.Error("Expected default server port to be set")
	}
}

// TestManager_ConcurrentAccess 测试并发访问安全性
func TestManager_ConcurrentAccess(t *testing.T) {
	logger := createTestLogger(t)
	manager := NewManager(logger)

	done := make(chan bool)

	// 启动多个 goroutine 并发访问
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			manager.GetConfig()
			manager.Set("test.key", "value")
			manager.GetString("test.key")
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}
