package logger

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AutoRotationManager 自动日志轮转管理器
type AutoRotationManager struct {
	config         LogRotationConfig
	filePath       string
	fileName       string
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	rotationTicker *time.Ticker
	cleanupTicker  *time.Ticker
}

// NewAutoRotationManager 创建自动轮转管理器
func NewAutoRotationManager() *AutoRotationManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &AutoRotationManager{
		config: getLogRotationConfig(),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动自动轮转管理器
func (arm *AutoRotationManager) Start() error {
	// 设置文件路径
	arm.filePath = getLogFilePath()
	arm.fileName = getLogFileName()

	// 启动轮转检查（每小时检查一次）
	arm.rotationTicker = time.NewTicker(1 * time.Hour)
	arm.wg.Add(1)
	go arm.rotationWorker()

	// 启动清理检查（每天凌晨2点执行）
	arm.cleanupTicker = time.NewTicker(24 * time.Hour)
	arm.wg.Add(1)
	go arm.cleanupWorker()

	Info("自动日志轮转管理器已启动")
	return nil
}

// Stop 停止自动轮转管理器
func (arm *AutoRotationManager) Stop() {
	if arm.cancel != nil {
		arm.cancel()
	}

	if arm.rotationTicker != nil {
		arm.rotationTicker.Stop()
	}

	if arm.cleanupTicker != nil {
		arm.cleanupTicker.Stop()
	}

	arm.wg.Wait()
	Info("自动日志轮转管理器已停止")
}

// rotationWorker 轮转工作协程
func (arm *AutoRotationManager) rotationWorker() {
	defer arm.wg.Done()

	for {
		select {
		case <-arm.ctx.Done():
			return
		case <-arm.rotationTicker.C:
			arm.checkAndRotate()
		}
	}
}

// cleanupWorker 清理工作协程
func (arm *AutoRotationManager) cleanupWorker() {
	defer arm.wg.Done()

	// 等待到下一个凌晨2点
	now := time.Now()
	nextCleanup := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	if nextCleanup.Before(now) {
		nextCleanup = nextCleanup.Add(24 * time.Hour)
	}

	// 等待到清理时间
	time.Sleep(time.Until(nextCleanup))

	for {
		select {
		case <-arm.ctx.Done():
			return
		case <-arm.cleanupTicker.C:
			arm.performCleanup()
		}
	}
}

// checkAndRotate 检查并执行轮转
func (arm *AutoRotationManager) checkAndRotate() {
	if logRotator == nil {
		return
	}

	// 检查是否需要轮转
	if logRotator.shouldRotate() {
		Info("检测到需要轮转日志文件")
		if err := logRotator.rotate(); err != nil {
			Error("日志轮转失败: " + err.Error())
		} else {
			Info("日志轮转完成")
		}
	}
}

// performCleanup 执行清理操作
func (arm *AutoRotationManager) performCleanup() {
	Info("开始执行自动日志清理")

	// 创建临时轮转器用于清理
	tempRotator := NewLogRotator(arm.config)
	tempRotator.SetFilePath(arm.filePath, arm.fileName)

	// 执行清理
	if err := tempRotator.cleanup(); err != nil {
		Error("自动日志清理失败: " + err.Error())
	} else {
		Info("自动日志清理完成")
	}
}

// GetRotationStatus 获取轮转状态信息
func (arm *AutoRotationManager) GetRotationStatus() map[string]interface{} {
	status := map[string]interface{}{
		"enabled":       arm.rotationTicker != nil,
		"config":        arm.config,
		"file_path":     arm.filePath,
		"file_name":     arm.fileName,
		"next_rotation": "每小时检查",
		"next_cleanup":  "每天凌晨2点",
	}

	if logRotator != nil {
		status["current_file_size"] = logRotator.fileSize
		status["last_rotate"] = logRotator.lastRotate
	}

	return status
}

// 全局自动轮转管理器
var autoRotationManager *AutoRotationManager

// StartAutoRotation 启动自动轮转
func StartAutoRotation() error {
	if autoRotationManager != nil {
		return fmt.Errorf("自动轮转管理器已经在运行")
	}

	autoRotationManager = NewAutoRotationManager()
	return autoRotationManager.Start()
}

// StopAutoRotation 停止自动轮转
func StopAutoRotation() {
	if autoRotationManager != nil {
		autoRotationManager.Stop()
		autoRotationManager = nil
	}
}

// GetAutoRotationStatus 获取自动轮转状态
func GetAutoRotationStatus() map[string]interface{} {
	if autoRotationManager == nil {
		return map[string]interface{}{
			"enabled": false,
			"message": "自动轮转未启动",
		}
	}

	return autoRotationManager.GetRotationStatus()
}
