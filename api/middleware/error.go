package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
)

// Recovery panic 恢复中间件
// 记录完整的错误堆栈，返回友好的错误信息给客户端
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		traceID := c.GetString("traceId")

		// 记录完整的错误堆栈
		logger.Error("panic recovered",
			zap.String("traceId", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Any("error", recovered),
			zap.String("stack", string(debug.Stack())),
		)

		// 返回友好的错误信息（不暴露 panic 详情）
		ResponseError(c, logger, &model.APIError{
			Code:    http.StatusInternalServerError,
			Message: "服务器内部错误，请稍后重试",
		}, http.StatusInternalServerError)
	})
}

// ResponseError 统一错误响应
// 参数：
//   - c: gin 上下文
//   - logger: 日志记录器
//   - err: 错误对象
//   - httpCode: HTTP 状态码
func ResponseError(c *gin.Context, logger *zap.Logger, err error, httpCode int) {
	traceID := c.GetString("traceId")

	var apiError *model.APIError

	// 类型断言处理不同类型的错误
	switch e := err.(type) {
	case *model.APIError:
		apiError = e
		if httpCode == 0 {
			httpCode = apiError.Code
		}

	case *errors.StatusError:
		// K8s API 错误
		httpCode = int(e.ErrStatus.Code)
		if httpCode == 0 {
			httpCode = http.StatusInternalServerError
		}
		apiError = &model.APIError{
			Code:    httpCode,
			Message: e.ErrStatus.Message,
			Details: e.ErrStatus.Reason,
		}

	default:
		// 其他错误
		if httpCode == 0 {
			httpCode = http.StatusInternalServerError
		}
		apiError = &model.APIError{
			Code:    httpCode,
			Message: http.StatusText(httpCode),
			Details: e.Error(),
		}
	}

	// 确保 httpCode 有效
	if httpCode <= 0 {
		httpCode = http.StatusInternalServerError
	}

	// 记录日志
	logFields := []zap.Field{
		zap.String("traceId", traceID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Int("httpCode", httpCode),
		zap.String("errorMessage", apiError.Message),
	}

	if httpCode >= 500 {
		logger.Error("server error", append(logFields, zap.Error(err))...)
	} else {
		logger.Warn("client error", logFields...)
	}

	// 返回错误响应
	c.JSON(httpCode, model.APIResponse{
		Code:      apiError.Code,
		Message:   apiError.Message,
		Data:      apiError.Details,
		TraceID:   traceID,
		Timestamp: time.Now().Unix(),
	})
}

// ResponseSuccess 统一成功响应
// 参数：
//   - c: gin 上下文
//   - data: 响应数据
//   - message: 成功消息
//   - page: 分页信息（可选）
func ResponseSuccess(c *gin.Context, data interface{}, message string, page *model.PageMeta) {
	c.JSON(http.StatusOK, model.APIResponse{
		Code:      model.CodeSuccess,
		Message:   message,
		Data:      data,
		TraceID:   c.GetString("traceId"),
		Timestamp: time.Now().Unix(),
		Page:      page,
	})
}
