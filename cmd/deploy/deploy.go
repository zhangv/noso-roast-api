package main

//必须用main包
import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"hahacking/utils"
)

const (
	serverUser  = "root"            // 修改为服务器用户名
	serverHost  = "api.noso.cafe"   // 修改为服务器IP
	remotePath  = "/home2/noso-api" // 服务器上应用的路径
	serviceName = "nosoapi"         // systemd 服务名
	// API 检查配置
	apiBaseURL = "https://api.noso.cafe"
	apiTimeout = 10 * time.Second
)

/**
 * 使用方法：
 * 部署新版本：go run deploy/deploy.go
 * 回滚到上一个版本：go run deploy/deploy.go -r 1
 *  ** -r 0 表示回滚到当前最新版本
 * 显示日志：go run deploy/deploy.go -l
 * 运行测试：go run deploy/deploy.go -t
 * 构建应用：go run deploy/deploy.go -b
 * 冒烟测试：go run deploy/deploy.go -st
 *
 * 注意：日志轮转和清理现在完全自动化，无需手动干预
 * - 自动轮转：每小时检查文件大小和时间
 * - 自动清理：每天凌晨2点清理旧日志
 * - 配置：通过 config/env.json 中的 log 部分配置
 */
var ReleaseFile string
var ReleaseNo string      //build no
var ReleaseVersion string //版本号

func main() {
	ReleaseFile = serviceName

	// 解析命令行参数
	rollback := flag.Int("r", -1, "回滚到第几个版本（1表示上一个版本,0表示当前最新版本）")
	runtestsOnly := flag.Bool("t", false, "是否只运行测试")
	runtests := flag.String("tc", "all", "是否运行测试(all|model|service|common|utils|scheduler|middleware)")
	runbuild := flag.Bool("b", false, "是否运行构建")
	smokeTest := flag.Bool("st", false, "是否运行冒烟测试")
	showlog := flag.Bool("l", false, "是否显示日志")
	releaseFlag := flag.String("release", "patch", "是否发布版本(patch|minor|major)")
	flag.Parse()

	if *rollback >= 0 {
		Rollback(*rollback)
		return
	}

	if *runtestsOnly {
		RunTests(*runtests)
		return
	}

	if *runbuild {
		BuildApp()
		return
	}
	if *smokeTest {
		SmokeTest()
		return
	}
	if *showlog {
		CheckLog()
		return
	}
	// 获取版本号
	vm := utils.NewVersionManager()
	ReleaseVersion = vm.GetVersion()

	if *releaseFlag != "" {
		newVersion, err := vm.IncrementVersion(*releaseFlag)
		if err != nil {
			fmt.Printf("❌ 版本增加失败: %v\n", err)
			os.Exit(1)
		}
		ReleaseVersion = newVersion

		// 保存新版本号到文件
		if err := vm.SetVersion(ReleaseVersion); err != nil {
			fmt.Printf("⚠️  保存版本号失败: %v\n", err)
		}

		// 创建 Git 标签
		if err := vm.CreateGitTag(ReleaseVersion); err != nil {
			fmt.Printf("⚠️  创建 Git 标签失败: %v\n", err)
		} else {
			fmt.Printf("🏷️  Git 标签已创建: v%s\n", ReleaseVersion)
		}

		fmt.Printf("📦 版本已更新为: %s\n", ReleaseVersion)
	}
	fmt.Println("🚀 开始构建、上传并重启 Go Web App...")
	var sshCommand string
	var output string
	var err error

	//step 0: 运行测试
	if err := RunTests(*runtests); err != nil {
		// fmt.Println("❌ 测试失败:", err)
		return
	}

	// Step 1: 编译 Go Web 应用
	if err := BuildApp(); err != nil {
		// fmt.Println("❌ 构建失败:", err)
		return
	}

	// Step 2: 上传到服务器
	UploadBinary()
	// 上传 config 目录下的所有文件
	configPath := "config/*"
	remoteConfigPath := fmt.Sprintf("%s@%s:%s/config/", serverUser, serverHost, remotePath)
	fmt.Println("📤 正在上传 config 目录下的所有文件到服务器:", remoteConfigPath)
	if err := runCommand(fmt.Sprintf("scp -r %s %s", configPath, remoteConfigPath)); err != nil {
		fmt.Println("❌ 上传 config 文件失败:", err)
		return
	}
	fmt.Println("✅ 上传成功!")

	// 2.1 建立软链接
	releasePath := fmt.Sprintf("%s/releases", remotePath)
	sshCommand = fmt.Sprintf(
		"sudo ln -sf %s/%s %s/%s",
		releasePath, ReleaseFile, remotePath, serviceName,
	)
	fmt.Println("🔗 正在创建软链接:", sshCommand)
	if output, err = runRemoteCommand(sshCommand); err != nil {
		fmt.Println("❌ 创建软链接失败:", err)
		return
	}
	fmt.Println(output)
	fmt.Println("✅ 软链接创建成功!")

	sshCommand = fmt.Sprintf("sudo systemctl stop %s", serviceName)
	if output, err = runRemoteCommand(sshCommand); err != nil {
		fmt.Println("❌ 服务停止失败:", err)
		return
	}
	fmt.Println(output)
	fmt.Println("🔄 服务停止成功!")

	// Step 3: 远程重启服务
	fmt.Println("🔄 正在重启服务器上的服务:", serviceName)

	reloadCmd := fmt.Sprintf(
		"chmod +x %s/%s && sudo systemctl restart %s && sudo systemctl status %s --no-pager",
		releasePath, ReleaseFile, serviceName, serviceName,
	)
	if output, err = runRemoteCommand(reloadCmd); err != nil {
		fmt.Println("❌ 服务重启失败:", err)
		return
	}

	fmt.Println(output)
	fmt.Println("✅ 远程服务已成功重启!")

	// Step 4: 清理旧版本文件
	CleanupOldVersions()
	//清理本地的旧版本文件
	CleanupOldLocalVersions()

	fmt.Println("🎉 部署完成!")

	fmt.Println("⏳ 等待服务启动...")
	time.Sleep(5 * time.Second)
	CheckLog()
	if err := SmokeTest(); err != nil {
		fmt.Println("❌ 冒烟测试失败:", err)
		return
	}

}
func CheckLog() {
	fmt.Println("🔨 正在检查日志...")
	cmd := fmt.Sprintf("tail -n 20 %s/logs/app.log", remotePath)
	// cmd := fmt.Sprintf("less %s/logs/app.log ", remotePath)
	output, err := runRemoteCommand(cmd)
	if err != nil {
		fmt.Println("❌ 检查日志失败:", err)
		return
	}
	fmt.Println(output)
	fmt.Println("✅ 日志检查完成!")
}

