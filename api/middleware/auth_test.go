package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestExtractTokenFromWebSocket(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid protocol", "k8svision.auth, abc123", "abc123"},
		{"valid protocol extra spaces", "k8svision.auth,  abc123  ", "abc123"},
		{"empty header", "", ""},
		{"single token", "abc123", "abc123"},
		{"invalid protocol format", "other.protocol, abc123", "other.protocol, abc123"},
		{"only protocol no token", "k8svision.auth, ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTokenFromWebSocket(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTokenFromRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		authHeader    string
		wsProtocol    string
		queryToken    string
		expectedToken string
	}{
		{"authorization header", "Bearer abc123", "", "", "Bearer abc123"},
		{"websocket protocol", "", "k8svision.auth, xyz789", "", "xyz789"},
		{"query parameter", "", "", "querytoken123", "querytoken123"},
		{"priority: auth header first", "Bearer abc123", "k8svision.auth, xyz789", "querytoken123", "Bearer abc123"},
		{"priority: ws protocol second", "", "k8svision.auth, xyz789", "querytoken123", "xyz789"},
		{"no token", "", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)
			c.Request = &http.Request{}
			c.Request.Header = make(map[string][]string)

			if tt.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.authHeader)
			}
			if tt.wsProtocol != "" {
				c.Request.Header.Set("Sec-WebSocket-Protocol", tt.wsProtocol)
			}
			if tt.queryToken != "" {
				c.Request.URL = &url.URL{Path: "/", RawQuery: "token=" + tt.queryToken}
			} else {
				c.Request.URL = &url.URL{Path: "/"}
			}

			token := getTokenFromRequest(c)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

// ---- mock ConfigProvider ----
type mockConfigProvider struct {
	secret []byte
}

func (m *mockConfigProvider) GetJWTSecret() []byte {
	return m.secret
}

// ---- helpers ----
func signToken(claims jwt.MapClaims, secret []byte) string {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := t.SignedString(secret)
	return s
}

func defaultClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"username": "admin",
		"iss":      JWTIssuer,
		"aud":      JWTAudience,
		"jti":      "test-jti",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}
}

// ---- NewJWTMiddleware / GetBlacklist ----
func TestNewJWTMiddleware(t *testing.T) {
	m := NewJWTMiddleware([]byte("secret"), zap.NewNop())
	require.NotNil(t, m)
	assert.NotNil(t, m.blacklist)
	assert.Equal(t, []byte("secret"), m.secret)
}

func TestGetBlacklist(t *testing.T) {
	m := NewJWTMiddleware([]byte("s"), zap.NewNop())
	bl := m.GetBlacklist()
	assert.NotNil(t, bl)
	assert.Equal(t, 0, bl.Size())
}

// ---- validateClaims ----
func TestValidateClaims_Success(t *testing.T) {
	err := validateClaims(defaultClaims())
	assert.NoError(t, err)
}

func TestValidateClaims_MissingUsername(t *testing.T) {
	c := jwt.MapClaims{"iss": JWTIssuer, "aud": JWTAudience}
	err := validateClaims(c)
	assert.EqualError(t, err, "token missing username")
}

func TestValidateClaims_EmptyUsername(t *testing.T) {
	c := jwt.MapClaims{"username": "", "iss": JWTIssuer, "aud": JWTAudience}
	err := validateClaims(c)
	assert.EqualError(t, err, "token missing username")
}

func TestValidateClaims_MissingIssuer(t *testing.T) {
	c := jwt.MapClaims{"username": "admin", "aud": JWTAudience}
	err := validateClaims(c)
	assert.EqualError(t, err, "token missing issuer")
}

func TestValidateClaims_InvalidIssuer(t *testing.T) {
	c := jwt.MapClaims{"username": "admin", "iss": "hacker", "aud": JWTAudience}
	err := validateClaims(c)
	assert.EqualError(t, err, "invalid token issuer")
}

func TestValidateClaims_MissingAudience(t *testing.T) {
	c := jwt.MapClaims{"username": "admin", "iss": JWTIssuer}
	err := validateClaims(c)
	assert.EqualError(t, err, "token missing audience")
}

func TestValidateClaims_InvalidAudience(t *testing.T) {
	c := jwt.MapClaims{"username": "admin", "iss": JWTIssuer, "aud": "hacker"}
	err := validateClaims(c)
	assert.EqualError(t, err, "invalid token audience")
}

// ---- getJWTSecret ----
func TestGetJWTSecret_NilProvider(t *testing.T) {
	_, err := getJWTSecret(nil)
	assert.EqualError(t, err, "config provider not initialized")
}

func TestGetJWTSecret_EmptySecret(t *testing.T) {
	p := &mockConfigProvider{secret: []byte{}}
	_, err := getJWTSecret(p)
	assert.EqualError(t, err, "JWT secret not configured")
}

func TestGetJWTSecret_Success(t *testing.T) {
	p := &mockConfigProvider{secret: []byte("mysecret")}
	s, err := getJWTSecret(p)
	assert.NoError(t, err)
	assert.Equal(t, []byte("mysecret"), s)
}

// ---- verifyToken ----
func TestVerifyToken_Success(t *testing.T) {
	p := &mockConfigProvider{secret: []byte("supersecret")}
	token := signToken(defaultClaims(), p.secret)

	claims, err := verifyToken(token, p)
	assert.NoError(t, err)
	assert.Equal(t, "admin", claims["username"])
	assert.Equal(t, "test-jti", claims["jti"])
}

