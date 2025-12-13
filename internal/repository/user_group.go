package repository

import (
	"dashgo/internal/model"

	"gorm.io/gorm"
)

type UserGroupRepository struct {
	db *gorm.DB
}

func NewUserGroupRepository(db *gorm.DB) *UserGroupRepository {
	return &UserGroupRepository{db: db}
}

// Create 创建用户组
func (r *UserGroupRepository) Create(group *model.UserGroup) error {
	return r.db.Create(group).Error
}

// Update 更新用户组
func (r *UserGroupRepository) Update(group *model.UserGroup) error {
	return r.db.Save(group).Error
}

// Delete 删除用户组
func (r *UserGroupRepository) Delete(id int64) error {
	return r.db.Delete(&model.UserGroup{}, id).Error
}

// FindByID 根据ID查找用户组
func (r *UserGroupRepository) FindByID(id int64) (*model.UserGroup, error) {
	var group model.UserGroup
	err := r.db.First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetAll 获取所有用户组
func (r *UserGroupRepository) GetAll() ([]model.UserGroup, error) {
	var groups []model.UserGroup
	err := r.db.Order("sort ASC, id ASC").Find(&groups).Error
	return groups, err
}

// Count 统计用户组总数
func (r *UserGroupRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.UserGroup{}).Count(&count).Error
	return count, err
}

// GetByName 根据名称查找用户组
func (r *UserGroupRepository) GetByName(name string) (*model.UserGroup, error) {
	var group model.UserGroup
	err := r.db.Where("name = ?", name).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// AddServerToGroup 为用户组添加节点
func (r *UserGroupRepository) AddServerToGroup(groupID, serverID int64) error {
	group, err := r.FindByID(groupID)
	if err != nil {
		return err
	}

	// 检查是否已存在
	serverIDs := group.GetServerIDsAsInt64()
	for _, id := range serverIDs {
		if id == serverID {
			return nil // 已存在，不重复添加
		}
	}

	// 添加新节点
	group.ServerIDs = append(group.ServerIDs, serverID)
	return r.Update(group)
}

// RemoveServerFromGroup 从用户组移除节点
func (r *UserGroupRepository) RemoveServerFromGroup(groupID, serverID int64) error {
	group, err := r.FindByID(groupID)
	if err != nil {
		return err
	}

	// 过滤掉要删除的节点
	newServerIDs := make(model.JSONArray, 0)
	for _, v := range group.ServerIDs {
		if id, ok := v.(float64); ok && int64(id) != serverID {
			newServerIDs = append(newServerIDs, v)
		} else if id, ok := v.(int64); ok && id != serverID {
			newServerIDs = append(newServerIDs, v)
		}
	}

	group.ServerIDs = newServerIDs
	return r.Update(group)
}

// AddPlanToGroup 为用户组添加套餐
func (r *UserGroupRepository) AddPlanToGroup(groupID, planID int64) error {
	group, err := r.FindByID(groupID)
	if err != nil {
		return err
	}

	// 检查是否已存在
	planIDs := group.GetPlanIDsAsInt64()
	for _, id := range planIDs {
		if id == planID {
			return nil // 已存在，不重复添加
		}
	}

	// 添加新套餐
	group.PlanIDs = append(group.PlanIDs, planID)
	return r.Update(group)
}

// RemovePlanFromGroup 从用户组移除套餐
func (r *UserGroupRepository) RemovePlanFromGroup(groupID, planID int64) error {
	group, err := r.FindByID(groupID)
	if err != nil {
		return err
	}

	// 过滤掉要删除的套餐
	newPlanIDs := make(model.JSONArray, 0)
	for _, v := range group.PlanIDs {
		if id, ok := v.(float64); ok && int64(id) != planID {
			newPlanIDs = append(newPlanIDs, v)
		} else if id, ok := v.(int64); ok && id != planID {
			newPlanIDs = append(newPlanIDs, v)
		}
	}

	group.PlanIDs = newPlanIDs
	return r.Update(group)
}
