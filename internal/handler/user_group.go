package handler

import (
	"net/http"
	"strconv"
	"time"

	"xboard/internal/model"
	"xboard/internal/service"

	"github.com/gin-gonic/gin"
)

// ==================== 用户组管理 ====================

// AdminListUserGroups 获取用户组列表
func AdminListUserGroups(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		groups, err := services.UserGroup.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 返回详细信息
		result := make([]map[string]interface{}, 0, len(groups))
		for _, group := range groups {
			result = append(result, services.UserGroup.GetGroupInfo(&group))
		}

		c.JSON(http.StatusOK, gin.H{"data": result})
	}
}

// AdminGetUserGroup 获取用户组详情
func AdminGetUserGroup(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		group, err := services.UserGroup.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user group not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": services.UserGroup.GetGroupInfo(group)})
	}
}

// AdminCreateUserGroup 创建用户组
func AdminCreateUserGroup(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string  `json:"name" binding:"required"`
			Description string  `json:"description"`
			ServerIDs   []int64 `json:"server_ids"`
			PlanIDs     []int64 `json:"plan_ids"`
			Sort        int     `json:"sort"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 转换为 JSONArray
		serverIDs := make(model.JSONArray, len(req.ServerIDs))
		for i, id := range req.ServerIDs {
			serverIDs[i] = id
		}

		planIDs := make(model.JSONArray, len(req.PlanIDs))
		for i, id := range req.PlanIDs {
			planIDs[i] = id
		}

		group := &model.UserGroup{
			Name:        req.Name,
			Description: req.Description,
			ServerIDs:   serverIDs,
			PlanIDs:     planIDs,
			Sort:        req.Sort,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		}

		if err := services.UserGroup.Create(group); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": services.UserGroup.GetGroupInfo(group)})
	}
}

// AdminUpdateUserGroup 更新用户组
func AdminUpdateUserGroup(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		group, err := services.UserGroup.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user group not found"})
			return
		}

		var req struct {
			Name        string  `json:"name"`
			Description string  `json:"description"`
			ServerIDs   []int64 `json:"server_ids"`
			PlanIDs     []int64 `json:"plan_ids"`
			Sort        int     `json:"sort"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 更新字段
		if req.Name != "" {
			group.Name = req.Name
		}
		group.Description = req.Description
		group.Sort = req.Sort

		// 转换为 JSONArray
		if req.ServerIDs != nil {
			serverIDs := make(model.JSONArray, len(req.ServerIDs))
			for i, id := range req.ServerIDs {
				serverIDs[i] = id
			}
			group.ServerIDs = serverIDs
		}

		if req.PlanIDs != nil {
			planIDs := make(model.JSONArray, len(req.PlanIDs))
			for i, id := range req.PlanIDs {
				planIDs[i] = id
			}
			group.PlanIDs = planIDs
		}

		group.UpdatedAt = time.Now().Unix()

		if err := services.UserGroup.Update(group); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": services.UserGroup.GetGroupInfo(group)})
	}
}

// AdminDeleteUserGroup 删除用户组
func AdminDeleteUserGroup(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.UserGroup.Delete(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminSetUserGroupServers 设置用户组的节点列表
func AdminSetUserGroupServers(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		var req struct {
			ServerIDs []int64 `json:"server_ids" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.UserGroup.SetServersForGroup(id, req.ServerIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminSetUserGroupPlans 设置用户组的套餐列表
func AdminSetUserGroupPlans(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		var req struct {
			PlanIDs []int64 `json:"plan_ids" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.UserGroup.SetPlansForGroup(id, req.PlanIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminAddServerToUserGroup 为用户组添加节点
func AdminAddServerToUserGroup(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		var req struct {
			ServerID int64 `json:"server_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.UserGroup.AddServerToGroup(id, req.ServerID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminRemoveServerFromUserGroup 从用户组移除节点
func AdminRemoveServerFromUserGroup(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		serverID, _ := strconv.ParseInt(c.Param("server_id"), 10, 64)

		if err := services.UserGroup.RemoveServerFromGroup(id, serverID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminAddPlanToUserGroup 为用户组添加套餐
func AdminAddPlanToUserGroup(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		var req struct {
			PlanID int64 `json:"plan_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.UserGroup.AddPlanToGroup(id, req.PlanID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminRemovePlanFromUserGroup 从用户组移除套餐
func AdminRemovePlanFromUserGroup(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		planID, _ := strconv.ParseInt(c.Param("plan_id"), 10, 64)

		if err := services.UserGroup.RemovePlanFromGroup(id, planID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}
