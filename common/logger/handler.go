package logger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
)

type LevelFileHandler struct {
	handlers map[slog.Level]slog.Handler
	mu       sync.Mutex
}

func NewLevelFileHandler() (*LevelFileHandler, error) {
	h := &LevelFileHandler{
		handlers: make(map[slog.Level]slog.Handler),
	}

	// 初始化不同级别的日志文件
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	// 配置 INFO 级别日志（10MB，保留7天）
	infoLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "info.log"),
		MaxSize:    10,   // MB
		MaxAge:     7,    // 天
		MaxBackups: 3,    // 备份文件数
		Compress:   true, // 压缩旧日志
	}
	// h.handlers[slog.LevelInfo] = slog.NewTextHandler(infoLogger, &slog.HandlerOptions{
	// 	Level: slog.LevelInfo,
	// })
	h.handlers[slog.LevelInfo] = slog.NewJSONHandler(infoLogger, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	// 配置 ERROR 级别日志（10MB，保留7天）
	errorLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "error.log"),
		MaxSize:    10,   // MB
		MaxAge:     7,    // 天
		MaxBackups: 3,    // 备份文件数
		Compress:   true, // 压缩旧日志
	}
	h.handlers[slog.LevelError] = slog.NewTextHandler(errorLogger, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	// h.handlers[slog.LevelError] = slog.NewJSONHandler(errorLogger, &slog.HandlerOptions{
	// 	Level: slog.LevelError,
	// })
	return h, nil
}

func (h *LevelFileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for l := range h.handlers {
		if level >= l {
			return true
		}
	}
	return false
}

func (h *LevelFileHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	for level, handler := range h.handlers {
		if r.Level >= level {
			return handler.Handle(ctx, r)
		}
	}
	return nil
}

func (h *LevelFileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	newHandlers := make(map[slog.Level]slog.Handler, len(h.handlers))
	for level, handler := range h.handlers {
		newHandlers[level] = handler.WithAttrs(attrs)
	}
	return &LevelFileHandler{handlers: newHandlers, mu: h.mu}
}

func (h *LevelFileHandler) WithGroup(name string) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	newHandlers := make(map[slog.Level]slog.Handler, len(h.handlers))
	for level, handler := range h.handlers {
		newHandlers[level] = handler.WithGroup(name)
	}
	return &LevelFileHandler{handlers: newHandlers, mu: h.mu}
}

func main0() {
	// 初始化日志管理器
	handler, err := NewLevelFileHandler()
	if err != nil {
		panic(err)
	}
	logger := slog.New(handler)

	// 示例使用
	logger.Info("This is an info message", "user", "alice", "time", time.Now())
	logger.Error("This is an error message", "error", "timeout", "time", time.Now())

	// 模拟业务逻辑
	for i := 0; i < 5; i++ {
		if i%2 == 0 {
			logger.Info("Processing item", "item", i)
		} else {
			logger.Error("Failed to process item", "item", i, "reason", "network error")
		}
		time.Sleep(time.Second)
	}
}
