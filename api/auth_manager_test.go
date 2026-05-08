package api

import (
	"testing"
	"time"

	"github.com/nick0323/K8sVision/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newTestAuthManager() *AuthManager {
	logger, _ := zap.NewDevelopment()
	mgr := config.NewManager(logger)
	mgr.GetConfig().Auth.MaxLoginFail = 3
	mgr.GetConfig().Auth.LockDuration = 1 * time.Hour
	return NewAuthManager(logger, mgr)
}

func TestAuthManagerIsLocked(t *testing.T) {
	am := newTestAuthManager()

	t.Run("no attempts returns false", func(t *testing.T) {
		assert.False(t, am.IsLocked("user1", "10.0.0.1"))
	})

	t.Run("after max failures returns true", func(t *testing.T) {
		am.RecordFailure("user1", "10.0.0.1")
		am.RecordFailure("user1", "10.0.0.1")
		am.RecordFailure("user1", "10.0.0.1")
		assert.True(t, am.IsLocked("user1", "10.0.0.1"))
	})

	t.Run("different IP not locked", func(t *testing.T) {
		assert.False(t, am.IsLocked("user1", "10.0.0.2"))
	})

	t.Run("different user not locked", func(t *testing.T) {
		assert.False(t, am.IsLocked("user2", "10.0.0.1"))
	})
}

func TestAuthManagerRecordSuccess(t *testing.T) {
	am := newTestAuthManager()

	am.RecordFailure("user2", "10.0.0.1")
	am.RecordSuccess("user2", "10.0.0.1")
	assert.False(t, am.IsLocked("user2", "10.0.0.1"))
	assert.Equal(t, 3, am.GetRemainingAttempts("user2", "10.0.0.1"))
}

func TestAuthManagerGetRemainingAttempts(t *testing.T) {
	am := newTestAuthManager()

	t.Run("no attempts returns max", func(t *testing.T) {
		assert.Equal(t, 3, am.GetRemainingAttempts("newuser", "1.1.1.1"))
	})

	t.Run("after one failure returns 2", func(t *testing.T) {
		am.RecordFailure("user3", "10.0.0.1")
		assert.Equal(t, 2, am.GetRemainingAttempts("user3", "10.0.0.1"))
	})

	t.Run("after all failures returns 0", func(t *testing.T) {
		am.RecordFailure("user4", "10.0.0.1")
		am.RecordFailure("user4", "10.0.0.1")
		am.RecordFailure("user4", "10.0.0.1")
		assert.Equal(t, 0, am.GetRemainingAttempts("user4", "10.0.0.1"))
	})

	t.Run("different user unaffected", func(t *testing.T) {
		assert.Equal(t, 3, am.GetRemainingAttempts("other-user", "10.0.0.1"))
	})
}

func TestAuthManagerGetLockTime(t *testing.T) {
	am := newTestAuthManager()

	t.Run("no attempts returns 0", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), am.GetLockTime("nolock", "1.1.1.1"))
	})

	t.Run("locked returns positive", func(t *testing.T) {
		am.RecordFailure("user5", "10.0.0.1")
		am.RecordFailure("user5", "10.0.0.1")
		am.RecordFailure("user5", "10.0.0.1")
		lockTime := am.GetLockTime("user5", "10.0.0.1")
		assert.Greater(t, lockTime, time.Duration(0))
	})

	t.Run("not locked returns 0", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), am.GetLockTime("user5", "10.0.0.2"))
	})
}

func TestAuthManagerGetStats(t *testing.T) {
	am := newTestAuthManager()
	am.RecordFailure("user-a", "10.0.0.1")
	am.RecordFailure("user-b", "10.0.0.2")

	stats := am.GetStats()
	assert.Equal(t, 2, stats["totalAttempts"])
	assert.Equal(t, 3, stats["maxFailCount"])
	assert.Equal(t, "1h0m0s", stats["lockDuration"])
}

func TestAuthManagerClose(t *testing.T) {
	am := newTestAuthManager()
	am.Close()
}

func TestAuthManagerShardDistribution(t *testing.T) {
	am := newTestAuthManager()
	shard1 := am.getShard("user-a", "10.0.0.1")
	shard2 := am.getShard("user-b", "10.0.0.2")
	assert.NotNil(t, shard1)
	assert.NotNil(t, shard2)
}
