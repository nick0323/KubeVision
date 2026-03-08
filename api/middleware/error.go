package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
)

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		traceID := c.GetString("traceId")

		logger.Error("panic recovered",
			zap.String("traceId", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Any("error", recovered),
			zap.String("stack", string(debug.Stack())),
		)

		ResponseError(c, logger, &model.APIError{
			Code:    http.StatusInternalServerError,
			Message: "服务器内部错误",
			Details: fmt.Sprintf("panic: %v", recovered),
		}, http.StatusInternalServerError)
	})
}

func ResponseError(c *gin.Context, logger *zap.Logger, err error, httpCode int) {
	traceID := c.GetString("traceId")

	var apiError *model.APIError

	switch e := err.(type) {
	case *model.APIError:
		apiError = e
	case *errors.StatusError:
		apiError = &model.APIError{
			Code:    httpCode,
			Message: http.StatusText(httpCode),
			Details: e.Error(),
		}
	default:
		apiError = &model.APIError{
			Code:    httpCode,
			Message: http.StatusText(httpCode),
			Details: e.Error(),
		}
	}

	logFields := []zap.Field{
		zap.String("traceId", traceID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Int("httpCode", httpCode),
		zap.Int("errorCode", apiError.Code),
		zap.String("errorMessage", apiError.Message),
	}

	if httpCode >= 500 {
		logger.Error("server error", append(logFields, zap.Error(err))...)
	} else {
		logger.Warn("client error", logFields...)
	}

	c.JSON(httpCode, model.APIResponse{
		Code:      apiError.Code,
		Message:   apiError.Message,
		Data:      apiError.Details,
		TraceID:   traceID,
		Timestamp: time.Now().Unix(),
	})
}

func ResponseSuccess(c *gin.Context, data interface{}, message string, page *model.PageMeta) {
	c.JSON(http.StatusOK, model.APIResponse{
		Code:      model.CodeSuccess,  // 使用 0 表示成功，而不是 HTTP 状态码
		Message:   message,
		Data:      data,
		TraceID:   c.GetString("traceId"),
		Timestamp: time.Now().Unix(),
		Page:      page,
	})
}
