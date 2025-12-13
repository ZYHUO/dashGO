package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dashgo/internal/model"
	"dashgo/internal/repository"
	"dashgo/pkg/cache"
	"dashgo/pkg/utils"
)

// HostService 主机服务
type HostService struct {
	hostRepo   *repository.HostRepository
	nodeRepo   *repository.ServerNodeRepository
	userRepo   *repository.UserRepository
	serverRepo *repository.ServerRepository
	cache      *cache.Client
}

func NewHostService(hostRepo *repository.HostRepository, nodeRepo *repository.ServerNodeRepository, userRepo *repository.UserRepository, serverRepo *repository.ServerRepository, cacheClient *cache.Client) *HostService {
	return &HostService{
		hostRepo:   hostRepo,
		nodeRepo:   nodeRepo,
		userRepo:   userRepo,
		serverRepo: serverRepo,
		cache:      cacheClient,
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

// UpdateHost 更新主机信息
func (s *HostService) UpdateHost(host *model.Host) error {
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
	// 先解除所有绑定到此主机的节点
	if err := s.serverRepo.UnbindFromHost(hostID); err != nil {
		return err
	}
	// 删除主机下的所有ServerNode（如果有的话）
	if err := s.nodeRepo.DeleteByHostID(hostID); err != nil {
		return err
	}
	return s.hostRepo.Delete(hostID)
}

// GetServersByHostID 获取绑定到主机的所有节点（只读，用于查询）
func (s *HostService) GetServersByHostID(hostID int64) ([]model.Server, error) {
	return s.serverRepo.GetByHostID(hostID)
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
	inbounds := make([]map[string]interface{}, 0)
	processedServerIDs := make(map[int64]bool) // 记录已处理的 Server ID，避免重复

	// 1. 从绑定到主机的Server 获取配置
	servers, err := s.serverRepo.GetByHostID(hostID)
	if err == nil {
		for _, server := range servers {
			if processedServerIDs[server.ID] {
				continue
			}
			inbound := s.buildInboundFromServer(&server)
			if inbound != nil {
				inbounds = append(inbounds, inbound)
				processedServerIDs[server.ID] = true
			}
		}
	}

	// 2. 告ServerNode 获取配置（兼容旧逻辑告
	nodes, err := s.nodeRepo.FindByHostID(hostID)
	if err == nil {
		for _, node := range nodes {
			inbound := s.buildInbound(&node)
			if inbound != nil {
				inbounds = append(inbounds, inbound)
			}
		}
	}

	// 获取主机信息，检查是否配置了 SOCKS 出口
	host, err := s.hostRepo.FindByID(hostID)
	if err != nil {
		return nil, err
	}

	// 构建 outbounds
	outbounds := []map[string]interface{}{
		{"type": "direct", "tag": "direct"},
		{"type": "block", "tag": "block"},
	}

	// 如果配置告SOCKS 出口，添告SOCKS outbound 并设置为默认出口
	finalOutbound := "direct"
	if host.SocksOutbound != nil && *host.SocksOutbound != "" {
		socksOutbound := s.parseSocksOutbound(*host.SocksOutbound)
		if socksOutbound != nil {
			outbounds = append([]map[string]interface{}{socksOutbound}, outbounds...)
			finalOutbound = "socks-out" // 使用 SOCKS 作为默认出口
		}
	}

	config := map[string]interface{}{
		"log": map[string]interface{}{
			"level":     "info",
			"timestamp": true,
		},
		"inbounds":  inbounds,
		"outbounds": outbounds,
		"route": map[string]interface{}{
			"rules": []map[string]interface{}{
				{"ip_is_private": true, "outbound": "block"},
			},
			"final": finalOutbound, // 使用配置的默认出告
		},
		"experimental": map[string]interface{}{
			"clash_api": map[string]interface{}{
				"external_controller": "127.0.0.1:9090",
			},
		},
	}

	return config, nil
}

// buildInboundFromServer 告Server 构建 inbound 配置
func (s *HostService) buildInboundFromServer(server *model.Server) map[string]interface{} {
	tag := server.Type + "-in-" + fmt.Sprintf("%d", server.ID)

	inbound := map[string]interface{}{
		"type":        server.Type,
		"tag":         tag,
		"listen":      "::",
		"listen_port": server.ServerPort,
	}

	// 合并协议设置
	for k, v := range server.ProtocolSettings {
		if k == "tls_settings" || k == "network_settings" || k == "tls" {
			continue
		}
		// sing-box 使用 method 而不告cipher
		if k == "cipher" {
			inbound["method"] = v
			continue
		}
		inbound[k] = v
	}

	// Shadowsocks 需要特殊处告
	if server.Type == model.ServerTypeShadowsocks {
		// 获取加密方式
		cipher := ""
		if c, ok := server.ProtocolSettings["method"].(string); ok {
			cipher = c
		} else if c, ok := server.ProtocolSettings["cipher"].(string); ok {
			cipher = c
		}

		// 确保 method 字段存在，删告cipher 字段
		inbound["method"] = cipher
		delete(inbound, "cipher")

		// 告SS2022 生成服务器密告
		if strings.HasPrefix(cipher, "2022-") {
			keySize := 16
			if cipher == "2022-blake3-aes-256-gcm" || cipher == "2022-blake3-chacha20-poly1305" {
				keySize = 32
			}
			inbound["password"] = utils.GetServerKey(server.CreatedAt, keySize)
		}
	}

	// TLS 设置
	if tls, ok := server.ProtocolSettings["tls_settings"].(map[string]interface{}); ok {
		inbound["tls"] = tls
	}

	// Transport 设置
	if transport, ok := server.ProtocolSettings["network_settings"].(map[string]interface{}); ok {
		inbound["transport"] = transport
	}

	// 用户列表初始化为告
	switch server.Type {
	case model.ServerTypeVmess, model.ServerTypeVless, model.ServerTypeTrojan, model.ServerTypeHysteria, model.ServerTypeTuic:
		inbound["users"] = []interface{}{}
	case model.ServerTypeShadowsocks:
		inbound["users"] = []interface{}{}
	}

	return inbound
}

// buildInbound 构建 inbound 配置
func (s *HostService) buildInbound(node *model.ServerNode) map[string]interface{} {
	// 使用节点自身的配告
	protocolSettings := node.ProtocolSettings
	tlsSettings := node.TLSSettings
	transportSettings := node.TransportSettings
	nodeType := node.Type
	createdAt := node.CreatedAt

	tag := nodeType + "-in-" + fmt.Sprintf("%d", node.ID)

	inbound := map[string]interface{}{
		"type":        nodeType,
		"tag":         tag,
		"listen":      "::",
		"listen_port": node.ListenPort,
	}

	// 合并协议设置
	for k, v := range protocolSettings {
		// 跳过不需要的字段
		if k == "tls_settings" || k == "network_settings" || k == "tls" {
			continue
		}
		// sing-box 使用 method 而不告cipher
		if k == "cipher" {
			inbound["method"] = v
			continue
		}
		inbound[k] = v
	}

	// Shadowsocks 需要特殊处告
	if nodeType == model.NodeTypeShadowsocks {
		// 获取加密方式
		cipher := ""
		if c, ok := protocolSettings["method"].(string); ok {
			cipher = c
		} else if c, ok := protocolSettings["cipher"].(string); ok {
			cipher = c
		}

		// 确保 method 字段存在
		inbound["method"] = cipher
		// 删除可能存在告cipher 字段（sing-box 不认识）
		delete(inbound, "cipher")

		// 告SS2022 生成服务器密告
		if strings.HasPrefix(cipher, "2022-") {
			keySize := 16
			if cipher == "2022-blake3-aes-256-gcm" || cipher == "2022-blake3-chacha20-poly1305" {
				keySize = 32
			}
			inbound["password"] = utils.GetServerKey(createdAt, keySize)
		}
	}

	// TLS 设置
	if len(tlsSettings) > 0 {
		inbound["tls"] = tlsSettings
	}

	// Transport 设置
	if len(transportSettings) > 0 {
		inbound["transport"] = transportSettings
	}

	// 用户列表初始化为告
	switch nodeType {
	case model.NodeTypeVMess, model.NodeTypeVLESS, model.NodeTypeTrojan, model.NodeTypeHysteria2, model.NodeTypeTUIC:
		inbound["users"] = []interface{}{}
	case model.NodeTypeShadowsocks:
		inbound["users"] = []interface{}{}
	case model.NodeTypeAnyTLS:
		inbound["users"] = []interface{}{}
	case model.NodeTypeShadowTLS:
		// ShadowTLS 需要特殊处告
		s.buildShadowTLSInbound(inbound, node)
	case model.NodeTypeNaive:
		inbound["users"] = []interface{}{}
	}

	return inbound
}

// buildShadowTLSInbound 构建 ShadowTLS inbound
func (s *HostService) buildShadowTLSInbound(inbound map[string]interface{}, node *model.ServerNode) {
	ps := node.ProtocolSettings

	// ShadowTLS v3 配置
	version := 3
	if v, ok := ps["version"].(float64); ok {
		version = int(v)
	}
	inbound["version"] = version

	// 握手服务告
	handshakeServer := "addons.mozilla.org"
	if hs, ok := ps["handshake_server"].(string); ok && hs != "" {
		handshakeServer = hs
	}
	handshakePort := 443
	if hp, ok := ps["handshake_port"].(float64); ok {
		handshakePort = int(hp)
	}
	inbound["handshake"] = map[string]interface{}{
		"server":      handshakeServer,
		"server_port": handshakePort,
	}

	// 严格模式
	if strictMode, ok := ps["strict_mode"].(bool); ok {
		inbound["strict_mode"] = strictMode
	} else {
		inbound["strict_mode"] = true
	}

	// 用户列表
	inbound["users"] = []interface{}{}

	// 删除不需要的字段
	delete(inbound, "handshake_server")
	delete(inbound, "handshake_port")
	delete(inbound, "detour_method")
}

// GetUsersForNode 获取节点可用的用户列告
func (s *HostService) GetUsersForNode(node *model.ServerNode) ([]map[string]interface{}, error) {
	// 使用节点自身的配告
	groupIDs := node.GetGroupIDsAsInt64()
	nodeType := node.Type
	protocolSettings := node.ProtocolSettings

	var users []model.User
	var err error

	if len(groupIDs) == 0 {
		// 如果没有设置组，获取所有可用用告
		users, err = s.userRepo.GetAllAvailableUsers()
	} else {
		users, err = s.userRepo.GetAvailableUsers(groupIDs)
	}

	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		userConfig := map[string]interface{}{}

		// 根据协议类型设置用户配置
		// sing-box 不同协议的用户字段不告
		switch nodeType {
		case model.NodeTypeShadowsocks:
			// SS 用户只需告name 告password
			userConfig["name"] = user.UUID[:8]
			userConfig["password"] = s.getSS2022UserKey(protocolSettings, &user)
		case model.NodeTypeVMess, model.NodeTypeVLESS:
			userConfig["name"] = user.UUID[:8]
			userConfig["uuid"] = user.UUID
		case model.NodeTypeTrojan:
			userConfig["name"] = user.UUID[:8]
			userConfig["password"] = user.UUID
		case model.NodeTypeHysteria2:
			userConfig["name"] = user.UUID[:8]
			userConfig["password"] = user.UUID
		case model.NodeTypeTUIC:
			userConfig["name"] = user.UUID[:8]
			userConfig["uuid"] = user.UUID
			userConfig["password"] = user.UUID
		case model.NodeTypeAnyTLS:
			userConfig["name"] = user.UUID[:8]
			userConfig["password"] = user.UUID
		case model.NodeTypeShadowTLS:
			userConfig["name"] = user.UUID[:8]
			userConfig["password"] = user.UUID
		case model.NodeTypeNaive:
			userConfig["username"] = user.UUID[:8]
			userConfig["password"] = user.UUID
		default:
			userConfig["name"] = user.UUID[:8]
			userConfig["password"] = user.UUID
		}

		result = append(result, userConfig)
	}

	return result, nil
}

// generateSS2022Password 生成 SS2022 密码
func (s *HostService) generateSS2022Password(node *model.ServerNode, user *model.User) string {
	return s.generateSS2022PasswordWithConfig(node.ProtocolSettings, node.CreatedAt, user)
}

// generateSS2022PasswordWithConfig 根据配置生成 SS2022 密码
// 返回格式: serverKey:userKey (用于客户端订告
func (s *HostService) generateSS2022PasswordWithConfig(ps model.JSONMap, createdAt int64, user *model.User) string {
	cipher := ""
	if c, ok := ps["method"].(string); ok {
		cipher = c
	} else if c, ok := ps["cipher"].(string); ok {
		cipher = c
	}

	return utils.GenerateSS2022Password(cipher, createdAt, user.UUID)
}

// getSS2022UserKey 获取 SS2022 用户密钥 (用于服务端用户列告
func (s *HostService) getSS2022UserKey(ps model.JSONMap, user *model.User) string {
	cipher := ""
	if c, ok := ps["method"].(string); ok {
		cipher = c
	} else if c, ok := ps["cipher"].(string); ok {
		cipher = c
	}

	return utils.GetSS2022UserPassword(cipher, user.UUID)
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
				"server_name": "addons.mozilla.org",
				"reality": map[string]interface{}{
					"enabled": true,
					"handshake": map[string]interface{}{
						"server":      "addons.mozilla.org",
						"server_port": 443,
					},
					"private_key": "", // Agent 自动生成
					"short_id":    []string{"0123456789abcdef"},
				},
			},
		}
	case model.NodeTypeVMess:
		return map[string]interface{}{
			"name":        "VMess节点",
			"listen_port": 443,
			"protocol_settings": map[string]interface{}{
				"security": "auto",
			},
			"transport_settings": map[string]interface{}{
				"type": "ws",
				"path": "/vmess",
				"headers": map[string]interface{}{
					"Host": "",
				},
			},
			"tls_settings": map[string]interface{}{
				"enabled":     false,
				"server_name": "",
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
	case model.NodeTypeTUIC:
		return map[string]interface{}{
			"name":        "TUIC节点",
			"listen_port": 443,
			"protocol_settings": map[string]interface{}{
				"congestion_control": "bbr",
			},
			"tls_settings": map[string]interface{}{
				"enabled": true,
				"alpn":    []string{"h3"},
			},
		}
	case model.NodeTypeAnyTLS:
		return map[string]interface{}{
			"name":        "AnyTLS节点",
			"listen_port": 443,
			"protocol_settings": map[string]interface{}{
				"padding_scheme": []interface{}{},
			},
			"tls_settings": map[string]interface{}{
				"enabled": true,
			},
		}
	case model.NodeTypeShadowTLS:
		return map[string]interface{}{
			"name":        "ShadowTLS节点",
			"listen_port": 443,
			"protocol_settings": map[string]interface{}{
				"version":          3,
				"handshake_server": "addons.mozilla.org",
				"handshake_port":   443,
				"strict_mode":      true,
				"detour_method":    "2022-blake3-aes-128-gcm",
			},
		}
	case model.NodeTypeNaive:
		return map[string]interface{}{
			"name":        "NaiveProxy节点",
			"listen_port": 443,
			"tls_settings": map[string]interface{}{
				"enabled": true,
				"acme": map[string]interface{}{
					"domain": "",
					"email":  "",
				},
			},
		}
	default:
		return map[string]interface{}{}
	}
}

