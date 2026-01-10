package logger

import (
	"fmt"
	"hahacking/config"
	"path/filepath"
	"strings"
	"time"
)

// getLogFilePath get the log file save path
func getLogFilePath() string {
	// 优先从 env.json 读取配置
	logConfig := config.Get("log")
	if logConfig != nil {
		if filename, ok := logConfig["filename"].(string); ok && filename != "" {
			// 如果配置了完整的文件路径，提取目录部分
			dir := filepath.Dir(filename)
			if dir != "." && dir != "" {
				return fmt.Sprintf("%s%s%s/", config.C.BasePath, config.C.AppSetting.RuntimeRootPath, dir)
			}
		}
	}
	// 回退到 app.ini 配置
	return fmt.Sprintf("%s%s%s", config.C.BasePath, config.C.AppSetting.RuntimeRootPath, config.C.AppSetting.LogSavePath)
}

// getLogFileName get the save name of the log file
func getLogFileName() string {
	// 在生产环境，优先使用配置文件中的 filename
	if config.ENV == "prd" {
		logConfig := config.Get("log")
		if logConfig != nil {
			if filename, ok := logConfig["filename"].(string); ok && filename != "" {
				// 如果配置的是完整路径（包含目录），只提取文件名
				// 例如 "logs/app.log" -> "app.log"
				if idx := strings.LastIndex(filename, "/"); idx >= 0 {
					return filename[idx+1:]
				}
				return filename
			}
		}
		// 如果没有配置 filename，使用固定的 app.log
		return "app.log"
	}

	// 开发环境使用时间戳文件名
	return fmt.Sprintf("%s%s.%s",
		config.C.AppSetting.LogSaveName,
		time.Now().Format(config.C.AppSetting.TimeFormat),
		config.C.AppSetting.LogFileExt,
	)
}
