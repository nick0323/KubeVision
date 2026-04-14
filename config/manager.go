package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nick0323/K8sVision/model"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Manager 配置管理器
type Manager struct {
	config     *model.Config
	viper      *viper.Viper
	logger     *zap.Logger
	mutex      sync.RWMutex
	watcher    *fsnotify.Watcher
	configFile string
	ctx        context.Context
	cancel     context.CancelFunc
	onChangeCb []func(*model.Config)
	onChangeMu sync.RWMutex
}

// OnChangeCallback 配置变更回调函数类型
type OnChangeCallback func(*model.Config)

// ManagerConfig 配置管理器配置
type ManagerConfig struct {
	ConfigFile string
	Logger     *zap.Logger
	OnChange   OnChangeCallback // 可选：配置变更回调
}

// NewManager 创建配置管理器
func NewManager(logger *zap.Logger) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		config: model.DefaultConfig(),
		viper:  viper.New(),
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// NewManagerWithConfig 使用配置创建配置管理器
func NewManagerWithConfig(cfg ManagerConfig) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		config:     model.DefaultConfig(),
		viper:      viper.New(),
		logger:     cfg.Logger,
		configFile: cfg.ConfigFile,
		ctx:        ctx,
		cancel:     cancel,
	}
	if cfg.OnChange != nil {
		m.onChangeCb = append(m.onChangeCb, cfg.OnChange)
	}
	return m
}

// Load 加载配置文件
func (m *Manager) Load(configFile string) error {
	m.configFile = configFile

	// 设置配置文件
	if configFile != "" {
		m.viper.SetConfigFile(configFile)
	} else {
		m.viper.SetConfigName("config")
		m.viper.AddConfigPath(".")
		m.viper.AddConfigPath("./config")
		m.viper.AddConfigPath("/etc/k8svision")
	}

	// 设置环境变量前缀
	m.viper.SetEnvPrefix("K8SVISION")
	m.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	m.viper.AutomaticEnv()

	// 读取配置文件
	if err := m.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			m.logger.Warn("配置文件未找到，使用默认配置", zap.String("configFile", configFile))
		} else {
			return fmt.Errorf("读取配置文件失败：%w", err)
		}
	} else {
		m.logger.Info("配置文件加载成功", zap.String("configFile", m.viper.ConfigFileUsed()))
	}

	// 解析配置（使用自定义解码钩子）
	if err := m.viper.Unmarshal(m.config); err != nil {
		return fmt.Errorf("解析配置失败：%w", err)
	}

	// 应用环境变量覆盖
	m.applyEnvironmentOverrides()

	// 验证配置
	if err := m.config.Validate(); err != nil {
		return fmt.Errorf("配置验证失败：%w", err)
	}

	m.logger.Info("配置加载完成",
		zap.String("server", m.config.GetServerAddress()),
		zap.String("logLevel", m.config.Log.Level),
	)

	return nil
}

// Watch 监听配置文件变化（使用 viper 内置功能）
func (m *Manager) Watch() error {
	if m.configFile == "" {
		return nil
	}

	// 使用 viper 内置的配置监听
	m.viper.WatchConfig()
	m.viper.OnConfigChange(func(e fsnotify.Event) {
		m.logger.Info("检测到配置文件变化，重新加载配置", zap.String("file", e.Name))
		if err := m.reload(); err != nil {
			m.logger.Error("重新加载配置失败", zap.Error(err))
		} else {
			m.notifyOnChange()
		}
	})

	m.logger.Info("配置文件监听已启动", zap.String("configFile", m.configFile))
	return nil
}

// notifyOnChange 通知配置变更回调
func (m *Manager) notifyOnChange() {
	m.onChangeMu.RLock()
	defer m.onChangeMu.RUnlock()

	config := m.GetConfig()
	for _, cb := range m.onChangeCb {
		go cb(config)
	}
}

// RegisterOnChange 注册配置变更回调
func (m *Manager) RegisterOnChange(cb OnChangeCallback) {
	m.onChangeMu.Lock()
	defer m.onChangeMu.Unlock()
	m.onChangeCb = append(m.onChangeCb, cb)
}

// Close 关闭配置管理器
func (m *Manager) Close() error {
	m.cancel()
	if m.watcher != nil {
		return m.watcher.Close()
	}
	return nil
}

// GetConfig 获取配置（深拷贝）
func (m *Manager) GetConfig() *model.Config {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.config.DeepCopy()
}

// reload 重新加载配置
func (m *Manager) reload() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 重新读取配置文件
	if err := m.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("重新读取配置文件失败：%w", err)
	}

	// 解析配置
	if err := m.viper.Unmarshal(m.config); err != nil {
		return fmt.Errorf("重新解析配置失败：%w", err)
	}

	// 应用环境变量覆盖
	m.applyEnvironmentOverrides()

	// 验证配置
	if err := m.config.Validate(); err != nil {
		return fmt.Errorf("重新验证配置失败：%w", err)
	}

	m.logger.Info("配置重新加载完成")
	return nil
}

