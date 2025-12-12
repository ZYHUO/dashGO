package service

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"dashgo/internal/config"
	"dashgo/internal/model"
	"dashgo/internal/repository"
	"dashgo/pkg/cache"
	"dashgo/pkg/utils"
)

type ServerService struct {
	serverRepo *repository.ServerRepository
	userRepo   *repository.UserRepository
	cache      *cache.Client
	cfg        *config.Config
}

func NewServerService(serverRepo *repository.ServerRepository, userRepo *repository.UserRepository, cache *cache.Client, cfg *config.Config) *ServerService {
	return &ServerService{
		serverRepo: serverRepo,
		userRepo:   userRepo,
		cache:      cache,
		cfg:        cfg,
	}
}

// GetAllServers 获取所有服务器
func (s *ServerService) GetAllServers() ([]model.Server, error) {
	return s.serverRepo.GetAllServers()
}

// GetAvailableServers 获取用户可用的服务器列表
func (s *ServerService) GetAvailableServers(user *model.User) ([]ServerInfo, error) {
	var servers []model.Server
	var err error

	if user.GroupID != nil {
		servers, err = s.serverRepo.GetAvailableServers(*user.GroupID)
	} else {
		// 没有用户组的用户获取所有公开节点
		servers, err = s.serverRepo.GetPublicServers()
	}

	if err != nil {
		return nil, err
	}

	result := make([]ServerInfo, 0, len(servers))
	for _, server := range servers {
		info := s.BuildServerInfo(&server, user)
		result = append(result, info)
	}
	return result, nil
}

// ServerInfo 服务器信息（包含用户密码等）
type ServerInfo struct {
	model.Server
	Password string `json:"password"`
	Ports    string `json:"ports,omitempty"`
}

func (s *ServerService) BuildServerInfo(server *model.Server, user *model.User) ServerInfo {
	info := ServerInfo{
		Server: *server,
	}

	// 处理端口范围
	if strings.Contains(server.Port, "-") {
		info.Ports = server.Port
		port := utils.RandomPort(server.Port)
		info.Port = strconv.Itoa(port)
	}

	// 生成用户密码
	info.Password = s.generateServerPassword(server, user)

	return info
}

// generateServerPassword 生成服务器密�?(用于客户端订�?
func (s *ServerService) generateServerPassword(server *model.Server, user *model.User) string {
	if server.Type != model.ServerTypeShadowsocks {
		return user.UUID
	}

	// Shadowsocks 2022 cipher
	cipher := ""
	if ps, ok := server.ProtocolSettings["cipher"]; ok {
		cipher, _ = ps.(string)
	}
	if cipher == "" {
		if ps, ok := server.ProtocolSettings["method"]; ok {
			cipher, _ = ps.(string)
		}
	}

	// 使用统一的密码生成函�?
	return utils.GenerateSS2022Password(cipher, server.CreatedAt, user.UUID)
}

// GetAvailableUsers 获取节点可用的用户列�?
func (s *ServerService) GetAvailableUsers(server *model.Server) ([]NodeUser, error) {
	groupIDs := server.GetGroupIDsAsInt64()
	if len(groupIDs) == 0 {
		return []NodeUser{}, nil
	}

	users, err := s.userRepo.GetAvailableUsers(groupIDs)
	if err != nil {
		return nil, err
	}

	result := make([]NodeUser, 0, len(users))
	for _, user := range users {
		nodeUser := NodeUser{
			ID:          user.ID,
			UUID:        user.UUID,
			SpeedLimit:  user.SpeedLimit,
			DeviceLimit: user.DeviceLimit,
		}
		result = append(result, nodeUser)
	}
	return result, nil
}

// NodeUser 节点用户信息
type NodeUser struct {
	ID          int64 `json:"id"`
	UUID        string `json:"uuid"`
	SpeedLimit  *int   `json:"speed_limit,omitempty"`
	DeviceLimit *int   `json:"device_limit,omitempty"`
}