func TestVerifyToken_InvalidSignature(t *testing.T) {
	p := &mockConfigProvider{secret: []byte("supersecret")}
	wrongSecret := []byte("wrongsecret")
	token := signToken(defaultClaims(), wrongSecret)

	_, err := verifyToken(token, p)
	assert.Error(t, err)
}

func TestVerifyToken_InvalidAlgorithm(t *testing.T) {
	p := &mockConfigProvider{secret: []byte("supersecret")}
	tk := jwt.NewWithClaims(jwt.SigningMethodNone, defaultClaims())
	token, _ := tk.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := verifyToken(token, p)
	assert.Error(t, err)
}

func TestVerifyToken_MalformedToken(t *testing.T) {
	p := &mockConfigProvider{secret: []byte("supersecret")}
	_, err := verifyToken("not.a.token", p)
	assert.Error(t, err)
}

func TestVerifyToken_EmptyToken(t *testing.T) {
	p := &mockConfigProvider{secret: []byte("supersecret")}
	_, err := verifyToken("", p)
	assert.Error(t, err)
}

func TestVerifyToken_ClaimsValidationFails(t *testing.T) {
	p := &mockConfigProvider{secret: []byte("supersecret")}
	claims := jwt.MapClaims{"username": "admin", "iss": "wrong"}
	token := signToken(claims, p.secret)

	_, err := verifyToken(token, p)
	assert.EqualError(t, err, "invalid token issuer")
}

// ---- verifyAndSetClaims ----
func TestVerifyAndSetClaims_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	p := &mockConfigProvider{secret: []byte("secret")}
	token := signToken(defaultClaims(), p.secret)
	mw := NewJWTMiddleware([]byte("secret"), zap.NewNop())

	username, jti, err := mw.verifyAndSetClaims(c, token, p)
	assert.NoError(t, err)
	assert.Equal(t, "admin", username)
	assert.Equal(t, "test-jti", jti)
	assert.Equal(t, "admin", c.GetString("username"))
	assert.Equal(t, "test-jti", c.GetString("jti"))
}

func TestVerifyAndSetClaims_NoJTI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	p := &mockConfigProvider{secret: []byte("secret")}
	claims := jwt.MapClaims{"username": "admin", "iss": JWTIssuer, "aud": JWTAudience}
	token := signToken(claims, p.secret)
	mw := NewJWTMiddleware([]byte("secret"), zap.NewNop())

	username, jti, err := mw.verifyAndSetClaims(c, token, p)
	assert.NoError(t, err)
	assert.Equal(t, "admin", username)
	assert.Equal(t, "", jti)
	assert.Equal(t, "admin", c.GetString("username"))
	_, exists := c.Get("jti")
	assert.False(t, exists)
}

// ---- AuthMiddleware ----
func TestAuthMiddleware_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	m := NewJWTMiddleware([]byte("secret"), zap.NewNop())
	handler := m.AuthMiddleware(&mockConfigProvider{secret: []byte("secret")})
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
}

func TestAuthMiddleware_EmptyBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer ")

	m := NewJWTMiddleware([]byte("secret"), zap.NewNop())
	handler := m.AuthMiddleware(&mockConfigProvider{secret: []byte("secret")})
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Token cannot be empty")
	assert.True(t, c.IsAborted())
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer invalidtoken")

	m := NewJWTMiddleware([]byte("secret"), zap.NewNop())
	handler := m.AuthMiddleware(&mockConfigProvider{secret: []byte("secret")})
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
}

func TestAuthMiddleware_BlacklistedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	p := &mockConfigProvider{secret: []byte("secret")}
	claims := defaultClaims()
	claims["jti"] = "revoked-jti"
	token := signToken(claims, p.secret)

	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	m := NewJWTMiddleware([]byte("secret"), zap.NewNop())
	m.blacklist.Add("revoked-jti", time.Now().Add(time.Hour))

	handler := m.AuthMiddleware(p)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "revoked")
	assert.True(t, c.IsAborted())
}

func TestAuthMiddleware_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	p := &mockConfigProvider{secret: []byte("secret")}
	token := signToken(defaultClaims(), p.secret)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	var called bool
	m := NewJWTMiddleware([]byte("secret"), zap.NewNop())
	handler := m.AuthMiddleware(p)
	// wrap so we can verify c.Next() was called
	next := func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	}
	ginHandler := func(c *gin.Context) {
		handler(c)
		if !c.IsAborted() {
			next(c)
		}
	}
	ginHandler(c)

	assert.True(t, called, "next handler should be called")
	assert.Equal(t, "admin", c.GetString("username"))
}

// ---- VerifyToken (exported) ----
func TestVerifyToken_Exported_Success(t *testing.T) {
	claims := defaultClaims()
	token := signToken(claims, []byte("secret"))

	result, err := VerifyToken(token, []byte("secret"))
	assert.NoError(t, err)
	assert.Equal(t, "admin", result["username"])
}

func TestVerifyToken_Exported_EmptySecret(t *testing.T) {
	_, err := VerifyToken("some.token", []byte{})
	assert.EqualError(t, err, "JWT secret not configured")
}

func TestVerifyToken_Exported_Malformed(t *testing.T) {
	_, err := VerifyToken("bad", []byte("secret"))
	assert.Error(t, err)
}
