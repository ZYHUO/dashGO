package service

import (
	"fmt"
	"dashgo/internal/model"

	"gorm.io/gorm"
)

// AgentVersionService Agent 版本服务
type AgentVersionService struct {
	db *gorm.DB
}

// NewAgentVersionService 创建 Agent 版本服务
func NewAgentVersionService(db *gorm.DB) *AgentVersionService {
	return &AgentVersionService{db: db}
}

// GetLatestVersion 获取最新版�?
func (s *AgentVersionService) GetLatestVersion() (*model.AgentVersion, error) {
	var version model.AgentVersion
	err := s.db.Where("is_latest = ?", true).First(&version).Error
	if err != nil {
		// 如果没有设置最新版本，返回默认配置
		if err == gorm.ErrRecordNotFound {
			return &model.AgentVersion{
				Version:      "v1.0.0",
				DownloadURL:  "https://download.sharon.wiki/xboard-agent-linux-amd64",
				SHA256:       "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				FileSize:     6090936,
				Strategy:     "manual",
				ReleaseNotes: "XBoard Agent v1.0.0",
			}, nil
		}
		return nil, err
	}
	return &version, nil
}

// GetByVersion 根据版本号获�?
func (s *AgentVersionService) GetByVersion(version string) (*model.AgentVersion, error) {
	var v model.AgentVersion
	err := s.db.Where("version = ?", version).First(&v).Error
	return &v, err
}

// Create 创建版本
func (s *AgentVersionService) Create(version *model.AgentVersion) error {
	return s.db.Create(version).Error
}

// Update 更新版本
func (s *AgentVersionService) Update(version *model.AgentVersion) error {
	return s.db.Save(version).Error
}

// SetLatest 设置为最新版�?
func (s *AgentVersionService) SetLatest(versionID int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 取消所有版本的 latest 标记
		if err := tx.Model(&model.AgentVersion{}).Where("is_latest = ?", true).Update("is_latest", false).Error; err != nil {
			return err
		}
		// 设置指定版本�?latest
		return tx.Model(&model.AgentVersion{}).Where("id = ?", versionID).Update("is_latest", true).Error
	})
}

// List 获取版本列表
func (s *AgentVersionService) List(page, pageSize int) ([]model.AgentVersion, int64, error) {
	var versions []model.AgentVersion
	var total int64

	query := s.db.Model(&model.AgentVersion{})
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&versions).Error
	
	return versions, total, err
}

// Delete 删除版本
func (s *AgentVersionService) Delete(versionID int64) error {
	// 检查是否是最新版�?
	var version model.AgentVersion
	if err := s.db.First(&version, versionID).Error; err != nil {
		return err
	}
	if version.IsLatest {
		return fmt.Errorf("cannot delete the latest version")
	}
	return s.db.Delete(&model.AgentVersion{}, versionID).Error
}

// RecordUpdateLog 记录更新日志
func (s *AgentVersionService) RecordUpdateLog(log *model.AgentUpdateLog) error {
	return s.db.Create(log).Error
}

// GetUpdateLogs 获取更新日志
func (s *AgentVersionService) GetUpdateLogs(hostID int64, page, pageSize int) ([]model.AgentUpdateLog, int64, error) {
	var logs []model.AgentUpdateLog
	var total int64

	query := s.db.Model(&model.AgentUpdateLog{})
	if hostID > 0 {
		query = query.Where("host_id = ?", hostID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	
	return logs, total, err
}
