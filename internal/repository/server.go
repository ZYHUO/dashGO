package repository

import (
	"fmt"

	"xboard/internal/model"

	"gorm.io/gorm"
)

type ServerRepository struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

func (r *ServerRepository) Create(server *model.Server) error {
	return r.db.Create(server).Error
}

func (r *ServerRepository) Update(server *model.Server) error {
	return r.db.Save(server).Error
}

func (r *ServerRepository) Delete(id int64) error {
	return r.db.Delete(&model.Server{}, id).Error
}

func (r *ServerRepository) FindByID(id int64) (*model.Server, error) {
	var server model.Server
	err := r.db.First(&server, id).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (r *ServerRepository) FindByCode(serverType, code string) (*model.Server, error) {
	var server model.Server
	err := r.db.Where("type = ? AND code = ?", serverType, code).First(&server).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

// GetAllServers 获取所有服务器
func (r *ServerRepository) GetAllServers() ([]model.Server, error) {
	var servers []model.Server
	err := r.db.Order("sort ASC").Find(&servers).Error
	return servers, err
}

// GetAvailableServers 获取指定权限组的可用服务器
func (r *ServerRepository) GetAvailableServers(groupID int64) ([]model.Server, error) {
	var servers []model.Server
	// 使用 JSON_CONTAINS 查询包含指定 group_id 的服务器
	// JSON_CONTAINS 需要传入 JSON 格式的值
	groupIDJSON := fmt.Sprintf("[%d]", groupID)
	err := r.db.
		Where("(JSON_CONTAINS(group_ids, ?) OR group_ids IS NULL OR group_ids = '[]' OR group_ids = '' OR JSON_LENGTH(group_ids) = 0)", groupIDJSON).
		Where("`show` = ?", true).
		Order("sort ASC").
		Find(&servers).Error
	return servers, err
}

// GetPublicServers 获取所有公开的服务器（不限制用户组）
func (r *ServerRepository) GetPublicServers() ([]model.Server, error) {
	var servers []model.Server
	err := r.db.
		Where("`show` = ?", true).
		Order("sort ASC").
		Find(&servers).Error
	return servers, err
}

// GetServersByType 按类型获取服务器
func (r *ServerRepository) GetServersByType(serverType string) ([]model.Server, error) {
	var servers []model.Server
	err := r.db.Where("type = ?", serverType).Order("sort ASC").Find(&servers).Error
	return servers, err
}

func (r *ServerRepository) List(page, pageSize int) ([]model.Server, int64, error) {
	var servers []model.Server
	var total int64

	r.db.Model(&model.Server{}).Count(&total)
	err := r.db.Order("sort ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&servers).Error
	return servers, total, err
}

// Count 统计服务器总数
func (r *ServerRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Server{}).Count(&count).Error
	return count, err
}

// GetByHostID 获取绑定到指定主机的所有节点
func (r *ServerRepository) GetByHostID(hostID int64) ([]model.Server, error) {
	var servers []model.Server
	err := r.db.Where("host_id = ?", hostID).Order("sort ASC").Find(&servers).Error
	return servers, err
}

// UpdateHostID 更新节点的主机绑定
func (r *ServerRepository) UpdateHostID(serverID int64, hostID *int64) error {
	return r.db.Model(&model.Server{}).Where("id = ?", serverID).Update("host_id", hostID).Error
}

// UnbindFromHost 解除节点与主机的绑定
func (r *ServerRepository) UnbindFromHost(hostID int64) error {
	return r.db.Model(&model.Server{}).Where("host_id = ?", hostID).Update("host_id", nil).Error
}

// ServerGroup Repository
type ServerGroupRepository struct {
	db *gorm.DB
}

func NewServerGroupRepository(db *gorm.DB) *ServerGroupRepository {
	return &ServerGroupRepository{db: db}
}

func (r *ServerGroupRepository) FindByID(id int64) (*model.ServerGroup, error) {
	var group model.ServerGroup
	err := r.db.First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *ServerGroupRepository) GetAll() ([]model.ServerGroup, error) {
	var groups []model.ServerGroup
	err := r.db.Find(&groups).Error
	return groups, err
}

// ServerRoute Repository
type ServerRouteRepository struct {
	db *gorm.DB
}

func NewServerRouteRepository(db *gorm.DB) *ServerRouteRepository {
	return &ServerRouteRepository{db: db}
}

func (r *ServerRouteRepository) FindByIDs(ids []int64) ([]model.ServerRoute, error) {
	var routes []model.ServerRoute
	err := r.db.Where("id IN ?", ids).Find(&routes).Error
	return routes, err
}

func (r *ServerRouteRepository) GetAll() ([]model.ServerRoute, error) {
	var routes []model.ServerRoute
	err := r.db.Find(&routes).Error
	return routes, err
}