// applyEnvironmentOverrides 应用环境变量覆盖
func (m *Manager) applyEnvironmentOverrides() {
	// 服务器配置
	if port := os.Getenv("K8SVISION_SERVER_PORT"); port != "" {
		m.config.Server.Port = port
	}
	if host := os.Getenv("K8SVISION_SERVER_HOST"); host != "" {
		m.config.Server.Host = host
	}

	// Kubernetes 配置
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		m.config.Kubernetes.Kubeconfig = kubeconfig
	}
	if apiServer := os.Getenv("K8SVISION_KUBERNETES_APISERVER"); apiServer != "" {
		m.config.Kubernetes.APIServer = apiServer
	}
	if token := os.Getenv("K8SVISION_KUBERNETES_TOKEN"); token != "" {
		m.config.Kubernetes.Token = token
	}

	// JWT 配置
	if secret := os.Getenv("K8SVISION_JWT_SECRET"); secret != "" {
		m.config.JWT.Secret = secret
	}

	// 日志配置
	if level := os.Getenv("K8SVISION_LOG_LEVEL"); level != "" {
		m.config.Log.Level = level
	}

	// 认证配置
	if username := os.Getenv("K8SVISION_AUTH_USERNAME"); username != "" {
		m.config.Auth.Username = username
	}
	if password := os.Getenv("K8SVISION_AUTH_PASSWORD"); password != "" {
		m.config.Auth.Password = password
	}
	if maxFail := os.Getenv("K8SVISION_AUTH_MAX_FAIL"); maxFail != "" {
		if val, err := parseInt(maxFail); err == nil {
			m.config.Auth.MaxLoginFail = val
		}
	}
	if lockMinutes := os.Getenv("K8SVISION_AUTH_LOCK_MINUTES"); lockMinutes != "" {
		if val, err := parseInt(lockMinutes); err == nil {
			m.config.Auth.LockDuration = time.Duration(val) * time.Minute
		}
	}
}

// parseInt 解析整数
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// GetString 获取字符串配置
func (m *Manager) GetString(key string) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.viper.GetString(key)
}

// GetInt 获取整数配置
func (m *Manager) GetInt(key string) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.viper.GetInt(key)
}

// GetBool 获取布尔配置
func (m *Manager) GetBool(key string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.viper.GetBool(key)
}

// GetDuration 获取时间配置
func (m *Manager) GetDuration(key string) time.Duration {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.viper.GetDuration(key)
}

// Set 设置配置值（同时更新 viper 和内存中的 config）
func (m *Manager) Set(key string, value interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.viper.Set(key, value)

	// 同步更新内存中的 config 对象
	parts := strings.Split(key, ".")
	if len(parts) == 2 {
		switch parts[0] {
		case "auth":
			switch parts[1] {
			case "password":
				if s, ok := value.(string); ok {
					m.config.Auth.Password = s
				}
			case "username":
				if s, ok := value.(string); ok {
					m.config.Auth.Username = s
				}
			}
		case "jwt":
			switch parts[1] {
			case "secret":
				if s, ok := value.(string); ok {
					m.config.JWT.Secret = s
				}
			}
		}
	}
}

// WriteConfig 写入配置文件
func (m *Manager) WriteConfig() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.viper.WriteConfig()
}

// WriteConfigWithBackup 写入配置文件并在覆盖前创建备份
func (m *Manager) WriteConfigWithBackup() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cfgPath := m.viper.ConfigFileUsed()
	if cfgPath == "" {
		cfgPath = m.configFile
	}

	// 创建备份
	if cfgPath != "" {
		if err := m.createBackup(cfgPath); err != nil {
			m.logger.Warn("创建配置备份失败", zap.Error(err))
		}
	}

	return m.viper.WriteConfig()
}

// createBackup 创建配置文件备份
func (m *Manager) createBackup(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	backupPath := path + ".bak"
	return os.WriteFile(backupPath, data, 0600)
}

// SetAndWrite 原子更新：先设置键值，再写入配置（带备份）
func (m *Manager) SetAndWrite(key string, value interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.viper.Set(key, value)
	return m.viper.WriteConfig()
}

// GetConfigFile 返回当前配置文件路径
func (m *Manager) GetConfigFile() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if m.configFile != "" {
		return m.configFile
	}
	return m.viper.ConfigFileUsed()
}

// GetJWTSecret 获取 JWT 密钥
func (m *Manager) GetJWTSecret() []byte {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return []byte(m.config.JWT.Secret)
}

// GetAuthConfig 获取认证配置
func (m *Manager) GetAuthConfig() *model.AuthConfig {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return &m.config.Auth
}

// UpdateLogger 更新 logger 实例
func (m *Manager) UpdateLogger(newLogger *zap.Logger) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.logger = newLogger
}
