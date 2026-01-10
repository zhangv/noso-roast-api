package logger

import (
	"os"
	"testing"
)

func TestLogLevelFiltering(t *testing.T) {
	// 测试不同日志级别的过滤功能

	// 设置日志级别为INFO
	SetLogLevel(INFO)

	// 验证级别检查
	if !IsLevelEnabled(INFO) {
		t.Error("INFO级别应该被启用")
	}
	if !IsLevelEnabled(ERROR) {
		t.Error("ERROR级别应该被启用")
	}
	if IsLevelEnabled(DEBUG) {
		t.Error("DEBUG级别不应该被启用")
	}

	// 设置日志级别为ERROR
	SetLogLevel(ERROR)

	if !IsLevelEnabled(ERROR) {
		t.Error("ERROR级别应该被启用")
	}
	if !IsLevelEnabled(FATAL) {
		t.Error("FATAL级别应该被启用")
	}
	if IsLevelEnabled(INFO) {
		t.Error("INFO级别不应该被启用")
	}
	if IsLevelEnabled(DEBUG) {
		t.Error("DEBUG级别不应该被启用")
	}
}

func TestLogLevelFromConfig(t *testing.T) {
	// 测试从配置文件读取日志级别
	level := getLogLevelFromConfig()

	// 验证返回的级别是有效的
	if level < DEBUG || level > FATAL {
		t.Errorf("无效的日志级别: %d", level)
	}
}

func TestLogLevelConstants(t *testing.T) {
	// 测试日志级别常量
	if DEBUG != 0 {
		t.Error("DEBUG应该是0")
	}
	if INFO != 1 {
		t.Error("INFO应该是1")
	}
	if WARNING != 2 {
		t.Error("WARNING应该是2")
	}
	if ERROR != 3 {
		t.Error("ERROR应该是3")
	}
	if FATAL != 4 {
		t.Error("FATAL应该是4")
	}
}

func TestLogLevelFlags(t *testing.T) {
	// 测试日志级别标志字符串
	expectedFlags := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

	if len(levelFlags) != len(expectedFlags) {
		t.Errorf("levelFlags长度不匹配: 期望%d, 实际%d", len(expectedFlags), len(levelFlags))
	}

	for i, flag := range expectedFlags {
		if levelFlags[i] != flag {
			t.Errorf("levelFlags[%d]不匹配: 期望%s, 实际%s", i, flag, levelFlags[i])
		}
	}
}

func TestLogLevelComparison(t *testing.T) {
	// 测试日志级别比较
	if DEBUG >= INFO {
		t.Error("DEBUG应该小于INFO")
	}
	if INFO >= WARNING {
		t.Error("INFO应该小于WARNING")
	}
	if WARNING >= ERROR {
		t.Error("WARNING应该小于ERROR")
	}
	if ERROR >= FATAL {
		t.Error("ERROR应该小于FATAL")
	}

}

func TestLogLevelThreadSafety(t *testing.T) {
	// 测试日志级别的线程安全性
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			SetLogLevel(DEBUG)
			GetLogLevel()
			IsLevelEnabled(INFO)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestLogLevelDefault(t *testing.T) {
	// 测试默认日志级别
	SetLogLevel(DEBUG)

	if GetLogLevel() != DEBUG {
		t.Error("默认日志级别应该是DEBUG")
	}

	// 验证所有级别都被启用
	if !IsLevelEnabled(DEBUG) {
		t.Error("DEBUG级别应该被启用")
	}
	if !IsLevelEnabled(INFO) {
		t.Error("INFO级别应该被启用")
	}
	if !IsLevelEnabled(WARNING) {
		t.Error("WARNING级别应该被启用")
	}
	if !IsLevelEnabled(ERROR) {
		t.Error("ERROR级别应该被启用")
	}
	if !IsLevelEnabled(FATAL) {
		t.Error("FATAL级别应该被启用")
	}
}

func TestLogLevelFatal(t *testing.T) {
	// 测试FATAL级别的特殊行为
	SetLogLevel(FATAL)

	// 只有FATAL级别应该被启用
	if IsLevelEnabled(DEBUG) {
		t.Error("DEBUG级别不应该被启用")
	}
	if IsLevelEnabled(INFO) {
		t.Error("INFO级别不应该被启用")
	}
	if IsLevelEnabled(WARNING) {
		t.Error("WARNING级别不应该被启用")
	}
	if IsLevelEnabled(ERROR) {
		t.Error("ERROR级别不应该被启用")
	}
	if !IsLevelEnabled(FATAL) {
		t.Error("FATAL级别应该被启用")
	}
}

// 测试主函数，用于手动验证日志输出
func TestLogOutput(t *testing.T) {
	// 跳过这个测试，除非需要手动验证
	if os.Getenv("TEST_LOG_OUTPUT") != "1" {
		t.Skip("跳过日志输出测试，设置TEST_LOG_OUTPUT=1来运行")
	}

	// 初始化日志系统
	Setup()

	// 测试不同级别的日志输出
	Debug("这是DEBUG日志")
	Info("这是INFO日志")
	Warn("这是WARN日志")
	Error("这是ERROR日志")
	// Fatal("这是FATAL日志") // 注意：这会终止程序

	// 测试日志级别过滤
	SetLogLevel(INFO)
	Debug("这条DEBUG日志不应该输出")
	Info("这条INFO日志应该输出")
	Warn("这条WARN日志应该输出")
	Error("这条ERROR日志应该输出")

	SetLogLevel(ERROR)
	Debug("这条DEBUG日志不应该输出")
	Info("这条INFO日志不应该输出")
	Warn("这条WARN日志不应该输出")
	Error("这条ERROR日志应该输出")
}
