package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"xboard/internal/model"
	"xboard/internal/repository"
	"xboard/pkg/utils"
)

// HostService 主机服务
type HostService struct {
	hostRepo *repository.HostRepository
	nodeRepo *repository.ServerNodeRepository
	userRepo *repository.UserRepository
}

func NewHostService(hostRepo *repository.HostRepository, nodeRepo *repository.ServerNodeRepository, userRepo *repository.UserRepository) *HostService {
	return &HostService{
		hostRepo: hostRepo,
		nodeRepo: nodeRepo,
		userRepo: userRepo,
	}
}

// CreateHost 创建主机
func (s *HostService) CreateHost(name string) (*model.Host, error) {
	token := generateHostToken()
	host := &model.Host{
		Name:      name,
		Token:     token,
		AgentPort: 9999,
		Status:    model.HostStatusOffline,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	if err := s.hostRepo.Create(host); err != nil {
		return nil, err
	}
	return host, nil
}

// GetByToken 根据 Token 获取主机
func (s *HostService) GetByToken(token string) (*model.Host, error) {
	return s.hostRepo.FindByToken(token)
}

// GetByID 根据 ID 获取主机
func (s *HostService) GetByID(id int64) (*model.Host, error) {
	return s.hostRepo.FindByID(id)
}

// GetAll 获取所有主机
func (s *HostService) GetAll() ([]model.Host, error) {
	return s.hostRepo.GetAll()
}

// UpdateHeartbeat 更新心跳
func (s *HostService) UpdateHeartbeat(hostID int64, ip string, systemInfo map[string]interface{}) error {
	host, err := s.hostRepo.FindByID(hostID)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	host.LastHeartbeat = &now
	host.IP = ip
	host.Status = model.HostStatusOnline
	host.SystemInfo = systemInfo
	return s.hostRepo.Update(host)
}

// ResetToken 重置主机 Token
func (s *HostService) ResetToken(hostID int64) (string, error) {
	host, err := s.hostRepo.FindByID(hostID)
	if err != nil {
		return "", err
	}
	host.Token = generateHostToken()
	if err := s.hostRepo.Update(host); err != nil {
		return "", err
	}
	return host.Token, nil
}

// Delete 删除主机
func (s *HostService) Delete(hostID int64) error {
	// 先删除主机下的所有节点
	if err := s.nodeRepo.DeleteByHostID(hostID); err != nil {
		return err
	}
	return s.hostRepo.Delete(hostID)
}

// CreateNode 创建节点
func (s *HostService) CreateNode(node *model.ServerNode) error {
	node.CreatedAt = time.Now().Unix()
	node.UpdatedAt = time.Now().Unix()
	return s.nodeRepo.Create(node)
}

// UpdateNode 更新节点
func (s *HostService) UpdateNode(node *model.ServerNode) error {
	node.UpdatedAt = time.Now().Unix()
	return s.nodeRepo.Update(node)
}

// DeleteNode 删除节点
func (s *HostService) DeleteNode(nodeID int64) error {
	return s.nodeRepo.Delete(nodeID)
}

// GetNodesByHostID 获取主机下的所有节点
func (s *HostService) GetNodesByHostID(hostID int64) ([]model.ServerNode, error) {
	return s.nodeRepo.FindByHostID(hostID)
}

// GetNodeByID 根据 ID 获取节点
func (s *HostService) GetNodeByID(nodeID int64) (*model.ServerNode, error) {
	return s.nodeRepo.FindByID(nodeID)
}

// GenerateSingBoxConfig 生成 sing-box 配置
func (s *HostService) GenerateSingBoxConfig(hostID int64) (map[string]interface{}, error) {
	nodes, err := s.nodeRepo.FindByHostID(hostID)
	if err != nil {
		return nil, err
	}

	inbounds := make([]map[string]interface{}, 0)

	for _, node := range nodes {
		inbound := s.buildInbound(&node)
		if inbound != nil {
			inbounds = append(inbounds, inbound)
		}
	}

	config := map[string]interface{}{
		"log": map[string]interface{}{
			"level":     "info",
			"timestamp": true,
		},
		"inbounds": inbounds,
		"outbounds": []map[string]interface{}{
			{"type": "direct", "tag": "direct"},
			{"type": "block", "tag": "block"},
		},
		"route": map[string]interface{}{
			"rules": []map[string]interface{}{
				{"ip_is_private": true, "outbound": "block"},
			},
			"final": "direct",
		},
	}

	return config, nil
}

// buildInbound 构建 inbound 配置
func (s *HostService) buildInbound(node *model.ServerNode) map[string]interface{} {
	tag := node.Type + "-in-" + fmt.Sprintf("%d", node.ID)

	inbound := map[string]interface{}{
		"type":        node.Type,
		"tag":         tag,
		"listen":      "::",
		"listen_port": node.ListenPort,
	}

	// 合并协议设置
	for k, v := range node.ProtocolSettings {
		inbound[k] = v
	}

	// Shadowsocks 2022 需要服务器密钥
	if node.Type == model.NodeTypeShadowsocks {
		cipher := ""
		if c, ok := node.ProtocolSettings["method"].(string); ok {
			cipher = c
		}
		// 为 SS2022 生成服务器密钥
		if strings.HasPrefix(cipher, "2022-") {
			keySize := 16
			if cipher == "2022-blake3-aes-256-gcm" || cipher == "2022-blake3-chacha20-poly1305" {
				keySize = 32
			}
			inbound["password"] = utils.GetServerKey(node.CreatedAt, keySize)
		}
	}

	// TLS 设置
	if len(node.TLSSettings) > 0 {
		inbound["tls"] = node.TLSSettings
	}

	// Transport 设置
	if len(node.TransportSettings) > 0 {
		inbound["transport"] = node.TransportSettings
	}

	// 用户列表初始化为空
	switch node.Type {
	case model.NodeTypeVMess, model.NodeTypeVLESS, model.NodeTypeTrojan, model.NodeTypeHysteria2, model.NodeTypeTUIC:
		inbound["users"] = []interface{}{}
	case model.NodeTypeShadowsocks:
		inbound["users"] = []interface{}{}
	}

	return inbound
}

// GetUsersForNode 获取节点可用的用户列表
func (s *HostService) GetUsersForNode(node *model.ServerNode) ([]map[string]interface{}, error) {
	groupIDs := node.GetGroupIDsAsInt64()
	
	var users []model.User
	var err error
	
	if len(groupIDs) == 0 {
		// 如果没有设置组，获取所有可用用户
		users, err = s.userRepo.GetAllAvailableUsers()
	} else {
		users, err = s.userRepo.GetAvailableUsers(groupIDs)
	}
	
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		userConfig := map[string]interface{}{
			"username": user.UUID,
		}

		// 根据协议类型设置密码字段
		switch node.Type {
		case model.NodeTypeShadowsocks:
			userConfig["password"] = s.generateSS2022Password(node, &user)
		case model.NodeTypeVMess, model.NodeTypeVLESS:
			userConfig["uuid"] = user.UUID
		case model.NodeTypeTrojan, model.NodeTypeHysteria2, model.NodeTypeTUIC:
			userConfig["password"] = user.UUID
		}

		result = append(result, userConfig)
	}

	return result, nil
}

