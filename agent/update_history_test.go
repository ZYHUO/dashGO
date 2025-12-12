package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewUpdateHistory(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "update-history.json")

	// 测试创建新的更新历史管理�?
	uh, err := NewUpdateHistory(historyFile)
	if err != nil {
		t.Fatalf("NewUpdateHistory failed: %v", err)
	}

	if uh == nil {
		t.Fatal("UpdateHistory is nil")
	}

	if uh.filePath != historyFile {
		t.Errorf("Expected filePath %s, got %s", historyFile, uh.filePath)
	}

	if len(uh.records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(uh.records))
	}
}

func TestAddRecord(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "update-history.json")

	uh, err := NewUpdateHistory(historyFile)
	if err != nil {
		t.Fatalf("NewUpdateHistory failed: %v", err)
	}

	// 添加一条记�?
	record := UpdateRecord{
		FromVersion:  "v1.0.0",
		ToVersion:    "v1.1.0",
		Status:       "success",
		ErrorMessage: "",
	}

	err = uh.AddRecord(record)
	if err != nil {
		t.Fatalf("AddRecord failed: %v", err)
	}

	// 验证记录已添�?
	if len(uh.records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(uh.records))
	}

	// 验证时间戳已自动设置
	if uh.records[0].Timestamp.IsZero() {
		t.Error("Timestamp should be set automatically")
	}

	// 验证文件已创�?
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		t.Error("History file should be created")
	}
}

func TestAddRecordWithTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "update-history.json")

	uh, err := NewUpdateHistory(historyFile)
	if err != nil {
		t.Fatalf("NewUpdateHistory failed: %v", err)
	}

	// 添加一条带时间戳的记录
	timestamp := time.Date(2024, 12, 11, 10, 0, 0, 0, time.UTC)
	record := UpdateRecord{
		Timestamp:    timestamp,
		FromVersion:  "v1.0.0",
		ToVersion:    "v1.1.0",
		Status:       "success",
		ErrorMessage: "",
	}

	err = uh.AddRecord(record)
	if err != nil {
		t.Fatalf("AddRecord failed: %v", err)
	}

	// 验证时间戳保持不�?
	if !uh.records[0].Timestamp.Equal(timestamp) {
		t.Errorf("Expected timestamp %v, got %v", timestamp, uh.records[0].Timestamp)
	}
}

func TestGetRecords(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "update-history.json")

	uh, err := NewUpdateHistory(historyFile)
	if err != nil {
		t.Fatalf("NewUpdateHistory failed: %v", err)
	}

	// 添加多条记录
	records := []UpdateRecord{
		{
			Timestamp:   time.Now().Add(-3 * time.Hour),
			FromVersion: "v1.0.0",
			ToVersion:   "v1.1.0",
			Status:      "success",
		},
		{
			Timestamp:   time.Now().Add(-2 * time.Hour),
			FromVersion: "v1.1.0",
			ToVersion:   "v1.2.0",
			Status:      "failed",
		},
		{
			Timestamp:   time.Now().Add(-1 * time.Hour),
			FromVersion: "v1.1.0",
			ToVersion:   "v1.2.0",
			Status:      "success",
		},
	}

	for _, record := range records {
		if err := uh.AddRecord(record); err != nil {
			t.Fatalf("AddRecord failed: %v", err)
		}
	}

	// 测试获取所有记�?
	allRecords := uh.GetRecords(0)
	if len(allRecords) != 3 {
		t.Errorf("Expected 3 records, got %d", len(allRecords))
	}

	// 验证记录按时间降序排列（最新的在前�?
	for i := 0; i < len(allRecords)-1; i++ {
		if allRecords[i].Timestamp.Before(allRecords[i+1].Timestamp) {
			t.Error("Records should be sorted by timestamp in descending order")
		}
	}

	// 测试获取限制数量的记�?
	limitedRecords := uh.GetRecords(2)
	if len(limitedRecords) != 2 {
		t.Errorf("Expected 2 records, got %d", len(limitedRecords))
	}

	// 验证返回的是最新的记录
	if limitedRecords[0].ToVersion != "v1.2.0" {
		t.Errorf("Expected first record to be v1.2.0, got %s", limitedRecords[0].ToVersion)
	}
}

func TestGetRecordsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "update-history.json")

	uh, err := NewUpdateHistory(historyFile)
	if err != nil {
		t.Fatalf("NewUpdateHistory failed: %v", err)
	}

	// 测试空记录列�?
	records := uh.GetRecords(10)
	if len(records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(records))
	}
}

func TestCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "update-history.json")

	uh, err := NewUpdateHistory(historyFile)
	if err != nil {
		t.Fatalf("NewUpdateHistory failed: %v", err)
	}

	// 添加不同时间的记�?
	now := time.Now()
	records := []UpdateRecord{
		{
			Timestamp:   now.AddDate(0, 0, -40), // 40 天前
			FromVersion: "v1.0.0",
			ToVersion:   "v1.1.0",
			Status:      "success",
		},
		{
			Timestamp:   now.AddDate(0, 0, -20), // 20 天前
			FromVersion: "v1.1.0",
			ToVersion:   "v1.2.0",
			Status:      "success",
		},
		{
			Timestamp:   now.AddDate(0, 0, -5), // 5 天前
			FromVersion: "v1.2.0",
			ToVersion:   "v1.3.0",
			Status:      "success",
		},
	}

	for _, record := range records {
		if err := uh.AddRecord(record); err != nil {
			t.Fatalf("AddRecord failed: %v", err)
		}
	}

	// 清理 30 天前的记�?
	err = uh.Cleanup(30)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// 验证只保留了 30 天内的记�?
	if len(uh.records) != 2 {
		t.Errorf("Expected 2 records after cleanup, got %d", len(uh.records))
	}

	// 验证保留的是正确的记�?
	for _, record := range uh.records {
		if record.FromVersion == "v1.0.0" {
			t.Error("Old record (40 days) should be removed")
		}
	}
}

func TestCleanupInvalidDays(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "update-history.json")

	uh, err := NewUpdateHistory(historyFile)
	if err != nil {
		t.Fatalf("NewUpdateHistory failed: %v", err)
	}

	// 测试无效的天�?
	err = uh.Cleanup(0)
	if err == nil {
		t.Error("Cleanup should fail with days = 0")
	}

	err = uh.Cleanup(-1)
	if err == nil {
		t.Error("Cleanup should fail with negative days")
	}
}

func TestPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "update-history.json")

	// 创建第一个实例并添加记录
	uh1, err := NewUpdateHistory(historyFile)
	if err != nil {
		t.Fatalf("NewUpdateHistory failed: %v", err)
	}

	record := UpdateRecord{
		Timestamp:   time.Now(),
		FromVersion: "v1.0.0",
		ToVersion:   "v1.1.0",
		Status:      "success",
	}

	err = uh1.AddRecord(record)
	if err != nil {
		t.Fatalf("AddRecord failed: %v", err)
	}

	// 创建第二个实例，应该能加载之前的记录
	uh2, err := NewUpdateHistory(historyFile)
	if err != nil {
		t.Fatalf("NewUpdateHistory failed: %v", err)
	}

	if len(uh2.records) != 1 {
		t.Errorf("Expected 1 record after reload, got %d", len(uh2.records))
	}

	if uh2.records[0].FromVersion != "v1.0.0" {
		t.Errorf("Expected FromVersion v1.0.0, got %s", uh2.records[0].FromVersion)
	}
}

func TestMultipleRecords(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "update-history.json")

	uh, err := NewUpdateHistory(historyFile)
	if err != nil {
		t.Fatalf("NewUpdateHistory failed: %v", err)
	}

	// 添加多条不同状态的记录
	records := []UpdateRecord{
		{
			FromVersion: "v1.0.0",
			ToVersion:   "v1.1.0",
			Status:      "success",
		},
		{
			FromVersion:  "v1.1.0",
			ToVersion:    "v1.2.0",
			Status:       "failed",
			ErrorMessage: "download failed: connection timeout",
		},
		{
			FromVersion: "v1.1.0",
			ToVersion:   "v1.2.0",
			Status:      "success",
		},
		{
			FromVersion:  "v1.2.0",
			ToVersion:    "v1.3.0",
			Status:       "rollback",
			ErrorMessage: "verification failed",
		},
	}

	for _, record := range records {
		if err := uh.AddRecord(record); err != nil {
			t.Fatalf("AddRecord failed: %v", err)
		}
	}

	// 验证所有记录都已添�?
	if len(uh.records) != 4 {
		t.Errorf("Expected 4 records, got %d", len(uh.records))
	}

	// 验证失败记录包含错误信息
	allRecords := uh.GetRecords(0)
	failedCount := 0
	for _, record := range allRecords {
		if record.Status == "failed" || record.Status == "rollback" {
			failedCount++
			if record.ErrorMessage == "" {
				t.Error("Failed/rollback record should have error message")
			}
		}
	}

	if failedCount != 2 {
		t.Errorf("Expected 2 failed/rollback records, got %d", failedCount)
	}
}
