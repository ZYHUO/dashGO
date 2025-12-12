package main

import (
	"testing"
	"time"
)

// TestAgent_UpdateFields 测试 Agent 的更新相关字�?
func TestAgent_UpdateFields(t *testing.T) {
	// 测试启用自动更新
	agent := NewAgent(false, true, 3600)
	
	if !agent.autoUpdate {
		t.Error("Expected autoUpdate to be true")
	}
	
	if agent.updateCheckInterval != 3600*time.Second {
		t.Errorf("Expected updateCheckInterval to be 3600s, got %v", agent.updateCheckInterval)
	}
	
	if agent.updating {
		t.Error("Expected updating to be false initially")
	}
	
	// 测试禁用自动更新
	agent2 := NewAgent(false, false, 0)
	
	if agent2.autoUpdate {
		t.Error("Expected autoUpdate to be false")
	}
	
	if agent2.updateCheckInterval != 0 {
		t.Errorf("Expected updateCheckInterval to be 0, got %v", agent2.updateCheckInterval)
	}
}

// TestAgent_UpdateMutex 测试更新互斥�?
func TestAgent_UpdateMutex(t *testing.T) {
	agent := NewAgent(false, true, 3600)
	
	// 设置 updating 标志�?true，模拟正在进行的更新
	agent.updateMutex.Lock()
	agent.updating = true
	agent.updateMutex.Unlock()
	
	// 尝试第二次更新应该被阻止
	updateInfo := &UpdateInfo{
		LatestVersion: "v1.1.0",
		DownloadURL:   "https://example.com/download",
		SHA256:        "abc123",
		FileSize:      1024,
		Strategy:      "auto",
	}
	
	// 尝试更新应该立即返回错误
	err := agent.performUpdate(updateInfo)
	if err == nil {
		t.Error("Expected error due to concurrent update, but got nil")
	}
	
	if err.Error() != "更新已在进行�? {
		t.Errorf("Expected '更新已在进行�? error, got: %v", err)
	} else {
		t.Logf("�?Concurrent update correctly blocked: %v", err)
	}
}

// TestAgent_VersionLogging 测试启动时记录版�?
func TestAgent_VersionLogging(t *testing.T) {
	agent := NewAgent(false, true, 3600)
	
	currentVersion := agent.versionManager.GetCurrentVersion()
	if currentVersion != Version {
		t.Errorf("Expected version %s, got %s", Version, currentVersion)
	}
	
	if currentVersion == "" {
		t.Error("Version should not be empty")
	}
	
	t.Logf("Current version: %s", currentVersion)
}

// TestAgent_UpdateCheckInterval 测试更新检查间隔配�?
func TestAgent_UpdateCheckInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval int
		expected time.Duration
	}{
		{
			name:     "1 hour",
			interval: 3600,
			expected: 3600 * time.Second,
		},
		{
			name:     "30 minutes",
			interval: 1800,
			expected: 1800 * time.Second,
		},
		{
			name:     "5 minutes",
			interval: 300,
			expected: 300 * time.Second,
		},
		{
			name:     "disabled",
			interval: 0,
			expected: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewAgent(false, true, tt.interval)
			
			if agent.updateCheckInterval != tt.expected {
				t.Errorf("Expected interval %v, got %v", tt.expected, agent.updateCheckInterval)
			}
		})
	}
}
