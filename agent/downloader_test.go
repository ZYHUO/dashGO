package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDownloader_Download_Success 测试成功下载
func TestDownloader_Download_Success(t *testing.T) {
	// 创建测试服务告
	testContent := "test file content for download"
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	// 创建临时目录
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test-download.txt")

	// 创建下载器（使用测试服务器的客户端）
	downloader := NewDownloader()
	downloader.client = server.Client()

	// 测试下载
	progressCalled := false
	err := downloader.Download(server.URL, destPath, func(downloaded, total int64) {
		progressCalled = true
		if downloaded < 0 || total < 0 {
			t.Errorf("Invalid progress values: downloaded=%d, total=%d", downloaded, total)
		}
	})

	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	if !progressCalled {
		t.Error("Progress callback was not called")
	}

	// 验证文件内容
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Downloaded content mismatch: got %q, want %q", string(content), testContent)
	}
}

// TestDownloader_Download_HTTPSOnly 测试只允告HTTPS
func TestDownloader_Download_HTTPSOnly(t *testing.T) {
	downloader := NewDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	// 测试 HTTP URL（应该失败）
	err := downloader.Download("http://example.com/file", destPath, nil)
	if err == nil {
		t.Error("Expected error for HTTP URL, got nil")
	}
	if !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("Expected HTTPS error, got: %v", err)
	}

	// 测试无效 URL（应该失败）
	err = downloader.Download("ftp://example.com/file", destPath, nil)
	if err == nil {
		t.Error("Expected error for FTP URL, got nil")
	}
}

// TestDownloader_Download_ServerError 测试服务器错告
func TestDownloader_Download_ServerError(t *testing.T) {
	// 创建返回 500 错误的测试服务器
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	downloader := NewDownloader()
	downloader.client = server.Client()

	err := downloader.Download(server.URL, destPath, nil)
	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
	if !strings.Contains(err.Error(), "status code") {
		t.Errorf("Expected status code error, got: %v", err)
	}
}

// TestDownloader_Download_InvalidPath 测试无效文件路径
func TestDownloader_Download_InvalidPath(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	defer server.Close()

	downloader := NewDownloader()
	downloader.client = server.Client()

	// 使用不存在的目录
	err := downloader.Download(server.URL, "/nonexistent/path/file.txt", nil)
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

// TestDownloader_Download_ProgressCallback 测试进度回调
func TestDownloader_Download_ProgressCallback(t *testing.T) {
	// 创建较大的测试内告
	testContent := strings.Repeat("A", 100*1024) // 100KB
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	downloader := NewDownloader()
	downloader.client = server.Client()

	var lastDownloaded int64
	var lastTotal int64
	callCount := 0

	err := downloader.Download(server.URL, destPath, func(downloaded, total int64) {
		callCount++
		lastDownloaded = downloaded
		lastTotal = total

		// 验证进度递增
		if downloaded < 0 {
			t.Errorf("Downloaded should be non-negative: %d", downloaded)
		}
		if total != int64(len(testContent)) {
			t.Errorf("Total size mismatch: got %d, want %d", total, len(testContent))
		}
	})

	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	if callCount == 0 {
		t.Error("Progress callback was never called")
	}

	if lastDownloaded != int64(len(testContent)) {
		t.Errorf("Final downloaded size mismatch: got %d, want %d", lastDownloaded, len(testContent))
	}

	if lastTotal != int64(len(testContent)) {
		t.Errorf("Total size mismatch: got %d, want %d", lastTotal, len(testContent))
	}
}

// TestDownloader_DownloadWithRetry_Success 测试重试成功
func TestDownloader_DownloadWithRetry_Success(t *testing.T) {
	attemptCount := 0
	testContent := "test content"

	// 创建前两次失败，第三次成功的服务告
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	downloader := NewDownloader()
	downloader.client = server.Client()
	downloader.retryDelay = 10 * time.Millisecond // 缩短测试时间

	err := downloader.DownloadWithRetry(server.URL, destPath)
	if err != nil {
		t.Fatalf("DownloadWithRetry failed: %v", err)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}

	// 验证文件内容
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Content mismatch: got %q, want %q", string(content), testContent)
	}
}

// TestDownloader_DownloadWithRetry_MaxRetriesExceeded 测试超过最大重试次告
func TestDownloader_DownloadWithRetry_MaxRetriesExceeded(t *testing.T) {
	attemptCount := 0

	// 创建始终失败的服务器
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	downloader := NewDownloader()
	downloader.client = server.Client()
	downloader.retryDelay = 10 * time.Millisecond // 缩短测试时间

	err := downloader.DownloadWithRetry(server.URL, destPath)
	if err == nil {
		t.Error("Expected error after max retries, got nil")
	}

	// 应该尝试 1 次初告+ 3 次重告= 4 告
	expectedAttempts := downloader.maxRetries + 1
	if attemptCount != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attemptCount)
	}

	if !strings.Contains(err.Error(), "retries") {
		t.Errorf("Expected retry error message, got: %v", err)
	}
}

// TestDownloader_DownloadWithRetry_FirstAttemptSuccess 测试第一次就成功
func TestDownloader_DownloadWithRetry_FirstAttemptSuccess(t *testing.T) {
	attemptCount := 0
	testContent := "success on first try"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	downloader := NewDownloader()
	downloader.client = server.Client()

	err := downloader.DownloadWithRetry(server.URL, destPath)
	if err != nil {
		t.Fatalf("DownloadWithRetry failed: %v", err)
	}

	if attemptCount != 1 {
		t.Errorf("Expected 1 attempt, got %d", attemptCount)
	}

	// 验证文件内容
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Content mismatch: got %q, want %q", string(content), testContent)
	}
}

// TestDownloader_Download_NoProgressCallback 测试不提供进度回告
func TestDownloader_Download_NoProgressCallback(t *testing.T) {
	testContent := "test without callback"
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	downloader := NewDownloader()
	downloader.client = server.Client()

	// 不提供进度回告
	err := downloader.Download(server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Downloaded file does not exist")
	}
}

// TestDownloader_Download_LargeFile 测试下载大文告
func TestDownloader_Download_LargeFile(t *testing.T) {
	// 创建 1MB 的测试内告
	testContent := strings.Repeat("X", 1024*1024)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "large-file.bin")

	downloader := NewDownloader()
	downloader.client = server.Client()

	progressUpdates := 0
	err := downloader.Download(server.URL, destPath, func(downloaded, total int64) {
		progressUpdates++
	})

	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// 验证文件大小
	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() != int64(len(testContent)) {
		t.Errorf("File size mismatch: got %d, want %d", info.Size(), len(testContent))
	}

	// 应该有多次进度更告
	if progressUpdates < 2 {
		t.Errorf("Expected multiple progress updates, got %d", progressUpdates)
	}
}
