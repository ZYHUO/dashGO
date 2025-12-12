package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// UpdateRecord 更新记录
type UpdateRecord struct {
	Timestamp    time.Time `json:"timestamp"`
	FromVersion  string    `json:"from_version"`
	ToVersion    string    `json:"to_version"`
	Status       string    `json:"status"` // "success", "failed", "rollback"
	ErrorMessage string    `json:"error_message,omitempty"`
}

// UpdateHistory 更新历史管理�?
type UpdateHistory struct {
	filePath string
	records  []UpdateRecord
}

// updateHistoryFile 更新历史文件结构
type updateHistoryFile struct {
	Records []UpdateRecord `json:"records"`
}

// NewUpdateHistory 创建更新历史管理�?
func NewUpdateHistory(filePath string) (*UpdateHistory, error) {
	uh := &UpdateHistory{
		filePath: filePath,
		records:  make([]UpdateRecord, 0),
	}

	// 尝试加载现有记录
	if err := uh.load(); err != nil {
		// 如果文件不存在，这是正常�?
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load update history: %w", err)
		}
	}

	return uh, nil
}

// AddRecord 添加更新记录
func (uh *UpdateHistory) AddRecord(record UpdateRecord) error {
	// 如果记录没有时间戳，使用当前时间
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	// 添加到记录列�?
	uh.records = append(uh.records, record)

	// 持久化到文件
	if err := uh.save(); err != nil {
		return fmt.Errorf("failed to save update history: %w", err)
	}

	return nil
}

// GetRecords 获取更新记录
// limit: 返回的最大记录数�? 表示返回所有记�?
func (uh *UpdateHistory) GetRecords(limit int) []UpdateRecord {
	// 按时间戳降序排序（最新的在前�?
	sorted := make([]UpdateRecord, len(uh.records))
	copy(sorted, uh.records)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.After(sorted[j].Timestamp)
	})

	// 如果 limit �?0 或大于记录数，返回所有记�?
	if limit <= 0 || limit > len(sorted) {
		return sorted
	}

	// 返回�?limit 条记�?
	return sorted[:limit]
}

// Cleanup 清理旧记�?
// days: 保留最近多少天的记�?
func (uh *UpdateHistory) Cleanup(days int) error {
	if days <= 0 {
		return fmt.Errorf("days must be positive")
	}

	// 计算截止时间
	cutoff := time.Now().AddDate(0, 0, -days)

	// 过滤出需要保留的记录
	kept := make([]UpdateRecord, 0)
	removed := 0
	for _, record := range uh.records {
		if record.Timestamp.After(cutoff) {
			kept = append(kept, record)
		} else {
			removed++
		}
	}

	// 如果没有记录被删除，直接返回
	if removed == 0 {
		return nil
	}

	// 更新记录列表
	uh.records = kept

	// 持久化到文件
	if err := uh.save(); err != nil {
		return fmt.Errorf("failed to save after cleanup: %w", err)
	}

	return nil
}

// load 从文件加载更新历�?
func (uh *UpdateHistory) load() error {
	// 读取文件
	data, err := os.ReadFile(uh.filePath)
	if err != nil {
		return err
	}

	// 解析 JSON
	var historyFile updateHistoryFile
	if err := json.Unmarshal(data, &historyFile); err != nil {
		return fmt.Errorf("failed to parse update history file: %w", err)
	}

	uh.records = historyFile.Records
	return nil
}

// save 保存更新历史到文�?
func (uh *UpdateHistory) save() error {
	// 确保目录存在
	dir := filepath.Dir(uh.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 构建文件结构
	historyFile := updateHistoryFile{
		Records: uh.records,
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(historyFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal update history: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(uh.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write update history file: %w", err)
	}

	return nil
}