// GetAllNodes 获取所有节告
func (s *HostService) GetAllNodes() ([]model.ServerNode, error) {
	return s.nodeRepo.GetAll()
}

// AgentConfig Agent 配置
type AgentConfig struct {
	SingBoxConfig map[string]interface{} `json:"singbox_config"`
	Nodes         []AgentNodeConfig      `json:"nodes"`
}

// AgentNodeConfig Agent 节点配置
type AgentNodeConfig struct {
	ID    int64                    `json:"id"`
	Type  string                   `json:"type"`
	Port  int                      `json:"port"`
	Tag   string                   `json:"tag"`
	Users []map[string]interface{} `json:"users"`
}

// GetAgentConfig 获取 Agent 完整配置
func (s *HostService) GetAgentConfig(hostID int64) (*AgentConfig, error) {
	config, err := s.GenerateSingBoxConfig(hostID)
	if err != nil {
		return nil, err
	}

	nodeConfigs := make([]AgentNodeConfig, 0)
	processedServerIDs := make(map[int64]bool) // 记录已处理的 Server ID，避免重告

	// 1. 从绑定到主机告Server 获取配置
	servers, err := s.serverRepo.GetByHostID(hostID)
	if err == nil {
		for _, server := range servers {
			if processedServerIDs[server.ID] {
				continue
			}
			users, _ := s.GetUsersForServer(&server)
			nodeConfigs = append(nodeConfigs, AgentNodeConfig{
				ID:    server.ID,
				Type:  server.Type,
				Port:  server.ServerPort,
				Tag:   server.Type + "-in-" + fmt.Sprintf("%d", server.ID),
				Users: users,
			})
			processedServerIDs[server.ID] = true
		}
	}

	// 2. 告ServerNode 获取配置（兼容旧逻辑告
	nodes, err := s.nodeRepo.FindByHostID(hostID)
	if err == nil {
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
	}

	return &AgentConfig{
		SingBoxConfig: config,
		Nodes:         nodeConfigs,
	}, nil
}

