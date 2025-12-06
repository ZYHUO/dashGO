package handler

import (
	"net/http"
	"strconv"
	"time"

	"xboard/internal/model"
	"xboard/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminStatsOverview 获取统计概览
func AdminStatsOverview(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := services.Stats.GetOverview()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": stats})
	}
}

// ==================== 用户管理 ====================

// AdminListUsers 获取用户列表
func AdminListUsers(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		search := c.Query("search")

		users, total, err := services.Stats.GetUserList(search, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": users, "total": total})
	}
}

// AdminGetUser 获取用户详情
func AdminGetUser(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		user, err := services.User.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": services.User.GetUserInfo(user)})
	}
}

// AdminUpdateUser 更新用户
func AdminUpdateUser(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		var req struct {
			Email          string  `json:"email"`
			Balance        *int64  `json:"balance"`
			PlanID         *int64  `json:"plan_id"`
			TransferEnable *int64  `json:"transfer_enable"`
			ExpiredAt      *int64  `json:"expired_at"`
			Banned         *bool   `json:"banned"`
			IsAdmin        *bool   `json:"is_admin"`
			IsStaff        *bool   `json:"is_staff"`
			Password       string  `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Stats.UpdateUser(id, req.Email, req.Balance, req.PlanID, req.TransferEnable, req.ExpiredAt, req.Banned, req.IsAdmin, req.IsStaff, req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminDeleteUser 删除用户
func AdminDeleteUser(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.Stats.DeleteUser(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminResetUserTraffic 重置用户流量
func AdminResetUserTraffic(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.Stats.ResetUserTraffic(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// ==================== 节点管理 ====================

// AdminListServers 获取服务器列表
func AdminListServers(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		servers, err := services.Server.GetAllServers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": servers})
	}
}

// AdminCreateServer 创建服务器
func AdminCreateServer(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name             string                 `json:"name" binding:"required"`
			Type             string                 `json:"type" binding:"required"`
			Host             string                 `json:"host" binding:"required"`
			Port             string                 `json:"port" binding:"required"`
			Rate             float64                `json:"rate"`
			Show             bool                   `json:"show"`
			Tags             []string               `json:"tags"`
			GroupID          []int64                `json:"group_id"`
			ProtocolSettings map[string]interface{} `json:"protocol_settings"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 转换 Tags 为 JSONArray
		tags := make(model.JSONArray, len(req.Tags))
		for i, t := range req.Tags {
			tags[i] = t
		}

		// 转换 GroupID 为 JSONArray
		groupIDs := make(model.JSONArray, len(req.GroupID))
		for i, g := range req.GroupID {
			groupIDs[i] = g
		}

		server := &model.Server{
			Name:             req.Name,
			Type:             req.Type,
			Host:             req.Host,
			Port:             req.Port,
			Rate:             req.Rate,
			Show:             req.Show,
			Tags:             tags,
			GroupIDs:         groupIDs,
			ProtocolSettings: model.JSONMap(req.ProtocolSettings),
			CreatedAt:        time.Now().Unix(),
			UpdatedAt:        time.Now().Unix(),
		}

		if server.Rate == 0 {
			server.Rate = 1
		}

		if err := services.Server.CreateServer(server); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": server})
	}
}

