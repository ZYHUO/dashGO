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
			node, err := services.Host.GetNodeByID(nodeData.ID)
			if err != nil {
				continue
			}
			for _, userData := range nodeData.Users {
				if userData.Upload == 0 && userData.Download == 0 {
					continue
				}
				user, err := services.User.GetByUUID(userData.Username)
				if err != nil {
					continue
				}
				u := int64(float64(userData.Upload) * node.Rate)
				d := int64(float64(userData.Download) * node.Rate)
				services.User.UpdateTraffic(user.ID, u, d)
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

func getHostFromContext(c *gin.Context) *model.Host {
	host, exists := c.Get("host")
	if !exists {
		return nil
	}
	return host.(*model.Host)
}