func SmokeTest() error {
	fmt.Println("🔨 正在运行冒烟测试...")
	if err := checkHTTPAPI(); err != nil {
		fmt.Println("❌ HTTP API 检查失败:", err)
		return err
	}
	fmt.Println("✅ HTTP API 检查通过!")
	fmt.Println("🎉 冒烟测试完成!")
	return nil
}

// checkHTTPAPI 检查 HTTP API 是否正常工作
func checkHTTPAPI() error {
	// 定义要检查的 API 端点
	endpoints := []struct {
		name   string
		url    string
		method string
	}{
		{"健康检查", apiBaseURL + "/ping", "GET"},
		// 可以根据需要添加更多端点
	}

	client := &http.Client{
		Timeout: apiTimeout,
	}

	for _, endpoint := range endpoints {
		fmt.Printf("🔍 检查 %s: %s\n", endpoint.name, endpoint.url)

		req, err := http.NewRequest(endpoint.method, endpoint.url, nil)
		if err != nil {
			return fmt.Errorf("创建请求失败 %s: %v", endpoint.name, err)
		}

		// 添加必要的请求头
		req.Header.Set("User-Agent", "Deploy-SmokeTest/1.0")
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("请求失败 %s: %v", endpoint.name, err)
		}
		defer resp.Body.Close()

		// 读取响应体
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取响应失败 %s: %v", endpoint.name, err)
		}

		// 检查状态码
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("%s 返回错误状态码: %d, 响应: %s",
				endpoint.name, resp.StatusCode, string(body))
		}

		fmt.Printf("✅ %s 检查通过 (状态码: %d)\n", endpoint.name, resp.StatusCode)

		// 打印响应内容（限制长度）
		responseStr := string(body)
		if len(responseStr) > 200 {
			responseStr = responseStr[:200] + "..."
		}
		fmt.Printf("📄 响应内容: %s\n", responseStr)
		if ReleaseNo != "" {
			if strings.Contains(responseStr, ReleaseNo) {
				fmt.Printf("✅ %s 检查通过 (包含版本号: %s)\n", endpoint.name, ReleaseNo)
			} else {
				return fmt.Errorf("%s 不包含版本号: %s", endpoint.name, ReleaseNo)
			}
		}
	}

	return nil
}