// GetUsersForServer 获取 Server 可用的用户列告
func (s *HostService) GetUsersForServer(server *model.Server) ([]map[string]interface{}, error) {
	groupIDs := server.GetGroupIDsAsInt64()

	var users []model.User
	var err error

	if len(groupIDs) == 0 {
		users, err = s.userRepo.GetAllAvailableUsers()
	} else {
		users, err = s.userRepo.GetAvailableUsers(groupIDs)
	}

	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		userConfig := map[string]interface{}{}

		// sing-box 不同协议的用户字段不告
		switch server.Type {
		case model.ServerTypeShadowsocks:
			userConfig["name"] = user.UUID[:8]
			userConfig["password"] = s.getSS2022UserKeyForServer(server, &user)
		case model.ServerTypeVmess, model.ServerTypeVless:
			userConfig["name"] = user.UUID[:8]
			userConfig["uuid"] = user.UUID
		case model.ServerTypeTrojan, model.ServerTypeHysteria, model.ServerTypeTuic:
			userConfig["name"] = user.UUID[:8]
			userConfig["password"] = user.UUID
		default:
			userConfig["name"] = user.UUID[:8]
			userConfig["password"] = user.UUID
		}

		result = append(result, userConfig)
	}

	return result, nil
}

// getSS2022UserKeyForServer 获取 Server 告SS2022 用户密钥 (仅用户密钥，用于服务告
func (s *HostService) getSS2022UserKeyForServer(server *model.Server, user *model.User) string {
	cipher := ""
	if c, ok := server.ProtocolSettings["method"].(string); ok {
		cipher = c
	} else if c, ok := server.ProtocolSettings["cipher"].(string); ok {
		cipher = c
	}

	return utils.GetSS2022UserPassword(cipher, user.UUID)
}

