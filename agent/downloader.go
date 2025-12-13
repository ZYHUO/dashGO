package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Downloader 文件下载告
type Downloader struct {
	client            *http.Client
	maxRetries        int
	retryDelay        time.Duration
	securityValidator *SecurityValidator
}

// NewDownloader 创建下载告
func NewDownloader() *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 5 * time.Minute, // 5分钟超时
		},
		maxRetries:        3,
		retryDelay:        2 * time.Second,
		securityValidator: NewSecurityValidator(),
	}
}

// ProgressCallback 下载进度回调函数
type ProgressCallback func(downloaded, total int64)

// Download 下载文件
// url: 下载地址
// destPath: 目标文件路径
// progressCallback: 进度回调函数（可选）
func (d *Downloader) Download(url, destPath string, progressCallback ProgressCallback) error {
	// Security validation: validate URL
	if err := d.securityValidator.ValidateDownloadURL(url); err != nil {
		return fmt.Errorf("URL validation failed: %w", err)
	}

	// Security validation: validate destination path
	if err := d.securityValidator.ValidateFilePath(destPath); err != nil {
		return fmt.Errorf("file path validation failed: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 发送请告
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 创建临时文件
	tmpPath := destPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// 获取文件总大告
	totalSize := resp.ContentLength
	var downloaded int64

	// 创建缓冲区用于流式下告
	buf := make([]byte, 32*1024) // 32KB 缓冲告
	
	// 下载文件
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			// 写入文件
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				os.Remove(tmpPath)
				return fmt.Errorf("failed to write file: %w", writeErr)
			}
			
			downloaded += int64(n)
			
			// 调用进度回调
			if progressCallback != nil {
				progressCallback(downloaded, totalSize)
			}
		}
		
		if err == io.EOF {
			break
		}
		if err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("failed to read response: %w", err)
		}
	}

	// 确保数据写入磁盘
	if err := out.Sync(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to sync file: %w", err)
	}

	// 关闭文件
	out.Close()

	// 重命名临时文件为目标文件
	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// DownloadWithRetry 带重试的下载
// 最多重告maxRetries 次，每次重试之间延迟递增
func (d *Downloader) DownloadWithRetry(url, destPath string) error {
	var lastErr error
	
	for attempt := 0; attempt <= d.maxRetries; attempt++ {
		if attempt > 0 {
			// 计算重试延迟（指数退避）
			delay := d.retryDelay * time.Duration(1<<uint(attempt-1))
			fmt.Printf("告等待 %v 后重告(告%d/%d 告...\n", delay, attempt, d.maxRetries)
			time.Sleep(delay)
		}
		
		// 尝试下载
		err := d.Download(url, destPath, func(downloaded, total int64) {
			if total > 0 {
				percentage := float64(downloaded) / float64(total) * 100
				fmt.Printf("\r📥 下载进度: %.2f%% (%.2f MB / %.2f MB)", 
					percentage,
					float64(downloaded)/1024/1024,
					float64(total)/1024/1024)
			} else {
				fmt.Printf("\r📥 已下告 %.2f MB", float64(downloaded)/1024/1024)
			}
		})
		
		if err == nil {
			fmt.Println() // 换行
			return nil
		}
		
		lastErr = err
		fmt.Printf("\n告下载失败: %v\n", err)
	}
	
	return fmt.Errorf("download failed after %d retries: %w", d.maxRetries, lastErr)
}
