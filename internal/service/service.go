package service

import (
	"dashgo/internal/config"
	"dashgo/internal/repository"
	"dashgo/pkg/cache"
)

type Services struct {
	User         *UserService
	Server       *ServerService
	Plan         *PlanService
	Order        *OrderService
	Auth         *AuthService
	Setting      *SettingService
	Ticket       *TicketService
	Mail         *MailService
	Telegram     *TelegramService
	NodeSync     *NodeSyncService
	Payment      *PaymentService
	Coupon       *CouponService
	Invite       *InviteService
	Notice       *NoticeService
	Knowledge    *KnowledgeService
	Stats        *StatsService
	Scheduler    *SchedulerService
	Host         *HostService
	ServerGroup  *ServerGroupService
	UserGroup    *UserGroupService
	Traffic      *TrafficService
	AgentVersion *AgentVersionService
}

func NewServices(repos *repository.Repositories, cache *cache.Client, cfg *config.Config) *Services {
	settingService := NewSettingService(repos.Setting, cache)
	mailService := NewMailService(cfg.Mail)
	telegramService := NewTelegramService(cfg.Telegram)
	telegramService.SetRepositories(repos.User, repos.Setting)
	serverService := NewServerService(repos.Server, repos.User, cache, cfg)
	orderService := NewOrderService(repos.Order, repos.User, repos.Plan, repos.Coupon)
	userGroupService := NewUserGroupService(repos.UserGroup, repos.Server, repos.Plan, repos.User)

	// 设置 UserGroupService 告ServerService 依赖
	userGroupService.SetServerService(serverService)

	return &Services{
		User:         NewUserService(repos.User, cache),
		Server:       serverService,
		Plan:         NewPlanService(repos.Plan, repos.User),
		Order:        orderService,
		Auth:         NewAuthService(repos.User, cfg.JWT),
		Setting:      settingService,
		Ticket:       NewTicketService(repos.Ticket, repos.TicketMessage, repos.User, mailService, telegramService),
		Mail:         mailService,
		Telegram:     telegramService,
		NodeSync:     NewNodeSyncService(repos.Server, repos.User, repos.Stat, cfg),
		Payment:      NewPaymentService(repos.Payment, repos.Order, orderService),
		Coupon:       NewCouponService(repos.Coupon, repos.Order),
		Invite:       NewInviteService(repos.InviteCode, repos.User, repos.CommissionLog),
		Notice:       NewNoticeService(repos.Notice),
		Knowledge:    NewKnowledgeService(repos.Knowledge),
		Stats:        NewStatsService(repos.User, repos.Order, repos.Server, repos.Stat, repos.Ticket),
		Scheduler:    NewSchedulerService(repos.User, repos.Order, repos.Stat, mailService, telegramService),
		Host:         NewHostService(repos.Host, repos.ServerNode, repos.User, repos.Server, cache),
		ServerGroup:  NewServerGroupService(repos.ServerGroup),
		UserGroup:    userGroupService,
		Traffic:      NewTrafficService(repos.User, mailService),
		AgentVersion: NewAgentVersionService(repos.DB),
	}
}
