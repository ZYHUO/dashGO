package service

import (
	"fmt"
	"time"

	"dashgo/internal/model"
	portpkg "dashgo/internal/port"
	"dashgo/internal/repository"
)

// PortService 端口管理服务
type PortService struct {
	manager    *portpkg.PortManager
	repository *repository.PortRepository
}

// NewPortService 创建端口服务实例
func NewPortService(portRange model.PortRange, repo *repository.PortRepository) (*PortService, error) {
	if err := portRange.Validate(); err != nil {
		return nil, fmt.Errorf("invalid port range: %w", err)
	}

	manager := portpkg.NewPortManager(portpkg.PortRange{
		Start: portRange.Start,
		End:   portRange.End,
	})

	service := &PortService{
		manager:    manager,
		repository: repo,
	}

	// 初始化时从数据库恢复端口分配状态
	if err := service.restorePortAllocations(); err != nil {
		return nil, fmt.Errorf("failed to restore port allocations: %w", err)
	}

	return service, nil
}

// AllocatePort 分配端口
func (s *PortService) AllocatePort(nodeID int64, purpose string) (int, error) {
	port, err := s.manager.AllocatePort(nodeID, purpose)
	if err != nil {
		return 0, err
	}

	// 保存到数据库
	allocation := &model.PortAllocation{
		Port:        port,
		NodeID:      nodeID,
		Purpose:     purpose,
		Status:      "allocated",
		AllocatedAt: time.Now(),
	}

	if err := s.repository.CreateAllocation(allocation); err != nil {
		// 如果数据库保存失败，回滚内存中的分配
		s.manager.ReleasePort(port)
		return 0, fmt.Errorf("failed to save port allocation: %w", err)
	}

	return port, nil
}

// AllocateSpecificPort 分配指定端口
func (s *PortService) AllocateSpecificPort(port int, nodeID int64, purpose string) error {
	if err := s.manager.AllocateSpecificPort(port, nodeID, purpose); err != nil {
		return err
	}

	// 保存到数据库
	allocation := &model.PortAllocation{
		Port:        port,
		NodeID:      nodeID,
		Purpose:     purpose,
		Status:      "allocated",
		AllocatedAt: time.Now(),
	}

	if err := s.repository.CreateAllocation(allocation); err != nil {
		// 如果数据库保存失败，回滚内存中的分配
		s.manager.ReleasePort(port)
		return fmt.Errorf("failed to save port allocation: %w", err)
	}

	return nil
}

// ReleasePort 释放端口
func (s *PortService) ReleasePort(port int) error {
	if err := s.manager.ReleasePort(port); err != nil {
		return err
	}

	// 更新数据库
	if err := s.repository.ReleasePort(port); err != nil {
		return fmt.Errorf("failed to update port release in database: %w", err)
	}

	return nil
}

// ReleasePortsByNode 释放节点的所有端口
func (s *PortService) ReleasePortsByNode(nodeID int64) error {
	if err := s.manager.ReleasePortsByNode(nodeID); err != nil {
		return err
	}

	// 更新数据库
	if err := s.repository.ReleasePortsByNode(nodeID); err != nil {
		return fmt.Errorf("failed to update port releases in database: %w", err)
	}

	return nil
}

// IsPortAvailable 检查端口是否可用
func (s *PortService) IsPortAvailable(port int) bool {
	return s.manager.IsPortAvailable(port)
}

// GetUsedPorts 获取已使用的端口列表
func (s *PortService) GetUsedPorts() []int {
	return s.manager.GetUsedPorts()
}

// GetPortAllocation 获取端口分配信息
func (s *PortService) GetPortAllocation(port int) (*portpkg.PortAllocation, error) {
	// 先从内存中获取
	if allocation := s.manager.GetPortAllocation(port); allocation != nil {
		return allocation, nil
	}

	// 如果内存中没有，从数据库获取
	dbAllocation, err := s.repository.GetAllocation(port)
	if err != nil {
		return nil, err
	}

	// 转换为内存模型
	var releasedAt *time.Time
	if dbAllocation.ReleasedAt != nil {
		releasedAt = dbAllocation.ReleasedAt
	}
	
	return &portpkg.PortAllocation{
		ID:          dbAllocation.ID,
		Port:        dbAllocation.Port,
		NodeID:      dbAllocation.NodeID,
		Purpose:     dbAllocation.Purpose,
		Status:      dbAllocation.Status,
		AllocatedAt: dbAllocation.AllocatedAt,
		ReleasedAt:  releasedAt,
	}, nil
}

// GetNodePorts 获取节点的所有端口
func (s *PortService) GetNodePorts(nodeID int64) []int {
	return s.manager.GetNodePorts(nodeID)
}

// GetMetrics 获取端口使用统计
func (s *PortService) GetMetrics() (*portpkg.PortUsageMetrics, error) {
	return s.manager.GetMetrics(), nil
}

// ScanAndRecordUsedPorts 扫描并记录当前系统使用的端口
func (s *PortService) ScanAndRecordUsedPorts() error {
	return s.manager.ScanAndRecordUsedPorts()
}

// GetPortHistory 获取端口分配历史
func (s *PortService) GetPortHistory(port int, limit int) ([]model.PortAllocation, error) {
	return s.repository.GetPortHistory(port, limit)
}

// GetNodePortHistory 获取节点端口分配历史
func (s *PortService) GetNodePortHistory(nodeID int64, limit int) ([]model.PortAllocation, error) {
	return s.repository.GetNodePortHistory(nodeID, limit)
}

// CleanupOldRecords 清理旧的端口分配记录
func (s *PortService) CleanupOldRecords(olderThan time.Time) error {
	return s.repository.CleanupOldRecords(olderThan)
}

// restorePortAllocations 从数据库恢复端口分配状态
func (s *PortService) restorePortAllocations() error {
	// 获取所有已分配的端口
	allocatedPorts, err := s.repository.GetAllocatedPorts()
	if err != nil {
		return err
	}

	// 恢复到内存管理器中
	for _, port := range allocatedPorts {
		allocation, err := s.repository.GetAllocation(port)
		if err != nil {
			continue // 跳过错误的记录
		}

		// 直接在内存中标记为已分配，不调用AllocateSpecificPort避免重复写数据库
		s.manager.AllocateSpecificPort(port, allocation.NodeID, allocation.Purpose)
	}

	return nil
}

// ValidatePortRange 验证端口范围
func (s *PortService) ValidatePortRange() error {
	return s.manager.ValidatePortRange()
}