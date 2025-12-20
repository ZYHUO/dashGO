package service

import (
	"encoding/json"
	"fmt"
	"time"

	"dashgo/internal/config"
	"dashgo/internal/model"
	"dashgo/internal/protocol"
	"dashgo/internal/repository"
)

// NodeService 节点管理服务
type NodeService struct {
	portService     *PortService
	configGenerator *config.ConfigGenerator
	templateManager *config.TemplateManager
	adapterRegistry *protocol.AdapterRegistry
	serverRepo      *repository.ServerRepository
	userRepo        *repository.UserRepository
}

// NodeCreateRequest 节点创建请求
type NodeCreateRequest struct {
	Name            string                 `json:"name" validate:"required"`
	Protocol        string                 `json:"protocol" validate:"required"`
	Host            string                 `json:"host" validate:"required"`
	Port            int                    `json:"port" validate:"required,min=1,max=65535"`
	LocalPort       int                    `json:"local_port,omitempty"` // 本地监听端口，可选
	Params          map[string]interface{} `json:"params"`
	AutoAllocatePort bool                  `json:"auto_allocate_port"` // 是否自动分配本地端口
}

// NodeUpdateRequest 节点更新请求
type NodeUpdateRequest struct {
	Name      *string                `json:"name,omitempty"`
	Host      *string                `json:"host,omitempty"`
	Port      *int                   `json:"port,omitempty"`
	LocalPort *int                   `json:"local_port,omitempty"`
	Params    map[string]interface{} `json:"params,omitempty"`
}

