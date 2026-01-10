package middleware

import (
	"bytes"
	"fmt"
	"hahacking/common/logger"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// CustomLogger 自定义Gin日志中间件，使用统一的日志系统
func CustomLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 构建日志消息
		var statusColor, methodColor, resetColor string
		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}

		// 使用统一的日志格式
		logMessage := fmt.Sprintf("[GIN] %v | %3d %s | %13v | %15s | %s %s %s %s",
			// param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			statusColor, param.StatusCode, resetColor,
			param.Latency,
			param.ClientIP,
			methodColor, param.Method, resetColor,
			param.Path,
		)

		// 如果有错误，添加到日志中
		if param.ErrorMessage != "" {
			logMessage += " | " + param.ErrorMessage
		}

		// 使用统一的日志系统记录
		logger.Info(logMessage)
		return ""
	})
}

// CustomLoggerWithBody 带请求体的日志中间件
func CustomLoggerWithBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 读取请求体
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			// 重新设置请求体，因为读取后需要重置
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 记录请求开始
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 记录请求结束
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 构建日志消息
		// logMessage := fmt.Sprintf("[GIN] %v | %3d | %13v | %15s | %s %s",
		// time.Now().Format("2006/01/02 - 15:04:05"),
		logMessage := fmt.Sprintf("[GIN] | %3d | %13v | %15s | %s %s",
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)

		if raw != "" {
			logMessage += "?" + raw
		}

		// 如果有请求体，添加到日志中（限制长度）
		if len(bodyBytes) > 0 && len(bodyBytes) < 1024 {
			logMessage += " | Body: " + string(bodyBytes)
		}

		// 如果有错误，添加到日志中
		if errorMessage != "" {
			logMessage += " | Error: " + errorMessage
		}

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			logger.Error(logMessage)
		} else if statusCode >= 400 {
			logger.Warn(logMessage)
		} else {
			logger.Info(logMessage)
		}
	}
}
