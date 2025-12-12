package handler

import (
	"net/http"
	"os"
	"path/filepath"

	"dashgo/internal/config"
	"dashgo/internal/middleware"
	"dashgo/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, services *service.Services, cfg *config.Config) {
	// 公共中间�?
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// 静态文件服�?
	staticPath := "web/dist"
	if _, err := os.Stat(staticPath); err == nil {
		r.Static("/assets", filepath.Join(staticPath, "assets"))
		r.StaticFile("/favicon.ico", filepath.Join(staticPath, "favicon.ico"))
		
		// SPA 路由支持
		r.NoRoute(func(c *gin.Context) {
			// API 路由返回 404
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.File(filepath.Join(staticPath, "index.html"))
		})
	}

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Guest routes (无需认证)
		guest := v1.Group("/guest")
		{
			guest.POST("/register", GuestRegister(services))
			guest.POST("/login", GuestLogin(services))
			guest.GET("/plans", GuestGetPlans(services))
		}

		// Passport routes (认证相关)
		passport := v1.Group("/passport")
		{
			passport.POST("/auth/login", PassportLogin(services))
			passport.POST("/auth/register", PassportRegister(services))
		}

		// User routes (需要用户认�?
		user := v1.Group("/user")
		user.Use(middleware.JWTAuth(services.Auth))
		{
			user.GET("/info", UserInfo(services))
			user.GET("/subscribe", UserSubscribe(services))
			user.POST("/reset_token", UserResetToken(services))
			user.POST("/reset_uuid", UserResetUUID(services))
			user.POST("/change_password", UserChangePassword(services))
			user.GET("/orders", UserOrders(services))
			user.POST("/order/create", UserCreateOrder(services))
			user.POST("/order/cancel", UserCancelOrder(services))

			// Ticket routes
			user.GET("/tickets", UserTickets(services))
			user.GET("/ticket/:id", UserTicketDetail(services))
			user.POST("/ticket/create", UserCreateTicket(services))
			user.POST("/ticket/:id/reply", UserReplyTicket(services))
			user.POST("/ticket/:id/close", UserCloseTicket(services))

			// Payment routes
			user.POST("/order/pay", CreatePayment(services))
			user.GET("/order/check", CheckPaymentStatus(services))
			user.POST("/coupon/check", CheckCoupon(services))

			// Invite routes
			user.GET("/invite", GetInviteInfo(services))
			user.POST("/invite/generate", GenerateInviteCode(services))
			user.GET("/invite/commission", GetCommissionLogs(services))
			user.POST("/invite/withdraw", WithdrawCommission(services))
		}

		// Client routes (订阅获取)
		client := v1.Group("/client")
		{
			client.GET("/subscribe", ClientSubscribe(services))
		}

		// Payment routes
		payment := v1.Group("/payment")
		{
			payment.GET("/methods", GetPaymentMethods(services))
			payment.POST("/notify/:uuid", PaymentNotify(services))
		}

		// Public routes
		v1.GET("/notices", GetNotices(services))
		v1.GET("/knowledge", GetKnowledge(services))
		v1.GET("/knowledge/categories", GetKnowledgeCategories(services))
		v1.GET("/guest/settings", GetPublicSettings(services))

		// Email verification
		v1.POST("/guest/send_email_code", GuestSendEmailCode(services))

		// Telegram webhook
		v1.POST("/telegram/webhook", TelegramWebhook(services))

		// Agent routes (主机对接)
		agent := v1.Group("/agent")
		agent.Use(AgentAuth(services))
		{
			agent.POST("/heartbeat", AgentHeartbeat(services))
			agent.GET("/config", AgentGetConfig(services))
			agent.POST("/traffic", AgentReportTraffic(services))
			agent.GET("/users", AgentGetUsers(services))
			agent.POST("/sync", AgentSyncStatus(services))
			agent.GET("/version", AgentGetVersion(services))
			agent.POST("/update-status", AgentUpdateStatus(services))
		}

		// Server routes (节点通信)
		server := v1.Group("/server")
		server.Use(middleware.NodeAuth(cfg.Node.Token))
		server.Use(SetServerContext(services))
		{
			// UniProxy 兼容接口
			uniProxy := server.Group("/UniProxy")
			{
				uniProxy.GET("/config", ServerConfig(services))
				uniProxy.GET("/user", ServerUsers(services))
				uniProxy.POST("/push", ServerPush(services))
				uniProxy.POST("/alive", ServerAlive(services))
				uniProxy.GET("/alivelist", ServerAliveList(services))
				uniProxy.POST("/status", ServerStatus(services))
			}
		}
	}

	// API v2 (Admin)
	v2 := r.Group("/api/v2")
	{
		admin := v2.Group("/admin")
		admin.Use(middleware.JWTAuth(services.Auth))
		admin.Use(middleware.AdminAuth())
		{
			// Server management
			admin.GET("/servers", AdminListServers(services))
			admin.POST("/server", AdminCreateServer(services))
			admin.PUT("/server/:id", AdminUpdateServer(services))
			admin.DELETE("/server/:id", AdminDeleteServer(services))
			admin.GET("/server/:id/status", AdminGetServerStatus(services))
			admin.POST("/server/:id/sync", AdminSyncServerUsers(services))

			// User management
			admin.GET("/users", AdminListUsers(services))
			admin.GET("/user/:id", AdminGetUser(services))
			admin.PUT("/user/:id", AdminUpdateUser(services))
			admin.DELETE("/user/:id", AdminDeleteUser(services))
			admin.POST("/user/:id/reset_traffic", AdminResetUserTraffic(services))

			// Plan management
			admin.GET("/plans", AdminListPlans(services))
			admin.POST("/plan", AdminCreatePlan(services))
			admin.PUT("/plan/:id", AdminUpdatePlan(services))
			admin.DELETE("/plan/:id", AdminDeletePlan(services))

			// Order management
			admin.GET("/orders", AdminListOrders(services))
			admin.GET("/order/:id", AdminGetOrder(services))
			admin.PUT("/order/:id/status", AdminUpdateOrderStatus(services))

			// Settings
			admin.GET("/settings", AdminGetSettings(services))
			admin.POST("/settings", AdminUpdateSettings(services))

			// Ticket management
			admin.GET("/tickets", AdminListTickets(services))
			admin.GET("/ticket/:id", AdminTicketDetail(services))
			admin.POST("/ticket/:id/reply", AdminReplyTicket(services))
			admin.POST("/ticket/:id/close", AdminCloseTicket(services))

			// Statistics
			admin.GET("/stats/overview", AdminStatsOverview(services))
			admin.GET("/stats/order", AdminOrderStats(services))
			admin.GET("/stats/user", AdminUserStats(services))
			admin.GET("/stats/traffic", AdminTrafficStats(services))
			admin.GET("/stats/server_ranking", AdminServerRanking(services))
			admin.GET("/stats/user_ranking", AdminUserRanking(services))

			// Notice management
			admin.GET("/notices", AdminListNotices(services))
			admin.POST("/notice", AdminCreateNotice(services))
			admin.PUT("/notice/:id", AdminUpdateNotice(services))
			admin.DELETE("/notice/:id", AdminDeleteNotice(services))

			// Knowledge management
			admin.GET("/knowledge", AdminListKnowledge(services))
			admin.POST("/knowledge", AdminCreateKnowledge(services))
			admin.PUT("/knowledge/:id", AdminUpdateKnowledge(services))
			admin.DELETE("/knowledge/:id", AdminDeleteKnowledge(services))

			// Coupon management
			admin.GET("/coupons", AdminListCoupons(services))
			admin.POST("/coupon", AdminCreateCoupon(services))
			admin.PUT("/coupon/:id", AdminUpdateCoupon(services))
			admin.DELETE("/coupon/:id", AdminDeleteCoupon(services))

			// Payment management
			admin.GET("/payments", AdminListPayments(services))
			admin.POST("/payment", AdminCreatePayment(services))
			admin.PUT("/payment/:id", AdminUpdatePayment(services))

			// Server Group management (用户组管�?
			admin.GET("/server_groups", AdminListServerGroups(services))
			admin.POST("/server_group", AdminCreateServerGroup(services))
			admin.PUT("/server_group/:id", AdminUpdateServerGroup(services))
			admin.DELETE("/server_group/:id", AdminDeleteServerGroup(services))

			// Host management (主机管理)
			admin.GET("/hosts", AdminListHosts(services))
			admin.POST("/host", AdminCreateHost(services))
			admin.PUT("/host/:id", AdminUpdateHost(services))
			admin.DELETE("/host/:id", AdminDeleteHost(services))
			admin.POST("/host/:id/reset_token", AdminResetHostToken(services))
			admin.GET("/host/:id/config", AdminGetHostConfig(services))

			// Node management (节点管理)
			admin.GET("/nodes", AdminListNodes(services))
			admin.POST("/node", AdminCreateNode(services))
			admin.PUT("/node/:id", AdminUpdateNode(services))
			admin.DELETE("/node/:id", AdminDeleteNode(services))
			admin.GET("/node/default", AdminGetDefaultNodeConfig(services))

			// Agent Version management (Agent 版本管理)
			admin.GET("/agent/versions", AdminListAgentVersions(services))
			admin.POST("/agent/version", AdminCreateAgentVersion(services))
			admin.PUT("/agent/version/:id", AdminUpdateAgentVersion(services))
			admin.DELETE("/agent/version/:id", AdminDeleteAgentVersion(services))
			admin.POST("/agent/version/:id/set_latest", AdminSetLatestAgentVersion(services))
			admin.GET("/agent/update_logs", AdminListAgentUpdateLogs(services))

			// Site settings (站点设置)
			admin.GET("/site/settings", AdminGetSiteSettings(services))
			admin.POST("/site/settings", AdminUpdateSiteSettings(services))

			// Telegram settings
			admin.GET("/telegram/settings", AdminGetTelegramSettings(services))
			admin.POST("/telegram/settings", AdminUpdateTelegramSettings(services))
			admin.POST("/telegram/webhook", AdminSetTelegramWebhook(services))

			// Traffic statistics (流量统计)
			admin.GET("/traffic/overview", AdminTrafficOverview(services))
			admin.GET("/traffic/servers", AdminServerTrafficOverview(services))
			admin.GET("/traffic/daily", AdminDailyTrafficStats(services))
			admin.GET("/traffic/user/:id", AdminUserTrafficDetail(services))

			// User Group management (用户组管�?- 新架�?
			admin.GET("/user-groups", AdminListUserGroups(services))
			admin.GET("/user-group/:id", AdminGetUserGroup(services))
			admin.POST("/user-group", AdminCreateUserGroup(services))
			admin.PUT("/user-group/:id", AdminUpdateUserGroup(services))
			admin.DELETE("/user-group/:id", AdminDeleteUserGroup(services))
			
			// User Group - Server management (用户组节点管�?
			admin.POST("/user-group/:id/servers", AdminSetUserGroupServers(services))
			admin.POST("/user-group/:id/server", AdminAddServerToUserGroup(services))
			admin.DELETE("/user-group/:id/server/:server_id", AdminRemoveServerFromUserGroup(services))
			
			// User Group - Plan management (用户组套餐管�?
			admin.POST("/user-group/:id/plans", AdminSetUserGroupPlans(services))
			admin.POST("/user-group/:id/plan", AdminAddPlanToUserGroup(services))
			admin.DELETE("/user-group/:id/plan/:plan_id", AdminRemovePlanFromUserGroup(services))

			// Traffic management (流量管理 - 新功�?
			admin.GET("/traffic/stats", AdminGetTrafficStats(services))
			admin.GET("/traffic/warnings", AdminGetTrafficWarnings(services))
			admin.POST("/traffic/reset/:id", AdminResetTraffic(services))
			admin.POST("/traffic/reset-all", AdminResetAllTraffic(services))
			admin.GET("/traffic/detail/:id", AdminGetUserTrafficDetail(services))
			admin.POST("/traffic/warning/:id", AdminSendTrafficWarning(services))
			admin.POST("/traffic/warnings/send", AdminBatchSendTrafficWarnings(services))
			admin.POST("/traffic/autoban", AdminAutobanOverTrafficUsers(services))
		}
	}
}
