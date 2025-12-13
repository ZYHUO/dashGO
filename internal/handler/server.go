package handler

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dashgo/internal/model"
	"dashgo/internal/service"

	"github.com/gin-gonic/gin"
)

// ServerConfig 获取节点配置
func ServerConfig(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		server := getServerFromContext(c)
		if server == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "server not found"})
			return
		}

		config := services.Server.GetServerConfig(server)

		// ETag 支持
		configJSON, _ := json.Marshal(config)
		etag := fmt.Sprintf(`"%x"`, sha1.Sum(configJSON))

		if c.GetHeader("If-None-Match") == etag {
			c.Status(http.StatusNotModified)
			return
		}

		c.Header("ETag", etag)
		c.JSON(http.StatusOK, config)
	}
}

// ServerUsers 获取可用用户列表
func ServerUsers(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		server := getServerFromContext(c)
		if server == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "server not found"})
			return
		}

		// 更新节点检查时�?
		services.Server.UpdateServerStatus(server.ID, server.Type, "check")

		users, err := services.Server.GetAvailableUsers(server)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := gin.H{"users": users}

		// ETag 支持
		responseJSON, _ := json.Marshal(response)
		etag := fmt.Sprintf(`"%x"`, sha1.Sum(responseJSON))

		if strings.Contains(c.GetHeader("If-None-Match"), etag) {
			c.Status(http.StatusNotModified)
			return
		}

		c.Header("ETag", etag)
		c.JSON(http.StatusOK, response)
	}
}

// ServerPush 流量上报
func ServerPush(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		server := getServerFromContext(c)
		if server == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "server not found"})
			return
		}

		// 解析流量数据 [[user_id, upload, download], ...]
		var data [][]int64
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data format"})
			return
		}

		if len(data) == 0 {
			c.JSON(http.StatusOK, gin.H{"data": true})
			return
		}

		// 更新在线用户�?
		services.Server.UpdateOnlineUsers(server.ID, server.Type, len(data))

		// 更新推送时�?
		services.Server.UpdateServerStatus(server.ID, server.Type, "push")

		// 处理流量数据
		trafficData := make(map[int64][2]int64)
		for _, item := range data {
			if len(item) >= 3 {
				userID := item[0]
				trafficData[userID] = [2]int64{item[1], item[2]}
			}
		}

		if err := services.User.TrafficFetch(server, trafficData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// ServerAlive 在线状态上�?
func ServerAlive(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		server := getServerFromContext(c)
		if server == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "server not found"})
			return
		}

		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
			return
		}

		// TODO: 处理在线数据
		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// ServerAliveList 获取在线用户列表
func ServerAliveList(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现在线用户列表
		c.JSON(http.StatusOK, gin.H{"alive": map[string]interface{}{}})
	}
}

// ServerStatus 节点状态上�?
func ServerStatus(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		server := getServerFromContext(c)
		if server == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "server not found"})
			return
		}

		var status struct {
			CPU  float64 `json:"cpu"`
			Mem  struct {
				Total int64 `json:"total"`
				Used  int64 `json:"used"`
			} `json:"mem"`
			Swap struct {
				Total int64 `json:"total"`
				Used  int64 `json:"used"`
			} `json:"swap"`
			Disk struct {
				Total int64 `json:"total"`
				Used  int64 `json:"used"`
			} `json:"disk"`
		}

		if err := c.ShouldBindJSON(&status); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
			return
		}

		statusData := map[string]interface{}{
			"cpu":        status.CPU,
			"mem":        map[string]int64{"total": status.Mem.Total, "used": status.Mem.Used},
			"swap":       map[string]int64{"total": status.Swap.Total, "used": status.Swap.Used},
			"disk":       map[string]int64{"total": status.Disk.Total, "used": status.Disk.Used},
			"updated_at": time.Now().Unix(),
		}

		if err := services.Server.UpdateLoadStatus(server.ID, server.Type, statusData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true, "code": 0, "message": "success"})
	}
}

// getServerFromContext 从上下文获取服务器信息
func getServerFromContext(c *gin.Context) *model.Server {
	// 先尝试从 context 获取（由中间件设置）
	if server, ok := c.Get("server"); ok {
		return server.(*model.Server)
	}
	return nil
}

// SetServerContext 设置服务器上下文（供中间件使用）
func SetServerContext(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		nodeID := c.Query("node_id")
		nodeType := c.Query("node_type")

		if nodeID == "" {
			nodeID = c.GetHeader("X-Node-ID")
		}
		if nodeType == "" {
			nodeType = c.GetHeader("X-Node-Type")
		}

		if nodeID == "" {
			c.Next()
			return
		}

		id, err := strconv.ParseInt(nodeID, 10, 64)
		if err != nil {
			c.Next()
			return
		}

		server, err := services.Server.FindServer(id, nodeType)
		if err == nil && server != nil {
			c.Set("server", server)
		}

		c.Next()
	}
}
