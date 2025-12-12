package handler

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"xboard/internal/model"
	"xboard/internal/service"

	"github.com/gin-gonic/gin"
)

// AgentAuth Agent 认证中间件
func AgentAuth(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		host, err := services.Host.GetByToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("host", host)
		c.Next()
	}
}

// AgentHeartbeat Agent 心跳
func AgentHeartbeat(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := getHostFromContext(c)
		if host == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req struct {
			SystemInfo map[string]interface{} `json:"system_info"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Host.UpdateHeartbeat(host.ID, c.ClientIP(), req.SystemInfo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	}
}

// AgentGetConfig 获取配置
func AgentGetConfig(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := getHostFromContext(c)
		if host == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		config, err := services.Host.GetAgentConfig(host.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": config})
	}
}

// AgentGetUsers 获取节点用户（支持增量同步）
// 注意：此接口返回的是 sing-box 格式的用户配置，包含 name 和 password
func AgentGetUsers(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := getHostFromContext(c)
		if host == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		nodeID, _ := strconv.ParseInt(c.Query("node_id"), 10, 64)
		nodeType := c.Query("type") // server 或 node
		lastHash := c.Query("hash")

		if nodeID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "node_id required"})
			return
		}

		var users []map[string]interface{}

		// 根据类型获取用户
		if nodeType == "server" {
			// 从 Server 获取用户
			server, err := services.Server.FindServer(nodeID, "")
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
				return
			}
			// 验证 Server 属于该主机
			if server.HostID == nil || *server.HostID != host.ID {
				c.JSON(http.StatusForbidden, gin.H{"error": "server not belong to this host"})
				return
			}
			var userErr error
			users, userErr = services.Host.GetUsersForServer(server)
			if userErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": userErr.Error()})
				return
			}
		} else {
			// 从 ServerNode 获取用户
			node, nodeErr := services.Host.GetNodeByID(nodeID)
			if nodeErr != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
				return
			}
			// 验证节点属于该主机
			if node.HostID != host.ID {
				c.JSON(http.StatusForbidden, gin.H{"error": "node not belong to this host"})
				return
			}
			var userErr error
			users, userErr = services.Host.GetUsersForNode(node)
			if userErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": userErr.Error()})
				return
			}
		}

		// 计算哈希
		usersJSON, _ := json.Marshal(users)
		currentHash := fmt.Sprintf("%x", md5.Sum(usersJSON))

		// 如果哈希相同，返回无变化
		if lastHash != "" && currentHash == lastHash {
			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"hash":       currentHash,
					"has_change": false,
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"hash":       currentHash,
				"has_change": true,
				"users":      users,
			},
		})
	}
}

// AgentSyncStatus Agent 同步状态
func AgentSyncStatus(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := getHostFromContext(c)
		if host == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req struct {
			Nodes []struct {
				ID          int64 `json:"id"`
				OnlineUsers int   `json:"online_users"`
				Status      struct {
					CPU    float64 `json:"cpu"`
					Memory float64 `json:"memory"`
					Disk   float64 `json:"disk"`
				} `json:"status"`
			} `json:"nodes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 更新节点状态
		for _, nodeData := range req.Nodes {
			services.Server.UpdateOnlineUsers(nodeData.ID, "", nodeData.OnlineUsers)
			services.Server.UpdateLoadStatus(nodeData.ID, "", map[string]interface{}{
				"cpu":    nodeData.Status.CPU,
				"memory": nodeData.Status.Memory,
				"disk":   nodeData.Status.Disk,
			})
		}

		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	}
}

