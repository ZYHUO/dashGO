package service

import (
	"fmt"

	"dashgo/internal/config"
	"dashgo/internal/model"
	"dashgo/internal/protocol"
	"dashgo/internal/repository"

	"gorm.io/gorm"
)

// IntegrationService 系统集成服务
type IntegrationService struct {
	db              *gorm.DB
	portService     *PortService
	nodeService     *NodeService
	configGenerator *config.ConfigGenerator
	templateManager *config.TemplateManager
	validator       *config.ValidationEngine
	protocolHandler *protocol.ProtocolHandler
	repositories    *repository.Repositories
}

// IntegrationConfig 集成配置
type IntegrationConfig struct {
	PortRange model.PortRange `json:"port_range"`
	LogLevel  string          `json:"log_level"`
}

// NewIntegrationService 创建集成服务
func NewIntegrationService(db *gorm.DB, cfg IntegrationConfig) (*IntegrationService, error) {
	// 初始化仓库
	repositories := repository.NewRepositories(db)

	// 初始化端口服务
	portService, err := NewPortService(cfg.PortRange, repositories.Port)
	if err != nil {
		return nil, fmt.Errorf("failed to create port service: %w", err)
	}

	// 初始化配置生成器
	configGenerator := config.NewConfigGenerator()

	// 初始化模板管理器
	templateManager := config.NewTemplateManager("./templates")

	// 初始化验证引擎
	validator := config.NewValidationEngine()

	// 初始化协议处理器
	protocolHandler := protocol.NewProtocolHandler()

	// 初始化节点服务
	nodeService := NewNodeService(
		portService,
		configGenerator,
		templateManager,
		protocol.NewAdapterRegistry(),
		repositories.Server,
		repositories.User,
	)

	return &IntegrationService{
		db:              db,
		portService:     portService,
		nodeService:     nodeService,
		configGenerator: configGenerator,
		templateManager: templateManager,
		validator:       validator,
		protocolHandler: protocolHandler,
		repositories:    repositories,
	}, nil
}

// GetPortService 获取端口服务
func (is *IntegrationService) GetPortService() *PortService {
	return is.portService
}

// GetNodeService 获取节点服务
func (is *IntegrationService) GetNodeService() *NodeService {
	return is.nodeService
}

// GetConfigGenerator 获取配置生成器
func (is *IntegrationService) GetConfigGenerator() *config.ConfigGenerator {
	return is.configGenerator
}

// GetTemplateManager 获取模板管理器
func (is *IntegrationService) GetTemplateManager() *config.TemplateManager {
	return is.templateManager
}

// GetValidator 获取验证引擎
func (is *IntegrationService) GetValidator() *config.ValidationEngine {
	return is.validator
}

// GetProtocolHandler 获取协议处理器
func (is *IntegrationService) GetProtocolHandler() *protocol.ProtocolHandler {
	return is.protocolHandler
}

// GetRepositories 获取仓库集合
func (is *IntegrationService) GetRepositories() *repository.Repositories {
	return is.repositories
}

// Initialize 初始化系统
func (is *IntegrationService) Initialize() error {
	// 扫描并记录当前系统使用的端口
	if err := is.portService.ScanAndRecordUsedPorts(); err != nil {
		return fmt.Errorf("failed to scan used ports: %w", err)
	}

	// 加载默认模板
	// 这里可以从文件系统加载模板

	return nil
}

// HealthCheck 健康检查
func (is *IntegrationService) HealthCheck() error {
	// 检查数据库连接
	sqlDB, err := is.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// 检查端口服务
	if err := is.portService.ValidatePortRange(); err != nil {
		return fmt.Errorf("port service validation failed: %w", err)
	}

	return nil
}

// GetSystemStatus 获取系统状态
func (is *IntegrationService) GetSystemStatus() (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// 端口使用统计
	portMetrics, err := is.portService.GetMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to get port metrics: %w", err)
	}
	status["port_metrics"] = portMetrics

	// 节点统计
	nodes, err := is.nodeService.ListNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	status["node_count"] = len(nodes)

	// 协议统计
	protocolStats := make(map[string]int)
	for _, node := range nodes {
		protocolStats[node.Protocol]++
	}
	status["protocol_stats"] = protocolStats

	// 模板统计
	templates := is.templateManager.ListTemplates()
	status["template_count"] = len(templates)

	return status, nil
}

// Cleanup 清理资源
func (is *IntegrationService) Cleanup() error {
	// 清理旧的配置历史记录（保留最近30天）
	// 这里可以实现清理逻辑

	return nil
}