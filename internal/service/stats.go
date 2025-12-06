package service

import (
	"time"

	"xboard/internal/repository"
)

// StatsService 统计服务
type StatsService struct {
	userRepo   *repository.UserRepository
	orderRepo  *repository.OrderRepository
	serverRepo *repository.ServerRepository
	statRepo   *repository.StatRepository
	ticketRepo *repository.TicketRepository
}

func NewStatsService(
	userRepo *repository.UserRepository,
	orderRepo *repository.OrderRepository,
	serverRepo *repository.ServerRepository,
	statRepo *repository.StatRepository,
	ticketRepo *repository.TicketRepository,
) *StatsService {
	return &StatsService{
		userRepo:   userRepo,
		orderRepo:  orderRepo,
		serverRepo: serverRepo,
		statRepo:   statRepo,
		ticketRepo: ticketRepo,
	}
}

// GetOverview 获取概览统计
func (s *StatsService) GetOverview() (map[string]interface{}, error) {
	// 用户统计
	totalUsers, _ := s.userRepo.Count()
	activeUsers, _ := s.userRepo.CountActive()

	// 订单统计
	totalOrders, _ := s.orderRepo.Count()
	todayOrders, todayIncome, _ := s.orderRepo.GetTodayStats()
	monthOrders, monthIncome, _ := s.orderRepo.GetMonthStats()

	// 服务器统计
	totalServers, _ := s.serverRepo.Count()

	// 工单统计
	pendingTickets, _ := s.ticketRepo.CountPending()

	return map[string]interface{}{
		"user": map[string]interface{}{
			"total":  totalUsers,
			"active": activeUsers,
		},
		"order": map[string]interface{}{
			"total":        totalOrders,
			"today_count":  todayOrders,
			"today_income": todayIncome,
			"month_count":  monthOrders,
			"month_income": monthIncome,
		},
		"server": map[string]interface{}{
			"total": totalServers,
		},
		"ticket": map[string]interface{}{
			"pending": pendingTickets,
		},
	}, nil
}

// GetOrderStats 获取订单统计
func (s *StatsService) GetOrderStats(startAt, endAt int64) ([]map[string]interface{}, error) {
	stats, err := s.statRepo.GetOrderStats(startAt, endAt)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(stats))
	for _, stat := range stats {
		result = append(result, map[string]interface{}{
			"date":        time.Unix(stat.RecordAt, 0).Format("2006-01-02"),
			"order_count": stat.OrderCount,
			"order_total": stat.OrderTotal,
			"paid_count":  stat.PaidCount,
			"paid_total":  stat.PaidTotal,
		})
	}

	return result, nil
}

// GetUserStats 获取用户统计
func (s *StatsService) GetUserStats(startAt, endAt int64) ([]map[string]interface{}, error) {
	stats, err := s.statRepo.GetOrderStats(startAt, endAt) // 使用 Stat 模型
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(stats))
	for _, stat := range stats {
		result = append(result, map[string]interface{}{
			"date":           time.Unix(stat.RecordAt, 0).Format("2006-01-02"),
			"register_count": stat.RegisterCount,
			"invite_count":   stat.InviteCount,
		})
	}

	return result, nil
}

// GetTrafficStats 获取流量统计
func (s *StatsService) GetTrafficStats(startAt, endAt int64) ([]map[string]interface{}, error) {
	stats, err := s.statRepo.GetServerTrafficStats(startAt, endAt)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(stats))
	for _, stat := range stats {
		result = append(result, map[string]interface{}{
			"date":     time.Unix(stat.RecordAt, 0).Format("2006-01-02"),
			"upload":   stat.U,
			"download": stat.D,
			"total":    stat.U + stat.D,
		})
	}

	return result, nil
}

// GetServerRanking 获取服务器排行
func (s *StatsService) GetServerRanking(limit int) ([]map[string]interface{}, error) {
	rankings, err := s.statRepo.GetServerRanking(limit)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(rankings))
	for _, r := range rankings {
		server, _ := s.serverRepo.FindByID(r.ServerID)
		name := ""
		if server != nil {
			name = server.Name
		}
		result = append(result, map[string]interface{}{
			"server_id":   r.ServerID,
			"server_name": name,
			"upload":      r.U,
			"download":    r.D,
			"total":       r.U + r.D,
		})
	}

	return result, nil
}

// GetUserRanking 获取用户流量排行
func (s *StatsService) GetUserRanking(limit int) ([]map[string]interface{}, error) {
	rankings, err := s.statRepo.GetUserRanking(limit)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(rankings))
	for _, r := range rankings {
		user, _ := s.userRepo.FindByID(r.UserID)
		email := ""
		if user != nil {
			email = user.Email
		}
		result = append(result, map[string]interface{}{
			"user_id":  r.UserID,
			"email":    email,
			"upload":   r.U,
			"download": r.D,
			"total":    r.U + r.D,
		})
	}

	return result, nil
}

// GetRealtimeStats 获取实时统计
func (s *StatsService) GetRealtimeStats() (map[string]interface{}, error) {
	// 在线用户数（最近 5 分钟有流量的用户）
	onlineUsers, _ := s.userRepo.CountOnline(5 * 60)

	// 今日流量
	todayStart := time.Now().Truncate(24 * time.Hour).Unix()
	todayTraffic, _ := s.statRepo.GetTotalTraffic(todayStart, time.Now().Unix())

	return map[string]interface{}{
		"online_users":  onlineUsers,
		"today_traffic": todayTraffic,
	}, nil
}
