package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// setupTestUpdater 创建测试用的 Updater
func setupTestUpdater(t *testing.T) (*Updater, string) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// 创建模拟的可执行文件
	execPath := filepath.Join(tmpDir, "test-agent")
	if runtime.GOOS == "windows" {
		execPath += ".exe"
	}

	// 写入测试内容
	if err := os.WriteFile(execPath, []byte("original version"), 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create test executable: %v", err)
	}

	base := filepath.Base(execPath)
	updater := &Updater{
		execPath:   execPath,
		backupPath: filepath.Join(tmpDir, base+".old"),
		newPath:    filepath.Join(tmpDir, base+".new"),
	}

	return updater, tmpDir
}

// TestNewUpdater 测试创建 Updater
func TestNewUpdater(t *testing.T) {
	updater, err := NewUpdater()
	if err != nil {
		t.Fatalf("NewUpdater failed: %v", err)
	}

	if updater.execPath == "" {
		t.Error("execPath should not be empty")
	}

	if updater.backupPath == "" {
		t.Error("backupPath should not be empty")
	}

	if updater.newPath == "" {
		t.Error("newPath should not be empty")
	}

	// 验证路径格式
	if !filepath.IsAbs(updater.execPath) {
		t.Error("execPath should be absolute")
	}
}

// TestBackup 测试备份功能
func TestBackup(t *testing.T) {
	updater, tmpDir := setupTestUpdater(t)
	defer os.RemoveAll(tmpDir)

	// 执行备份
	if err := updater.Backup(); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// 验证备份文件存在
	if _, err := os.Stat(updater.backupPath); err != nil {
		t.Errorf("Backup file not found: %v", err)
	}

	// 验证原文件不存在
	if _, err := os.Stat(updater.execPath); err == nil {
		t.Error("Original file should not exist after backup")
	}

	// 验证备份文件内容
	content, err := os.ReadFile(updater.backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}
	if string(content) != "original version" {
		t.Errorf("Backup content mismatch: got %s, want 'original version'", string(content))
	}
}

// TestBackupRemovesOldBackup 测试备份时删除旧备份
func TestBackupRemovesOldBackup(t *testing.T) {
	updater, tmpDir := setupTestUpdater(t)
	defer os.RemoveAll(tmpDir)

	// 创建旧的备份文件
	if err := os.WriteFile(updater.backupPath, []byte("old backup"), 0644); err != nil {
		t.Fatalf("Failed to create old backup: %v", err)
	}

	// 执行备份
	if err := updater.Backup(); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// 验证备份文件内容是新告
	content, err := os.ReadFile(updater.backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}
	if string(content) != "original version" {
		t.Errorf("Backup should contain new content, got: %s", string(content))
	}
}

