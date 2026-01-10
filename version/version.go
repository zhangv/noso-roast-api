package version

// BuildInfo 构建信息
type BuildInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
}

// 这些变量会在构建时被注入
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// GetBuildInfo 获取构建信息
func GetBuildInfo() *BuildInfo {
	return &BuildInfo{
		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
	}
}