func CleanupOldLocalVersions() {
	fmt.Println("🧹 正在清理本地旧版本文件...")

	// 获取所有以服务名开头的文件，按时间排序
	listCmd := fmt.Sprintf("ls -t %s* | tail -n +6", serviceName)
	cmd := exec.Command("bash", "-c", listCmd)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("❌ 获取本地文件列表失败:", err)
		return
	}

	// 计算要删除的文件
	filesToDelete := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(filesToDelete) == 0 || (len(filesToDelete) == 1 && filesToDelete[0] == "") {
		fmt.Println("✅ 没有需要清理的本地旧文件")
		return
	}

	// 删除旧文件
	for _, file := range filesToDelete {
		if err := os.Remove(file); err != nil {
			fmt.Printf("❌ 删除文件 %s 失败: %v\n", file, err)
			continue
		}
	}

	fmt.Printf("✅ 成功清理了 %d 个本地旧版本文件\n", len(filesToDelete))
}

func UploadBinary() {
	releasePath := fmt.Sprintf("%s/releases", remotePath)
	//创建releases目录
	runCommand(fmt.Sprintf("ssh %s@%s 'mkdir -p %s'", serverUser, serverHost, releasePath))
	remoteFile := fmt.Sprintf("%s@%s:%s/%s", serverUser, serverHost, releasePath, ReleaseFile)
	fmt.Println("📤 正在上传到服务器:", remoteFile)
	if err := runCommand(fmt.Sprintf("scp %s %s", ReleaseFile, remoteFile)); err != nil {
		fmt.Println("❌ 上传失败:", err)
		return
	}
	fmt.Println("✅ 上传成功!")
}

func BuildApp() error {
	ReleaseNo = time.Now().Format("20060102150405")
	ReleaseFile += ReleaseNo
	ldflags := "-s -w" //-s: 移除符号表（symbol table），这会减小可执行文件的大小，但会使调试变得困难。-w: 移除DWARF调试信息，进一步减小可执行文件的大小。
	gitCommit := "$(git rev-parse --short HEAD)"

	// 使用全局版本号变量
	version := ReleaseVersion

	ldflags += fmt.Sprintf(" -X 'hahacking/version.BuildTime=%s' -X 'hahacking/version.Version=%s' -X 'hahacking/version.GitCommit=%s'", ReleaseNo, version, gitCommit)
	cmd := fmt.Sprintf("CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags=\"%s\" -buildvcs=false -o %s", ldflags, ReleaseFile)
	fmt.Println("🔨 正在编译应用:", cmd)
	if err := runCommand(cmd); err != nil {
		fmt.Println("❌ 构建失败:", err)
		return err
	}
	fmt.Println("✅ 构建完成!")
	return nil
}

func RunTests(tests string) error {
	fmt.Println("🔨 正在运行测试...")
	// 跳过 cmd 目录，避免编译错误
	if tests == "all" {
		// tests = "./models/... ./service/... ./common/... ./utils/... ./scheduler/... ./middleware/..."
		tests = "./..."
	}
	cmd := "go test -coverprofile=coverage.out " + tests
	if err := runCommand(cmd); err != nil {
		fmt.Println("❌ 测试失败:", err)
		return err
	}
	fmt.Println("✅ 测试完成!")
	fmt.Println("📊 正在生成覆盖率报告...")
	cmd = "go tool cover -html=coverage.out -o coverage.html"
	if err := runCommand(cmd); err != nil {
		fmt.Println("❌ 生成覆盖率报表失败:", err)
		return err
	}

	cmd = "go tool cover -func=coverage.out | grep total: | awk '{print $3}'"
	var outBuf, errBuf bytes.Buffer
	runCommandOutput(cmd, &outBuf, &errBuf)
	fmt.Println("测试覆盖率:", outBuf.String())
	return nil
}
func runRemoteCommand(cmd string) (string, error) {
	remoteCmd := "ssh %s@%s '%s'"
	remoteCmd = fmt.Sprintf(remoteCmd, serverUser, serverHost, cmd)
	log.Println("🔄 正在执行远程命令:", remoteCmd)
	execCmd := exec.Command("bash", "-c", remoteCmd)
	output, err := execCmd.Output()
	return string(output), err
}

