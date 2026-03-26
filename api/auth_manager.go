package api

import (
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/nick0323/K8sVision/config"
	"go.uber.org/zap"
)

// 认证配置常量
const (
	AuthShardCount      = 32               // 分片数量
	AuthCleanupInterval = 10 * time.Minute // 清理间隔
	AuthExpiryDuration  = time.Hour        // 过期时间
)

type LoginAttempt struct {
	FailCount int       `json:"failCount"`
	LockUntil time.Time `json:"lockUntil"`
	LastFail  time.Time `json:"lastFail"`
}

type AuthManager struct {
	shards [AuthShardCount]*authShard // 使用分片锁减少竞争
	logger *zap.Logger
	config *config.Manager
	stopCh chan struct{}
}

type authShard struct {
	attempts map[string]*LoginAttempt
	mutex    sync.RWMutex
}

func NewAuthManager(logger *zap.Logger, configMgr *config.Manager) *AuthManager {
	am := &AuthManager{
		logger: logger,
		config: configMgr,
		stopCh: make(chan struct{}),
	}

	// 初始化所有分片
	for i := range am.shards {
		am.shards[i] = &authShard{
			attempts: make(map[string]*LoginAttempt),
		}
	}

	go am.startCleanup()
	return am
}

// getShard 根据用户名和IP获取对应的分片
func (am *AuthManager) getShard(username, ip string) *authShard {
	key := fmt.Sprintf("%s|%s", username, ip)
	h := fnv.New32a()
	h.Write([]byte(key))
	return am.shards[h.Sum32()%32]
}

func (am *AuthManager) IsLocked(username, ip string) bool {
	shard := am.getShard(username, ip)
	key := am.makeKey(username, ip)

	shard.mutex.RLock()
	defer shard.mutex.RUnlock()

	attempt, exists := shard.attempts[key]
	if !exists {
		return false
	}

	if time.Now().After(attempt.LockUntil) {
		go am.clearAttempt(username, ip)
		return false
	}

	authConfig := am.config.GetAuthConfig()
	return attempt.FailCount >= authConfig.MaxLoginFail
}

func (am *AuthManager) RecordFailure(username, ip string) {
	shard := am.getShard(username, ip)
	key := am.makeKey(username, ip)
	authConfig := am.config.GetAuthConfig()

	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	attempt, exists := shard.attempts[key]
	if !exists {
		attempt = &LoginAttempt{}
		shard.attempts[key] = attempt
	}

	attempt.FailCount++
	attempt.LastFail = time.Now()

	if attempt.FailCount >= authConfig.MaxLoginFail {
		attempt.LockUntil = time.Now().Add(authConfig.LockDuration)
		am.logger.Warn("用户已被锁定",
			zap.String("username", username),
			zap.String("ip", ip),
			zap.Int("failCount", attempt.FailCount),
			zap.Time("lockUntil", attempt.LockUntil),
		)
	}
}

func (am *AuthManager) RecordSuccess(username, ip string) {
	am.clearAttempt(username, ip)
}

func (am *AuthManager) GetRemainingAttempts(username, ip string) int {
	shard := am.getShard(username, ip)
	key := am.makeKey(username, ip)
	authConfig := am.config.GetAuthConfig()

	shard.mutex.RLock()
	defer shard.mutex.RUnlock()

	attempt, exists := shard.attempts[key]
	if !exists {
		return authConfig.MaxLoginFail
	}

	remaining := authConfig.MaxLoginFail - attempt.FailCount
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

func (am *AuthManager) GetLockTime(username, ip string) time.Duration {
	shard := am.getShard(username, ip)
	key := am.makeKey(username, ip)

	shard.mutex.RLock()
	defer shard.mutex.RUnlock()

	attempt, exists := shard.attempts[key]
	if !exists {
		return 0
	}

	if time.Now().After(attempt.LockUntil) {
		return 0
	}

	return time.Until(attempt.LockUntil)
}

func (am *AuthManager) GetStats() map[string]interface{} {
	totalAttempts := 0
	lockedCount := 0

	for _, shard := range am.shards {
		shard.mutex.RLock()
		totalAttempts += len(shard.attempts)

		now := time.Now()
		authConfig := am.config.GetAuthConfig()
		for _, attempt := range shard.attempts {
			if now.Before(attempt.LockUntil) {
				if attempt.FailCount >= authConfig.MaxLoginFail {
					lockedCount++
				}
			}
		}
		shard.mutex.RUnlock()
	}

	authConfig := am.config.GetAuthConfig()
	return map[string]interface{}{
		"totalAttempts": totalAttempts,
		"lockedUsers":   lockedCount,
		"maxFailCount":  authConfig.MaxLoginFail,
		"lockDuration":  authConfig.LockDuration.String(),
	}
}

func (am *AuthManager) Close() {
	close(am.stopCh)
	am.logger.Info("认证管理器已关闭")
}

func (am *AuthManager) makeKey(username, ip string) string {
	return fmt.Sprintf("%s|%s", username, ip)
}

func (am *AuthManager) clearAttempt(username, ip string) {
	shard := am.getShard(username, ip)
	key := am.makeKey(username, ip)

	shard.mutex.Lock()
	defer shard.mutex.Unlock()
	delete(shard.attempts, key)
}

func (am *AuthManager) startCleanup() {
	ticker := time.NewTicker(AuthCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			am.cleanup()
		case <-am.stopCh:
			return
		}
	}
}

func (am *AuthManager) cleanup() {
	cleaned := 0

	for _, shard := range am.shards {
		shard.mutex.Lock()
		now := time.Now()

		for key, attempt := range shard.attempts {
			if now.After(attempt.LockUntil) && now.Sub(attempt.LastFail) > AuthExpiryDuration {
				delete(shard.attempts, key)
				cleaned++
			}
		}
		shard.mutex.Unlock()
	}

	if cleaned > 0 {
		am.logger.Debug("清理过期登录记录", zap.Int("count", cleaned))
	}
}
