package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPasswordManager(t *testing.T) {
	pm := NewPasswordManager()
	assert.NotNil(t, pm)
	assert.Empty(t, pm.passwordHistory)
}

func TestHashPassword(t *testing.T) {
	pm := NewPasswordManager()

	t.Run("hash and verify", func(t *testing.T) {
		hashed, err := pm.HashPassword("mySecureP@ss1")
		assert.NoError(t, err)
		assert.NotEmpty(t, hashed)
		assert.True(t, pm.VerifyPassword("mySecureP@ss1", hashed))
		assert.False(t, pm.VerifyPassword("wrongPassword", hashed))
	})

	t.Run("different salts produce different hashes", func(t *testing.T) {
		h1, _ := pm.HashPassword("test123")
		h2, _ := pm.HashPassword("test123")
		assert.NotEqual(t, h1, h2)
	})
}

func TestCompare(t *testing.T) {
	pm := NewPasswordManager()
	hashed, _ := pm.HashPassword("pass123")
	assert.True(t, pm.Compare("pass123", hashed))
	assert.False(t, pm.Compare("wrong", hashed))
}

func TestGeneratePassword(t *testing.T) {
	pm := NewPasswordManager()

	t.Run("default length", func(t *testing.T) {
		pwd, err := pm.GeneratePassword(0)
		assert.NoError(t, err)
		assert.Equal(t, 12, len(pwd))
	})

	t.Run("custom length", func(t *testing.T) {
		pwd, err := pm.GeneratePassword(20)
		assert.NoError(t, err)
		assert.Equal(t, 20, len(pwd))
	})

	t.Run("max length cap", func(t *testing.T) {
		pwd, err := pm.GeneratePassword(200)
		assert.NoError(t, err)
		assert.Equal(t, 128, len(pwd))
	})
}

func TestValidatePasswordStrength(t *testing.T) {
	pm := NewPasswordManager()

	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"too short", "Ab1!", false},
		{"valid complex", "MyStr0ng#Xz", true},
		{"only lowercase", "abcdefgh", false},
		{"two char types", "ABCD1234", false},
		{"weak password", "password123", false},
		{"consecutive numbers", "abcd1234!", false},
		{"repeated characters", "aaaaaB1!", false},
		{"max length exceeded", string(make([]byte, 129)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := pm.ValidatePasswordStrength(tt.password)
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestPasswordHistory(t *testing.T) {
	pm := NewPasswordManager()

	t.Run("add and check", func(t *testing.T) {
		pwd := "N3wP@ssword!"
		hashed, _ := pm.HashPassword(pwd)
		pm.AddToPasswordHistory(hashed)
		assert.True(t, pm.IsPasswordInHistory(pwd, hashed))
	})

	t.Run("history size limit", func(t *testing.T) {
		pm2 := NewPasswordManager()
		for i := 0; i < 10; i++ {
			hashed, _ := pm2.HashPassword("TestP@ss" + string(rune('0'+i)))
			pm2.AddToPasswordHistory(hashed)
		}
		assert.LessOrEqual(t, len(pm2.passwordHistory), 5)
	})

	t.Run("not in history", func(t *testing.T) {
		pm2 := NewPasswordManager()
		hashed, _ := pm2.HashPassword("Un1que@Pwd")
		assert.False(t, pm2.IsPasswordInHistory("N3wP@ssword!", hashed))
	})
}

func TestHasConsecutiveNumbers(t *testing.T) {
	pm := NewPasswordManager()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"4 consecutive digits", "abc1234def", true},
		{"3 consecutive digits not enough", "abc123def", false},
		{"no consecutive", "a1b2c3d", false},
		{"less than 3", "ab12cd", false},
		{"no numbers", "abcdef", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, pm.hasConsecutiveNumbers(tt.input))
		})
	}
}

func TestHasRepeatedCharacters(t *testing.T) {
	pm := NewPasswordManager()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"repeated a", "aaaaB1!", true},
		{"no repeat", "Abcd1234!", false},
		{"single char returns true", "a", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, pm.hasRepeatedCharacters(tt.input))
		})
	}
}