// CleanupOldVersions 清理远程服务器上的旧版本文件，只保留最近3个文件
func CleanupOldVersions() {
	fmt.Println("🧹 正在清理旧版本文件...")
	listCmd := "cd %s && ls -t %s* | tail -n +4"
	releasePath := fmt.Sprintf("%s/releases", remotePath)
	listCmd = fmt.Sprintf(listCmd, releasePath, serviceName)
	output, err := runRemoteCommand(listCmd)
	if err != nil {
		fmt.Println("❌ 获取文件列表失败:", err)
		return
	}

	// 计算要删除的文件数量
	filesToDelete := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(filesToDelete) == 0 || (len(filesToDelete) == 1 && filesToDelete[0] == "") {
		fmt.Println("✅ 没有需要清理的旧文件")
		return
	}

	// 执行删除命令
	cleanupCmd := fmt.Sprintf("%s | xargs -r rm -f", listCmd)
	_, err = runRemoteCommand(cleanupCmd)
	if err != nil {
		fmt.Println("❌ 清理旧版本文件失败:", err)
		return
	}

	fmt.Printf("✅ 成功清理了 %d 个旧版本文件\n", len(filesToDelete))
}

// Rollback 回滚到指定版本
func Rollback(version int) {
	fmt.Printf("🔄 正在回滚到第 %d 个版本...\n", version)
	releasePath := fmt.Sprintf("%s/releases", remotePath)
	// 获取远程服务器上的版本列表
	listCmd := fmt.Sprintf(
		"ssh %s@%s 'cd %s && ls -t %s*'",
		serverUser, serverHost, releasePath, serviceName,
	)

	cmd := exec.Command("bash", "-c", listCmd)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("❌ 获取版本列表失败:", err)
		return
	}

	// 解析版本列表
	versions := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(versions) == 0 || (len(versions) == 1 && versions[0] == "") {
		fmt.Println("❌ 没有找到可回滚的版本")
		return
	}

	// 检查版本号是否有效
	if version > len(versions) {
		fmt.Printf("❌ 版本号无效，当前共有 %d 个版本\n", len(versions))
		return
	}

	// 获取目标版本文件名
	targetVersion := versions[version]

	// 停止当前服务
	stopCmd := fmt.Sprintf(
		"ssh %s@%s 'sudo systemctl stop %s'",
		serverUser, serverHost, serviceName,
	)
	if err := runCommand(stopCmd); err != nil {
		fmt.Println("❌ 停止服务失败:", err)
		return
	}

	// 创建软链接到目标版本
	linkCmd := fmt.Sprintf(
		"ssh %s@%s 'sudo ln -sf %s/%s %s/%s'",
		serverUser, serverHost, releasePath, targetVersion, releasePath, serviceName,
	)
	if err := runCommand(linkCmd); err != nil {
		fmt.Println("❌ 创建软链接失败:", err)
		return
	}

	// 重启服务
	restartCmd := fmt.Sprintf(
		"ssh %s@%s 'chmod +x %s/%s && sudo systemctl restart %s && sudo systemctl status %s --no-pager'",
		serverUser, serverHost, releasePath, targetVersion, serviceName, serviceName,
	)
	if err := runCommand(restartCmd); err != nil {
		fmt.Println("❌ 重启服务失败:", err)
		return
	}

	fmt.Printf("✅ 已成功回滚到版本: %s\n", targetVersion)
}

func runCommand(cmd string) error {
	return runCommandOutput(cmd, nil, nil)
}

// 运行 shell 命令
func runCommandOutput(cmd string, output *bytes.Buffer, errout *bytes.Buffer) error {
	// fmt.Println("▶️ 执行命令:", cmd)
	command := exec.Command("bash", "-c", cmd)
	if output == nil {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	} else {
		command.Stdout = output
		command.Stderr = errout
	}
	return command.Run()
}
