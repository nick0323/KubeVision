package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ==================== 常量定义 ====================

const (
	// K8s  ConfigMap 配置
	ConfigMapNamespace = "k8svision-system"
	ConfigMapName      = "k8svision-config"
	ConfigDataKey      = "auth-config.json"

	// 标签
	LabelApp       = "app.kubernetes.io/name"
	LabelComponent = "app.kubernetes.io/component"

	// 默认值
	defaultSyncInterval = 5 * time.Minute
	defaultTimeout      = 30 * time.Second
)

// ==================== 配置存储 ====================

// AuthConfigData 认证配置数据（用于存储）
type AuthConfigData struct {
	Username        string        `json:"username"`
	PasswordHash    string        `json:"passwordHash"`
	MaxLoginFail    int           `json:"maxLoginFail"`
	LockDuration    time.Duration `json:"lockDuration"`
	SessionTimeout  time.Duration `json:"sessionTimeout"`
	EnableRateLimit bool          `json:"enableRateLimit"`
	RateLimit       int           `json:"rateLimit"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	UpdatedBy       string        `json:"updatedBy"`
}

// ToAuthConfig 转换为认证配置
func (d *AuthConfigData) ToAuthConfig() *model.AuthConfig {
	return &model.AuthConfig{
		Username:        d.Username,
		Password:        d.PasswordHash,
		MaxLoginFail:    d.MaxLoginFail,
		LockDuration:    d.LockDuration,
		SessionTimeout:  d.SessionTimeout,
		EnableRateLimit: d.EnableRateLimit,
		RateLimit:       d.RateLimit,
	}
}

// FromAuthConfig 从认证配置转换
func (d *AuthConfigData) FromAuthConfig(cfg *model.AuthConfig, username string) {
	d.Username = cfg.Username
	d.PasswordHash = cfg.Password
	d.MaxLoginFail = cfg.MaxLoginFail
	d.LockDuration = cfg.LockDuration
	d.SessionTimeout = cfg.SessionTimeout
	d.EnableRateLimit = cfg.EnableRateLimit
	d.RateLimit = cfg.RateLimit
	d.UpdatedAt = time.Now()
	d.UpdatedBy = username
}

// ==================== ConfigMap 存储管理器 ====================

// ConfigMapStore ConfigMap 存储管理器
type ConfigMapStore struct {
	clientset     *kubernetes.Clientset
	logger        *zap.Logger
	namespace     string
	configMapName string
	syncInterval  time.Duration
	timeout       time.Duration

	// 本地缓存
	localConfig *AuthConfigData
	cacheMu     sync.RWMutex

	// 后台同步
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewConfigMapStore 创建 ConfigMap 存储管理器
func NewConfigMapStore(
	clientset *kubernetes.Clientset,
	logger *zap.Logger,
	namespace string,
) *ConfigMapStore {
	store := &ConfigMapStore{
		clientset:     clientset,
		logger:        logger,
		namespace:     namespace,
		configMapName: ConfigMapName,
		syncInterval:  defaultSyncInterval,
		timeout:       defaultTimeout,
		stopCh:        make(chan struct{}),
	}

	// 如果 namespace 为空，使用默认值
	if store.namespace == "" {
		store.namespace = ConfigMapNamespace
	}

	return store
}

// SetSyncInterval 设置同步间隔
func (s *ConfigMapStore) SetSyncInterval(interval time.Duration) {
	s.syncInterval = interval
}

// Start 启动后台同步
func (s *ConfigMapStore) Start() error {
	// 首次加载
	if err := s.LoadConfig(); err != nil {
		s.logger.Warn("initial config load failed, will retry", zap.Error(err))
	}

	// 启动后台同步
	s.wg.Add(1)
	go s.syncLoop()

	s.logger.Info("configmap store started",
		zap.String("namespace", s.namespace),
		zap.String("configmap", s.configMapName),
		zap.Duration("syncInterval", s.syncInterval),
	)

	return nil
}

// Stop 停止后台同步
func (s *ConfigMapStore) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	s.logger.Info("configmap store stopped")
}

// syncLoop 后台同步循环
func (s *ConfigMapStore) syncLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.LoadConfig(); err != nil {
				s.logger.Error("config sync failed", zap.Error(err))
			} else {
				s.logger.Debug("config synced")
			}
		case <-s.stopCh:
			return
		}
	}
}

// LoadConfig 从 ConfigMap 加载配置
func (s *ConfigMapStore) LoadConfig() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	// 获取 ConfigMap
	cm, err := s.clientset.CoreV1().ConfigMaps(s.namespace).Get(ctx, s.configMapName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// ConfigMap 不存在，创建默认的
			return s.createDefaultConfigMap(ctx)
		}
		return fmt.Errorf("failed to get configmap: %w", err)
	}

	// 解析配置数据
	configData, exists := cm.Data[ConfigDataKey]
	if !exists {
		return fmt.Errorf("config data key not found: %s", ConfigDataKey)
	}

	var config AuthConfigData
	if err := json.Unmarshal([]byte(configData), &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 更新本地缓存
	s.cacheMu.Lock()
	s.localConfig = &config
	s.cacheMu.Unlock()

	s.logger.Info("config loaded from configmap",
		zap.String("username", config.Username),
		zap.Time("updatedAt", config.UpdatedAt),
	)

	return nil
}

// SaveConfig 保存配置到 ConfigMap
func (s *ConfigMapStore) SaveConfig(config *AuthConfigData, username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	// 更新配置
	config.UpdatedAt = time.Now()
	config.UpdatedBy = username

	// 序列化配置
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 获取或创建 ConfigMap
	cm, err := s.clientset.CoreV1().ConfigMaps(s.namespace).Get(ctx, s.configMapName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return s.createConfigMap(ctx, config)
		}
		return fmt.Errorf("failed to get configmap: %w", err)
	}

	// 更新 ConfigMap
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	cm.Data[ConfigDataKey] = string(configData)

	// 添加标签
	if cm.Labels == nil {
		cm.Labels = make(map[string]string)
	}
	cm.Labels[LabelApp] = "k8svision"
	cm.Labels[LabelComponent] = "auth-config"

	// 更新
	_, err = s.clientset.CoreV1().ConfigMaps(s.namespace).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update configmap: %w", err)
	}

	// 更新本地缓存
	s.cacheMu.Lock()
	s.localConfig = config
	s.cacheMu.Unlock()

	s.logger.Info("config saved to configmap",
		zap.String("username", config.Username),
		zap.String("updatedBy", username),
	)

	return nil
}

// GetConfig 获取配置（从本地缓存）
func (s *ConfigMapStore) GetConfig() *AuthConfigData {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	if s.localConfig == nil {
		return nil
	}

	// 返回副本
	config := *s.localConfig
	return &config
}

// GetAuthConfig 获取认证配置
func (s *ConfigMapStore) GetAuthConfig() *model.AuthConfig {
	config := s.GetConfig()
	if config == nil {
		authConfig := model.DefaultConfig().Auth
		return &authConfig
	}
	return config.ToAuthConfig()
}

// UpdatePassword 更新密码
func (s *ConfigMapStore) UpdatePassword(newPasswordHash, username string) error {
	config := s.GetConfig()
	if config == nil {
		return fmt.Errorf("config not loaded")
	}

	config.PasswordHash = newPasswordHash
	return s.SaveConfig(config, username)
}

// createDefaultConfigMap 创建默认 ConfigMap
func (s *ConfigMapStore) createDefaultConfigMap(ctx context.Context) error {
	defaultConfig := model.DefaultConfig().Auth
	configData := &AuthConfigData{}
	configData.FromAuthConfig(&defaultConfig, "system")

	return s.createConfigMap(ctx, configData)
}

// createConfigMap 创建 ConfigMap
func (s *ConfigMapStore) createConfigMap(ctx context.Context, config *AuthConfigData) error {
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.configMapName,
			Namespace: s.namespace,
			Labels: map[string]string{
				LabelApp:       "k8svision",
				LabelComponent: "auth-config",
			},
		},
		Data: map[string]string{
			ConfigDataKey: string(configData),
		},
	}

	_, err = s.clientset.CoreV1().ConfigMaps(s.namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create configmap: %w", err)
	}

	// 更新本地缓存
	s.cacheMu.Lock()
	s.localConfig = config
	s.cacheMu.Unlock()

	s.logger.Info("configmap created",
		zap.String("namespace", s.namespace),
		zap.String("name", s.configMapName),
	)

	return nil
}

// ==================== ConfigMap 存储管理器（全局单例） ====================

var (
	globalConfigStore *ConfigMapStore
	storeOnce         sync.Once
)

// GetConfigStore 获取全局配置存储
func GetConfigStore() *ConfigMapStore {
	return globalConfigStore
}

// InitConfigStore 初始化全局配置存储
func InitConfigStore(
	clientset *kubernetes.Clientset,
	logger *zap.Logger,
	namespace string,
) error {
	var initErr error
	storeOnce.Do(func() {
		globalConfigStore = NewConfigMapStore(clientset, logger, namespace)
		initErr = globalConfigStore.Start()
	})
	return initErr
}

// ==================== 密码管理器 ====================

// passwordManager 密码管理器
type passwordManager struct{}

// newPasswordEncoder 创建密码管理器
func newPasswordEncoder() *passwordManager {
	return &passwordManager{}
}

// HashPassword 哈希密码
func (pm *passwordManager) hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("生成盐失败：%w", err)
	}

	passwordWithSalt := password + base64.URLEncoding.EncodeToString(salt)

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(passwordWithSalt), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败：%w", err)
	}

	return base64.URLEncoding.EncodeToString(salt) + ":" + string(hashedBytes), nil
}

// VerifyPassword 验证密码
func (pm *passwordManager) verifyPassword(password, hashedPassword string) bool {
	parts := strings.Split(hashedPassword, ":")
	if len(parts) != 2 {
		return false
	}

	salt, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}

	passwordWithSalt := password + base64.URLEncoding.EncodeToString(salt)

	err = bcrypt.CompareHashAndPassword([]byte(parts[1]), []byte(passwordWithSalt))
	return err == nil
}

// ==================== 密码编码器 ====================

// PasswordEncoder 密码编码器
type PasswordEncoder struct {
	pm *passwordManager
}

// NewPasswordEncoder 创建密码编码器
func NewPasswordEncoder() *PasswordEncoder {
	return &PasswordEncoder{
		pm: newPasswordEncoder(),
	}
}

// Encode 编码密码
func (e *PasswordEncoder) Encode(password string) (string, error) {
	return e.pm.hashPassword(password)
}

// Verify 验证密码
func (e *PasswordEncoder) Verify(password, hashedPassword string) bool {
	return e.pm.verifyPassword(password, hashedPassword)
}

// EncodeBase64  Base64 编码（用于 ConfigMap）
func (e *PasswordEncoder) EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64  Base64 解码
func (e *PasswordEncoder) DecodeBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}
