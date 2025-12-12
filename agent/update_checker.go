package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// UpdateInfo 更新信息
type UpdateInfo struct {
	LatestVersion string `json:"latest_version"`
	DownloadURL   string `json:"download_url"`
	SHA256        string `json:"sha256"`
	FileSize      int64  `json:"file_size"`
	Strategy      string `json:"strategy"` // "auto" or "manual"
	ReleaseNotes  string `json:"release_notes"`
}

// UpdateChecker 更新检查器
type UpdateChecker struct {
	panelURL          string
	token             string
	client            *http.Client
	versionManager    *VersionManager
	securityValidator *SecurityValidator
}

// NewUpdateChecker 创建更新检查器
func NewUpdateChecker(panelURL, token string, versionManager *VersionManager) *UpdateChecker {
	return &UpdateChecker{
		panelURL:          panelURL,
		token:             token,
		client:            &http.Client{Timeout: 30 * time.Second},
		versionManager:    versionManager,
		securityValidator: NewSecurityValidator(),
	}
}

// CheckUpdate 检查更�?
func (uc *UpdateChecker) CheckUpdate(currentVersion string) (*UpdateInfo, error) {
	// Validate token before making request
	if err := uc.securityValidator.ValidateToken(uc.token); err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	url := uc.panelURL + "/api/v1/agent/version"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", uc.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := uc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data UpdateInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Validate the update info received from server
	if err := uc.securityValidator.ValidateUpdateInfo(&result.Data, uc.token); err != nil {
		return nil, fmt.Errorf("update info validation failed: %w", err)
	}

	return &result.Data, nil
}

// ShouldUpdate 判断是否需要更�?
func (uc *UpdateChecker) ShouldUpdate(updateInfo *UpdateInfo) (bool, error) {
	if updateInfo == nil {
		return false, fmt.Errorf("update info is nil")
	}

	// 比较版本�?
	cmp, err := uc.versionManager.CompareVersion(updateInfo.LatestVersion)
	if err != nil {
		return false, fmt.Errorf("failed to compare versions: %w", err)
	}

	// 如果当前版本更旧（cmp == -1），则需要更�?
	return cmp == -1, nil
}
