package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/model"
)

// ==================== InitWebSocketUpgrader 测试 ====================

// TestInitWebSocketUpgrader_EmptyConfig 测试空配置
func TestInitWebSocketUpgrader_EmptyConfig(t *testing.T) {
	// 保存原始 upgrader
	originalUpgrader := upgrader

	// 初始化空配置
	InitWebSocketUpgrader([]string{})

	// 验证 CheckOrigin 仍然拒绝所有
	result := upgrader.CheckOrigin(&http.Request{})
	if result != false {
		t.Error("Expected CheckOrigin to return false for empty config")
	}

	// 恢复原始 upgrader
	upgrader = originalUpgrader
}

// TestInitWebSocketUpgrader_Wildcard 测试通配符配置
func TestInitWebSocketUpgrader_Wildcard(t *testing.T) {
	// 保存原始 upgrader
	originalUpgrader := upgrader
	defer func() { upgrader = originalUpgrader }()

	// 初始化通配符配置
	InitWebSocketUpgrader([]string{"*"})

	// 验证允许所有源
	req := &http.Request{
		Header: http.Header{"Origin": []string{"https://evil.com"}},
	}
	result := upgrader.CheckOrigin(req)
	if result != true {
		t.Error("Expected CheckOrigin to return true for wildcard config")
	}

	// 验证空源也允许
	req2 := &http.Request{}
	result2 := upgrader.CheckOrigin(req2)
	if result2 != true {
		t.Error("Expected CheckOrigin to return true for empty origin with wildcard")
	}
}

// TestInitWebSocketUpgrader_SpecificOrigin 测试特定源配置
func TestInitWebSocketUpgrader_SpecificOrigin(t *testing.T) {
	// 保存原始 upgrader
	originalUpgrader := upgrader
	defer func() { upgrader = originalUpgrader }()

	// 初始化特定源配置
	InitWebSocketUpgrader([]string{"https://example.com"})

	// 验证允许的源
	req := &http.Request{
		Header: http.Header{"Origin": []string{"https://example.com"}},
	}
	result := upgrader.CheckOrigin(req)
	if result != true {
		t.Error("Expected CheckOrigin to return true for allowed origin")
	}

	// 验证不允许的源
	req2 := &http.Request{
		Header: http.Header{"Origin": []string{"https://evil.com"}},
	}
	result2 := upgrader.CheckOrigin(req2)
	if result2 != false {
		t.Error("Expected CheckOrigin to return false for disallowed origin")
	}
}

// TestInitWebSocketUpgrader_MultipleOrigins 测试多源配置
func TestInitWebSocketUpgrader_MultipleOrigins(t *testing.T) {
	// 保存原始 upgrader
	originalUpgrader := upgrader
	defer func() { upgrader = originalUpgrader }()

	// 初始化多源配置
	InitWebSocketUpgrader([]string{"https://example.com", "https://test.com"})

	// 验证第一个允许的源
	req1 := &http.Request{
		Header: http.Header{"Origin": []string{"https://example.com"}},
	}
	if !upgrader.CheckOrigin(req1) {
		t.Error("Expected CheckOrigin to return true for first allowed origin")
	}

	// 验证第二个允许的源
	req2 := &http.Request{
		Header: http.Header{"Origin": []string{"https://test.com"}},
	}
	if !upgrader.CheckOrigin(req2) {
		t.Error("Expected CheckOrigin to return true for second allowed origin")
	}

	// 验证不允许的源
	req3 := &http.Request{
		Header: http.Header{"Origin": []string{"https://evil.com"}},
	}
	if upgrader.CheckOrigin(req3) {
		t.Error("Expected CheckOrigin to return false for disallowed origin")
	}
}

// ==================== ParsePaginationParams 测试 ====================

