package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"hahacking/config"
)

// LogRotationConfig 日志轮转配置
type LogRotationConfig struct {
	MaxSize    int // 单个日志文件的最大大小（MB）
	MaxAge     int // 日志文件的最大保留天数
	MaxBackups int // 保留的日志文件备份数量
}

// LogRotator 日志轮转器
type LogRotator struct {
	config     LogRotationConfig
	file       *os.File
	filePath   string
	fileName   string
	fileSize   int64
	mutex      sync.Mutex
	lastRotate time.Time
}

// NewLogRotator 创建新的日志轮转器
func NewLogRotator(config LogRotationConfig) *LogRotator {
	return &LogRotator{
		config:     config,
		lastRotate: time.Now(),
	}
}

// Write 实现io.Writer接口，支持日志轮转
func (lr *LogRotator) Write(p []byte) (n int, err error) {
	lr.mutex.Lock()
	defer lr.mutex.Unlock()

	// 检查是否需要轮转
	if lr.shouldRotate() {
		if err := lr.rotate(); err != nil {
			return 0, err
		}
	}

	// 写入日志
	if lr.file != nil {
		n, err = lr.file.Write(p)
		if err == nil {
			lr.fileSize += int64(n)
		}
		return n, err
	}

	return 0, fmt.Errorf("log file not opened")
}

// shouldRotate 检查是否需要轮转
func (lr *LogRotator) shouldRotate() bool {
	if lr.file == nil {
		return false
	}

	// 检查文件大小
	if lr.config.MaxSize > 0 && lr.fileSize >= int64(lr.config.MaxSize*1024*1024) {
		return true
	}

	// 检查时间（每天轮转一次）
	now := time.Now()
	if now.Day() != lr.lastRotate.Day() || now.Month() != lr.lastRotate.Month() || now.Year() != lr.lastRotate.Year() {
		return true
	}

	return false
}

// rotate 执行日志轮转
func (lr *LogRotator) rotate() error {
	if lr.file != nil {
		lr.file.Close()
	}

	// 重命名当前日志文件
	if lr.file != nil {
		oldPath := filepath.Join(lr.filePath, lr.fileName)
		if _, err := os.Stat(oldPath); err == nil {
			// 生成带时间戳的备份文件名
			timestamp := time.Now().Format("20060102-150405")
			ext := filepath.Ext(lr.fileName)
			name := strings.TrimSuffix(lr.fileName, ext)
			backupName := fmt.Sprintf("%s.%s%s", name, timestamp, ext)
			backupPath := filepath.Join(lr.filePath, backupName)

			if err := os.Rename(oldPath, backupPath); err != nil {
				return fmt.Errorf("failed to rename log file: %v", err)
			}
		}
	}

	// 清理旧文件
	if err := lr.cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup old log files: %v", err)
	}

	// 创建新的日志文件
	if err := lr.createNewFile(); err != nil {
		return fmt.Errorf("failed to create new log file: %v", err)
	}

	lr.lastRotate = time.Now()
	return nil
}

// cleanup 清理旧的日志文件
func (lr *LogRotator) cleanup() error {
	if lr.config.MaxAge <= 0 && lr.config.MaxBackups <= 0 {
		return nil
	}

	// 获取所有日志文件
	files, err := lr.getLogFiles()
	if err != nil {
		return err
	}

	// 按修改时间排序（最新的在前）
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	// 删除超过最大保留天数的文件
	if lr.config.MaxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -lr.config.MaxAge)
		for i := len(files) - 1; i >= 0; i-- {
			file := files[i]
			if file.ModTime().Before(cutoff) {
				filePath := filepath.Join(lr.filePath, file.Name())
				if err := os.Remove(filePath); err != nil {
					// 如果文件不存在，忽略错误
					if !os.IsNotExist(err) {
						return fmt.Errorf("failed to remove old log file %s: %v", filePath, err)
					}
				}
				// 从切片中移除已删除的文件
				files = append(files[:i], files[i+1:]...)
			}
		}
	}

	// 删除超过最大备份数量的文件
	if lr.config.MaxBackups > 0 && len(files) > lr.config.MaxBackups {
		for i := lr.config.MaxBackups; i < len(files); i++ {
			filePath := filepath.Join(lr.filePath, files[i].Name())
			if err := os.Remove(filePath); err != nil {
				// 如果文件不存在，忽略错误
				if !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove excess log file %s: %v", filePath, err)
				}
			}
		}
	}

	return nil
}

// getLogFiles 获取所有日志文件
func (lr *LogRotator) getLogFiles() ([]os.FileInfo, error) {
	var files []os.FileInfo

	entries, err := os.ReadDir(lr.filePath)
	if err != nil {
		return nil, err
	}

	// 获取当前日志文件的基础名称（不含扩展名）
	baseName := strings.TrimSuffix(lr.fileName, filepath.Ext(lr.fileName))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 检查是否是日志文件（以基础名称开头）
		if strings.HasPrefix(entry.Name(), baseName) {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			files = append(files, info)
		}
	}

	return files, nil
}

// createNewFile 创建新的日志文件
func (lr *LogRotator) createNewFile() error {
	// 确保目录存在
	if err := os.MkdirAll(lr.filePath, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// 创建新文件
	filePath := filepath.Join(lr.filePath, lr.fileName)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}

	lr.file = file
	lr.fileSize = 0

	// 获取文件大小
	if stat, err := file.Stat(); err == nil {
		lr.fileSize = stat.Size()
	}

	return nil
}

// Close 关闭日志文件
func (lr *LogRotator) Close() error {
	lr.mutex.Lock()
	defer lr.mutex.Unlock()

	if lr.file != nil {
		return lr.file.Close()
	}
	return nil
}

// SetFilePath 设置文件路径
func (lr *LogRotator) SetFilePath(filePath, fileName string) {
	lr.mutex.Lock()
	defer lr.mutex.Unlock()

	lr.filePath = filePath
	lr.fileName = fileName
}

// getLogRotationConfig 从配置文件获取日志轮转配置
func getLogRotationConfig() LogRotationConfig {
	logConfig := config.Get("log")
	rotationConfig := LogRotationConfig{
		MaxSize:    100, // 默认100MB
		MaxAge:     7,   // 默认7天
		MaxBackups: 10,  // 默认10个备份
	}

	if logConfig == nil {
		return rotationConfig
	}

	// 读取maxsize
	if maxsize, ok := logConfig["maxsize"].(float64); ok {
		rotationConfig.MaxSize = int(maxsize)
	} else if maxsize, ok := logConfig["maxsize"].(int); ok {
		rotationConfig.MaxSize = maxsize
	}

	// 读取maxage
	if maxage, ok := logConfig["maxage"].(float64); ok {
		rotationConfig.MaxAge = int(maxage)
	} else if maxage, ok := logConfig["maxage"].(int); ok {
		rotationConfig.MaxAge = maxage
	}

	// 读取maxbackups
	if maxbackups, ok := logConfig["maxbackups"].(float64); ok {
		rotationConfig.MaxBackups = int(maxbackups)
	} else if maxbackups, ok := logConfig["maxbackups"].(int); ok {
		rotationConfig.MaxBackups = maxbackups
	}

	return rotationConfig
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return strconv.FormatInt(size, 10) + " B"
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
