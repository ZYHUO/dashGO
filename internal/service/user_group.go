package service

import (
	"errors"
	"time"

	"dashgo/internal/model"
	"dashgo/internal/repository"
)

type UserGroupService struct {
	groupRepo     *repository.UserGroupRepository
	serverRepo    *repository.ServerRepository
	planRepo      *repository.PlanRepository
	userRepo      *repository.UserRepository
	serverService *ServerService
}

func NewUserGroupService(
	groupRepo *repository.UserGroupRepository,
	serverRepo *repository.ServerRepository,
	planRepo *repository.PlanRepository,
	userRepo *repository.UserRepository,
) *UserGroupService {
	return &UserGroupService{
		groupRepo:  groupRepo,
		serverRepo: serverRepo,
		planRepo:   planRepo,
		userRepo:   userRepo,
	}
}

// SetServerService 设置 ServerService（用于构�?ServerInfo�?
func (s *UserGroupService) SetServerService(serverService *ServerService) {
	s.serverService = serverService
}

// Create 创建用户�?
func (s *UserGroupService) Create(group *model.UserGroup) error {
	group.CreatedAt = time.Now().Unix()
	group.UpdatedAt = time.Now().Unix()
	return s.groupRepo.Create(group)
}

// Update 更新用户�?
func (s *UserGroupService) Update(group *model.UserGroup) error {
	group.UpdatedAt = time.Now().Unix()
	return s.groupRepo.Update(group)
}

// Delete 删除用户�?
func (s *UserGroupService) Delete(id int64) error {
	// 检查是否有用户使用该组
	count, err := s.userRepo.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		// 这里应该检查具体有多少用户在这个组，简化处�?
		// 实际应该添加 CountByGroupID 方法
		return errors.New("该用户组下还有用户，无法删除")
	}
	return s.groupRepo.Delete(id)
}

// GetByID 根据ID获取用户�?
func (s *UserGroupService) GetByID(id int64) (*model.UserGroup, error) {
	return s.groupRepo.FindByID(id)
}

// GetAll 获取所有用户组
func (s *UserGroupService) GetAll() ([]model.UserGroup, error) {
	return s.groupRepo.GetAll()
}

// GetGroupInfo 获取用户组详细信息（包含节点和套餐列表）
func (s *UserGroupService) GetGroupInfo(group *model.UserGroup) map[string]interface{} {
	info := map[string]interface{}{
		"id":          group.ID,
		"name":        group.Name,
		"description": group.Description,
		"sort":        group.Sort,
		"created_at":  group.CreatedAt,
		"updated_at":  group.UpdatedAt,
		// 注意：不再返�?default_transfer_enable 等字�?
		// 流量、速度、设备限制应该由套餐决定
	}

	// 获取节点列表
	serverIDs := group.GetServerIDsAsInt64()
	servers := make([]map[string]interface{}, 0)
	for _, serverID := range serverIDs {
		server, err := s.serverRepo.FindByID(serverID)
		if err == nil {
			servers = append(servers, map[string]interface{}{
				"id":   server.ID,
				"name": server.Name,
				"type": server.Type,
				"host": server.Host,
			})
		}
	}
	info["servers"] = servers
	info["server_ids"] = serverIDs

	// 获取套餐列表
	planIDs := group.GetPlanIDsAsInt64()
	plans := make([]map[string]interface{}, 0)
	for _, planID := range planIDs {
		plan, err := s.planRepo.FindByID(planID)
		if err == nil {
			plans = append(plans, map[string]interface{}{
				"id":              plan.ID,
				"name":            plan.Name,
				"transfer_enable": plan.TransferEnable,
			})
		}
	}
	info["plans"] = plans
	info["plan_ids"] = planIDs

	return info
}

