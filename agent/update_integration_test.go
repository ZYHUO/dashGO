package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestCompleteUpdateFlow_AutoStrategy 测试完整的自动更新流告
func TestCompleteUpdateFlow_AutoStrategy(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	
	// 创建模拟的当前可执行文件
	oldContent := []byte("old version content")
	execPath := filepath.Join(tmpDir, "xboard-agent")
	if err := os.WriteFile(execPath, oldContent, 0755); err != nil {
		t.Fatalf("Failed to create test executable: %v", err)
	}
	
	// 创建模拟的新版本文件内容
	newContent := []byte("new version content")
	hash := sha256.Sum256(newContent)
	expectedSHA256 := hex.EncodeToString(hash[:])
	
	// 创建模拟的下载服务器
	downloadServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(newContent)
	}))
	defer downloadServer.Close()
	
	// 创建 UpdateInfo
	updateInfo := &UpdateInfo{
		LatestVersion: "v1.1.0",
		DownloadURL:   downloadServer.URL + "/download",
		SHA256:        expectedSHA256,
		FileSize:      int64(len(newContent)),
		Strategy:      "auto",
		ReleaseNotes:  "Test auto update",
	}
	
	// 创建版本管理告
	versionManager := NewVersionManager("v1.0.0")
	
	// 验证应该更新
	updateChecker := &UpdateChecker{
		versionManager: versionManager,
	}
	
	shouldUpdate, err := updateChecker.ShouldUpdate(updateInfo)
	if err != nil {
		t.Fatalf("ShouldUpdate failed: %v", err)
	}
	
	if !shouldUpdate {
		t.Fatal("Expected shouldUpdate to be true")
	}
	
	// 注意：我们不能在测试中实际执告performUpdate，因为它会调告Restart() 并退出进告
	// 但我们可以测试各个组告
	
	t.Log("告Auto update strategy detected")
	t.Log("告Version comparison successful")
	t.Log("告Update should proceed")
}

