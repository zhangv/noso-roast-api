package logger

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"hahacking/config"
)

var setupOnce sync.Once

func setup() {
	setupOnce.Do(func() {
		config.Setup("../../", "test")
	})
}

func TestLogRotationConfig(t *testing.T) {
	setup()
	// 测试日志轮转配置读取
	config := getLogRotationConfig()

	// 验证配置值在合理范围内
	if config.MaxSize <= 0 {
		t.Error("MaxSize应该大于0")
	}
	if config.MaxAge <= 0 {
		t.Error("MaxAge应该大于0")
	}
	if config.MaxBackups <= 0 {
		t.Error("MaxBackups应该大于0")
	}

	t.Logf("日志轮转配置：MaxSize=%dMB, MaxAge=%d天, MaxBackups=%d个",
		config.MaxSize, config.MaxAge, config.MaxBackups)
}

func TestLogRotatorCreation(t *testing.T) {
	// 测试日志轮转器创建
	config := LogRotationConfig{
		MaxSize:    10, // 10MB
		MaxAge:     7,  // 7天
		MaxBackups: 5,  // 5个备份
	}

	rotator := NewLogRotator(config)
	if rotator == nil {
		t.Error("日志轮转器创建失败")
	}

	if rotator.config.MaxSize != 10 {
		t.Error("MaxSize配置不正确")
	}
	if rotator.config.MaxAge != 7 {
		t.Error("MaxAge配置不正确")
	}
	if rotator.config.MaxBackups != 5 {
		t.Error("MaxBackups配置不正确")
	}
}

func TestLogRotatorFileCreation(t *testing.T) {
	// 测试日志文件创建
	config := LogRotationConfig{
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 5,
	}

	rotator := NewLogRotator(config)

	// 设置测试路径
	testPath := "test_logs"
	testFile := "test.log"
	rotator.SetFilePath(testPath, testFile)

	// 创建日志文件
	err := rotator.createNewFile()
	if err != nil {
		t.Errorf("创建日志文件失败: %v", err)
	}

	// 验证文件是否存在
	filePath := filepath.Join(testPath, testFile)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("日志文件未创建")
	}

	// 清理测试文件
	rotator.Close()
	os.RemoveAll(testPath)
}

func TestLogRotatorWrite(t *testing.T) {
	// 测试日志写入
	config := LogRotationConfig{
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 5,
	}

	rotator := NewLogRotator(config)
	testPath := "test_logs"
	testFile := "test.log"
	rotator.SetFilePath(testPath, testFile)

	// 创建日志文件
	err := rotator.createNewFile()
	if err != nil {
		t.Fatalf("创建日志文件失败: %v", err)
	}
	defer func() {
		rotator.Close()
		os.RemoveAll(testPath)
	}()

	// 写入测试日志
	testData := []byte("这是一条测试日志\n")
	n, err := rotator.Write(testData)
	if err != nil {
		t.Errorf("写入日志失败: %v", err)
	}
	if n != len(testData) {
		t.Errorf("写入字节数不匹配: 期望%d, 实际%d", len(testData), n)
	}

	// 验证文件大小
	if rotator.fileSize != int64(len(testData)) {
		t.Errorf("文件大小不匹配: 期望%d, 实际%d", len(testData), rotator.fileSize)
	}
}

func TestLogRotatorShouldRotate(t *testing.T) {
	// 测试轮转条件检查
	config := LogRotationConfig{
		MaxSize:    1, // 1MB
		MaxAge:     7,
		MaxBackups: 5,
	}

	rotator := NewLogRotator(config)
	testPath := "test_logs"
	testFile := "test.log"
	rotator.SetFilePath(testPath, testFile)

	// 创建日志文件
	err := rotator.createNewFile()
	if err != nil {
		t.Fatalf("创建日志文件失败: %v", err)
	}
	defer func() {
		rotator.Close()
		os.RemoveAll(testPath)
	}()

	// 初始状态不应该轮转
	if rotator.shouldRotate() {
		t.Error("初始状态不应该需要轮转")
	}

	// 模拟文件大小超过限制
	rotator.fileSize = 2 * 1024 * 1024 // 2MB
	if !rotator.shouldRotate() {
		t.Error("文件大小超过限制时应该需要轮转")
	}

	// 重置文件大小，模拟时间变化
	rotator.fileSize = 0
	rotator.lastRotate = time.Now().AddDate(0, 0, -1) // 昨天
	if !rotator.shouldRotate() {
		t.Error("时间变化时应该需要轮转")
	}
}

func TestFormatFileSize(t *testing.T) {
	// 测试文件大小格式化
	testCases := []struct {
		size     int64
		expected string
	}{
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1500, "1.5 KB"},
		{1500 * 1024, "1.5 MB"},
	}

	for _, tc := range testCases {
		result := formatFileSize(tc.size)
		if result != tc.expected {
			t.Errorf("文件大小格式化错误: 输入%d, 期望%s, 实际%s", tc.size, tc.expected, result)
		}
	}
}

