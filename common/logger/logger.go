package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"hahacking/config"
)

type Level int

var (
	F *os.File

	DefaultPrefix      = ""
	DefaultCallerDepth = 2

	logger     *log.Logger
	logPrefix  = ""
	levelFlags = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

	// 添加多输出支持
	multiWriter io.Writer

	// 全局日志级别
	currentLevel Level = DEBUG
	levelMutex   sync.RWMutex

	// 日志轮转器
	logRotator *LogRotator
)

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
	FATAL
)

// Setup initialize the log instance
func Setup() {
	filePath := getLogFilePath()
	fileName := getLogFileName()

	// 从配置文件读取日志级别
	level := getLogLevelFromConfig()
	SetLogLevel(level)

	// 创建多输出writer，同时输出到文件和控制台
	var writers []io.Writer
	writers = append(writers, os.Stdout)

	if config.ENV == "prd" {
		// 生产环境启用日志轮转
		setupLogRotation()
		if logRotator != nil && logRotator.file != nil {
			// 将日志轮转器的文件句柄加入到 writers
			writers = append(writers, logRotator)
		}
		setupGracefulShutdown()
	}

	multiWriter = io.MultiWriter(writers...)
	logger = log.New(multiWriter, DefaultPrefix, log.LstdFlags)

	// 输出初始化信息
	logger.Println("日志系统初始化完成，文件路径：", filePath+fileName, "日志级别：", levelFlags[level])

	// 如果启用了日志轮转，输出轮转配置信息并启动自动轮转管理器
	if config.ENV == "prd" && logRotator != nil {
		rotationConfig := getLogRotationConfig()
		logger.Printf("日志轮转配置：最大大小=%dMB，最大保留天数=%d天，最大备份数=%d个",
			rotationConfig.MaxSize, rotationConfig.MaxAge, rotationConfig.MaxBackups)

		// 启动自动轮转管理器
		if err := StartAutoRotation(); err != nil {
			logger.Printf("启动自动轮转失败: %v", err)
		} else {
			logger.Println("自动日志轮转已启动")
		}
	}
}
func setupLogRotation() {
	filePath := getLogFilePath()
	fileName := getLogFileName()

	// 获取日志轮转配置
	rotationConfig := getLogRotationConfig()

	// 创建日志轮转器
	logRotator = NewLogRotator(rotationConfig)
	logRotator.SetFilePath(filePath, fileName)

	// 创建初始日志文件
	if err := logRotator.createNewFile(); err != nil {
		log.Fatalf("logging.Setup err: %v", err)
	}

	// 获取文件句柄用于兼容性
	F = logRotator.file
	// 注意：此时 logger 还未创建，日志信息会在 Setup() 函数中输出
}

// GetWriter 获取日志writer，供其他组件使用
func GetWriter() io.Writer {
	return multiWriter
}

// GetFileWriter 获取文件writer
func GetFileWriter() io.Writer {
	return F
}

// GetLogRotationInfo 获取日志轮转信息
func GetLogRotationInfo() map[string]interface{} {
	if logRotator == nil {
		return nil
	}

	info := make(map[string]interface{})
	info["max_size"] = logRotator.config.MaxSize
	info["max_age"] = logRotator.config.MaxAge
	info["max_backups"] = logRotator.config.MaxBackups
	info["current_file_size"] = formatFileSize(logRotator.fileSize)
	info["file_path"] = logRotator.filePath
	info["file_name"] = logRotator.fileName

	return info
}

// SetupWithLevel 使用指定日志级别初始化日志系统
func SetupWithLevel(level Level) {
	SetLogLevel(level)
	Setup()
}

// SetLogLevel 设置日志级别
func SetLogLevel(level Level) {
	levelMutex.Lock()
	defer levelMutex.Unlock()
	currentLevel = level
}

// GetLogLevel 获取当前日志级别
func GetLogLevel() Level {
	levelMutex.RLock()
	defer levelMutex.RUnlock()
	return currentLevel
}

// IsLevelEnabled 检查指定级别是否启用
func IsLevelEnabled(level Level) bool {
	levelMutex.RLock()
	defer levelMutex.RUnlock()
	return level >= currentLevel
}

// Debug output logs at debug level
func Debug(v ...interface{}) {
	if !IsLevelEnabled(DEBUG) {
		return
	}
	setPrefix(DEBUG)
	logger.Println(v...)
}
func Debugf(format string, v ...interface{}) {
	if !IsLevelEnabled(DEBUG) {
		return
	}
	setPrefix(DEBUG)
	logger.Printf(format, v...)
}

// Info output logs at info level
func Info(v ...interface{}) {
	if !IsLevelEnabled(INFO) {
		return
	}
	setPrefix(INFO)
	logger.Println(v...)
}
func Infof(format string, v ...interface{}) {
	if !IsLevelEnabled(INFO) {
		return
	}
	setPrefix(INFO)
	logger.Printf(format, v...)
}

// Warn output logs at warn level
func Warn(v ...interface{}) {
	if !IsLevelEnabled(WARNING) {
		return
	}
	setPrefix(WARNING)
	logger.Println(v...)
}
func Warnf(format string, v ...interface{}) {
	if !IsLevelEnabled(WARNING) {
		return
	}
	setPrefix(WARNING)
	logger.Printf(format, v...)
}

// Error output logs at error level
func Error(v ...interface{}) {
	if !IsLevelEnabled(ERROR) {
		return
	}
	setPrefix(ERROR)
	logger.Println(v...)
}
func Errorf(format string, v ...interface{}) {
	if !IsLevelEnabled(ERROR) {
		return
	}
	setPrefix(ERROR)
	logger.Printf(format, v...)
}

// Fatal output logs at fatal level
func Fatal(v ...interface{}) {
	if !IsLevelEnabled(FATAL) {
		return
	}
	setPrefix(FATAL)
	logger.Println(v...)
}
func Fatalf(format string, v ...interface{}) {
	if !IsLevelEnabled(FATAL) {
		return
	}
	setPrefix(FATAL)
	logger.Printf(format, v...)
}

// setPrefix set the prefix of the log output
func setPrefix(level Level) {
	_, file, line, ok := runtime.Caller(DefaultCallerDepth)
	if ok {
		logPrefix = fmt.Sprintf("[%s][%s:%d]", levelFlags[level], filepath.Base(file), line)
	} else {
		logPrefix = fmt.Sprintf("[%s]", levelFlags[level])
	}

	logger.SetPrefix(logPrefix)
}

// getLogLevelFromConfig 从配置文件读取日志级别
func getLogLevelFromConfig() Level {
	logConfig := config.Get("log")
	if logConfig == nil {
		return DEBUG // 默认DEBUG级别
	}

	levelStr, ok := logConfig["level"].(string)
	if !ok {
		return DEBUG // 默认DEBUG级别
	}

	levelStr = strings.ToLower(levelStr)
	switch levelStr {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn", "warning":
		return WARNING
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	default:
		return DEBUG // 默认DEBUG级别
	}
}

// setupGracefulShutdown 设置优雅关闭
func setupGracefulShutdown() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		Info("收到关闭信号，正在优雅关闭日志系统...")
		StopAutoRotation()
		if logRotator != nil {
			logRotator.Close()
		}
		Info("日志系统已关闭")
		os.Exit(0)
	}()
}