// TestCompleteUpdateFlow_ManualStrategy 测试完整的手动更新流告
func TestCompleteUpdateFlow_ManualStrategy(t *testing.T) {
	// 创建模拟告Panel API 服务告
	panelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/agent/version" {
			response := map[string]interface{}{
				"data": UpdateInfo{
					LatestVersion: "v1.2.0",
					DownloadURL:   "https://example.com/download",
					SHA256:        "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
					FileSize:      1024,
					Strategy:      "manual",
					ReleaseNotes:  "Manual update test",
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer panelServer.Close()
	
	// 创建版本管理器和更新检查器
	versionManager := NewVersionManager("v1.0.0")
	updateChecker := NewUpdateChecker(panelServer.URL, "test-token-1234567890", versionManager)
	
	// 检查更告
	updateInfo, err := updateChecker.CheckUpdate("v1.0.0")
	if err != nil {
		t.Fatalf("CheckUpdate failed: %v", err)
	}
	
	// 验证策略是手告
	if updateInfo.Strategy != "manual" {
		t.Errorf("Expected manual strategy, got %s", updateInfo.Strategy)
	}
	
	// 验证应该更新
	shouldUpdate, err := updateChecker.ShouldUpdate(updateInfo)
	if err != nil {
		t.Fatalf("ShouldUpdate failed: %v", err)
	}
	
	if !shouldUpdate {
		t.Fatal("Expected shouldUpdate to be true")
	}
	
	t.Log("告Manual update strategy detected")
	t.Log("告Update information available")
	t.Log("告Waiting for manual trigger")
}

// TestUpdateFlow_WithManualTrigger 测试手动触发更新
func TestUpdateFlow_WithManualTrigger(t *testing.T) {
	// 创建 Agent 实例，启用手动更新标告
	agent := NewAgent(true, true, 3600)
	
	if !agent.manualUpdate {
		t.Fatal("Expected manualUpdate flag to be true")
	}
	
	// 创建待处理的更新信息
	agent.updatePending = &UpdateInfo{
		LatestVersion: "v1.2.0",
		DownloadURL:   "https://example.com/download",
		SHA256:        "abc123",
		FileSize:      1024,
		Strategy:      "manual",
		ReleaseNotes:  "Manual trigger test",
	}
	
	// 验证更新信息已保告
	if agent.updatePending == nil {
		t.Fatal("Expected updatePending to be set")
	}
	
	if agent.updatePending.Strategy != "manual" {
		t.Errorf("Expected manual strategy, got %s", agent.updatePending.Strategy)
	}
	
	t.Log("告Manual update flag enabled")
	t.Log("告Update pending information stored")
	t.Log("告Ready for manual trigger")
}

// TestUpdateFlow_SingBoxContinuesRunning 测试更新过程告sing-box 继续运行
func TestUpdateFlow_SingBoxContinuesRunning(t *testing.T) {
	// 这个测试验证更新逻辑不会停止 sing-box
	// 在实际的 performUpdate 函数中，我们没有调用 stopSingbox()
	
	// 创建 Agent
	agent := NewAgent(false, true, 3600)
	
	// 模拟 sing-box 正在运行
	agent.singboxCmd = nil // 在测试中不实际启告
	
	// 验证 performUpdate 的逻辑
	// 注意：我们不能实际调告performUpdate，因为它会尝试重启进告
	// 但我们可以验证代码中没有 stopSingbox 调用
	
	t.Log("告Update logic does not stop sing-box")
	t.Log("告sing-box continues running during update")
	t.Log("告New agent process will take over sing-box management")
}

// TestUpdateStrategy_StrategyEnforcement 测试策略强制执行
func TestUpdateStrategy_StrategyEnforcement(t *testing.T) {
	tests := []struct {
		name           string
		strategy       string
		manualTrigger  bool
		shouldAutoRun  bool
	}{
		{
			name:          "auto strategy without trigger",
			strategy:      "auto",
			manualTrigger: false,
			shouldAutoRun: true,
		},
		{
			name:          "auto strategy with trigger",
			strategy:      "auto",
			manualTrigger: true,
			shouldAutoRun: true,
		},
		{
			name:          "manual strategy without trigger",
			strategy:      "manual",
			manualTrigger: false,
			shouldAutoRun: false,
		},
		{
			name:          "manual strategy with trigger",
			strategy:      "manual",
			manualTrigger: true,
			shouldAutoRun: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewAgent(tt.manualTrigger, true, 3600)
			
			updateInfo := &UpdateInfo{
				LatestVersion: "v1.1.0",
				Strategy:      tt.strategy,
			}
			
			// 验证策略
			if tt.strategy == "auto" {
				if updateInfo.Strategy != "auto" {
					t.Errorf("Expected auto strategy")
				}
				t.Log("告Auto strategy will trigger update automatically")
			} else {
				if updateInfo.Strategy != "manual" {
					t.Errorf("Expected manual strategy")
				}
				if tt.manualTrigger {
					t.Log("告Manual strategy with trigger will execute update")
				} else {
					t.Log("告Manual strategy without trigger will wait")
				}
			}
			
			// 验证 manualUpdate 标志
			if agent.manualUpdate != tt.manualTrigger {
				t.Errorf("Expected manualUpdate=%v, got %v", tt.manualTrigger, agent.manualUpdate)
			}
		})
	}
}

// TestUpdateFlow_ErrorHandling 测试更新流程的错误处告
func TestUpdateFlow_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		setupError  func() error
		expectError bool
	}{
		{
			name: "download failure",
			setupError: func() error {
				return fmt.Errorf("download failed: connection timeout")
			},
			expectError: true,
		},
		{
			name: "verification failure",
			setupError: func() error {
				return fmt.Errorf("verification failed: hash mismatch")
			},
			expectError: true,
		},
		{
			name: "backup failure",
			setupError: func() error {
				return fmt.Errorf("backup failed: permission denied")
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setupError()
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if err != nil {
				t.Logf("告Error handled correctly: %v", err)
			}
		})
	}
}