// generateSS2022Password 生成 SS2022 密码
func (s *HostService) generateSS2022Password(node *model.ServerNode, user *model.User) string {
	cipher := ""
	if c, ok := node.ProtocolSettings["method"].(string); ok {
		cipher = c
	}

	switch cipher {
	case "2022-blake3-aes-128-gcm":
		serverKey := utils.GetServerKey(node.CreatedAt, 16)
		userKey := utils.UUIDToBase64(user.UUID, 16)
		return serverKey + ":" + userKey
	case "2022-blake3-aes-256-gcm", "2022-blake3-chacha20-poly1305":
		serverKey := utils.GetServerKey(node.CreatedAt, 32)
		userKey := utils.UUIDToBase64(user.UUID, 32)
		return serverKey + ":" + userKey
	default:
		return user.UUID
	}
}

// GetDefaultNodeConfig 获取默认节点配置（带完整默认值）
func (s *HostService) GetDefaultNodeConfig(nodeType string) map[string]interface{} {
	switch nodeType {
	case model.NodeTypeShadowsocks:
		return map[string]interface{}{
			"name":        "SS2022节点",
			"listen_port": 8388,
			"protocol_settings": map[string]interface{}{
				"method": "2022-blake3-aes-128-gcm",
			},
		}
	case model.NodeTypeVLESS:
		return map[string]interface{}{
			"name":        "VLESS Reality节点",
			"listen_port": 443,
			"protocol_settings": map[string]interface{}{
				"flow": "xtls-rprx-vision",
			},
			"tls_settings": map[string]interface{}{
				"enabled":     true,
				"server_name": "www.microsoft.com",
				"reality": map[string]interface{}{
					"enabled": true,
					"handshake": map[string]interface{}{
						"server":      "www.microsoft.com",
						"server_port": 443,
					},
					"private_key": "", // Agent 自动生成
					"short_id":    []string{"0123456789abcdef"},
				},
			},
		}
	case model.NodeTypeTrojan:
		return map[string]interface{}{
			"name":        "Trojan节点",
			"listen_port": 443,
			"tls_settings": map[string]interface{}{
				"enabled":     true,
				"server_name": "",
				"acme": map[string]interface{}{
					"domain": "",
					"email":  "",
				},
			},
		}
	case model.NodeTypeHysteria2:
		return map[string]interface{}{
			"name":        "Hysteria2节点",
			"listen_port": 443,
			"protocol_settings": map[string]interface{}{
				"up_mbps":   100,
				"down_mbps": 100,
			},
			"tls_settings": map[string]interface{}{
				"enabled":     true,
				"server_name": "",
			},
		}
	default:
		return map[string]interface{}{}
	}
}