// TestParsePaginationParams_Default 测试默认参数
func TestParsePaginationParams_Default(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		URL: mustParseURL("http://localhost/api/resources"),
	}

	params := ParsePaginationParams(c)

	if params.Limit != model.DefaultPageSize {
		t.Errorf("Expected default limit %d, got %d", model.DefaultPageSize, params.Limit)
	}
	if params.Offset != model.DefaultPageOffset {
		t.Errorf("Expected default offset %d, got %d", model.DefaultPageOffset, params.Offset)
	}
	if params.Search != "" {
		t.Errorf("Expected empty search, got %s", params.Search)
	}
	if params.SortBy != "name" {
		t.Errorf("Expected default sortBy 'name', got %s", params.SortBy)
	}
	if params.SortOrder != "asc" {
		t.Errorf("Expected default sortOrder 'asc', got %s", params.SortOrder)
	}
}

// TestParsePaginationParams_Custom 测试自定义参数
func TestParsePaginationParams_Custom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		URL: mustParseURL("http://localhost/api/resources?limit=50&offset=10&search=test&namespace=default&sortBy=age&sortOrder=desc"),
	}

	params := ParsePaginationParams(c)

	if params.Limit != 50 {
		t.Errorf("Expected limit 50, got %d", params.Limit)
	}
	if params.Offset != 10 {
		t.Errorf("Expected offset 10, got %d", params.Offset)
	}
	if params.Search != "test" {
		t.Errorf("Expected search 'test', got %s", params.Search)
	}
	if params.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got %s", params.Namespace)
	}
	if params.SortBy != "age" {
		t.Errorf("Expected sortBy 'age', got %s", params.SortBy)
	}
	if params.SortOrder != "desc" {
		t.Errorf("Expected sortOrder 'desc', got %s", params.SortOrder)
	}
}

// TestParsePaginationParams_MaxLimit 测试最大限制
func TestParsePaginationParams_MaxLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		URL: mustParseURL("http://localhost/api/resources?limit=1000"),
	}

	params := ParsePaginationParams(c)

	if params.Limit != model.MaxPageSize {
		t.Errorf("Expected max limit %d, got %d", model.MaxPageSize, params.Limit)
	}
}

// TestParsePaginationParams_InvalidParams 测试无效参数
func TestParsePaginationParams_InvalidParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		URL: mustParseURL("http://localhost/api/resources?limit=-1&offset=-5"),
	}

	params := ParsePaginationParams(c)

	// 负数 limit 应该使用默认值
	if params.Limit != model.DefaultPageSize {
		t.Errorf("Expected default limit for negative value, got %d", params.Limit)
	}
	// 负数 offset 应该使用默认值
	if params.Offset != model.DefaultPageOffset {
		t.Errorf("Expected default offset for negative value, got %d", params.Offset)
	}
}

// ==================== GetRequestContext 测试 ====================

// TestGetRequestContext 测试获取请求上下文
func TestGetRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// 创建一个有效的请求
	c.Request = httptest.NewRequest("GET", "http://localhost/test", nil)

	result := GetRequestContext(c)
	if result == nil {
		t.Error("Expected non-nil context")
	}
}

// TestGetRequestContext_NoContext 测试无上下文情况
func TestGetRequestContext_NoContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{}

	result := GetRequestContext(c)
	if result == nil {
		t.Error("Expected non-nil context (should fallback to background)")
	}
}

// ==================== GetTraceID 测试 ====================

// TestGetTraceID_FromHeader 测试从 header 获取 trace ID
func TestGetTraceID_FromHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "http://localhost/api/test", nil)
	c.Request.Header.Set("X-Trace-ID", "trace-123")

	result := GetTraceID(c)
	if result != "trace-123" {
		t.Errorf("Expected trace ID 'trace-123', got %s", result)
	}
}

// TestGetTraceID_FromContext 测试从上下文获取 trace ID
func TestGetTraceID_FromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{}
	c.Set("traceId", "trace-456")

	result := GetTraceID(c)
	if result != "trace-456" {
		t.Errorf("Expected trace ID 'trace-456', got %s", result)
	}
}

// TestGetTraceID_Empty 测试空 trace ID
func TestGetTraceID_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{}

	result := GetTraceID(c)
	if result != "" {
		t.Errorf("Expected empty trace ID, got %s", result)
	}
}

// ==================== Paginate 测试 ====================

