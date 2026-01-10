package logger

import (
	"go.uber.org/zap"
)

var (
	Zaplogger *zap.Logger
	Sugar     *zap.SugaredLogger
)

// Setup initialize the log instance
func InitZap() {
	Zaplogger, err := zap.NewDevelopment()
	if err != nil {
		panic("Failed to initialize zap logger: " + err.Error())
	}
	Sugar = Zaplogger.Sugar()
	defer Zaplogger.Sync() // 确保日志在程序退出前写入
}