// GetServerConfig 获取节点配置（供节点端调用）
func (s *ServerService) GetServerConfig(server *model.Server) map[string]interface{} {
	ps := server.ProtocolSettings
	config := map[string]interface{}{
		"protocol":    server.Type,
		"listen_ip":   "0.0.0.0",
		"server_port": server.ServerPort,
		"network":     ps["network"],
	}

	if ns, ok := ps["network_settings"]; ok {
		config["networkSettings"] = ns
	}

	switch server.Type {
	case model.ServerTypeShadowsocks:
		config["cipher"] = ps["cipher"]
		if plugin, ok := ps["plugin"]; ok {
			config["plugin"] = plugin
		}
		if pluginOpts, ok := ps["plugin_opts"]; ok {
			config["plugin_opts"] = pluginOpts
		}
		// Server key for 2022 ciphers
		cipher, _ := ps["cipher"].(string)
		switch cipher {
		case "2022-blake3-aes-128-gcm":
			config["server_key"] = utils.GetServerKey(server.CreatedAt, 16)
		case "2022-blake3-aes-256-gcm":
			config["server_key"] = utils.GetServerKey(server.CreatedAt, 32)
		}

	case model.ServerTypeVmess:
		config["tls"] = ps["tls"]

	case model.ServerTypeTrojan:
		config["host"] = server.Host
		config["server_name"] = ps["server_name"]

	case model.ServerTypeVless:
		config["tls"] = ps["tls"]
		config["flow"] = ps["flow"]
		tls, _ := ps["tls"].(float64)
		if int(tls) == 2 {
			config["tls_settings"] = ps["reality_settings"]
		} else {
			config["tls_settings"] = ps["tls_settings"]
		}

	case model.ServerTypeHysteria:
		config["version"] = ps["version"]
		config["host"] = server.Host
		if tls, ok := ps["tls"].(map[string]interface{}); ok {
			config["server_name"] = tls["server_name"]
		}
		if bw, ok := ps["bandwidth"].(map[string]interface{}); ok {
			config["up_mbps"] = bw["up"]
			config["down_mbps"] = bw["down"]
		}
		if obfs, ok := ps["obfs"].(map[string]interface{}); ok {
			if open, _ := obfs["open"].(bool); open {
				config["obfs"] = obfs["type"]
				config["obfs-password"] = obfs["password"]
			}
		}

	case model.ServerTypeTuic:
		config["version"] = ps["version"]
		if tls, ok := ps["tls"].(map[string]interface{}); ok {
			config["server_name"] = tls["server_name"]
		}
		config["congestion_control"] = ps["congestion_control"]
	}

	config["base_config"] = map[string]interface{}{
		"push_interval": s.cfg.Node.PushInterval,
		"pull_interval": s.cfg.Node.PullInterval,
	}

	return config
}

// UpdateServerStatus 更新节点状�?
func (s *ServerService) UpdateServerStatus(serverID int64, serverType string, statusType string) error {
	key := ""
	switch statusType {
	case "check":
		key = cache.ServerLastCheckAtKey(strings.ToUpper(serverType), serverID)
	case "push":
		key = cache.ServerLastPushAtKey(strings.ToUpper(serverType), serverID)
	}
	return s.cache.Set(key, time.Now().Unix(), time.Hour)
}

// UpdateOnlineUsers 更新在线用户�?
func (s *ServerService) UpdateOnlineUsers(serverID int64, serverType string, count int) error {
	key := cache.ServerOnlineUserKey(strings.ToUpper(serverType), serverID)
	return s.cache.Set(key, count, time.Hour)
}

// UpdateLoadStatus 更新节点负载状�?
func (s *ServerService) UpdateLoadStatus(serverID int64, serverType string, status map[string]interface{}) error {
	key := cache.ServerLoadStatusKey(strings.ToUpper(serverType), serverID)
	data, _ := json.Marshal(status)
	return s.cache.Set(key, string(data), time.Hour)
}

// FindServer 查找服务�?
func (s *ServerService) FindServer(serverID int64, serverType string) (*model.Server, error) {
	if serverType != "" {
		return s.serverRepo.FindByCode(serverType, strconv.FormatInt(serverID, 10))
	}
	return s.serverRepo.FindByID(serverID)
}

// CreateServer 创建服务�?
func (s *ServerService) CreateServer(server *model.Server) error {
	return s.serverRepo.Create(server)
}

// UpdateServer 更新服务�?
func (s *ServerService) UpdateServer(server *model.Server) error {
	return s.serverRepo.Update(server)
}

// DeleteServer 删除服务�?
func (s *ServerService) DeleteServer(id int64) error {
	return s.serverRepo.Delete(id)
}
