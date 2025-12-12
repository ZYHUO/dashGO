package service

import (
	"time"

	"dashgo/internal/model"
	"dashgo/internal/repository"
)

// ServerGroupService 用户组服�?
type ServerGroupService struct {
	groupRepo *repository.ServerGroupRepository
}

func NewServerGroupService(groupRepo *repository.ServerGroupRepository) *ServerGroupService {
	return &ServerGroupService{
		groupRepo: groupRepo,
	}
}

// GetAll 获取所有用户组
func (s *ServerGroupService) GetAll() ([]model.ServerGroup, error) {
	return s.groupRepo.GetAll()
}

// GetByID 根据 ID 获取用户�?
func (s *ServerGroupService) GetByID(id int64) (*model.ServerGroup, error) {
	return s.groupRepo.FindByID(id)
}

// Create 创建用户�?
func (s *ServerGroupService) Create(name string) (*model.ServerGroup, error) {
	group := &model.ServerGroup{
		Name:      name,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	if err := s.groupRepo.Create(group); err != nil {
		return nil, err
	}
	return group, nil
}

// Update 更新用户�?
func (s *ServerGroupService) Update(id int64, name string) error {
	group, err := s.groupRepo.FindByID(id)
	if err != nil {
		return err
	}
	group.Name = name
	group.UpdatedAt = time.Now().Unix()
	return s.groupRepo.Update(group)
}

// Delete 删除用户�?
func (s *ServerGroupService) Delete(id int64) error {
	return s.groupRepo.Delete(id)
}