// TestPaginate_Normal 测试正常分页
func TestPaginate_Normal(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}

	result := Paginate(items, 0, 2)
	if len(result) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result))
	}
	if result[0] != "a" || result[1] != "b" {
		t.Errorf("Expected [a, b], got %v", result)
	}

	result = Paginate(items, 2, 2)
	if len(result) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result))
	}
	if result[0] != "c" || result[1] != "d" {
		t.Errorf("Expected [c, d], got %v", result)
	}
}

// TestPaginate_OutOfBounds 测试越界分页
func TestPaginate_OutOfBounds(t *testing.T) {
	items := []string{"a", "b", "c"}

	// offset 超出范围
	result := Paginate(items, 10, 2)
	if len(result) != 0 {
		t.Errorf("Expected 0 items for out of bounds offset, got %d", len(result))
	}

	// 负数 offset
	result = Paginate(items, -1, 2)
	if len(result) != 0 {
		t.Errorf("Expected 0 items for negative offset, got %d", len(result))
	}

	// limit 为 0
	result = Paginate(items, 0, 0)
	if len(result) != 0 {
		t.Errorf("Expected 0 items for zero limit, got %d", len(result))
	}
}

// TestPaginate_Partial 测试部分分页
func TestPaginate_Partial(t *testing.T) {
	items := []string{"a", "b", "c"}

	// 请求超过剩余数量
	result := Paginate(items, 2, 10)
	if len(result) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result))
	}
	if result[0] != "c" {
		t.Errorf("Expected [c], got %v", result)
	}
}

// ==================== GenericSearchFilter 测试 ====================

// TestGenericSearchFilter_EmptySearch 测试空搜索
func TestGenericSearchFilter_EmptySearch(t *testing.T) {
	items := []model.Pod{
		{Name: "pod1", Namespace: "default"},
		{Name: "pod2", Namespace: "kube-system"},
	}

	result := GenericSearchFilter(items, "")
	if len(result) != 2 {
		t.Errorf("Expected 2 items for empty search, got %d", len(result))
	}
}

// TestGenericSearchFilter_Match 测试匹配搜索
func TestGenericSearchFilter_Match(t *testing.T) {
	items := []model.Pod{
		{Name: "pod1", Namespace: "default"},
		{Name: "pod2", Namespace: "kube-system"},
		{Name: "pod3", Namespace: "default"},
	}

	result := GenericSearchFilter(items, "pod1")
	if len(result) != 1 {
		t.Errorf("Expected 1 item for 'pod1' search, got %d", len(result))
	}

	result = GenericSearchFilter(items, "default")
	if len(result) != 2 {
		t.Errorf("Expected 2 items for 'default' search, got %d", len(result))
	}
}

// ==================== SortItems 测试 ====================

// TestSortItems_Asc 测试升序排序
func TestSortItems_Asc(t *testing.T) {
	items := []model.Pod{
		{Name: "pod3"},
		{Name: "pod1"},
		{Name: "pod2"},
	}

	result := SortItems(items, "name", "asc")
	if result[0].Name != "pod1" || result[1].Name != "pod2" || result[2].Name != "pod3" {
		t.Errorf("Expected sorted order, got %v", result)
	}
}

// TestSortItems_Desc 测试降序排序
func TestSortItems_Desc(t *testing.T) {
	items := []model.Pod{
		{Name: "pod1"},
		{Name: "pod3"},
		{Name: "pod2"},
	}

	result := SortItems(items, "name", "desc")
	if result[0].Name != "pod3" || result[1].Name != "pod2" || result[2].Name != "pod1" {
		t.Errorf("Expected reverse sorted order, got %v", result)
	}
}

// ==================== ExtractTokenFromRequest 测试 ====================

// TestExtractTokenFromRequest_WebSocketProtocol 测试从 WebSocket 协议头获取 token
func TestExtractTokenFromRequest_WebSocketProtocol(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "http://localhost/api", nil)
	c.Request.Header.Set("Sec-WebSocket-Protocol", "Bearer token123")

	result := ExtractTokenFromRequest(c)
	if result != "Bearer token123" {
		t.Errorf("Expected 'Bearer token123', got %s", result)
	}
}

func TestExtractTokenFromRequest_WebSocketProtocolAuthPair(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "http://localhost/api", nil)
	c.Request.Header.Set("Sec-WebSocket-Protocol", "k8svision.auth, token123")

	result := ExtractTokenFromRequest(c)
	if result != "token123" {
		t.Errorf("Expected 'token123', got %s", result)
	}
}

