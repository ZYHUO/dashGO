package repository

import (
	"dashgo/internal/model"

	"gorm.io/gorm"
)

// HostRepository 主机仓库
type HostRepository struct {
	db *gorm.DB
}

func NewHostRepository(db *gorm.DB) *HostRepository {
	return &HostRepository{db: db}
}

func (r *HostRepository) Create(host *model.Host) error {
	return r.db.Create(host).Error
}

func (r *HostRepository) Update(host *model.Host) error {
	return r.db.Save(host).Error
}

func (r *HostRepository) Delete(id int64) error {
	return r.db.Delete(&model.Host{}, id).Error
}

func (r *HostRepository) FindByID(id int64) (*model.Host, error) {
	var host model.Host
	err := r.db.First(&host, id).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func (r *HostRepository) FindByToken(token string) (*model.Host, error) {
	var host model.Host
	err := r.db.Where("token = ?", token).First(&host).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func (r *HostRepository) GetAll() ([]model.Host, error) {
	var hosts []model.Host
	err := r.db.Order("created_at DESC").Find(&hosts).Error
	return hosts, err
}

// ServerNodeRepository 节点仓库
type ServerNodeRepository struct {
	db *gorm.DB
}

func NewServerNodeRepository(db *gorm.DB) *ServerNodeRepository {
	return &ServerNodeRepository{db: db}
}

func (r *ServerNodeRepository) Create(node *model.ServerNode) error {
	return r.db.Create(node).Error
}

func (r *ServerNodeRepository) Update(node *model.ServerNode) error {
	return r.db.Save(node).Error
}

func (r *ServerNodeRepository) Delete(id int64) error {
	return r.db.Delete(&model.ServerNode{}, id).Error
}

func (r *ServerNodeRepository) DeleteByHostID(hostID int64) error {
	return r.db.Where("host_id = ?", hostID).Delete(&model.ServerNode{}).Error
}

func (r *ServerNodeRepository) FindByID(id int64) (*model.ServerNode, error) {
	var node model.ServerNode
	err := r.db.First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *ServerNodeRepository) FindByHostID(hostID int64) ([]model.ServerNode, error) {
	var nodes []model.ServerNode
	err := r.db.Where("host_id = ?", hostID).Order("sort ASC, created_at ASC").Find(&nodes).Error
	return nodes, err
}

func (r *ServerNodeRepository) GetAll() ([]model.ServerNode, error) {
	var nodes []model.ServerNode
	err := r.db.Order("created_at DESC").Find(&nodes).Error
	return nodes, err
}
