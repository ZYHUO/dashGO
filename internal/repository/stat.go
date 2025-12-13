package repository

import (
	"dashgo/internal/model"
	"time"

	"gorm.io/gorm"
)

type StatRepository struct {
	db *gorm.DB
}

func NewStatRepository(db *gorm.DB) *StatRepository {
	return &StatRepository{db: db}
}

// RecordUserTraffic 记录用户流量统计
func (r *StatRepository) RecordUserTraffic(userID int64, serverRate float64, u, d int64, recordType string) error {
	now := time.Now()
	var recordAt int64
	switch recordType {
	case "d": // daily
		recordAt = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	case "m": // monthly
		recordAt = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()
	default:
		recordAt = now.Unix()
	}

	var stat model.StatUser
	err := r.db.Where("user_id = ? AND server_rate = ? AND record_at = ?", userID, serverRate, recordAt).First(&stat).Error
	if err == gorm.ErrRecordNotFound {
		stat = model.StatUser{
			UserID:     userID,
			ServerRate: serverRate,
			U:          u,
			D:          d,
			RecordType: recordType,
			RecordAt:   recordAt,
		}
		return r.db.Create(&stat).Error
	}
	if err != nil {
		return err
	}
	return r.db.Model(&stat).Updates(map[string]interface{}{
		"u": gorm.Expr("u + ?", u),
		"d": gorm.Expr("d + ?", d),
	}).Error
}

// RecordServerTraffic 记录节点流量统计
func (r *StatRepository) RecordServerTraffic(serverID int64, serverType string, u, d int64, recordType string) error {
	now := time.Now()
	var recordAt int64
	switch recordType {
	case "d": // daily
		recordAt = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	case "m": // monthly
		recordAt = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()
	default:
		recordAt = now.Unix()
	}

	var stat model.StatServer
	err := r.db.Where("server_id = ? AND server_type = ? AND record_at = ?", serverID, serverType, recordAt).First(&stat).Error
	if err == gorm.ErrRecordNotFound {
		stat = model.StatServer{
			ServerID:   serverID,
			ServerType: serverType,
			U:          u,
			D:          d,
			RecordType: recordType,
			RecordAt:   recordAt,
		}
		return r.db.Create(&stat).Error
	}
	if err != nil {
		return err
	}
	return r.db.Model(&stat).Updates(map[string]interface{}{
		"u": gorm.Expr("u + ?", u),
		"d": gorm.Expr("d + ?", d),
	}).Error
}

// GetUserStats 获取用户流量统计
func (r *StatRepository) GetUserStats(userID int64, startAt, endAt int64) ([]model.StatUser, error) {
	var stats []model.StatUser
	err := r.db.Where("user_id = ? AND record_at >= ? AND record_at <= ?", userID, startAt, endAt).Find(&stats).Error
	return stats, err
}

// GetServerStats 获取节点流量统计
func (r *StatRepository) GetServerStats(serverID int64, startAt, endAt int64) ([]model.StatServer, error) {
	var stats []model.StatServer
	err := r.db.Where("server_id = ? AND record_at >= ? AND record_at <= ?", serverID, startAt, endAt).Find(&stats).Error
	return stats, err
}

// CreateOrUpdateStat 创建或更新统计
func (r *StatRepository) CreateOrUpdateStat(stat *model.Stat) error {
	var existing model.Stat
	err := r.db.Where("record_at = ? AND record_type = ?", stat.RecordAt, stat.RecordType).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(stat).Error
	}
	if err != nil {
		return err
	}
	stat.ID = existing.ID
	return r.db.Save(stat).Error
}

// GetOrderStats 获取订单统计
func (r *StatRepository) GetOrderStats(startAt, endAt int64) ([]model.Stat, error) {
	var stats []model.Stat
	err := r.db.Where("record_at >= ? AND record_at <= ? AND record_type = ?", startAt, endAt, "d").
		Order("record_at ASC").
		Find(&stats).Error
	return stats, err
}

// GetServerTrafficStats 获取服务器流量统计
func (r *StatRepository) GetServerTrafficStats(startAt, endAt int64) ([]model.StatServer, error) {
	var stats []model.StatServer
	err := r.db.Where("record_at >= ? AND record_at <= ?", startAt, endAt).
		Order("record_at ASC").
		Find(&stats).Error
	return stats, err
}

// GetServerRanking 获取服务器排名
func (r *StatRepository) GetServerRanking(limit int) ([]model.StatServer, error) {
	var stats []model.StatServer
	err := r.db.Model(&model.StatServer{}).
		Select("server_id, SUM(u) as u, SUM(d) as d").
		Group("server_id").
		Order("(SUM(u) + SUM(d)) DESC").
		Limit(limit).
		Find(&stats).Error
	return stats, err
}

// GetUserRanking 获取用户排名
func (r *StatRepository) GetUserRanking(limit int) ([]model.StatUser, error) {
	var stats []model.StatUser
	err := r.db.Model(&model.StatUser{}).
		Select("user_id, SUM(u) as u, SUM(d) as d").
		Group("user_id").
		Order("(SUM(u) + SUM(d)) DESC").
		Limit(limit).
		Find(&stats).Error
	return stats, err
}

// GetTotalTraffic 获取总流量
func (r *StatRepository) GetTotalTraffic(startAt, endAt int64) (int64, error) {
	var total int64
	err := r.db.Model(&model.StatServer{}).
		Where("record_at >= ? AND record_at <= ?", startAt, endAt).
		Select("COALESCE(SUM(u + d), 0)").
		Scan(&total).Error
	return total, err
}

// GetServerTotalTraffic 获取节点总流量
func (r *StatRepository) GetServerTotalTraffic(serverID int64) (model.StatServer, error) {
	var stat model.StatServer
	err := r.db.Model(&model.StatServer{}).
		Select("server_id, COALESCE(SUM(u), 0) as u, COALESCE(SUM(d), 0) as d").
		Where("server_id = ?", serverID).
		Group("server_id").
		First(&stat).Error
	if err == gorm.ErrRecordNotFound {
		return model.StatServer{ServerID: serverID, U: 0, D: 0}, nil
	}
	return stat, err
}

// CreateServerLog 创建流量日志
func (r *StatRepository) CreateServerLog(log *model.ServerLog) error {
	return r.db.Create(log).Error
}

// DeleteOldServerLogs 删除旧的流量日志
func (r *StatRepository) DeleteOldServerLogs(beforeTime int64) (int64, error) {
	result := r.db.Where("created_at < ?", beforeTime).Delete(&model.ServerLog{})
	return result.RowsAffected, result.Error
}

// DeleteOldUserStats 删除旧的用户统计（日统计）
func (r *StatRepository) DeleteOldUserStats(beforeTime int64) (int64, error) {
	result := r.db.Where("record_type = ? AND record_at < ?", "d", beforeTime).Delete(&model.StatUser{})
	return result.RowsAffected, result.Error
}

// DeleteOldServerStats 删除旧的节点统计（日统计）
func (r *StatRepository) DeleteOldServerStats(beforeTime int64) (int64, error) {
	result := r.db.Where("record_type = ? AND record_at < ?", "d", beforeTime).Delete(&model.StatServer{})
	return result.RowsAffected, result.Error
}