// AgentReportTraffic 上报流量
func AgentReportTraffic(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := getHostFromContext(c)
		if host == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req struct {
			Nodes []struct {
				ID    int64 `json:"id"`
				Users []struct {
					Username string `json:"username"`
					Upload   int64  `json:"upload"`
					Download int64  `json:"download"`
				} `json:"users"`
			} `json:"nodes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 处理流量
		for _, nodeData := range req.Nodes {
			// 获取节点信息和倍率
			var rate float64 = 1.0
			var serverType string = "unknown"
			var serverID int64 = nodeData.ID
			
			// 尝试从 Server 获取
			server, err := services.Server.FindServer(nodeData.ID, "")
			if err == nil && server != nil {
				rate = server.Rate
				serverType = server.Type
			} else {
				// 尝试从 ServerNode 获取
				node, err := services.Host.GetNodeByID(nodeData.ID)
				if err == nil && node != nil {
					rate = node.Rate
					serverType = node.Type
				}
			}

			// 处理每个用户的流量
			for _, userData := range nodeData.Users {
				if userData.Upload == 0 && userData.Download == 0 {
					continue
				}
				
				// Username 是 UUID 的前8位，使用前缀匹配
				user, err := services.User.GetByUUIDPrefix(userData.Username)
				if err != nil {
					continue
				}
				
				// 应用倍率
				u := int64(float64(userData.Upload) * rate)
				d := int64(float64(userData.Download) * rate)
				
				// 更新用户流量
				services.User.UpdateTraffic(user.ID, u, d)
				
				// 记录用户流量统计（日统计）
				services.NodeSync.RecordUserTrafficStat(user.ID, rate, u, d)
				
				// 记录流量日志
				services.NodeSync.RecordTrafficLog(user.ID, serverID, u, d, rate)
			}
			
			// 计算节点总流量
			var totalU, totalD int64
			for _, userData := range nodeData.Users {
				totalU += int64(float64(userData.Upload) * rate)
				totalD += int64(float64(userData.Download) * rate)
			}
			
			// 记录节点流量统计（日统计）
			if totalU > 0 || totalD > 0 {
				services.NodeSync.RecordServerTrafficStat(serverID, serverType, totalU, totalD)
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	}
}

// AdminListHosts 获取主机列表
func AdminListHosts(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		hosts, err := services.Host.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": hosts})
	}
}

// AdminCreateHost 创建主机
func AdminCreateHost(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		host, err := services.Host.CreateHost(req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": host})
	}
}

// AdminDeleteHost 删除主机
func AdminDeleteHost(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		if err := services.Host.Delete(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminResetHostToken 重置主机 Token
func AdminResetHostToken(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		token, err := services.Host.ResetToken(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"token": token}})
	}
}

// AdminUpdateHost 更新主机配置
func AdminUpdateHost(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		host, err := services.Host.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
			return
		}

		var req struct {
			Name          *string `json:"name"`
			SocksOutbound *string `json:"socks_outbound"` // SOCKS5 出口代理，格式：socks5://user:pass@host:port
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 更新字段
		if req.Name != nil {
			host.Name = *req.Name
		}
		if req.SocksOutbound != nil {
			host.SocksOutbound = req.SocksOutbound
		}

		if err := services.Host.UpdateHost(host); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": host})
	}
}

// AdminListNodes 获取节点列表
func AdminListNodes(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostID, _ := strconv.ParseInt(c.Query("host_id"), 10, 64)
		var nodes []model.ServerNode
		var err error

		if hostID > 0 {
			nodes, err = services.Host.GetNodesByHostID(hostID)
		} else {
			nodes, err = services.Host.GetAllNodes()
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": nodes})
	}
}

// AdminCreateNode 创建节点
func AdminCreateNode(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var node model.ServerNode
		if err := c.ShouldBindJSON(&node); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Host.CreateNode(&node); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": node})
	}
}

// AdminUpdateNode 更新节点
func AdminUpdateNode(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		node, err := services.Host.GetNodeByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
			return
		}

		if err := c.ShouldBindJSON(node); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Host.UpdateNode(node); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": node})
	}
}

// AdminDeleteNode 删除节点
func AdminDeleteNode(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		if err := services.Host.DeleteNode(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminGetHostConfig 获取主机配置预览
func AdminGetHostConfig(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		config, err := services.Host.GetAgentConfig(hostID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": config})
	}
}

// AdminGetDefaultNodeConfig 获取默认节点配置
func AdminGetDefaultNodeConfig(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		nodeType := c.Query("type")
		if nodeType == "" {
			nodeType = "shadowsocks"
		}
		config := services.Host.GetDefaultNodeConfig(nodeType)
		c.JSON(http.StatusOK, gin.H{"data": config})
	}
}

// AgentGetVersion 获取 Agent 版本信息
func AgentGetVersion(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := getHostFromContext(c)
		if host == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// 获取当前 Agent 版本（从请求头或查询参数）
		currentVersion := c.GetHeader("X-Agent-Version")
		if currentVersion == "" {
			currentVersion = c.Query("version")
		}

		// 从数据库获取最新版本信息
		version, err := services.AgentVersion.GetLatestVersion()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		versionInfo := gin.H{
			"latest_version": version.Version,
			"download_url":   version.DownloadURL,
			"sha256":         version.SHA256,
			"file_size":      version.FileSize,
			"strategy":       version.Strategy,
			"release_notes":  version.ReleaseNotes,
		}

		c.JSON(http.StatusOK, gin.H{"data": versionInfo})
	}
}

// AgentUpdateStatus 接收 Agent 更新状态通知
func AgentUpdateStatus(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := getHostFromContext(c)
		if host == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req struct {
			FromVersion  string `json:"from_version" binding:"required"`
			ToVersion    string `json:"to_version" binding:"required"`
			Status       string `json:"status" binding:"required"` // success, failed, rollback
			ErrorMessage string `json:"error_message"`
			Timestamp    string `json:"timestamp" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 记录更新日志到数据库
		log := &model.AgentUpdateLog{
			HostID:       host.ID,
			FromVersion:  req.FromVersion,
			ToVersion:    req.ToVersion,
			Status:       req.Status,
			ErrorMessage: req.ErrorMessage,
		}

		if err := services.AgentVersion.RecordUpdateLog(log); err != nil {
			fmt.Printf("⚠️  Failed to record update log: %v\n", err)
		}

		// 打印日志
		if req.Status == "success" {
			fmt.Printf("✅ Host %d (%s) updated successfully: %s -> %s\n",
				host.ID, host.Name, req.FromVersion, req.ToVersion)
		} else if req.Status == "failed" {
			fmt.Printf("❌ Host %d (%s) update failed: %s -> %s, error: %s\n",
				host.ID, host.Name, req.FromVersion, req.ToVersion, req.ErrorMessage)
		} else if req.Status == "rollback" {
			fmt.Printf("🔄 Host %d (%s) rolled back: %s -> %s, reason: %s\n",
				host.ID, host.Name, req.FromVersion, req.ToVersion, req.ErrorMessage)
		}

		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	}
}

