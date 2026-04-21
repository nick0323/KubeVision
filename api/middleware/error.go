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
			Message: "Internal server error, please try again later",
		}, http.StatusInternalServerError)
	})
}

func ResponseError(c *gin.Context, logger *zap.Logger, err error, httpCode int) {
	traceID := c.GetString("traceId")

	var apiError *model.APIError

	switch e := err.(type) {
	case *model.APIError:
		apiError = e
		if httpCode == 0 {
			httpCode = apiError.Code
		}
	case *errors.StatusError:
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
		if httpCode == 0 {
			httpCode = http.StatusInternalServerError
		}
		apiError = &model.APIError{
			Code:    httpCode,
			Message: http.StatusText(httpCode),
			Details: e.Error(),
		}
	}

	if httpCode <= 0 {
		httpCode = http.StatusInternalServerError
	}

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
		Code:      model.CodeSuccess,
		Message:   message,
		Data:      data,
		TraceID:   c.GetString("traceId"),
		Timestamp: time.Now().Unix(),
		Page:      page,
	})
}
