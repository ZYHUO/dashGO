package handler

import (
	"net/http"
	"time"

	"dashgo/internal/service"

	"github.com/gin-gonic/gin"
)

// GuestRegister 用户注册
func GuestRegister(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email      string `json:"email" binding:"required,email"`
			Password   string `json:"password" binding:"required,min=6"`
			InviteCode string `json:"invite_code"`
			EmailCode  string `json:"email_code"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 检查是否需要邮箱验告
		if services.Setting.GetBool(service.SettingMailVerify, false) {
			if req.EmailCode == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "请输入邮箱验证码"})
				return
			}
			// 验证邮箱验证码
			if !services.User.VerifyEmailCode(req.Email, req.EmailCode) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已过期"})
				return
			}
		}

		// 检查IP 注册限制
		clientIP := c.ClientIP()
		ipLimit := services.Setting.GetInt(service.SettingRegisterIPLimit, 0)
		if ipLimit > 0 {
			count, _ := services.User.CountByRegisterIP(clientIP)
			if count >= int64(ipLimit) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "告IP 注册次数已达上限"})
				return
			}
		}

		// 检查是否仅限邀请注册
		if services.Setting.GetBool(service.SettingRegisterInviteOnly, false) && req.InviteCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "仅限邀请注册"})
			return
		}

		// 处理邀请码
		var inviteUserID *int64
		if req.InviteCode != "" {
			inviteCode, err := services.Invite.ValidateInviteCode(req.InviteCode)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的邀请码"})
				return
			}
			inviteUserID = &inviteCode.UserID
		}

		user, err := services.User.RegisterWithIP(req.Email, req.Password, inviteUserID, clientIP)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 标记邀请码已使告
		if req.InviteCode != "" {
			services.Invite.UseInviteCode(req.InviteCode, user.ID)
		}

		token, err := services.Auth.GenerateToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"token": token,
			},
		})
	}
}

// GuestSendEmailCode 发送邮箱验证码
func GuestSendEmailCode(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 检查邮件是否配置
		if !services.Mail.IsConfigured() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "邮件服务未配置"})
			return
		}

		// 检查邮箱是否已注册
		existing, _ := services.User.GetByEmail(req.Email)
		if existing != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该邮箱已注册"})
			return
		}

		// 检查冷却时间
		cooldown := services.User.GetEmailCodeCooldown(req.Email)
		if cooldown > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请稍后再试", "cooldown": cooldown})
			return
		}

		// 生成验证码
		code := generateNumericCode(6)

		// 存储验证码
		if err := services.User.SetEmailCode(req.Email, code); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发送失败"})
			return
		}

		// 发送邮件
		if err := services.Mail.SendVerifyCode(req.Email, code); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发送失败，请稍后重试"})
			return
		}

		// 设置冷却时间
		services.User.SetEmailCodeCooldown(req.Email)

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// generateNumericCode 生成数字验证告
func generateNumericCode(length int) string {
	const digits = "0123456789"
	code := make([]byte, length)
	for i := range code {
		code[i] = digits[time.Now().UnixNano()%10]
		time.Sleep(time.Nanosecond)
	}
	return string(code)
}

// GuestLogin 用户登录
func GuestLogin(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := services.User.Login(req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, err := services.Auth.GenerateToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"token":    token,
				"is_admin": user.IsAdmin,
			},
		})
	}
}

// GuestGetPlans 获取可购买套餐列告
func GuestGetPlans(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		plans, err := services.Plan.GetAvailable()
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

// PassportLogin Passport 登录
func PassportLogin(services *service.Services) gin.HandlerFunc {
	return GuestLogin(services)
}

// PassportRegister Passport 注册
func PassportRegister(services *service.Services) gin.HandlerFunc {
	return GuestRegister(services)
}

// GetNotices 获取公告列表
func GetNotices(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		notices, err := services.Notice.GetPublic()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": notices})
	}
}

// GetKnowledge 获取知识库列告
func GetKnowledge(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.Query("category")
		items, err := services.Knowledge.GetPublic(category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items})
	}
}

// GetKnowledgeCategories 获取知识库分告
func GetKnowledgeCategories(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		categories, err := services.Knowledge.GetCategories()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": categories})
	}
}

// GetPublicSettings 获取公开设置
func GetPublicSettings(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		settings := services.Setting.GetPublicSettings()
		c.JSON(http.StatusOK, gin.H{"data": settings})
	}
}

// TelegramWebhook Telegram Webhook
func TelegramWebhook(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var update service.TelegramUpdate
		if err := c.ShouldBindJSON(&update); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Telegram.HandleUpdate(&update); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	}
}