// GetAllNodes 获取所有节点
func (s *HostService) GetAllNodes() ([]model.ServerNode, error) {
	return s.nodeRepo.GetAll()
}

// AgentConfig Agent 配置
type AgentConfig struct {
	SingBoxConfig map[string]interface{}   `json:"singbox_config"`
	Nodes         []AgentNodeConfig        `json:"nodes"`
}

// AgentNodeConfig Agent 节点配置
type AgentNodeConfig struct {
	ID       int64                    `json:"id"`
	Type     string                   `json:"type"`
	Port     int                      `json:"port"`
	Tag      string                   `json:"tag"`
	Users    []map[string]interface{} `json:"users"`
}

// GetAgentConfig 获取 Agent 完整配置
func (s *HostService) GetAgentConfig(hostID int64) (*AgentConfig, error) {
	config, err := s.GenerateSingBoxConfig(hostID)
	if err != nil {
		return nil, err
	}

	nodes, err := s.nodeRepo.FindByHostID(hostID)
	if err != nil {
		return nil, err
	}

	nodeConfigs := make([]AgentNodeConfig, 0, len(nodes))
	for _, node := range nodes {
		users, _ := s.GetUsersForNode(&node)
		nodeConfigs = append(nodeConfigs, AgentNodeConfig{
			ID:    node.ID,
			Type:  node.Type,
			Port:  node.ListenPort,
			Tag:   node.Type + "-in-" + fmt.Sprintf("%d", node.ID),
			Users: users,
		})
	}

	return &AgentConfig{
		SingBoxConfig: config,
		Nodes:         nodeConfigs,
	}, nil
}

// ToJSON 转换为 JSON
func (c *AgentConfig) ToJSON() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}

func generateHostToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
