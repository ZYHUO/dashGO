package handler

import (
	"net/http"
	"strconv"
	"time"

	"dashgo/internal/model"
	"dashgo/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminOrderStats 订单统计
func AdminOrderStats(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
		endAt := time.Now().Unix()
		startAt := endAt - int64(days*86400)

		stats, err := services.Stats.GetOrderStats(startAt, endAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": stats})
	}
}

// AdminUserStats 用户统计
func AdminUserStats(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
		endAt := time.Now().Unix()
		startAt := endAt - int64(days*86400)

		stats, err := services.Stats.GetUserStats(startAt, endAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": stats})
	}
}

// AdminTrafficStats 流量统计
func AdminTrafficStats(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
		endAt := time.Now().Unix()
		startAt := endAt - int64(days*86400)

		stats, err := services.Stats.GetTrafficStats(startAt, endAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": stats})
	}
}

// AdminServerRanking 服务器排告
func AdminServerRanking(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		ranking, err := services.Stats.GetServerRanking(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": ranking})
	}
}

// AdminUserRanking 用户排行
func AdminUserRanking(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		ranking, err := services.Stats.GetUserRanking(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": ranking})
	}
}

// AdminListNotices 公告列表
func AdminListNotices(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		notices, err := services.Notice.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": notices})
	}
}

// AdminCreateNotice 创建公告
func AdminCreateNotice(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var notice model.Notice
		if err := c.ShouldBindJSON(&notice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Notice.Create(&notice); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": notice})
	}
}

// AdminUpdateNotice 更新公告
func AdminUpdateNotice(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		notice, err := services.Notice.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "notice not found"})
			return
		}

		if err := c.ShouldBindJSON(notice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Notice.Update(notice); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": notice})
	}
}

// AdminDeleteNotice 删除公告
func AdminDeleteNotice(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.Notice.Delete(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminListKnowledge 知识库列告
func AdminListKnowledge(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := services.Knowledge.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": items})
	}
}

// AdminCreateKnowledge 创建知识库文告
func AdminCreateKnowledge(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var knowledge model.Knowledge
		if err := c.ShouldBindJSON(&knowledge); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Knowledge.Create(&knowledge); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": knowledge})
	}
}

// AdminUpdateKnowledge 更新知识库文告
func AdminUpdateKnowledge(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		knowledge, err := services.Knowledge.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "knowledge not found"})
			return
		}

		if err := c.ShouldBindJSON(knowledge); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Knowledge.Update(knowledge); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": knowledge})
	}
}

// AdminDeleteKnowledge 删除知识库文告
func AdminDeleteKnowledge(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.Knowledge.Delete(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminListCoupons 优惠券列告
func AdminListCoupons(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		coupons, err := services.Coupon.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": coupons})
	}
}

// AdminCreateCoupon 创建优惠告
func AdminCreateCoupon(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var coupon model.Coupon
		if err := c.ShouldBindJSON(&coupon); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Coupon.Create(&coupon); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": coupon})
	}
}

// AdminUpdateCoupon 更新优惠告
func AdminUpdateCoupon(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		coupon, err := services.Coupon.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "coupon not found"})
			return
		}

		if err := c.ShouldBindJSON(coupon); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Coupon.Update(coupon); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": coupon})
	}
}

// AdminDeleteCoupon 删除优惠告
func AdminDeleteCoupon(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.Coupon.Delete(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminListPayments 支付方式列表
func AdminListPayments(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		payments, err := services.Payment.GetEnabledPayments()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": payments})
	}
}

// AdminCreatePayment 创建支付方式
func AdminCreatePayment(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现创建支付方式
		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminUpdatePayment 更新支付方式
func AdminUpdatePayment(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现更新支付方式
		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// ==================== 用户组管告====================

// AdminListServerGroups 获取用户组列告
func AdminListServerGroups(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		groups, err := services.ServerGroup.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": groups})
	}
}

// AdminCreateServerGroup 创建用户告
func AdminCreateServerGroup(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		group, err := services.ServerGroup.Create(req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": group})
	}
}

// AdminUpdateServerGroup 更新用户告
func AdminUpdateServerGroup(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		var req struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.ServerGroup.Update(id, req.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminDeleteServerGroup 删除用户告
func AdminDeleteServerGroup(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		if err := services.ServerGroup.Delete(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}
