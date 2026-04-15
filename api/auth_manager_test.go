package api

import (
	"sync"
	"testing"

	"github.com/nick0323/K8sVision/config"
	"go.uber.org/zap"
)

// 辅助函数：创建测试 logger 和配置管理器
func createTestAuthComponents(t *testing.T) (*zap.Logger, *config.Manager) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	configMgr := config.NewManager(logger)
	// 设置认证配置
	configMgr.Set("auth.maxLoginFail", 5)
	configMgr.Set("auth.lockMinutes", 15)

	return logger, configMgr
}

// ==================== AuthManager 测试 ====================

// TestNewAuthManager 测试创建认证管理器
func TestNewAuthManager(t *testing.T) {
	logger, configMgr := createTestAuthComponents(t)
	manager := NewAuthManager(logger, configMgr)

	if manager == nil {
		t.Fatal("Expected AuthManager to be created")
	}
	if manager.logger == nil {
		t.Error("Expected logger to be set")
	}
	if manager.config == nil {
		t.Error("Expected config to be set")
	}

	// 验证所有分片已初始化
	for i, shard := range manager.shards {
		if shard == nil {
			t.Errorf("Expected shard %d to be initialized", i)
			continue
		}
		if shard.attempts == nil {
			t.Errorf("Expected shard %d attempts map to be initialized", i)
		}
	}

	manager.Close()
}

// TestAuthManager_IsLocked_NewUser 测试新用户未锁定
func TestAuthManager_IsLocked_NewUser(t *testing.T) {
	logger, configMgr := createTestAuthComponents(t)
	manager := NewAuthManager(logger, configMgr)
	defer manager.Close()

	locked := manager.IsLocked("newuser", "192.168.1.1")
	if locked {
		t.Error("Expected new user to not be locked")
	}
}

// TestAuthManager_RecordFailure 测试记录失败登录
func TestAuthManager_RecordFailure(t *testing.T) {
	logger, configMgr := createTestAuthComponents(t)
	manager := NewAuthManager(logger, configMgr)
	defer manager.Close()

	username := "testuser"
	ip := "192.168.1.1"

	// 记录 3 次失败
	manager.RecordFailure(username, ip)
	manager.RecordFailure(username, ip)
	manager.RecordFailure(username, ip)

	// 用户不应该被锁定（未达到阈值）
	locked := manager.IsLocked(username, ip)
	if locked {
		t.Error("Expected user to not be locked after 3 failures")
	}

	// 剩余尝试次数应该是 2
	remaining := manager.GetRemainingAttempts(username, ip)
	if remaining != 2 {
		t.Errorf("Expected 2 remaining attempts, got %d", remaining)
	}
}

// TestAuthManager_RecordFailure_LockUser 测试锁定用户
func TestAuthManager_RecordFailure_LockUser(t *testing.T) {
	logger, configMgr := createTestAuthComponents(t)
	manager := NewAuthManager(logger, configMgr)
	defer manager.Close()

	username := "lockuser"
	ip := "192.168.1.2"

	// 记录 5 次失败（达到阈值）
	for i := 0; i < 5; i++ {
		manager.RecordFailure(username, ip)
	}

	// 用户应该被锁定
	locked := manager.IsLocked(username, ip)
	if !locked {
		t.Error("Expected user to be locked after 5 failures")
	}

	// 剩余尝试次数应该是 0
	remaining := manager.GetRemainingAttempts(username, ip)
	if remaining != 0 {
		t.Errorf("Expected 0 remaining attempts, got %d", remaining)
	}

	// 锁定时间应该大于 0
	lockTime := manager.GetLockTime(username, ip)
	if lockTime <= 0 {
		t.Error("Expected lock time to be positive")
	}
}

// TestAuthManager_RecordSuccess 测试记录成功登录
func TestAuthManager_RecordSuccess(t *testing.T) {
	logger, configMgr := createTestAuthComponents(t)
	manager := NewAuthManager(logger, configMgr)
	defer manager.Close()

	username := "successuser"
	ip := "192.168.1.3"

	// 记录一些失败
	manager.RecordFailure(username, ip)
	manager.RecordFailure(username, ip)

	// 记录成功
	manager.RecordSuccess(username, ip)

	// 剩余尝试次数应该重置
	remaining := manager.GetRemainingAttempts(username, ip)
	if remaining != 5 {
		t.Errorf("Expected 5 remaining attempts after success, got %d", remaining)
	}

	// 不应该被锁定
	locked := manager.IsLocked(username, ip)
	if locked {
		t.Error("Expected user to not be locked after successful login")
	}
}

// TestAuthManager_GetStats 测试获取统计信息
func TestAuthManager_GetStats(t *testing.T) {
	logger, configMgr := createTestAuthComponents(t)
	manager := NewAuthManager(logger, configMgr)
	defer manager.Close()

	// 记录一些登录尝试
	manager.RecordFailure("user1", "192.168.1.1")
	manager.RecordFailure("user2", "192.168.1.2")
	manager.RecordSuccess("user3", "192.168.1.3")

	stats := manager.GetStats()

	if stats["totalAttempts"].(int) < 2 {
		t.Errorf("Expected at least 2 total attempts, got %d", stats["totalAttempts"])
	}
	if stats["maxFailCount"].(int) != 5 {
		t.Errorf("Expected maxFailCount to be 5, got %d", stats["maxFailCount"])
	}
}

