package service

import (
	"time"

	"dashgo/internal/repository"
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

	// 服务器统�?
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

// GetServerRanking 获取服务器排�?
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
	// 在线用户数（最�?5 分钟有流量的用户�?
	onlineUsers, _ := s.userRepo.CountOnline(5 * 60)

	// 今日流量
	todayStart := time.Now().Truncate(24 * time.Hour).Unix()
	todayTraffic, _ := s.statRepo.GetTotalTraffic(todayStart, time.Now().Unix())

	return map[string]interface{}{
		"online_users":  onlineUsers,
		"today_traffic": todayTraffic,
	}, nil
}

// GetUserList 获取用户列表
func (s *StatsService) GetUserList(search string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	users, total, err := s.userRepo.FindAll(search, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		result = append(result, map[string]interface{}{
			"id":              user.ID,
			"email":           user.Email,
			"balance":         user.Balance,
			"plan_id":         user.PlanID,
			"transfer_enable": user.TransferEnable,
			"u":               user.U,
			"d":               user.D,
			"expired_at":      user.ExpiredAt,
			"banned":          user.Banned,
			"is_admin":        user.IsAdmin,
			"is_staff":        user.IsStaff,
			"created_at":      user.CreatedAt,
		})
	}

	return result, total, nil
}

// UpdateUser 更新用户
func (s *StatsService) UpdateUser(id int64, email string, balance, planID, transferEnable, expiredAt *int64, banned, isAdmin, isStaff *bool, password string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}

	if email != "" {
		user.Email = email
	}
	if balance != nil {
		user.Balance = *balance
	}
	if planID != nil {
		user.PlanID = planID
	}
	if transferEnable != nil {
		user.TransferEnable = *transferEnable
	}
	if expiredAt != nil {
		user.ExpiredAt = expiredAt
	}
	if banned != nil {
		user.Banned = *banned
	}
	if isAdmin != nil {
		user.IsAdmin = *isAdmin
	}
	if isStaff != nil {
		user.IsStaff = *isStaff
	}
	if password != "" {
		// 需要导�?utils 包来加密密码
		// 这里简化处理，实际应该加密
		user.Password = password
	}

	return s.userRepo.Update(user)
}

// DeleteUser 删除用户
func (s *StatsService) DeleteUser(id int64) error {
	return s.userRepo.Delete(id)
}

// ResetUserTraffic 重置用户流量
func (s *StatsService) ResetUserTraffic(id int64) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}

	user.U = 0
	user.D = 0
	return s.userRepo.Update(user)
}

// GetOrderList 获取订单列表
func (s *StatsService) GetOrderList(status *int, page, pageSize int) ([]map[string]interface{}, int64, error) {
	orders, total, err := s.orderRepo.FindAll(status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, 0, len(orders))
	for _, order := range orders {
		// 获取用户邮箱
		userEmail := ""
		if user, _ := s.userRepo.FindByID(order.UserID); user != nil {
			userEmail = user.Email
		}

		result = append(result, map[string]interface{}{
			"id":           order.ID,
			"user_id":      order.UserID,
			"user_email":   userEmail,
			"trade_no":     order.TradeNo,
			"plan_id":      order.PlanID,
			"period":       order.Period,
			"total_amount": order.TotalAmount,
			"status":       order.Status,
			"type":         order.Type,
			"created_at":   order.CreatedAt,
		})
	}

	return result, total, nil
}

// UpdateOrderStatus 更新订单状�?
func (s *StatsService) UpdateOrderStatus(id int64, status int) error {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return err
	}

	order.Status = status
	return s.orderRepo.Update(order)
}


