package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"xboard/internal/config"
	"xboard/internal/model"
	"xboard/internal/repository"
)

// NodeSyncService 节点同步服务 - 与 sing-box SSMAPI 对接
type NodeSyncService struct {
	serverRepo *repository.ServerRepository
	userRepo   *repository.UserRepository
	statRepo   *repository.StatRepository
	cfg        *config.Config
	httpClient *http.Client
}

func NewNodeSyncService(
	serverRepo *repository.ServerRepository,
	userRepo *repository.UserRepository,
	statRepo *repository.StatRepository,
	cfg *config.Config,
) *NodeSyncService {
	return &NodeSyncService{
		serverRepo: serverRepo,
		userRepo:   userRepo,
		statRepo:   statRepo,
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SSMAPIUser sing-box SSMAPI 用户结构
type SSMAPIUser struct {
	UserName        string `json:"username"`
	Password        string `json:"uPSK,omitempty"`
	DownlinkBytes   int64  `json:"downlinkBytes"`
	UplinkBytes     int64  `json:"uplinkBytes"`
	DownlinkPackets int64  `json:"downlinkPackets"`
	UplinkPackets   int64  `json:"uplinkPackets"`
	TCPSessions     int64  `json:"tcpSessions"`
	UDPSessions     int64  `json:"udpSessions"`
}

// SSMAPIStats sing-box SSMAPI 统计结构
type SSMAPIStats struct {
	UplinkBytes     int64         `json:"uplinkBytes"`
	DownlinkBytes   int64         `json:"downlinkBytes"`
	UplinkPackets   int64         `json:"uplinkPackets"`
	DownlinkPackets int64         `json:"downlinkPackets"`
	TCPSessions     int64         `json:"tcpSessions"`
	UDPSessions     int64         `json:"udpSessions"`
	Users           []*SSMAPIUser `json:"users"`
}

// SSMAPIServerInfo sing-box 服务器信息
type SSMAPIServerInfo struct {
	Server     string `json:"server"`
	APIVersion string `json:"apiVersion"`
}

// NodeEndpoint 节点端点配置
type NodeEndpoint struct {
	Server      *model.Server
	BaseURL     string // SSMAPI 基础 URL，如 http://node:9000/ss
	BearerToken string // 可选的认证令牌
}

// GetServerInfo 获取服务器信息
func (s *NodeSyncService) GetServerInfo(endpoint NodeEndpoint) (*SSMAPIServerInfo, error) {
	url := fmt.Sprintf("%s/server/v1/", endpoint.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	s.setAuthHeader(req, endpoint)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var info SSMAPIServerInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

// ListUsers 获取节点上的用户列表
func (s *NodeSyncService) ListUsers(endpoint NodeEndpoint) ([]*SSMAPIUser, error) {
	url := fmt.Sprintf("%s/server/v1/users", endpoint.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	s.setAuthHeader(req, endpoint)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Users []*SSMAPIUser `json:"users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Users, nil
}

// AddUser 添加用户到节点
func (s *NodeSyncService) AddUser(endpoint NodeEndpoint, username, password string) error {
	url := fmt.Sprintf("%s/server/v1/users", endpoint.BaseURL)

	body, _ := json.Marshal(map[string]string{
		"username": username,
		"uPSK":     password,
	})

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	s.setAuthHeader(req, endpoint)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetUser 获取单个用户信息
func (s *NodeSyncService) GetUser(endpoint NodeEndpoint, username string) (*SSMAPIUser, error) {
	url := fmt.Sprintf("%s/server/v1/users/%s", endpoint.BaseURL, username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	s.setAuthHeader(req, endpoint)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var user SSMAPIUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser 更新用户密码
func (s *NodeSyncService) UpdateUser(endpoint NodeEndpoint, username, password string) error {
	url := fmt.Sprintf("%s/server/v1/users/%s", endpoint.BaseURL, username)

	body, _ := json.Marshal(map[string]string{
		"uPSK": password,
	})

	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	s.setAuthHeader(req, endpoint)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteUser 从节点删除用户
func (s *NodeSyncService) DeleteUser(endpoint NodeEndpoint, username string) error {
	url := fmt.Sprintf("%s/server/v1/users/%s", endpoint.BaseURL, username)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	s.setAuthHeader(req, endpoint)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetStats 获取流量统计
func (s *NodeSyncService) GetStats(endpoint NodeEndpoint, clear bool) (*SSMAPIStats, error) {
	url := fmt.Sprintf("%s/server/v1/stats", endpoint.BaseURL)
	if clear {
		url += "?clear=true"
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	s.setAuthHeader(req, endpoint)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var stats SSMAPIStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// SyncUsers 同步用户到节点
func (s *NodeSyncService) SyncUsers(endpoint NodeEndpoint) error {
	// 获取节点可用用户
	groupIDs := endpoint.Server.GetGroupIDsAsInt64()
	if len(groupIDs) == 0 {
		return nil
	}

	users, err := s.userRepo.GetAvailableUsers(groupIDs)
	if err != nil {
		return fmt.Errorf("get available users: %w", err)
	}

	// 获取当前节点上的用户
	currentUsers, err := s.ListUsers(endpoint)
	if err != nil {
		return fmt.Errorf("list node users: %w", err)
	}

	// 构建用户映射
	currentUserMap := make(map[string]*SSMAPIUser)
	for _, u := range currentUsers {
		currentUserMap[u.UserName] = u
	}

	expectedUserMap := make(map[string]*model.User)
	for i := range users {
		expectedUserMap[users[i].UUID] = &users[i]
	}

	// 添加新用户
	for uuid, user := range expectedUserMap {
		password := s.generatePassword(endpoint.Server, user)
		if existing, exists := currentUserMap[uuid]; exists {
			// 用户已存在，检查密码是否需要更新
			if existing.Password != password {
				if err := s.UpdateUser(endpoint, uuid, password); err != nil {
					log.Printf("[NodeSync] Failed to update user %s: %v", uuid, err)
				}
			}
		} else {
			// 添加新用户
			if err := s.AddUser(endpoint, uuid, password); err != nil {
				log.Printf("[NodeSync] Failed to add user %s: %v", uuid, err)
			}
		}
	}

	// 删除不存在的用户
	for uuid := range currentUserMap {
		if _, exists := expectedUserMap[uuid]; !exists {
			if err := s.DeleteUser(endpoint, uuid); err != nil {
				log.Printf("[NodeSync] Failed to delete user %s: %v", uuid, err)
			}
		}
	}

	return nil
}

// FetchAndProcessTraffic 获取并处理流量数据
func (s *NodeSyncService) FetchAndProcessTraffic(endpoint NodeEndpoint) error {
	stats, err := s.GetStats(endpoint, true)
	if err != nil {
		return fmt.Errorf("get stats: %w", err)
	}

	server := endpoint.Server
	rate := server.Rate
	if rate <= 0 {
		rate = 1
	}

	// 更新用户流量
	for _, userStat := range stats.Users {
		if userStat.UplinkBytes == 0 && userStat.DownlinkBytes == 0 {
			continue
		}

		user, err := s.userRepo.FindByUUID(userStat.UserName)
		if err != nil {
			continue // 用户不存在，跳过
		}

		// 应用倍率
		u := int64(float64(userStat.UplinkBytes) * rate)
		d := int64(float64(userStat.DownlinkBytes) * rate)

		// 更新用户流量
		if err := s.userRepo.UpdateTraffic(user.ID, u, d); err != nil {
			log.Printf("[NodeSync] Failed to update user traffic: %v", err)
		}

		// 记录统计
		if err := s.statRepo.RecordUserTraffic(user.ID, rate, u, d, "d"); err != nil {
			log.Printf("[NodeSync] Failed to record user traffic: %v", err)
		}
	}

	// 记录节点统计
	totalU := int64(float64(stats.UplinkBytes) * rate)
	totalD := int64(float64(stats.DownlinkBytes) * rate)
	if err := s.statRepo.RecordServerTraffic(server.ID, server.Type, totalU, totalD, "d"); err != nil {
		log.Printf("[NodeSync] Failed to record server traffic: %v", err)
	}

	return nil
}

// generatePassword 生成用户密码
func (s *NodeSyncService) generatePassword(server *model.Server, user *model.User) string {
	// 对于大多数协议，直接使用 UUID
	return user.UUID
}

// setAuthHeader 设置认证头
func (s *NodeSyncService) setAuthHeader(req *http.Request, endpoint NodeEndpoint) {
	if endpoint.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+endpoint.BearerToken)
	}
}

// StartSyncLoop 启动同步循环
// 注意：此功能已禁用，因为新架构使用 Agent 模式
// Agent 会主动向面板上报流量和用户同步
// 如果你的节点支持 SSMAPI 并且需要面板主动同步，可以在配置中启用
func (s *NodeSyncService) StartSyncLoop() {
	// 检查是否启用节点同步
	if !s.cfg.Node.EnableSync {
		log.Println("[NodeSync] Node sync is disabled, using Agent mode instead")
		return
	}

	// 用户同步间隔
	syncTicker := time.NewTicker(time.Duration(s.cfg.Node.PullInterval) * time.Second)
	// 流量获取间隔
	trafficTicker := time.NewTicker(time.Duration(s.cfg.Node.PushInterval) * time.Second)

	go func() {
		for {
			select {
			case <-syncTicker.C:
				s.syncAllNodes()
			case <-trafficTicker.C:
				s.fetchAllTraffic()
			}
		}
	}()
}

// syncAllNodes 同步所有节点
func (s *NodeSyncService) syncAllNodes() {
	servers, err := s.serverRepo.GetAllServers()
	if err != nil {
		log.Printf("[NodeSync] Failed to get servers: %v", err)
		return
	}

	for _, server := range servers {
		endpoint := s.buildEndpoint(&server)
		if endpoint.BaseURL == "" {
			continue
		}

		if err := s.SyncUsers(endpoint); err != nil {
			log.Printf("[NodeSync] Failed to sync users for server %s: %v", server.Name, err)
		}
	}
}

// fetchAllTraffic 获取所有节点流量
func (s *NodeSyncService) fetchAllTraffic() {
	servers, err := s.serverRepo.GetAllServers()
	if err != nil {
		log.Printf("[NodeSync] Failed to get servers: %v", err)
		return
	}

	for _, server := range servers {
		endpoint := s.buildEndpoint(&server)
		if endpoint.BaseURL == "" {
			continue
		}

		if err := s.FetchAndProcessTraffic(endpoint); err != nil {
			log.Printf("[NodeSync] Failed to fetch traffic for server %s: %v", server.Name, err)
		}
	}
}

// buildEndpoint 构建节点端点
func (s *NodeSyncService) buildEndpoint(server *model.Server) NodeEndpoint {
	endpoint := NodeEndpoint{
		Server: server,
	}

	// 从 protocol_settings 中获取 SSMAPI 配置
	if ps := server.ProtocolSettings; ps != nil {
		if apiURL, ok := ps["ssmapi_url"].(string); ok {
			endpoint.BaseURL = apiURL
		}
		if token, ok := ps["ssmapi_token"].(string); ok {
			endpoint.BearerToken = token
		}
	}

	// 如果没有配置，使用默认值
	if endpoint.BaseURL == "" {
		// 默认使用 http://host:9000/协议类型
		endpoint.BaseURL = fmt.Sprintf("http://%s:9000/%s", server.Host, server.Type)
	}

	return endpoint
}

// GetNodeStatus 获取节点状态
func (s *NodeSyncService) GetNodeStatus(server *model.Server) (map[string]interface{}, error) {
	// 简化实现：直接返回基本状态，不实际连接节点
	// 因为新架构使用 Agent 模式，节点状态由 Agent 心跳上报
	return map[string]interface{}{
		"online": true,
		"stats": map[string]interface{}{
			"uplink_bytes":   0,
			"downlink_bytes": 0,
			"tcp_sessions":   0,
			"udp_sessions":   0,
		},
	}, nil
}
