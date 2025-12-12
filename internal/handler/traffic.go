package handler

import (
	"net/http"
	"strconv"

	"dashgo/internal/service"

	"github.com/gin-gonic/gin"
)

// ==================== 流量管理 ====================

// AdminGetTrafficStats 获取流量统计
func AdminGetTrafficStats(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := services.Traffic.GetTrafficStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": stats})
	}
}

// AdminGetTrafficWarnings 获取流量预警用户
func AdminGetTrafficWarnings(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		threshold, _ := strconv.Atoi(c.DefaultQuery("threshold", "80"))

		users, err := services.Traffic.GetTrafficWarningUsers(threshold)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 构建返回数据
		result := make([]map[string]interface{}, 0, len(users))
		for _, user := range users {
			isOver, percentage := services.Traffic.CheckUserTrafficLimit(&user)
			result = append(result, map[string]interface{}{
				"id":              user.ID,
				"email":           user.Email,
				"upload":          user.U,
				"download":        user.D,
				"total_used":      user.U + user.D,
				"transfer_enable": user.TransferEnable,
				"usage_percent":   percentage,
				"is_over_limit":   isOver,
				"upload_gb":       float64(user.U) / 1024 / 1024 / 1024,
				"download_gb":     float64(user.D) / 1024 / 1024 / 1024,
				"total_gb":        float64(user.U+user.D) / 1024 / 1024 / 1024,
				"limit_gb":        float64(user.TransferEnable) / 1024 / 1024 / 1024,
			})
		}

		c.JSON(http.StatusOK, gin.H{"data": result, "total": len(result)})
	}
}

// AdminResetTraffic 重置用户流量
func AdminResetTraffic(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.Traffic.ResetUserTraffic(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminResetAllTraffic 重置所有用户流�?
func AdminResetAllTraffic(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		count, err := services.Traffic.ResetAllUsersTraffic()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": true,
			"message": "已重置流�?,
			"count": count,
		})
	}
}

// AdminGetUserTrafficDetail 获取用户流量详情
func AdminGetUserTrafficDetail(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		detail, err := services.Traffic.GetUserTrafficDetail(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": detail})
	}
}

// AdminSendTrafficWarning 发送流量预警通知
func AdminSendTrafficWarning(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		user, err := services.User.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		_, percentage := services.Traffic.CheckUserTrafficLimit(user)
		if err := services.Traffic.SendTrafficWarning(user, percentage); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true, "message": "预警通知已发�?})
	}
}

// AdminBatchSendTrafficWarnings 批量发送流量预�?
func AdminBatchSendTrafficWarnings(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		threshold, _ := strconv.Atoi(c.DefaultQuery("threshold", "80"))

		users, err := services.Traffic.GetTrafficWarningUsers(threshold)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		successCount := 0
		for _, user := range users {
			_, percentage := services.Traffic.CheckUserTrafficLimit(&user)
			if err := services.Traffic.SendTrafficWarning(&user, percentage); err == nil {
				successCount++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"data": true,
			"message": "批量发送完�?,
			"total": len(users),
			"success": successCount,
		})
	}
}

// AdminAutobanOverTrafficUsers 自动封禁超流量用�?
func AdminAutobanOverTrafficUsers(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		count, err := services.Traffic.AutoBanOverTrafficUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": true,
			"message": "已封禁超流量用户",
			"count": count,
		})
	}
}