// AdminUpdateServer 更新服务器
func AdminUpdateServer(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		server, err := services.Server.FindServer(id, "")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
			return
		}

		var req struct {
			Name             string                 `json:"name"`
			Type             string                 `json:"type"`
			Host             string                 `json:"host"`
			Port             string                 `json:"port"`
			Rate             float64                `json:"rate"`
			Show             bool                   `json:"show"`
			Tags             []string               `json:"tags"`
			GroupID          []int64                `json:"group_id"`
			ProtocolSettings map[string]interface{} `json:"protocol_settings"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 转换 Tags 为 JSONArray
		tags := make(model.JSONArray, len(req.Tags))
		for i, t := range req.Tags {
			tags[i] = t
		}

		// 转换 GroupID 为 JSONArray
		groupIDs := make(model.JSONArray, len(req.GroupID))
		for i, g := range req.GroupID {
			groupIDs[i] = g
		}

		server.Name = req.Name
		server.Type = req.Type
		server.Host = req.Host
		server.Port = req.Port
		server.Rate = req.Rate
		server.Show = req.Show
		server.Tags = tags
		server.GroupIDs = groupIDs
		server.ProtocolSettings = model.JSONMap(req.ProtocolSettings)
		server.UpdatedAt = time.Now().Unix()

		if err := services.Server.UpdateServer(server); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": server})
	}
}

// AdminDeleteServer 删除服务器
func AdminDeleteServer(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.Server.DeleteServer(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminGetServerStatus 获取服务器状态
func AdminGetServerStatus(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 直接返回在线状态，实际状态由 Agent 心跳管理
		// 节点管理页面的状态检测已改为即时返回
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"online": true,
				"stats": gin.H{
					"uplink_bytes":   0,
					"downlink_bytes": 0,
					"tcp_sessions":   0,
					"udp_sessions":   0,
				},
			},
		})
	}
}

// AdminSyncServerUsers 手动同步服务器用户
func AdminSyncServerUsers(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Agent 模式下，用户同步由 Agent 自动处理
		// 这里直接返回成功
		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// ==================== 套餐管理 ====================

// AdminListPlans 获取套餐列表
func AdminListPlans(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		plans, err := services.Plan.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		result := make([]map[string]interface{}, 0, len(plans))
		for _, plan := range plans {
			result = append(result, services.Plan.GetPlanInfo(&plan))
		}

		c.JSON(http.StatusOK, gin.H{"data": result})
	}
}

// AdminCreatePlan 创建套餐
func AdminCreatePlan(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name           string            `json:"name" binding:"required"`
			TransferEnable int64             `json:"transfer_enable"`
			SpeedLimit     *int              `json:"speed_limit"`
			DeviceLimit    *int              `json:"device_limit"`
			Prices         map[string]int64  `json:"prices"`
			Show           bool              `json:"show"`
			Sell           bool              `json:"sell"`
			GroupID        *int64            `json:"group_id"`
			Sort           int               `json:"sort"`
			Content        string            `json:"content"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		plan := &model.Plan{
			Name:           req.Name,
			TransferEnable: req.TransferEnable,
			SpeedLimit:     req.SpeedLimit,
			DeviceLimit:    req.DeviceLimit,
			Show:           req.Show,
			Sell:           req.Sell,
			GroupID:        req.GroupID,
			Sort:           req.Sort,
			Content:        req.Content,
			CreatedAt:      time.Now().Unix(),
			UpdatedAt:      time.Now().Unix(),
		}

		// 设置价格
		if req.Prices != nil {
			if v, ok := req.Prices["monthly"]; ok {
				plan.MonthPrice = &v
			}
			if v, ok := req.Prices["quarterly"]; ok {
				plan.QuarterPrice = &v
			}
			if v, ok := req.Prices["half_yearly"]; ok {
				plan.HalfYearPrice = &v
			}
			if v, ok := req.Prices["yearly"]; ok {
				plan.YearPrice = &v
			}
			if v, ok := req.Prices["two_yearly"]; ok {
				plan.TwoYearPrice = &v
			}
			if v, ok := req.Prices["three_yearly"]; ok {
				plan.ThreeYearPrice = &v
			}
			if v, ok := req.Prices["onetime"]; ok {
				plan.OnetimePrice = &v
			}
			if v, ok := req.Prices["reset"]; ok {
				plan.ResetPrice = &v
			}
		}

		if err := services.Plan.Create(plan); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": services.Plan.GetPlanInfo(plan)})
	}
}

