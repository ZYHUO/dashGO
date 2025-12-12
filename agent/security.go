package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// SecurityValidator handles security validation operations
type SecurityValidator struct {
	allowedDownloadHosts []string // 允许的下载域名白名单（可选）
}

// NewSecurityValidator creates a new SecurityValidator instance
func NewSecurityValidator() *SecurityValidator {
	return &SecurityValidator{
		allowedDownloadHosts: []string{}, // 空列表表示允许所�?HTTPS 域名
	}
}

// NewSecurityValidatorWithWhitelist creates a SecurityValidator with domain whitelist
func NewSecurityValidatorWithWhitelist(allowedHosts []string) *SecurityValidator {
	return &SecurityValidator{
		allowedDownloadHosts: allowedHosts,
	}
}

// ValidateDownloadURL validates that the download URL is secure
// Requirements: 5.1 - 验证下载 URL 必须�?HTTPS
func (sv *SecurityValidator) ValidateDownloadURL(downloadURL string) error {
	if downloadURL == "" {
		return fmt.Errorf("download URL cannot be empty")
	}

	// Parse the URL
	parsedURL, err := url.Parse(downloadURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Verify HTTPS scheme
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("download URL must use HTTPS protocol, got: %s", parsedURL.Scheme)
	}

	// Verify host is not empty
	if parsedURL.Host == "" {
		return fmt.Errorf("download URL must have a valid host")
	}

	// If whitelist is configured, verify host is in whitelist
	if len(sv.allowedDownloadHosts) > 0 {
		allowed := false
		for _, allowedHost := range sv.allowedDownloadHosts {
			if parsedURL.Host == allowedHost {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("download host %s is not in whitelist", parsedURL.Host)
		}
	}

	return nil
}

// ValidateFilePath validates that the file path is safe and prevents path traversal
// Requirements: 5.1 - 验证文件路径，防止路径遍�?
func (sv *SecurityValidator) ValidateFilePath(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Clean the path to resolve any .. or . components
	cleanPath := filepath.Clean(filePath)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal detected in file path: %s", filePath)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify the path doesn't escape to sensitive directories (Unix-like systems only)
	// This is a basic check - in production you might want to define allowed directories
	if filepath.Separator == '/' { // Unix-like systems
		if strings.HasPrefix(absPath, "/etc") || 
		   strings.HasPrefix(absPath, "/sys") || 
		   strings.HasPrefix(absPath, "/proc") {
			return fmt.Errorf("access to system directory is not allowed: %s", absPath)
		}
	}

	return nil
}

// ValidateFilePermissions validates that the file has appropriate permissions
// Requirements: 5.2 - 验证文件权限
func (sv *SecurityValidator) ValidateFilePermissions(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	mode := fileInfo.Mode()

	// Check if file is a regular file (not a directory, symlink, etc.)
	if !mode.IsRegular() {
		return fmt.Errorf("file is not a regular file: %s", filePath)
	}

	// Use platform-specific implementation
	return validateFilePermissionsImpl(filePath, fileInfo)
}

// ValidateToken validates that the authentication token is present and valid
// Requirements: 5.1 - 添加 Token 认证检�?
func (sv *SecurityValidator) ValidateToken(token string) error {
	if token == "" {
		return fmt.Errorf("authentication token cannot be empty")
	}

	// Check minimum token length (basic validation)
	if len(token) < 16 {
		return fmt.Errorf("authentication token is too short (minimum 16 characters)")
	}

	// Check for suspicious characters that might indicate injection attempts
	if strings.ContainsAny(token, "\n\r\x00") {
		return fmt.Errorf("authentication token contains invalid characters")
	}

	return nil
}

// ValidateUpdateInfo validates all security aspects of update information
func (sv *SecurityValidator) ValidateUpdateInfo(updateInfo *UpdateInfo, token string) error {
	if updateInfo == nil {
		return fmt.Errorf("update info cannot be nil")
	}

	// Validate token
	if err := sv.ValidateToken(token); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	// Validate download URL
	if err := sv.ValidateDownloadURL(updateInfo.DownloadURL); err != nil {
		return fmt.Errorf("download URL validation failed: %w", err)
	}

	// Validate SHA256 hash format
	if len(updateInfo.SHA256) != 64 {
		return fmt.Errorf("invalid SHA256 hash length: expected 64 characters, got %d", len(updateInfo.SHA256))
	}

	// Validate file size is reasonable (not zero, not too large)
	if updateInfo.FileSize <= 0 {
		return fmt.Errorf("invalid file size: %d", updateInfo.FileSize)
	}

	// Check for unreasonably large files (e.g., > 500MB)
	const maxFileSize = 500 * 1024 * 1024 // 500MB
	if updateInfo.FileSize > maxFileSize {
		return fmt.Errorf("file size too large: %d bytes (max: %d bytes)", updateInfo.FileSize, maxFileSize)
	}

	// Validate strategy
	if updateInfo.Strategy != "auto" && updateInfo.Strategy != "manual" {
		return fmt.Errorf("invalid update strategy: %s (must be 'auto' or 'manual')", updateInfo.Strategy)
	}

	return nil
}

// ValidateBeforeDownload performs all security checks before downloading
func (sv *SecurityValidator) ValidateBeforeDownload(downloadURL, destPath, token string) error {
	// Validate token
	if err := sv.ValidateToken(token); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	// Validate download URL
	if err := sv.ValidateDownloadURL(downloadURL); err != nil {
		return fmt.Errorf("URL validation failed: %w", err)
	}

	// Validate destination file path
	if err := sv.ValidateFilePath(destPath); err != nil {
		return fmt.Errorf("file path validation failed: %w", err)
	}

	return nil
}

// ValidateAfterDownload performs security checks after downloading
func (sv *SecurityValidator) ValidateAfterDownload(filePath string) error {
	// Validate file path again
	if err := sv.ValidateFilePath(filePath); err != nil {
		return fmt.Errorf("file path validation failed: %w", err)
	}

	// Validate file permissions
	if err := sv.ValidateFilePermissions(filePath); err != nil {
		return fmt.Errorf("file permissions validation failed: %w", err)
	}

	return nil
}

// ValidateDownloadedFile performs basic security checks on a downloaded file
// This is used for temporary files that will be renamed later
func (sv *SecurityValidator) ValidateDownloadedFile(filePath string) error {
	// Validate file path
	if err := sv.ValidateFilePath(filePath); err != nil {
		return fmt.Errorf("file path validation failed: %w", err)
	}

	// Check if file exists and is a regular file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("file is not a regular file: %s", filePath)
	}

	// Check if file is readable
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("file is not readable: %w", err)
	}
	file.Close()

	return nil
}
