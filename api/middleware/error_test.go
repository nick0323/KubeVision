package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Set("traceId", "test-trace-id")
	return c, w
}

func TestResponseSuccess(t *testing.T) {
	c, w := setupTestContext()

	ResponseSuccess(c, map[string]string{"key": "value"}, "success", nil)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResponseSuccessWithPage(t *testing.T) {
	c, w := setupTestContext()

	page := &model.PageMeta{Total: 100, Limit: 20, Offset: 0}
	ResponseSuccess(c, []string{"a"}, "list", page)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResponseError_APIError(t *testing.T) {
	c, w := setupTestContext()
	logger := zap.NewNop()

	apiErr := &model.APIError{Code: http.StatusBadRequest, Message: "bad request"}
	ResponseError(c, logger, apiErr, http.StatusBadRequest)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestResponseError_APIErrorZeroHTTPCode(t *testing.T) {
	c, w := setupTestContext()
	logger := zap.NewNop()

	apiErr := &model.APIError{Code: http.StatusNotFound, Message: "not found"}
	ResponseError(c, logger, apiErr, 0)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestResponseError_GenericError(t *testing.T) {
	c, w := setupTestContext()
	logger := zap.NewNop()

	ResponseError(c, logger, errors.New("something went wrong"), http.StatusInternalServerError)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestResponseError_GenericErrorZeroHTTPCode(t *testing.T) {
	c, w := setupTestContext()
	logger := zap.NewNop()

	ResponseError(c, logger, errors.New("error"), 0)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestResponseError_K8sStatusError(t *testing.T) {
	c, w := setupTestContext()
	logger := zap.NewNop()

	statusErr := k8serrors.NewConflict(schema.GroupResource{Group: "", Resource: "pods"}, "my-pod", errors.New("conflict"))
	ResponseError(c, logger, statusErr, 0)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestResponseError_K8sStatusErrorWithHTTPCode(t *testing.T) {
	c, w := setupTestContext()
	logger := zap.NewNop()

	statusErr := k8serrors.NewNotFound(schema.GroupResource{Group: "", Resource: "pods"}, "my-pod")
	ResponseError(c, logger, statusErr, http.StatusNotFound)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestResponseError_NegativeHTTPCode(t *testing.T) {
	c, w := setupTestContext()
	logger := zap.NewNop()

	ResponseError(c, logger, errors.New("error"), -1)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	handler := Recovery(logger)
	assert.NotNil(t, handler)
}
