package handler

import (
	"net/http"

	"dashgo/internal/service"

	"github.com/gin-gonic/gin"
)

// GetPaymentMethods 获取支付方式列表
func GetPaymentMethods(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		payments, err := services.Payment.GetEnabledPayments()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		result := make([]map[string]interface{}, 0, len(payments))
		for _, p := range payments {
			result = append(result, map[string]interface{}{
				"id":   p.ID,
				"name": p.Name,
				"icon": p.Icon,
			})
		}

		c.JSON(http.StatusOK, gin.H{"data": result})
	}
}

// CreatePayment 创建支付
func CreatePayment(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req struct {
			TradeNo   string `json:"trade_no" binding:"required"`
			PaymentID int64  `json:"payment_id"` // 移除 required�? 表示余额支付
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 如果 payment_id �?0，使用余额支�?
		if req.PaymentID == 0 {
			err := services.Payment.PayWithBalance(req.TradeNo, user.ID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": gin.H{"type": "balance", "paid": true}})
			return
		}

		result, err := services.Payment.CreatePayment(req.TradeNo, req.PaymentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": result})
	}
}

// PaymentNotify 支付回调
func PaymentNotify(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		paymentUUID := c.Param("uuid")

		// 获取所有参�?
		params := make(map[string]string)

		// GET 参数
		for k, v := range c.Request.URL.Query() {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}

		// POST 参数
		c.Request.ParseForm()
		for k, v := range c.Request.PostForm {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}

		if err := services.Payment.HandleCallback(paymentUUID, params); err != nil {
			c.String(http.StatusBadRequest, "fail")
			return
		}

		c.String(http.StatusOK, "success")
	}
}

// CheckPaymentStatus 检查支付状�?
func CheckPaymentStatus(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		tradeNo := c.Query("trade_no")
		if tradeNo == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "trade_no required"})
			return
		}

		paid, err := services.Payment.CheckPaymentStatus(tradeNo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": gin.H{"paid": paid}})
	}
}

// CheckCoupon 检查优惠券
func CheckCoupon(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req struct {
			Code   string `json:"code" binding:"required"`
			PlanID int64  `json:"plan_id" binding:"required"`
			Period string `json:"period" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		coupon, _, err := services.Coupon.CheckCoupon(req.Code, req.PlanID, req.Period, user.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 获取套餐价格来计算实际折�?
		plan, err := services.Plan.GetByID(req.PlanID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "plan not found"})
			return
		}

		price := plan.GetPriceByPeriod(req.Period)
		discount := services.Coupon.CalculateDiscount(coupon, price)

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"id":       coupon.ID,
				"name":     coupon.Name,
				"type":     coupon.Type,
				"value":    coupon.Value,
				"discount": discount,
			},
		})
	}
}



// GetInviteInfo 获取邀请信�?
func GetInviteInfo(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// 获取邀请码
		codes, _ := services.Invite.GetUserInviteCodes(user.ID)

		// 获取统计
		stats, _ := services.Invite.GetInviteStats(user.ID)

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"codes": codes,
				"stats": stats,
			},
		})
	}
}

// GenerateInviteCode 生成邀请码
func GenerateInviteCode(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		code, err := services.Invite.GenerateInviteCode(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": code})
	}
}

// GetCommissionLogs 获取佣金记录
func GetCommissionLogs(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		page := 1
		pageSize := 20

		logs, total, err := services.Invite.GetCommissionLogs(user.ID, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  logs,
			"total": total,
		})
	}
}

// WithdrawCommission 提现佣金
func WithdrawCommission(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if err := services.Invite.WithdrawCommission(user.ID, user.CommissionBalance); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}