// AdminListAgentVersions 获取 Agent 版本列表
func AdminListAgentVersions(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		versions, total, err := services.AgentVersion.List(page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"items": versions,
				"total": total,
				"page":  page,
			},
		})
	}
}

// AdminCreateAgentVersion 创建 Agent 版本
func AdminCreateAgentVersion(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.AgentVersion
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.AgentVersion.Create(&req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": req})
	}
}

// AdminUpdateAgentVersion 更新 Agent 版本
func AdminUpdateAgentVersion(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		
		version, err := services.AgentVersion.GetByVersion("")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
			return
		}

		if err := c.ShouldBindJSON(version); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		version.ID = id
		if err := services.AgentVersion.Update(version); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": version})
	}
}

// AdminSetLatestAgentVersion 设置最新版本
func AdminSetLatestAgentVersion(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.AgentVersion.SetLatest(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	}
}

// AdminDeleteAgentVersion 删除 Agent 版本
func AdminDeleteAgentVersion(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.AgentVersion.Delete(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminListAgentUpdateLogs 获取 Agent 更新日志
func AdminListAgentUpdateLogs(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostID, _ := strconv.ParseInt(c.Query("host_id"), 10, 64)
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		logs, total, err := services.AgentVersion.GetUpdateLogs(hostID, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"items": logs,
				"total": total,
				"page":  page,
			},
		})
	}
}

func getHostFromContext(c *gin.Context) *model.Host {
	host, exists := c.Get("host")
	if !exists {
		return nil
	}
	return host.(*model.Host)
}