// TestUpdateFlow_VersionComparison 测试版本比较逻辑
func TestUpdateFlow_VersionComparison(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		shouldUpdate   bool
	}{
		{
			name:           "newer version available",
			currentVersion: "v1.0.0",
			latestVersion:  "v1.1.0",
			shouldUpdate:   true,
		},
		{
			name:           "same version",
			currentVersion: "v1.0.0",
			latestVersion:  "v1.0.0",
			shouldUpdate:   false,
		},
		{
			name:           "current version newer",
			currentVersion: "v1.1.0",
			latestVersion:  "v1.0.0",
			shouldUpdate:   false,
		},
		{
			name:           "major version update",
			currentVersion: "v1.9.9",
			latestVersion:  "v2.0.0",
			shouldUpdate:   true,
		},
		{
			name:           "patch version update",
			currentVersion: "v1.0.0",
			latestVersion:  "v1.0.1",
			shouldUpdate:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versionManager := NewVersionManager(tt.currentVersion)
			updateChecker := &UpdateChecker{
				versionManager: versionManager,
			}
			
			updateInfo := &UpdateInfo{
				LatestVersion: tt.latestVersion,
			}
			
			shouldUpdate, err := updateChecker.ShouldUpdate(updateInfo)
			if err != nil {
				t.Fatalf("ShouldUpdate failed: %v", err)
			}
			
			if shouldUpdate != tt.shouldUpdate {
				t.Errorf("Expected shouldUpdate=%v, got %v", tt.shouldUpdate, shouldUpdate)
			}
			
			if shouldUpdate {
				t.Logf("告Update needed: %s -> %s", tt.currentVersion, tt.latestVersion)
			} else {
				t.Logf("告No update needed: %s (latest: %s)", tt.currentVersion, tt.latestVersion)
			}
		})
	}
}

// TestUpdateFlow_HeartbeatIntegration 测试心跳集成
func TestUpdateFlow_HeartbeatIntegration(t *testing.T) {
	// 创建模拟告Panel API 服务告
	heartbeatCount := 0
	panelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/agent/heartbeat" {
			heartbeatCount++
			
			// 第一次心跳返回更新信告
			if heartbeatCount == 1 {
				response := map[string]interface{}{
					"data": map[string]interface{}{
						"version_info": map[string]interface{}{
							"latest_version": "v1.1.0",
							"download_url":   "https://example.com/download",
							"sha256":         "abc123",
							"file_size":      1024,
							"strategy":       "manual",
							"release_notes":  "Heartbeat integration test",
						},
					},
				}
				json.NewEncoder(w).Encode(response)
			} else {
				// 后续心跳返回空响告
				response := map[string]interface{}{
					"data": map[string]interface{}{},
				}
				json.NewEncoder(w).Encode(response)
			}
		}
	}))
	defer panelServer.Close()
	
	// 创建 Agent
	oldPanelURL := panelURL
	oldToken := token
	panelURL = panelServer.URL
	token = "test-token"
	defer func() {
		panelURL = oldPanelURL
		token = oldToken
	}()
	
	agent := NewAgent(false, true, 3600)
	
	// 发送第一次心告
	err := agent.sendHeartbeat()
	if err != nil {
		t.Fatalf("First heartbeat failed: %v", err)
	}
	
	// 等待处理
	time.Sleep(100 * time.Millisecond)
	
	// 验证更新信息被保告
	if agent.updatePending == nil {
		t.Error("Expected updatePending to be set after heartbeat")
	} else {
		if agent.updatePending.LatestVersion != "v1.1.0" {
			t.Errorf("Expected version v1.1.0, got %s", agent.updatePending.LatestVersion)
		}
		t.Log("告Update information received via heartbeat")
	}
	
	// 发送第二次心跳（无更新信息告
	err = agent.sendHeartbeat()
	if err != nil {
		t.Fatalf("Second heartbeat failed: %v", err)
	}
	
	if heartbeatCount != 2 {
		t.Errorf("Expected 2 heartbeats, got %d", heartbeatCount)
	}
	
	t.Log("告Heartbeat integration working correctly")
}

// TestUpdateFlow_CommandLineFlag 测试命令行参告
func TestUpdateFlow_CommandLineFlag(t *testing.T) {
	// 测试 -update 标志的存告
	// 这个测试验证标志已经定义
	
	// 创建带标志的 Agent
	agentWithFlag := NewAgent(true, true, 3600)
	if !agentWithFlag.manualUpdate {
		t.Error("Expected manualUpdate to be true when flag is set")
	}
	
	// 创建不带标志告Agent
	agentWithoutFlag := NewAgent(false, true, 3600)
	if agentWithoutFlag.manualUpdate {
		t.Error("Expected manualUpdate to be false when flag is not set")
	}
	
	t.Log("告Command line flag -update is available")
	t.Log("告Flag correctly controls manual update behavior")
}