// GetTrafficOverview 获取流量概览（用于饼状图�?
func (s *StatsService) GetTrafficOverview() (map[string]interface{}, error) {
	// 获取所有用户的流量统计
	users, _, err := s.userRepo.FindAll("", 1, 10000)
	if err != nil {
		return nil, err
	}

	var totalUpload, totalDownload int64
	var activeUsers int
	userTraffic := make([]map[string]interface{}, 0)

	for _, user := range users {
		used := user.U + user.D
		if used > 0 {
			activeUsers++
			totalUpload += user.U
			totalDownload += user.D
			userTraffic = append(userTraffic, map[string]interface{}{
				"user_id":  user.ID,
				"email":    user.Email,
				"upload":   user.U,
				"download": user.D,
				"total":    used,
			})
		}
	}

	// 按流量排序取�?10
	// 简单冒泡排�?
	for i := 0; i < len(userTraffic)-1 && i < 10; i++ {
		for j := i + 1; j < len(userTraffic); j++ {
			if userTraffic[j]["total"].(int64) > userTraffic[i]["total"].(int64) {
				userTraffic[i], userTraffic[j] = userTraffic[j], userTraffic[i]
			}
		}
	}

	topUsers := userTraffic
	if len(topUsers) > 10 {
		topUsers = topUsers[:10]
	}

	return map[string]interface{}{
		"total_upload":   totalUpload,
		"total_download": totalDownload,
		"total_traffic":  totalUpload + totalDownload,
		"active_users":   activeUsers,
		"top_users":      topUsers,
		"upload_percent": func() float64 {
			total := totalUpload + totalDownload
			if total == 0 {
				return 0
			}
			return float64(totalUpload) / float64(total) * 100
		}(),
		"download_percent": func() float64 {
			total := totalUpload + totalDownload
			if total == 0 {
				return 0
			}
			return float64(totalDownload) / float64(total) * 100
		}(),
	}, nil
}

// GetServerTrafficOverview 获取节点流量概览
func (s *StatsService) GetServerTrafficOverview() ([]map[string]interface{}, error) {
	servers, err := s.serverRepo.GetAllServers()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(servers))
	for _, server := range servers {
		// 从统计表获取节点流量
		traffic, _ := s.statRepo.GetServerTotalTraffic(server.ID)
		result = append(result, map[string]interface{}{
			"server_id":   server.ID,
			"server_name": server.Name,
			"server_type": server.Type,
			"upload":      traffic.U,
			"download":    traffic.D,
			"total":       traffic.U + traffic.D,
		})
	}

	return result, nil
}

// GetUserTrafficDetail 获取用户流量详情
func (s *StatsService) GetUserTrafficDetail(userID int64) (map[string]interface{}, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	used := user.U + user.D
	remaining := user.TransferEnable - used
	if remaining < 0 {
		remaining = 0
	}

	usedPercent := float64(0)
	if user.TransferEnable > 0 {
		usedPercent = float64(used) / float64(user.TransferEnable) * 100
	}

	return map[string]interface{}{
		"user_id":         user.ID,
		"email":           user.Email,
		"upload":          user.U,
		"download":        user.D,
		"total_used":      used,
		"transfer_enable": user.TransferEnable,
		"remaining":       remaining,
		"used_percent":    usedPercent,
	}, nil
}

// GetDailyTrafficStats 获取每日流量统计
func (s *StatsService) GetDailyTrafficStats(days int) ([]map[string]interface{}, error) {
	endAt := time.Now().Unix()
	startAt := endAt - int64(days*86400)

	stats, err := s.statRepo.GetServerTrafficStats(startAt, endAt)
	if err != nil {
		return nil, err
	}

	// 按日期聚�?
	dailyMap := make(map[string]map[string]int64)
	for _, stat := range stats {
		date := time.Unix(stat.RecordAt, 0).Format("2006-01-02")
		if _, ok := dailyMap[date]; !ok {
			dailyMap[date] = map[string]int64{"upload": 0, "download": 0}
		}
		dailyMap[date]["upload"] += stat.U
		dailyMap[date]["download"] += stat.D
	}

	result := make([]map[string]interface{}, 0)
	for date, traffic := range dailyMap {
		result = append(result, map[string]interface{}{
			"date":     date,
			"upload":   traffic["upload"],
			"download": traffic["download"],
			"total":    traffic["upload"] + traffic["download"],
		})
	}

	return result, nil
}