// TestExtractTokenFromRequest_Authorization 测试从 Authorization header 获取 token
func TestExtractTokenFromRequest_Authorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Header: http.Header{"Authorization": []string{"Bearer token456"}},
	}

	result := ExtractTokenFromRequest(c)
	if result != "token456" {
		t.Errorf("Expected 'token456', got %s", result)
	}
}

// TestExtractTokenFromRequest_QueryParam 测试从 URL 参数获取 token
func TestExtractTokenFromRequest_QueryParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		URL: mustParseURL("http://localhost/api?token=token789"),
	}

	result := ExtractTokenFromRequest(c)
	if result != "token789" {
		t.Errorf("Expected 'token789', got %s", result)
	}
}

// TestExtractTokenFromRequest_Empty 测试空 token
func TestExtractTokenFromRequest_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "http://localhost/api", nil)

	result := ExtractTokenFromRequest(c)
	if result != "" {
		t.Errorf("Expected empty token, got %s", result)
	}
}

func TestBuildWebSocketUpgradeHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "http://localhost/api", nil)
	c.Request.Header.Set("Sec-WebSocket-Protocol", "k8svision.auth, token123")

	headers := buildWebSocketUpgradeHeaders(c)
	if headers == nil {
		t.Fatal("Expected websocket upgrade headers to be set")
	}
	if got := headers.Get("Sec-WebSocket-Protocol"); got != "k8svision.auth" {
		t.Errorf("Expected negotiated protocol 'k8svision.auth', got %s", got)
	}
}

// ==================== parseAgeToSeconds 测试 ====================

// TestParseAgeToSeconds 测试年龄解析
func TestParseAgeToSeconds(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"seconds", "30s", 30},
		{"minutes", "5m", 300},
		{"hours", "2h", 7200},
		{"days", "1d", 86400},
		{"empty", "", 0},
		{"dash", "-", 0},
		{"invalid", "invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAgeToSeconds(tt.input)
			if result != tt.expected {
				t.Errorf("parseAgeToSeconds(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// ==================== config_helpers 测试 ====================

// TestSetConfigManager 测试设置配置管理器
func TestSetConfigManager(t *testing.T) {
	_, configMgr := createTestAuthComponents(t)

	// 保存原始配置管理器
	original := configManager
	defer func() { configManager = original }()

	SetConfigManager(configMgr)

	if configManager != configMgr {
		t.Error("Expected configManager to be set")
	}
}

// TestGetAuthConfig 测试获取认证配置
func TestGetAuthConfig(t *testing.T) {
	_, configMgr := createTestAuthComponents(t)

	// 保存原始配置管理器
	original := configManager
	defer func() { configManager = original }()

	SetConfigManager(configMgr)

	authConfig := GetAuthConfig()
	if authConfig == nil {
		t.Error("Expected auth config to be returned")
	}
}

// TestGetAuthConfig_NilManager 测试配置管理器为空时
func TestGetAuthConfig_NilManager(t *testing.T) {
	// 保存原始配置管理器
	original := configManager
	defer func() { configManager = original }()

	configManager = nil

	authConfig := GetAuthConfig()
	if authConfig != nil {
		t.Error("Expected nil auth config when manager is nil")
	}
}

// TestGetUsernameFromContext 测试从上下文获取用户名
func TestGetUsernameFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "admin")

	result := GetUsernameFromContext(c)
	if result != "admin" {
		t.Errorf("Expected username 'admin', got %s", result)
	}
}

// TestGetUsernameFromContext_NotSet 测试用户名未设置
func TestGetUsernameFromContext_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	result := GetUsernameFromContext(c)
	if result != "" {
		t.Errorf("Expected empty username, got %s", result)
	}
}

// TestGetUsernameFromContext_WrongType 测试用户名类型错误
func TestGetUsernameFromContext_WrongType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", 123) // 错误类型

	result := GetUsernameFromContext(c)
	if result != "" {
		t.Errorf("Expected empty username for wrong type, got %s", result)
	}
}

// 辅助函数
func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
