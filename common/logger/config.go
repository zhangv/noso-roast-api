package logger

import (
	"hahacking/common/file"
	"io"
	"log"
	"os"
)

// LogConfig 日志配置结构
type LogConfig struct {
	// 输出目标配置
	OutputToFile    bool // 是否输出到文件
	OutputToConsole bool // 是否输出到控制台

	// 文件配置
	LogFilePath string // 日志文件路径
	LogFileName string // 日志文件名

	// 控制台配置
	EnableColors bool // 是否启用颜色

	// 日志级别配置
	MinLevel Level // 最小日志级别
}

// DefaultConfig 默认配置
func DefaultConfig() *LogConfig {
	return &LogConfig{
		OutputToFile:    true,
		OutputToConsole: true,
		LogFilePath:     "runtime/logs/",
		LogFileName:     "app.log",
		EnableColors:    true,
		MinLevel:        DEBUG,
	}
}

// SetupWithConfig 使用配置初始化日志系统
func SetupWithConfig(config *LogConfig) error {
	var writers []io.Writer

	// 设置日志级别
	SetLogLevel(config.MinLevel)

	// 添加控制台输出
	if config.OutputToConsole {
		writers = append(writers, os.Stdout)
	}

	// 添加文件输出
	if config.OutputToFile {
		filePath := config.LogFilePath
		fileName := config.LogFileName

		F, err := file.MustOpen(fileName, filePath)
		if err != nil {
			return err
		}

		writers = append(writers, F)
	}

	// 创建多输出writer
	if len(writers) == 0 {
		// 如果没有配置任何输出，默认输出到控制台
		writers = append(writers, os.Stdout)
	}

	multiWriter = io.MultiWriter(writers...)
	logger = log.New(multiWriter, DefaultPrefix, log.LstdFlags)

	return nil
}