// AdminUpdatePlan 更新套餐
func AdminUpdatePlan(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		plan, err := services.Plan.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}

		var req struct {
			Name           string            `json:"name"`
			TransferEnable int64             `json:"transfer_enable"`
			SpeedLimit     *int              `json:"speed_limit"`
			DeviceLimit    *int              `json:"device_limit"`
			Prices         map[string]int64  `json:"prices"`
			Show           bool              `json:"show"`
			Sell           bool              `json:"sell"`
			GroupID        *int64            `json:"group_id"`
			Sort           int               `json:"sort"`
			Content        string            `json:"content"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		plan.Name = req.Name
		plan.TransferEnable = req.TransferEnable
		plan.SpeedLimit = req.SpeedLimit
		plan.DeviceLimit = req.DeviceLimit
		plan.Show = req.Show
		plan.Sell = req.Sell
		plan.GroupID = req.GroupID
		plan.Sort = req.Sort
		plan.Content = req.Content
		plan.UpdatedAt = time.Now().Unix()

		// 更新价格
		if req.Prices != nil {
			if v, ok := req.Prices["monthly"]; ok {
				plan.MonthPrice = &v
			}
			if v, ok := req.Prices["quarterly"]; ok {
				plan.QuarterPrice = &v
			}
			if v, ok := req.Prices["half_yearly"]; ok {
				plan.HalfYearPrice = &v
			}
			if v, ok := req.Prices["yearly"]; ok {
				plan.YearPrice = &v
			}
			if v, ok := req.Prices["two_yearly"]; ok {
				plan.TwoYearPrice = &v
			}
			if v, ok := req.Prices["three_yearly"]; ok {
				plan.ThreeYearPrice = &v
			}
			if v, ok := req.Prices["onetime"]; ok {
				plan.OnetimePrice = &v
			}
			if v, ok := req.Prices["reset"]; ok {
				plan.ResetPrice = &v
			}
		}

		if err := services.Plan.Update(plan); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": services.Plan.GetPlanInfo(plan)})
	}
}

// AdminDeletePlan 删除套餐
func AdminDeletePlan(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.Plan.Delete(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// ==================== 订单管理 ====================

// AdminListOrders 获取订单列表
func AdminListOrders(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		status := c.Query("status")

		var statusPtr *int
		if status != "" {
			v, _ := strconv.Atoi(status)
			statusPtr = &v
		}

		orders, total, err := services.Stats.GetOrderList(statusPtr, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": orders, "total": total})
	}
}

// AdminGetOrder 获取订单详情
func AdminGetOrder(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		order, err := services.Order.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": order})
	}
}

// AdminUpdateOrderStatus 更新订单状态
func AdminUpdateOrderStatus(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		var req struct {
			Status int `json:"status" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Stats.UpdateOrderStatus(id, req.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// ==================== 工单管理 ====================

// AdminListTickets 获取工单列表
func AdminListTickets(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		var status *int
		if s := c.Query("status"); s != "" {
			v, _ := strconv.Atoi(s)
			status = &v
		}

		tickets, total, err := services.Ticket.GetAllTickets(status, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": tickets, "total": total})
	}
}

// AdminTicketDetail 获取工单详情
func AdminTicketDetail(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		detail, err := services.Ticket.GetTicketDetail(id, user.ID, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": detail})
	}
}

// AdminReplyTicket 管理员回复工单
func AdminReplyTicket(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		var req struct {
			Message string `json:"message" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		message, err := services.Ticket.ReplyTicket(id, user.ID, req.Message, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": message})
	}
}

// AdminCloseTicket 管理员关闭工单
func AdminCloseTicket(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.Ticket.CloseTicket(id, user.ID, true); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// ==================== 系统设置 ====================

// AdminGetSettings 获取系统设置
func AdminGetSettings(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		settings, err := services.Setting.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": settings})
	}
}

// AdminUpdateSettings 更新系统设置
func AdminUpdateSettings(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var settings map[string]string
		if err := c.ShouldBindJSON(&settings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		for key, value := range settings {
			if err := services.Setting.Set(key, value); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}