// NodeResponse 节点响应
type NodeResponse struct {
	ID          int64                  `json:"id"`
	Name        string                 `json:"name"`
	Protocol    string                 `json:"protocol"`
	Host        string                 `json:"host"`
	Port        int                    `json:"port"`        // 节点服务端口
	LocalPort   int                    `json:"local_port"`  // 本地监听端口
	Params      map[string]interface{} `json:"params"`
	Status      string                 `json:"status"`
	Config      string                 `json:"config,omitempty"` // 生成的配置
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ConfigHistory 配置历史记录
type ConfigHistory struct {
	ID        int64     `json:"id"`
	NodeID    int64     `json:"node_id"`
	Action    string    `json:"action"`    // "create", "update", "delete"
	OldConfig string    `json:"old_config,omitempty"`
	NewConfig string    `json:"new_config"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// NewNodeService 创建节点服务
func NewNodeService(
	portService *PortService,
	configGenerator *config.ConfigGenerator,
	templateManager *config.TemplateManager,
	adapterRegistry *protocol.AdapterRegistry,
	serverRepo *repository.ServerRepository,
	userRepo *repository.UserRepository,
) *NodeService {
	return &NodeService{
		portService:     portService,
		configGenerator: configGenerator,
		templateManager: templateManager,
		adapterRegistry: adapterRegistry,
		serverRepo:      serverRepo,
		userRepo:        userRepo,
	}
}

// CreateNode 创建节点
func (ns *NodeService) CreateNode(req *NodeCreateRequest) (*NodeResponse, error) {
	// 验证协议支持
	adapter, exists := ns.adapterRegistry.GetAdapter(req.Protocol)
	if !exists {
		return nil, fmt.Errorf("unsupported protocol: %s", req.Protocol)
	}

	// 验证协议参数
	if err := adapter.ValidateParams(req.Params); err != nil {
		return nil, fmt.Errorf("invalid protocol parameters: %w", err)
	}

	// 转换协议参数
	convertedParams, err := adapter.ConvertParams(req.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert protocol parameters: %w", err)
	}

	// 分配本地监听端口
	var localPort int
	if req.AutoAllocatePort || req.LocalPort == 0 {
		allocatedPort, err := ns.portService.AllocatePort(0, "inbound") // 使用0作为临时节点ID
		if err != nil {
			return nil, fmt.Errorf("failed to allocate local port: %w", err)
		}
		localPort = allocatedPort
	} else {
		localPort = req.LocalPort
		// 检查指定端口是否可用
		if !ns.portService.IsPortAvailable(localPort) {
			return nil, fmt.Errorf("local port %d is not available", localPort)
		}
	}

	// 创建服务器记录
	server := &model.Server{
		Name:     req.Name,
		Type:     req.Protocol,
		Host:             req.Host,
		Port:             fmt.Sprintf("%d", req.Port),
		ProtocolSettings: convertedParams,
	}

	// 保存到数据库
	if err := ns.serverRepo.Create(server); err != nil {
		// 如果创建失败，释放已分配的端口
		if req.AutoAllocatePort || req.LocalPort == 0 {
			ns.portService.ReleasePort(localPort)
		}
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	// 更新端口分配记录中的节点ID
	if req.AutoAllocatePort || req.LocalPort == 0 {
		// 先释放临时分配，再重新分配给正确的节点ID
		ns.portService.ReleasePort(localPort)
		if err := ns.portService.AllocateSpecificPort(localPort, server.ID, "inbound"); err != nil {
			return nil, fmt.Errorf("failed to update port allocation: %w", err)
		}
	} else {
		// 分配指定端口
		if err := ns.portService.AllocateSpecificPort(localPort, server.ID, "inbound"); err != nil {
			return nil, fmt.Errorf("failed to allocate specified port: %w", err)
		}
	}

	// 生成配置
	nodeConfig := config.NodeInfo{
		Name:     server.Name,
		Protocol: server.Type,
		Host:     server.Host,
		Port:     req.Port,
		Params:   convertedParams,
	}

	singboxConfig, err := ns.configGenerator.GenerateConfig([]config.NodeInfo{nodeConfig}, &config.GenerateOptions{
		LocalPort:   localPort,
		LogLevel:    "info",
		EnableDNS:   true,
		EnableRoute: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate config: %w", err)
	}

	configJSON, err := singboxConfig.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize config: %w", err)
	}

	// 记录配置历史
	ns.recordConfigHistory(server.ID, "create", "", configJSON, true, "")

	return &NodeResponse{
		ID:        server.ID,
		Name:      server.Name,
		Protocol:  server.Type,
		Host:      server.Host,
		Port:      req.Port,
		LocalPort: localPort,
		Params:    convertedParams,
		Config:    configJSON,
		CreatedAt: time.Unix(server.CreatedAt, 0),
		UpdatedAt: time.Unix(server.UpdatedAt, 0),
	}, nil
}

// UpdateNode 更新节点
func (ns *NodeService) UpdateNode(nodeID int64, req *NodeUpdateRequest) (*NodeResponse, error) {
	// 获取现有节点
	server, err := ns.serverRepo.FindByID(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	// 备份原始配置
	oldConfigData, _ := json.Marshal(server.ProtocolSettings)
	oldConfig := string(oldConfigData)

	// 更新字段
	updated := false
	if req.Name != nil && *req.Name != server.Name {
		server.Name = *req.Name
		updated = true
	}

	if req.Host != nil && *req.Host != server.Host {
		server.Host = *req.Host
		updated = true
	}

	var newPort int
	if req.Port != nil && *req.Port != parsePort(server.Port) {
		newPort = *req.Port
		server.Port = fmt.Sprintf("%d", newPort)
		updated = true
	} else {
		newPort = parsePort(server.Port)
	}

	// 处理本地端口更新
	currentPorts := ns.portService.GetNodePorts(nodeID)
	var currentLocalPort int
	if len(currentPorts) > 0 {
		currentLocalPort = currentPorts[0] // 假设只有一个本地端口
	}

	var newLocalPort int = currentLocalPort
	if req.LocalPort != nil && *req.LocalPort != currentLocalPort {
		// 检查新端口是否可用
		if !ns.portService.IsPortAvailable(*req.LocalPort) {
			return nil, fmt.Errorf("local port %d is not available", *req.LocalPort)
		}

		// 释放旧端口，分配新端口
		if currentLocalPort > 0 {
			ns.portService.ReleasePort(currentLocalPort)
		}
		
		if err := ns.portService.AllocateSpecificPort(*req.LocalPort, nodeID, "inbound"); err != nil {
			// 如果分配失败，尝试恢复旧端口
			if currentLocalPort > 0 {
				ns.portService.AllocateSpecificPort(currentLocalPort, nodeID, "inbound")
			}
			return nil, fmt.Errorf("failed to allocate new local port: %w", err)
		}
		
		newLocalPort = *req.LocalPort
		updated = true
	}

	// 更新协议参数
	if req.Params != nil {
		// 验证协议参数
		adapter, exists := ns.adapterRegistry.GetAdapter(server.Type)
		if !exists {
			return nil, fmt.Errorf("unsupported protocol: %s", server.Type)
		}

		if err := adapter.ValidateParams(req.Params); err != nil {
			return nil, fmt.Errorf("invalid protocol parameters: %w", err)
		}

		convertedParams, err := adapter.ConvertParams(req.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to convert protocol parameters: %w", err)
		}

		server.ProtocolSettings = convertedParams
		updated = true
	}

	if !updated {
		// 没有更新，返回当前状态
		return ns.getNodeResponse(server, newLocalPort)
	}

	// 保存更新
	if err := ns.serverRepo.Update(server); err != nil {
		return nil, fmt.Errorf("failed to update server: %w", err)
	}

	// 重新生成配置
	nodeConfig := config.NodeInfo{
		Name:     server.Name,
		Protocol: server.Type,
		Host:     server.Host,
		Port:     newPort,
		Params:   server.ProtocolSettings,
	}

	singboxConfig, err := ns.configGenerator.GenerateConfig([]config.NodeInfo{nodeConfig}, &config.GenerateOptions{
		LocalPort:   newLocalPort,
		LogLevel:    "info",
		EnableDNS:   true,
		EnableRoute: true,
	})
	if err != nil {
		ns.recordConfigHistory(nodeID, "update", oldConfig, "", false, err.Error())
		return nil, fmt.Errorf("failed to generate config: %w", err)
	}

	configJSON, err := singboxConfig.ToJSON()
	if err != nil {
		ns.recordConfigHistory(nodeID, "update", oldConfig, "", false, err.Error())
		return nil, fmt.Errorf("failed to serialize config: %w", err)
	}

	// 记录配置历史
	ns.recordConfigHistory(nodeID, "update", oldConfig, configJSON, true, "")

	return &NodeResponse{
		ID:        server.ID,
		Name:      server.Name,
		Protocol:  server.Type,
		Host:      server.Host,
		Port:      newPort,
		LocalPort: newLocalPort,
		Params:    server.ProtocolSettings,
		Config:    configJSON,
		CreatedAt: time.Unix(server.CreatedAt, 0),
		UpdatedAt: time.Unix(server.UpdatedAt, 0),
	}, nil
}

// DeleteNode 删除节点
func (ns *NodeService) DeleteNode(nodeID int64) error {
	// 获取节点信息
	server, err := ns.serverRepo.FindByID(nodeID)
	if err != nil {
		return fmt.Errorf("failed to get server: %w", err)
	}

	// 备份配置用于历史记录
	configData, _ := json.Marshal(server.ProtocolSettings)
	oldConfig := string(configData)

	// 释放所有相关端口
	if err := ns.portService.ReleasePortsByNode(nodeID); err != nil {
		return fmt.Errorf("failed to release ports: %w", err)
	}

	// 删除服务器记录
	if err := ns.serverRepo.Delete(nodeID); err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	// 记录配置历史
	ns.recordConfigHistory(nodeID, "delete", oldConfig, "", true, "")

	return nil
}

// GetNode 获取节点信息
func (ns *NodeService) GetNode(nodeID int64) (*NodeResponse, error) {
	server, err := ns.serverRepo.FindByID(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	// 获取本地端口
	ports := ns.portService.GetNodePorts(nodeID)
	var localPort int
	if len(ports) > 0 {
		localPort = ports[0]
	}

	return ns.getNodeResponse(server, localPort)
}

// ListNodes 列出所有节点
func (ns *NodeService) ListNodes() ([]*NodeResponse, error) {
	servers, err := ns.serverRepo.GetAllServers()
	if err != nil {
		return nil, fmt.Errorf("failed to get servers: %w", err)
	}

	var responses []*NodeResponse
	for _, server := range servers {
		// 获取本地端口
		ports := ns.portService.GetNodePorts(server.ID)
		var localPort int
		if len(ports) > 0 {
			localPort = ports[0]
		}

		response, err := ns.getNodeResponse(&server, localPort)
		if err != nil {
			continue // 跳过错误的节点
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GenerateNodeConfig 生成节点配置
func (ns *NodeService) GenerateNodeConfig(nodeID int64) (string, error) {
	server, err := ns.serverRepo.FindByID(nodeID)
	if err != nil {
		return "", fmt.Errorf("failed to get server: %w", err)
	}

	// 获取本地端口
	ports := ns.portService.GetNodePorts(nodeID)
	var localPort int
	if len(ports) > 0 {
		localPort = ports[0]
	}

	nodeConfig := config.NodeInfo{
		Name:     server.Name,
		Protocol: server.Type,
		Host:     server.Host,
		Port:     parsePort(server.Port),
		Params:   server.ProtocolSettings,
	}

	singboxConfig, err := ns.configGenerator.GenerateConfig([]config.NodeInfo{nodeConfig}, &config.GenerateOptions{
		LocalPort:   localPort,
		LogLevel:    "info",
		EnableDNS:   true,
		EnableRoute: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate config: %w", err)
	}

	return singboxConfig.ToJSON()
}

// RollbackNodeConfig 回滚节点配置
func (ns *NodeService) RollbackNodeConfig(nodeID int64) error {
	// 这里可以实现配置回滚逻辑
	// 从配置历史中获取上一个有效配置并应用
	return fmt.Errorf("rollback not implemented yet")
}

// getNodeResponse 构建节点响应
func (ns *NodeService) getNodeResponse(server *model.Server, localPort int) (*NodeResponse, error) {
	// 生成当前配置
	nodeConfig := config.NodeInfo{
		Name:     server.Name,
		Protocol: server.Type,
		Host:     server.Host,
		Port:     parsePort(server.Port),
		Params:   server.ProtocolSettings,
	}

	singboxConfig, err := ns.configGenerator.GenerateConfig([]config.NodeInfo{nodeConfig}, &config.GenerateOptions{
		LocalPort:   localPort,
		LogLevel:    "info",
		EnableDNS:   true,
		EnableRoute: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate config: %w", err)
	}

	configJSON, err := singboxConfig.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize config: %w", err)
	}

	return &NodeResponse{
		ID:        server.ID,
		Name:      server.Name,
		Protocol:  server.Type,
		Host:      server.Host,
		Port:      parsePort(server.Port),
		LocalPort: localPort,
		Params:    server.ProtocolSettings,
		Config:    configJSON,
		CreatedAt: time.Unix(server.CreatedAt, 0),
		UpdatedAt: time.Unix(server.UpdatedAt, 0),
	}, nil
}

// recordConfigHistory 记录配置历史
func (ns *NodeService) recordConfigHistory(nodeID int64, action, oldConfig, newConfig string, success bool, errorMsg string) {
	// 这里可以实现配置历史记录逻辑
	// 保存到数据库或日志文件
	history := ConfigHistory{
		NodeID:    nodeID,
		Action:    action,
		OldConfig: oldConfig,
		NewConfig: newConfig,
		Success:   success,
		Error:     errorMsg,
		Timestamp: time.Now(),
	}
	
	// TODO: 保存到数据库
	_ = history
}

// parsePort 解析端口字符串
func parsePort(portStr string) int {
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return port
}