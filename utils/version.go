package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// VersionManager 版本号管理器
type VersionManager struct {
	VersionFile string // VERSION 文件路径
	GitTag      string // Git tag 前缀
}

// NewVersionManager 创建版本号管理器
func NewVersionManager() *VersionManager {
	return &VersionManager{
		VersionFile: "VERSION.txt",
		GitTag:      "v",
	}
}

// GetVersion 获取版本号，按优先级顺序：
// 1. 环境变量 VERSION
// 2. Git tag (最新标签)
// 3. VERSION 文件
// 4. 默认版本
func (vm *VersionManager) GetVersion() string {
	// 1. 检查环境变量
	if version := os.Getenv("VERSION"); version != "" {
		return version
	}

	// 2. 尝试从 Git tag 获取
	if version := vm.GetVersionFromGitTag(); version != "" {
		return version
	}

	// 3. 尝试从 VERSION 文件获取
	if version := vm.GetVersionFromFile(); version != "" {
		return version
	}

	// 4. 返回默认版本
	return "1.0.0"
}

// GetVersionFromGitTag 从 Git tag 获取版本号
func (vm *VersionManager) GetVersionFromGitTag() string {
	// 获取最新的 tag
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	tag := strings.TrimSpace(string(output))

	// 移除 tag 前缀 (如 "v1.0.0" -> "1.0.0")
	tag = strings.TrimPrefix(tag, vm.GitTag)

	// 验证版本号格式 (x.y.z)
	if vm.isValidVersion(tag) {
		return tag
	}

	return ""
}

// GetVersionFromFile 从 VERSION 文件获取版本号
func (vm *VersionManager) GetVersionFromFile() string {
	file, err := os.Open(vm.VersionFile)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		version := strings.TrimSpace(scanner.Text())
		if vm.isValidVersion(version) {
			return version
		}
	}

	return ""
}

// isValidVersion 验证版本号格式
func (vm *VersionManager) isValidVersion(version string) bool {
	// 匹配语义化版本号格式 (x.y.z)
	pattern := `^\d+\.\d+\.\d+(-[a-zA-Z0-9\-\.]+)?(\+[a-zA-Z0-9\-\.]+)?$`
	matched, _ := regexp.MatchString(pattern, version)
	return matched
}

// GetGitCommit 获取 Git commit hash
func (vm *VersionManager) GetGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// GetBuildTime 获取构建时间
func (vm *VersionManager) GetBuildTime() string {
	return fmt.Sprintf("%d", getCurrentTimestamp())
}

// IncrementVersion 增加版本号
func (vm *VersionManager) IncrementVersion(versionType string) (string, error) {
	version := vm.GetVersion()

	// 解析版本号
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid version format: %s", version)
	}

	var major, minor, patch int
	if _, err := fmt.Sscanf(parts[0], "%d", &major); err != nil {
		return "", err
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &minor); err != nil {
		return "", err
	}
	if _, err := fmt.Sscanf(parts[2], "%d", &patch); err != nil {
		return "", err
	}

	// 根据类型增加版本号
	switch versionType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return "", fmt.Errorf("invalid version type: %s (must be major, minor, or patch)", versionType)
	}

	newVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	return newVersion, nil
}

// SetVersion 设置版本号到 VERSION 文件
func (vm *VersionManager) SetVersion(version string) error {
	if !vm.isValidVersion(version) {
		return fmt.Errorf("invalid version format: %s", version)
	}

	file, err := os.Create(vm.VersionFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(version + "\n")
	return err
}

// CreateGitTag 创建 Git tag
func (vm *VersionManager) CreateGitTag(version string) error {
	if !vm.isValidVersion(version) {
		return fmt.Errorf("invalid version format: %s", version)
	}

	tagName := vm.GitTag + version
	cmd := exec.Command("git", "tag", "-a", tagName, "-m", fmt.Sprintf("Release %s", version))
	return cmd.Run()
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() int64 {
	return 0 // 这个值会在构建时被替换
}