func TestLogRotatorCleanup(t *testing.T) {
	// 测试日志文件清理 - 基于MaxAge
	config := LogRotationConfig{
		MaxSize:    10,
		MaxAge:     1, // 1天
		MaxBackups: 2, // 2个备份
	}

	rotator := NewLogRotator(config)
	testPath := "test_logs"
	testFile := "test.log"
	rotator.SetFilePath(testPath, testFile)

	// 创建测试目录
	os.MkdirAll(testPath, 0755)
	defer os.RemoveAll(testPath)

	// 创建一些测试文件
	testFiles := []string{
		"test.20250101-120000.log",
		"test.20250102-120000.log",
		"test.20250103-120000.log",
		"test.20250104-120000.log",
	}

	for _, fileName := range testFiles {
		filePath := filepath.Join(testPath, fileName)
		file, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}
		file.Close()

		// 设置文件修改时间为过去
		pastTime := time.Now().AddDate(0, 0, -2)
		os.Chtimes(filePath, pastTime, pastTime)
	}

	// 等待一下确保文件时间戳生效
	time.Sleep(100 * time.Millisecond)

	// 执行清理
	err := rotator.cleanup()
	if err != nil {
		t.Errorf("清理失败: %v", err)
	}

	// 验证清理结果
	entries, err := os.ReadDir(testPath)
	if err != nil {
		t.Fatalf("读取目录失败: %v", err)
	}

	// 由于MaxAge=1天，所有文件都是2天前的，应该都被删除
	expectedCount := 0
	if len(entries) != expectedCount {
		t.Errorf("清理后文件数量不正确: 期望%d, 实际%d", expectedCount, len(entries))
		for _, entry := range entries {
			t.Logf("剩余文件: %s", entry.Name())
		}
	}
}

func TestLogRotatorCleanupMaxBackups(t *testing.T) {
	// 测试日志文件清理 - 基于MaxBackups
	config := LogRotationConfig{
		MaxSize:    10,
		MaxAge:     0, // 不基于时间清理
		MaxBackups: 2, // 2个备份
	}

	rotator := NewLogRotator(config)
	testPath := "test_logs"
	testFile := "test.log"
	rotator.SetFilePath(testPath, testFile)

	// 创建测试目录
	os.MkdirAll(testPath, 0755)
	defer os.RemoveAll(testPath)

	// 创建一些测试文件
	testFiles := []string{
		"test.20250101-120000.log",
		"test.20250102-120000.log",
		"test.20250103-120000.log",
		"test.20250104-120000.log",
	}

	for _, fileName := range testFiles {
		filePath := filepath.Join(testPath, fileName)
		file, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}
		file.Close()
	}

	// 执行清理
	err := rotator.cleanup()
	if err != nil {
		t.Errorf("清理失败: %v", err)
	}

	// 验证清理结果
	entries, err := os.ReadDir(testPath)
	if err != nil {
		t.Fatalf("读取目录失败: %v", err)
	}

	// 应该只剩下2个文件（MaxBackups=2）
	expectedCount := 2
	if len(entries) != expectedCount {
		t.Errorf("清理后文件数量不正确: 期望%d, 实际%d", expectedCount, len(entries))
		for _, entry := range entries {
			t.Logf("剩余文件: %s", entry.Name())
		}
	}
}

func TestLogRotatorGetLogFiles(t *testing.T) {
	// 测试获取日志文件列表
	config := LogRotationConfig{
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 5,
	}

	rotator := NewLogRotator(config)
	testPath := "test_logs"
	testFile := "test.log"
	rotator.SetFilePath(testPath, testFile)

	// 创建测试目录
	os.MkdirAll(testPath, 0755)
	defer os.RemoveAll(testPath)

	// 创建一些测试文件
	testFiles := []string{
		"test.20250101-120000.log",
		"test.20250102-120000.log",
		"other.log", // 不是日志文件
		"test.txt",  // 不是日志文件
	}

	for _, fileName := range testFiles {
		filePath := filepath.Join(testPath, fileName)
		file, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}
		file.Close()
	}

	// 获取日志文件列表
	files, err := rotator.getLogFiles()
	if err != nil {
		t.Errorf("获取日志文件列表失败: %v", err)
	}

	// 应该返回以"test"开头的文件（包括test.txt）
	expectedCount := 3
	if len(files) != expectedCount {
		t.Errorf("日志文件数量不正确: 期望%d, 实际%d", expectedCount, len(files))
		for _, file := range files {
			t.Logf("找到文件: %s", file.Name())
		}
	}

	for _, file := range files {
		if !strings.HasPrefix(file.Name(), "test") {
			t.Errorf("包含非日志文件: %s", file.Name())
		}
	}
}
