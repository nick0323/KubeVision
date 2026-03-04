package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nick0323/K8sVision/model"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Manager struct {
	config     *model.Config
	viper      *viper.Viper
	logger     *zap.Logger
	mutex      sync.RWMutex
	watcher    *fsnotify.Watcher
	configFile string
}

func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		config: model.DefaultConfig(),
		viper:  viper.New(),
		logger: logger,
	}
}

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
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	} else {
		m.logger.Info("配置文件加载成功", zap.String("configFile", m.viper.ConfigFileUsed()))
	}

	// 解析配置
	if err := m.viper.Unmarshal(m.config); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 应用环境变量覆盖
	m.applyEnvironmentOverrides()

	// 验证配置
	if err := m.config.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	m.logger.Info("配置加载完成",
		zap.String("server", m.config.GetServerAddress()),
		zap.String("logLevel", m.config.Log.Level),
	)

	return nil
}

// Watch 监听配置文件变化
func (m *Manager) Watch() error {
	if m.configFile == "" {
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("创建文件监听器失败: %w", err)
	}
	m.watcher = watcher

	// 监听配置文件目录
	configDir := filepath.Dir(m.configFile)
	if err := watcher.Add(configDir); err != nil {
		return fmt.Errorf("监听配置目录失败: %w", err)
	}

	go func() {
		// 简单去抖，避免同一次保存触发多次重载
		var lastReload time.Time
		const debounce = 500 * time.Millisecond
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if filepath.Base(event.Name) == filepath.Base(m.configFile) {
						if time.Since(lastReload) < debounce {
							// 忽略短时间内的重复事件
							continue
						}
						lastReload = time.Now()
						m.logger.Info("检测到配置文件变化，重新加载配置", zap.String("file", event.Name))
						if err := m.reload(); err != nil {
							m.logger.Error("重新加载配置失败", zap.Error(err))
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				m.logger.Error("配置文件监听错误", zap.Error(err))
			}
		}
	}()

	m.logger.Info("配置文件监听已启动", zap.String("configFile", m.configFile))
	return nil
}

// Close 关闭配置管理器
func (m *Manager) Close() error {
	if m.watcher != nil {
		return m.watcher.Close()
	}
	return nil
}

// GetConfig 获取配置
func (m *Manager) GetConfig() *model.Config {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	// 返回副本以防止外部修改
	configCopy := *m.config
	return &configCopy
}

// reload 重新加载配置
func (m *Manager) reload() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 重新读取配置文件
	if err := m.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("重新读取配置文件失败: %w", err)
	}

	// 解析配置
	if err := m.viper.Unmarshal(m.config); err != nil {
		return fmt.Errorf("重新解析配置失败: %w", err)
	}

	// 应用环境变量覆盖
	m.applyEnvironmentOverrides()

	// 验证配置
	if err := m.config.Validate(); err != nil {
		return fmt.Errorf("重新验证配置失败: %w", err)
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

	// Kubernetes配置
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		m.config.Kubernetes.Kubeconfig = kubeconfig
	}
	if apiServer := os.Getenv("K8SVISION_KUBERNETES_APISERVER"); apiServer != "" {
		m.config.Kubernetes.APIServer = apiServer
	}
	if token := os.Getenv("K8SVISION_KUBERNETES_TOKEN"); token != "" {
		m.config.Kubernetes.Token = token
	}

	// JWT配置
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

// parseInt 解析整数 - 使用标准库strconv.Atoi提供更好的性能和错误处理
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

// Set 设置配置值
func (m *Manager) Set(key string, value interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.viper.Set(key, value)
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
	if cfgPath != "" {
		// 创建简单备份
		if data, err := os.ReadFile(cfgPath); err == nil {
			_ = os.WriteFile(cfgPath+".bak", data, 0600)
		}
	}
	return m.viper.WriteConfig()
}

// SetAndWrite 原子更新：先设置键值，再写入配置（带备份）
func (m *Manager) SetAndWrite(key string, value interface{}) error {
	m.mutex.Lock()
	m.viper.Set(key, value)
	m.mutex.Unlock()
	return m.WriteConfigWithBackup()
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

// GetJWTSecret 获取JWT密钥
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

// UpdateLogger 更新logger实例（避免重复创建配置管理器）
func (m *Manager) UpdateLogger(newLogger *zap.Logger) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.logger = newLogger
}