// TestReplace 测试替换功能
func TestReplace(t *testing.T) {
	updater, tmpDir := setupTestUpdater(t)
	defer os.RemoveAll(tmpDir)

	// 先备告
	if err := updater.Backup(); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// 创建新版本文告
	if err := os.WriteFile(updater.newPath, []byte("new version"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	// 执行替换
	if err := updater.Replace(); err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	// 验证新文件已替换到原位置
	content, err := os.ReadFile(updater.execPath)
	if err != nil {
		t.Fatalf("Failed to read executable: %v", err)
	}
	if string(content) != "new version" {
		t.Errorf("Executable content mismatch: got %s, want 'new version'", string(content))
	}

	// 验证新文件已被移告
	if _, err := os.Stat(updater.newPath); err == nil {
		t.Error("New file should be removed after replace")
	}

	// 验证备份文件仍然存在
	if _, err := os.Stat(updater.backupPath); err != nil {
		t.Error("Backup file should still exist after replace")
	}

	// 验证可执行权限（Unix-like 系统告
	if runtime.GOOS != "windows" {
		info, err := os.Stat(updater.execPath)
		if err != nil {
			t.Fatalf("Failed to stat executable: %v", err)
		}
		mode := info.Mode()
		if mode&0111 == 0 {
			t.Error("Executable should have execute permission")
		}
	}
}

// TestReplaceWithoutBackup 测试没有备份时替换失告
func TestReplaceWithoutBackup(t *testing.T) {
	updater, tmpDir := setupTestUpdater(t)
	defer os.RemoveAll(tmpDir)

	// 创建新版本文告
	if err := os.WriteFile(updater.newPath, []byte("new version"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	// 尝试替换（应该失败，因为没有备份告
	if err := updater.Replace(); err == nil {
		t.Error("Replace should fail without backup")
	}
}

// TestReplaceWithoutNewFile 测试没有新文件时替换失败
func TestReplaceWithoutNewFile(t *testing.T) {
	updater, tmpDir := setupTestUpdater(t)
	defer os.RemoveAll(tmpDir)

	// 先备告
	if err := updater.Backup(); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// 尝试替换（应该失败，因为没有新文件）
	if err := updater.Replace(); err == nil {
		t.Error("Replace should fail without new file")
	}
}

// TestRollback 测试回滚功能
func TestRollback(t *testing.T) {
	updater, tmpDir := setupTestUpdater(t)
	defer os.RemoveAll(tmpDir)

	// 先备告
	if err := updater.Backup(); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// 创建一个损坏的新版告
	if err := os.WriteFile(updater.execPath, []byte("corrupted version"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// 执行回滚
	if err := updater.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// 验证原文件已恢复
	content, err := os.ReadFile(updater.execPath)
	if err != nil {
		t.Fatalf("Failed to read executable: %v", err)
	}
	if string(content) != "original version" {
		t.Errorf("Executable should be restored to original, got: %s", string(content))
	}

	// 验证备份文件已被移除
	if _, err := os.Stat(updater.backupPath); err == nil {
		t.Error("Backup file should be removed after rollback")
	}
}

// TestRollbackWithoutBackup 测试没有备份时回滚失告
func TestRollbackWithoutBackup(t *testing.T) {
	updater, tmpDir := setupTestUpdater(t)
	defer os.RemoveAll(tmpDir)

	// 尝试回滚（应该失败，因为没有备份告
	if err := updater.Rollback(); err == nil {
		t.Error("Rollback should fail without backup")
	}
}

// TestCleanupBackup 测试清理备份文件
func TestCleanupBackup(t *testing.T) {
	updater, tmpDir := setupTestUpdater(t)
	defer os.RemoveAll(tmpDir)

	// 创建备份文件
	if err := os.WriteFile(updater.backupPath, []byte("backup"), 0644); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// 清理备份
	if err := updater.CleanupBackup(); err != nil {
		t.Fatalf("CleanupBackup failed: %v", err)
	}

	// 验证备份文件已被删除
	if _, err := os.Stat(updater.backupPath); err == nil {
		t.Error("Backup file should be removed")
	}
}

// TestCleanupBackupWhenNotExists 测试清理不存在的备份文件
func TestCleanupBackupWhenNotExists(t *testing.T) {
	updater, tmpDir := setupTestUpdater(t)
	defer os.RemoveAll(tmpDir)

	// 清理不存在的备份（应该成功）
	if err := updater.CleanupBackup(); err != nil {
		t.Errorf("CleanupBackup should succeed when backup doesn't exist: %v", err)
	}
}

// TestCleanupNew 测试清理新版本文告
func TestCleanupNew(t *testing.T) {
	updater, tmpDir := setupTestUpdater(t)
	defer os.RemoveAll(tmpDir)

	// 创建新版本文告
	if err := os.WriteFile(updater.newPath, []byte("new version"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	// 清理新文告
	if err := updater.CleanupNew(); err != nil {
		t.Fatalf("CleanupNew failed: %v", err)
	}

	// 验证新文件已被删告
	if _, err := os.Stat(updater.newPath); err == nil {
		t.Error("New file should be removed")
	}
}

// TestCleanupNewWhenNotExists 测试清理不存在的新文告
func TestCleanupNewWhenNotExists(t *testing.T) {
	updater, tmpDir := setupTestUpdater(t)
	defer os.RemoveAll(tmpDir)

	// 清理不存在的新文件（应该成功告
	if err := updater.CleanupNew(); err != nil {
		t.Errorf("CleanupNew should succeed when new file doesn't exist: %v", err)
	}
}

// TestGetNewPath 测试获取新版本文件路告
func TestGetNewPath(t *testing.T) {
	updater, tmpDir := setupTestUpdater(t)
	defer os.RemoveAll(tmpDir)

	newPath := updater.GetNewPath()
	if newPath == "" {
		t.Error("GetNewPath should return non-empty path")
	}

	if newPath != updater.newPath {
		t.Errorf("GetNewPath mismatch: got %s, want %s", newPath, updater.newPath)
	}
}

// TestCompleteUpdateFlow 测试完整的更新流告
func TestCompleteUpdateFlow(t *testing.T) {
	updater, tmpDir := setupTestUpdater(t)
	defer os.RemoveAll(tmpDir)

	// 1. 备份当前版本
	if err := updater.Backup(); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// 2. 创建新版告
	if err := os.WriteFile(updater.newPath, []byte("new version"), 0755); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	// 3. 替换
	if err := updater.Replace(); err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	// 4. 验证新版告
	content, err := os.ReadFile(updater.execPath)
	if err != nil {
		t.Fatalf("Failed to read executable: %v", err)
	}
	if string(content) != "new version" {
		t.Errorf("Update failed: got %s, want 'new version'", string(content))
	}

	// 5. 清理备份
	if err := updater.CleanupBackup(); err != nil {
		t.Fatalf("CleanupBackup failed: %v", err)
	}

	// 验证备份已清告
	if _, err := os.Stat(updater.backupPath); err == nil {
		t.Error("Backup should be cleaned up")
	}
}

// TestFailedUpdateWithRollback 测试更新失败后的回滚
func TestFailedUpdateWithRollback(t *testing.T) {
	updater, tmpDir := setupTestUpdater(t)
	defer os.RemoveAll(tmpDir)

	// 1. 备份当前版本
	if err := updater.Backup(); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// 2. 模拟更新失败（创建损坏的文件告
	if err := os.WriteFile(updater.execPath, []byte("corrupted"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// 3. 回滚
	if err := updater.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// 4. 验证已恢复原版本
	content, err := os.ReadFile(updater.execPath)
	if err != nil {
		t.Fatalf("Failed to read executable: %v", err)
	}
	if string(content) != "original version" {
		t.Errorf("Rollback failed: got %s, want 'original version'", string(content))
	}
}