// ToJSON 转换告JSON
func (c *AgentConfig) ToJSON() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}

// parseSocksOutbound 解析 SOCKS 出口配置
// 支持格式告
//   - socks5://host:port
//   - socks5://user:pass@host:port
func (s *HostService) parseSocksOutbound(socksURL string) map[string]interface{} {
	// 简单解告SOCKS URL
	// 格式：socks5://[user:pass@]host:port

	if socksURL == "" {
		return nil
	}

	// 移除协议前缀
	socksURL = strings.TrimPrefix(socksURL, "socks5://")
	socksURL = strings.TrimPrefix(socksURL, "socks://")

	outbound := map[string]interface{}{
		"type": "socks",
		"tag":  "socks-out",
	}

	// 检查是否有认证信息
	var server string
	if strings.Contains(socksURL, "@") {
		parts := strings.SplitN(socksURL, "@", 2)
		auth := parts[0]
		server = parts[1]

		// 解析用户名和密码
		if strings.Contains(auth, ":") {
			authParts := strings.SplitN(auth, ":", 2)
			outbound["username"] = authParts[0]
			outbound["password"] = authParts[1]
		}
	} else {
		server = socksURL
	}

	// 解析服务器地址和端告
	if strings.Contains(server, ":") {
		parts := strings.SplitN(server, ":", 2)
		outbound["server"] = parts[0]
		outbound["server_port"], _ = strconv.Atoi(parts[1])
	} else {
		outbound["server"] = server
		outbound["server_port"] = 1080 // 默认端口
	}

	// 添加版本
	outbound["version"] = "5"

	return outbound
}

func generateHostToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
