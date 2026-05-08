package middleware

import (
	"sync"
	"time"
)

// TokenBlacklist 管理已注销的 JWT token
type TokenBlacklist struct {
	mu            sync.RWMutex
	tokens        map[string]time.Time // jti -> expire time
	maxSize       int
	cleanupTicker *time.Ticker
	stopCh        chan struct{}
	closeOnce     sync.Once
}

// NewTokenBlacklist 创建一个新的 token 黑名单
func NewTokenBlacklist(maxSize int) *TokenBlacklist {
	tb := &TokenBlacklist{
		tokens:  make(map[string]time.Time, maxSize),
		maxSize: maxSize,
		stopCh:  make(chan struct{}),
	}

	// 启动定期清理过期 token
	tb.cleanupTicker = time.NewTicker(5 * time.Minute)
	go tb.cleanupWorker()

	return tb
}

// Add 将 token 加入黑名单
func (tb *TokenBlacklist) Add(jti string, expireTime time.Time) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 如果黑名单已满，清理过期 token
	if len(tb.tokens) >= tb.maxSize {
		tb.cleanupExpired()
	}

	tb.tokens[jti] = expireTime
}

// IsBlacklisted 检查 token 是否在黑名单中
func (tb *TokenBlacklist) IsBlacklisted(jti string) bool {
	tb.mu.RLock()
	expireTime, exists := tb.tokens[jti]
	if !exists {
		tb.mu.RUnlock()
		return false
	}

	// 如果 token 已过期，从黑名单中移除
	if time.Now().After(expireTime) {
		tb.mu.RUnlock()
		tb.Remove(jti)
		return false
	}

	tb.mu.RUnlock()
	return true
}

// Remove 从黑名单中移除 token
func (tb *TokenBlacklist) Remove(jti string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	delete(tb.tokens, jti)
}

// cleanupExpired 清理已过期的 token
func (tb *TokenBlacklist) cleanupExpired() {
	now := time.Now()
	for jti, expireTime := range tb.tokens {
		if now.After(expireTime) {
			delete(tb.tokens, jti)
		}
	}
}

// cleanupWorker 定期清理过期 token
func (tb *TokenBlacklist) cleanupWorker() {
	for {
		select {
		case <-tb.cleanupTicker.C:
			tb.mu.Lock()
			tb.cleanupExpired()
			tb.mu.Unlock()
		case <-tb.stopCh:
			tb.cleanupTicker.Stop()
			return
		}
	}
}

// Close 关闭黑名单清理器
func (tb *TokenBlacklist) Close() {
	tb.closeOnce.Do(func() {
		close(tb.stopCh)
	})
}

// Size 返回黑名单中的 token 数量
func (tb *TokenBlacklist) Size() int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return len(tb.tokens)
}
