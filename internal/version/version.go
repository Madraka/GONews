package version

import (
	"fmt"
	"runtime"
)

var (
	// Version information - set during build
	Version   = "v1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
	GoVersion = runtime.Version()
)

// VersionInfo contains version information
type VersionInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
	GoVersion string `json:"go_version"`
}

// GetVersionInfo returns version information
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
		GoVersion: GoVersion,
	}
}

// PrintVersion prints version information
func PrintVersion() {
	info := GetVersionInfo()
	fmt.Printf("News API %s\n", info.Version)
	fmt.Printf("Build Time: %s\n", info.BuildTime)
	fmt.Printf("Git Commit: %s\n", info.GitCommit)
	fmt.Printf("Go Version: %s\n", info.GoVersion)
}
