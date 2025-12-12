package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// Updater 更新器，负责执行更新操作
type Updater struct {
	execPath          string // 当前可执行文件路�?
	backupPath        string // 备份文件路径
	newPath           string // 新版本文件路�?
	securityValidator *SecurityValidator
}

// NewUpdater 创建更新�?
func NewUpdater() (*Updater, error) {
	// 获取当前可执行文件路�?
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	// 解析符号链接，获取真实路�?
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve symlink: %w", err)
	}

	dir := filepath.Dir(execPath)
	base := filepath.Base(execPath)

	updater := &Updater{
		execPath:          execPath,
		backupPath:        filepath.Join(dir, base+".old"),
		newPath:           filepath.Join(dir, base+".new"),
		securityValidator: NewSecurityValidator(),
	}

	// Validate all paths for security
	if err := updater.securityValidator.ValidateFilePath(execPath); err != nil {
		return nil, fmt.Errorf("executable path validation failed: %w", err)
	}
	if err := updater.securityValidator.ValidateFilePath(updater.backupPath); err != nil {
		return nil, fmt.Errorf("backup path validation failed: %w", err)
	}
	if err := updater.securityValidator.ValidateFilePath(updater.newPath); err != nil {
		return nil, fmt.Errorf("new path validation failed: %w", err)
	}

	return updater, nil
}

// Backup 备份当前可执行文�?
func (u *Updater) Backup() error {
	// 检查当前文件是否存�?
	if _, err := os.Stat(u.execPath); err != nil {
		return fmt.Errorf("current executable not found: %w", err)
	}

	// 如果已存在旧的备份文件，先删�?
	if _, err := os.Stat(u.backupPath); err == nil {
		if err := os.Remove(u.backupPath); err != nil {
			return fmt.Errorf("failed to remove old backup: %w", err)
		}
	}

	// 使用 rename 进行备份（原子操作）
	if err := os.Rename(u.execPath, u.backupPath); err != nil {
		return fmt.Errorf("failed to backup current executable: %w", err)
	}

	return nil
}

// Replace 替换可执行文件（原子操作�?
func (u *Updater) Replace() error {
	// 检查新文件是否存在
	if _, err := os.Stat(u.newPath); err != nil {
		return fmt.Errorf("new executable not found: %w", err)
	}

	// Security validation: validate downloaded file (basic checks only, since it's a temp file)
	if err := u.securityValidator.ValidateDownloadedFile(u.newPath); err != nil {
		return fmt.Errorf("security validation failed for new file: %w", err)
	}

	// 检查备份文件是否存�?
	if _, err := os.Stat(u.backupPath); err != nil {
		return fmt.Errorf("backup file not found, cannot proceed with replace: %w", err)
	}

	// 使用 rename 替换文件（原子操作）
	if err := os.Rename(u.newPath, u.execPath); err != nil {
		// 替换失败，尝试回�?
		if rollbackErr := os.Rename(u.backupPath, u.execPath); rollbackErr != nil {
			return fmt.Errorf("failed to replace and rollback failed: replace error: %w, rollback error: %v", err, rollbackErr)
		}
		return fmt.Errorf("failed to replace executable (rolled back): %w", err)
	}

	// 设置可执行权限（Unix-like 系统�?
	if runtime.GOOS != "windows" {
		if err := os.Chmod(u.execPath, 0755); err != nil {
			// 权限设置失败，回�?
			os.Remove(u.execPath)
			os.Rename(u.backupPath, u.execPath)
			return fmt.Errorf("failed to set executable permission (rolled back): %w", err)
		}
	}

	// After rename, validate the final file has correct permissions
	if err := u.securityValidator.ValidateFilePermissions(u.execPath); err != nil {
		// Permission validation failed, rollback
		os.Remove(u.execPath)
		os.Rename(u.backupPath, u.execPath)
		return fmt.Errorf("final file permission validation failed (rolled back): %w", err)
	}

	return nil
}

// Rollback 回滚到备份版�?
func (u *Updater) Rollback() error {
	// 检查备份文件是否存�?
	if _, err := os.Stat(u.backupPath); err != nil {
		return fmt.Errorf("backup file not found: %w", err)
	}

	// 如果当前文件存在（可能是损坏的新版本），先删�?
	if _, err := os.Stat(u.execPath); err == nil {
		if err := os.Remove(u.execPath); err != nil {
			return fmt.Errorf("failed to remove current executable: %w", err)
		}
	}

	// 使用 rename 恢复备份（原子操作）
	if err := os.Rename(u.backupPath, u.execPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// Restart 重启 Agent
func (u *Updater) Restart() error {
	// 获取当前进程的参�?
	args := os.Args[1:]

	// 创建新进�?
	cmd := exec.Command(u.execPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// 启动新进�?
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start new process: %w", err)
	}

	// 等待新进程启�?
	time.Sleep(2 * time.Second)

	// 退出当前进�?
	os.Exit(0)

	return nil
}

// CleanupBackup 清理备份文件
func (u *Updater) CleanupBackup() error {
	// 检查备份文件是否存�?
	if _, err := os.Stat(u.backupPath); err != nil {
		if os.IsNotExist(err) {
			return nil // 备份文件不存在，无需清理
		}
		return fmt.Errorf("failed to check backup file: %w", err)
	}

	// 删除备份文件
	if err := os.Remove(u.backupPath); err != nil {
		return fmt.Errorf("failed to remove backup file: %w", err)
	}

	return nil
}

// CleanupNew 清理新版本文件（用于更新失败后的清理�?
func (u *Updater) CleanupNew() error {
	// 检查新文件是否存在
	if _, err := os.Stat(u.newPath); err != nil {
		if os.IsNotExist(err) {
			return nil // 新文件不存在，无需清理
		}
		return fmt.Errorf("failed to check new file: %w", err)
	}

	// 删除新文�?
	if err := os.Remove(u.newPath); err != nil {
		return fmt.Errorf("failed to remove new file: %w", err)
	}

	return nil
}

// GetNewPath 获取新版本文件路径（供下载器使用�?
func (u *Updater) GetNewPath() string {
	return u.newPath
}
