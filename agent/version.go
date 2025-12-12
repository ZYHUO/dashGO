package main

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// Version 当前 Agent 版本
const Version = "v1.0.0"

// VersionManager 版本管理�?
type VersionManager struct {
	currentVersion string
}

// NewVersionManager 创建版本管理�?
func NewVersionManager(version string) *VersionManager {
	return &VersionManager{
		currentVersion: version,
	}
}

// GetCurrentVersion 获取当前版本
func (vm *VersionManager) GetCurrentVersion() string {
	return vm.currentVersion
}

// ParseVersion 解析版本�?
func (vm *VersionManager) ParseVersion(version string) (*semver.Version, error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}
	return v, nil
}

// CompareVersion 比较两个版本�?
// 返回: -1 (当前版本更旧), 0 (版本相同), 1 (当前版本更新)
func (vm *VersionManager) CompareVersion(remote string) (int, error) {
	currentVer, err := vm.ParseVersion(vm.currentVersion)
	if err != nil {
		return 0, fmt.Errorf("failed to parse current version: %w", err)
	}

	remoteVer, err := vm.ParseVersion(remote)
	if err != nil {
		return 0, fmt.Errorf("failed to parse remote version: %w", err)
	}

	return currentVer.Compare(remoteVer), nil
}