// AddServerToGroup 为用户组添加节点
func (s *UserGroupService) AddServerToGroup(groupID, serverID int64) error {
	// 验证节点是否存在
	_, err := s.serverRepo.FindByID(serverID)
	if err != nil {
		return errors.New("节点不存�?)
	}
	return s.groupRepo.AddServerToGroup(groupID, serverID)
}

// RemoveServerFromGroup 从用户组移除节点
func (s *UserGroupService) RemoveServerFromGroup(groupID, serverID int64) error {
	return s.groupRepo.RemoveServerFromGroup(groupID, serverID)
}

// AddPlanToGroup 为用户组添加套餐
func (s *UserGroupService) AddPlanToGroup(groupID, planID int64) error {
	// 验证套餐是否存在
	_, err := s.planRepo.FindByID(planID)
	if err != nil {
		return errors.New("套餐不存�?)
	}
	return s.groupRepo.AddPlanToGroup(groupID, planID)
}

// RemovePlanFromGroup 从用户组移除套餐
func (s *UserGroupService) RemovePlanFromGroup(groupID, planID int64) error {
	return s.groupRepo.RemovePlanFromGroup(groupID, planID)
}

// SetServersForGroup 设置用户组的节点列表（覆盖）
func (s *UserGroupService) SetServersForGroup(groupID int64, serverIDs []int64) error {
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return err
	}

	// 转换�?JSONArray
	newServerIDs := make(model.JSONArray, len(serverIDs))
	for i, id := range serverIDs {
		newServerIDs[i] = id
	}

	group.ServerIDs = newServerIDs
	return s.groupRepo.Update(group)
}

// SetPlansForGroup 设置用户组的套餐列表（覆盖）
func (s *UserGroupService) SetPlansForGroup(groupID int64, planIDs []int64) error {
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return err
	}

	// 转换�?JSONArray
	newPlanIDs := make(model.JSONArray, len(planIDs))
	for i, id := range planIDs {
		newPlanIDs[i] = id
	}

	group.PlanIDs = newPlanIDs
	return s.groupRepo.Update(group)
}

// GetAvailableServersForUser 获取用户可访问的节点列表
func (s *UserGroupService) GetAvailableServersForUser(user *model.User) ([]ServerInfo, error) {
	if user.GroupID == nil || *user.GroupID == 0 {
		// 没有用户组，返回空列�?
		return []ServerInfo{}, nil
	}

	group, err := s.groupRepo.FindByID(*user.GroupID)
	if err != nil {
		return nil, err
	}

	serverIDs := group.GetServerIDsAsInt64()
	if len(serverIDs) == 0 {
		return []ServerInfo{}, nil
	}

	// 获取节点列表并构�?ServerInfo
	servers := make([]ServerInfo, 0)
	for _, serverID := range serverIDs {
		server, err := s.serverRepo.FindByID(serverID)
		if err == nil && server.Show {
			// 使用 ServerService 构建 ServerInfo
			if s.serverService != nil {
				serverInfo := s.serverService.BuildServerInfo(server, user)
				servers = append(servers, serverInfo)
			} else {
				// 如果没有 ServerService，创建基本的 ServerInfo
				servers = append(servers, ServerInfo{
					Server:   *server,
					Password: user.UUID,
				})
			}
		}
	}

	return servers, nil
}

// GetAvailablePlansForUser 获取用户可购买的套餐列表
func (s *UserGroupService) GetAvailablePlansForUser(user *model.User) ([]model.Plan, error) {
	if user.GroupID == nil || *user.GroupID == 0 {
		// 没有用户组，返回空列�?
		return []model.Plan{}, nil
	}

	group, err := s.groupRepo.FindByID(*user.GroupID)
	if err != nil {
		return nil, err
	}

	planIDs := group.GetPlanIDsAsInt64()
	if len(planIDs) == 0 {
		return []model.Plan{}, nil
	}

	// 获取套餐列表
	plans := make([]model.Plan, 0)
	for _, planID := range planIDs {
		plan, err := s.planRepo.FindByID(planID)
		if err == nil && plan.Show && plan.Sell {
			plans = append(plans, *plan)
		}
	}

	return plans, nil
}