// TestAuthManager_GetShard 测试分片分布
func TestAuthManager_GetShard(t *testing.T) {
	logger, configMgr := createTestAuthComponents(t)
	manager := NewAuthManager(logger, configMgr)
	defer manager.Close()

	// 相同用户名和 IP 应该映射到相同分片
	shard1 := manager.getShard("user1", "192.168.1.1")
	shard2 := manager.getShard("user1", "192.168.1.1")
	if shard1 != shard2 {
		t.Error("Expected same user+IP to map to same shard")
	}

	// 不同用户名应该可能映射到不同分片
	shard3 := manager.getShard("user2", "192.168.1.1")
	// 注意：可能相同也可能不同，这是哈希的特性

	// 不同 IP 应该可能映射到不同分片
	shard4 := manager.getShard("user1", "192.168.1.2")
	// 注意：可能相同也可能不同，这是哈希的特性

	_ = shard3
	_ = shard4
}

// TestAuthManager_ClearAttempt 测试清除尝试记录
func TestAuthManager_ClearAttempt(t *testing.T) {
	logger, configMgr := createTestAuthComponents(t)
	manager := NewAuthManager(logger, configMgr)
	defer manager.Close()

	username := "clearuser"
	ip := "192.168.1.4"

	// 记录一些失败
	manager.RecordFailure(username, ip)
	manager.RecordFailure(username, ip)

	// 清除
	manager.clearAttempt(username, ip)

	// 剩余尝试次数应该重置
	remaining := manager.GetRemainingAttempts(username, ip)
	if remaining != 5 {
		t.Errorf("Expected 5 remaining attempts after clear, got %d", remaining)
	}
}

// TestAuthManager_MakeKey 测试 key 生成
func TestAuthManager_MakeKey(t *testing.T) {
	logger, configMgr := createTestAuthComponents(t)
	manager := NewAuthManager(logger, configMgr)
	defer manager.Close()

	key1 := manager.makeKey("user1", "192.168.1.1")
	key2 := manager.makeKey("user1", "192.168.1.1")
	key3 := manager.makeKey("user2", "192.168.1.1")
	key4 := manager.makeKey("user1", "192.168.1.2")

	if key1 != key2 {
		t.Error("Expected same user+IP to generate same key")
	}
	if key1 == key3 {
		t.Error("Expected different users to generate different keys")
	}
	if key1 == key4 {
		t.Error("Expected different IPs to generate different keys")
	}
}

// TestAuthManager_ConcurrentAccess 测试并发访问安全性
func TestAuthManager_ConcurrentAccess(t *testing.T) {
	logger, configMgr := createTestAuthComponents(t)
	manager := NewAuthManager(logger, configMgr)
	defer manager.Close()

	var wg sync.WaitGroup

	// 启动多个 goroutine 并发访问
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			username := "concurrent-user"
			ip := "192.168.1.100"
			manager.RecordFailure(username, ip)
			manager.IsLocked(username, ip)
			manager.GetRemainingAttempts(username, ip)
		}(i)
	}

	wg.Wait()
	// 测试主要是确保不崩溃
}

// TestAuthManager_DifferentUsers 测试不同用户独立计数
func TestAuthManager_DifferentUsers(t *testing.T) {
	logger, configMgr := createTestAuthComponents(t)
	manager := NewAuthManager(logger, configMgr)
	defer manager.Close()

	// 用户 1 记录 3 次失败
	manager.RecordFailure("user1", "192.168.1.1")
	manager.RecordFailure("user1", "192.168.1.1")
	manager.RecordFailure("user1", "192.168.1.1")

	// 用户 2 记录 1 次失败
	manager.RecordFailure("user2", "192.168.1.1")

	// 验证独立计数
	remaining1 := manager.GetRemainingAttempts("user1", "192.168.1.1")
	remaining2 := manager.GetRemainingAttempts("user2", "192.168.1.1")

	if remaining1 != 2 {
		t.Errorf("Expected user1 to have 2 remaining attempts, got %d", remaining1)
	}
	if remaining2 != 4 {
		t.Errorf("Expected user2 to have 4 remaining attempts, got %d", remaining2)
	}
}

// TestAuthManager_DifferentIPs 测试不同 IP 独立计数
func TestAuthManager_DifferentIPs(t *testing.T) {
	logger, configMgr := createTestAuthComponents(t)
	manager := NewAuthManager(logger, configMgr)
	defer manager.Close()

	// 同一用户从不同 IP 登录
	manager.RecordFailure("user1", "192.168.1.1")
	manager.RecordFailure("user1", "192.168.1.1")
	manager.RecordFailure("user1", "192.168.1.2")

	// 验证独立计数
	remaining1 := manager.GetRemainingAttempts("user1", "192.168.1.1")
	remaining2 := manager.GetRemainingAttempts("user1", "192.168.1.2")

	if remaining1 != 3 {
		t.Errorf("Expected user1 from IP1 to have 3 remaining attempts, got %d", remaining1)
	}
	if remaining2 != 4 {
		t.Errorf("Expected user1 from IP2 to have 4 remaining attempts, got %d", remaining2)
	}
}

// TestAuthManager_LockExpiry 测试锁定过期
// 注：此测试主要验证锁定逻辑，不等待实际过期
func TestAuthManager_LockExpiry(t *testing.T) {
	logger, configMgr := createTestAuthComponents(t)

	manager := NewAuthManager(logger, configMgr)
	defer manager.Close()

	username := "expiryuser"
	ip := "192.168.1.5"

	// 记录 5 次失败（达到默认阈值）
	for i := 0; i < 5; i++ {
		manager.RecordFailure(username, ip)
	}

	// 用户应该被锁定
	locked := manager.IsLocked(username, ip)
	if !locked {
		t.Error("Expected user to be locked after 5 failures")
	}

	// 此测试主要验证锁定逻辑，不等待过期
	// 因为过期需要等待实际时间
}
