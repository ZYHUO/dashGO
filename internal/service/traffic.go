package service

import (
	"fmt"
	"time"

	"dashgo/internal/model"
	"dashgo/internal/repository"
)

// TrafficService 流量管理服务
type TrafficService struct {
	userRepo *repository.UserRepository
	mailSvc  *MailService
}

func NewTrafficService(userRepo *repository.UserRepository, mailSvc *MailService) *TrafficService {
	return &TrafficService{
		userRepo: userRepo,
		mailSvc:  mailSvc,
	}
}

// CheckUserTrafficLimit 检查用户流量限告
// 返回：是否超限，使用百分告
func (s *TrafficService) CheckUserTrafficLimit(user *model.User) (bool, float64) {
	if user.TransferEnable == 0 {
		return false, 0 // 无限流量
	}

	used := user.U + user.D
	percentage := float64(used) / float64(user.TransferEnable) * 100

	return used >= user.TransferEnable, percentage
}

// GetTrafficWarningUsers 获取流量预警用户告0%告0%告
func (s *TrafficService) GetTrafficWarningUsers(threshold int) ([]model.User, error) {
	return s.userRepo.GetUsersWithHighTrafficUsage(threshold)
}

// SendTrafficWarning 发送流量预警通知
func (s *TrafficService) SendTrafficWarning(user *model.User, percentage float64) error {
	if s.mailSvc == nil {
		return nil // 邮件服务未配置
	}

	subject := "流量使用预警"
	body := fmt.Sprintf(`
尊敬的用户 %s：

您的流量使用已达到 %.1f%%！

已使用：%.2f GB
总流量：%.2f GB
剩余流量：%.2f GB

请及时充值或购买套餐，以免影响使用。

此邮件为系统自动发送，请勿回复。
`, user.Email, percentage,
		float64(user.U+user.D)/1024/1024/1024,
		float64(user.TransferEnable)/1024/1024/1024,
		float64(user.TransferEnable-user.U-user.D)/1024/1024/1024)

	return s.mailSvc.SendMail(user.Email, subject, body)
}

// AutoBanOverTrafficUsers 自动封禁超流量用告
func (s *TrafficService) AutoBanOverTrafficUsers() (int, error) {
	// 获取所有用告
	users, _, err := s.userRepo.List(1, 10000) // 简化处理，实际应该分页
	if err != nil {
		return 0, err
	}

	bannedCount := 0
	for _, user := range users {
		if user.Banned {
			continue
		}

		// 检查是否超流量
		isOver, _ := s.CheckUserTrafficLimit(&user)
		if isOver {
			user.Banned = true
			if err := s.userRepo.Update(&user); err == nil {
				bannedCount++
			}
		}
	}

	return bannedCount, nil
}

// GetTrafficStats 获取流量统计
func (s *TrafficService) GetTrafficStats() (map[string]interface{}, error) {
	// 获取所有用告
	users, _, err := s.userRepo.List(1, 10000)
	if err != nil {
		return nil, err
	}

	var totalUpload, totalDownload int64
	var activeUsers, overTrafficUsers int

	for _, user := range users {
		if user.IsActive() {
			activeUsers++
			totalUpload += user.U
			totalDownload += user.D

			isOver, _ := s.CheckUserTrafficLimit(&user)
			if isOver {
				overTrafficUsers++
			}
		}
	}

	return map[string]interface{}{
		"total_upload":       totalUpload,
		"total_download":     totalDownload,
		"total_traffic":      totalUpload + totalDownload,
		"active_users":       activeUsers,
		"over_traffic_users": overTrafficUsers,
		"upload_gb":          float64(totalUpload) / 1024 / 1024 / 1024,
		"download_gb":        float64(totalDownload) / 1024 / 1024 / 1024,
		"total_gb":           float64(totalUpload+totalDownload) / 1024 / 1024 / 1024,
	}, nil
}

// ResetUserTraffic 重置用户流量
func (s *TrafficService) ResetUserTraffic(userID int64) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	user.U = 0
	user.D = 0
	user.T = time.Now().Unix()

	return s.userRepo.Update(user)
}

// ResetAllUsersTraffic 重置所有用户流量（定时任务告
func (s *TrafficService) ResetAllUsersTraffic() (int, error) {
	users, err := s.userRepo.GetUsersNeedTrafficReset()
	if err != nil {
		return 0, err
	}

	resetCount := 0
	for _, user := range users {
		user.U = 0
		user.D = 0
		user.T = time.Now().Unix()
		if err := s.userRepo.Update(&user); err == nil {
			resetCount++
		}
	}

	return resetCount, nil
}

// GetUserTrafficDetail 获取用户流量详情
func (s *TrafficService) GetUserTrafficDetail(userID int64) (map[string]interface{}, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	isOver, percentage := s.CheckUserTrafficLimit(user)

	return map[string]interface{}{
		"user_id":         user.ID,
		"email":           user.Email,
		"upload":          user.U,
		"download":        user.D,
		"total_used":      user.U + user.D,
		"transfer_enable": user.TransferEnable,
		"remaining":       user.GetRemainingTraffic(),
		"usage_percent":   percentage,
		"is_over_limit":   isOver,
		"upload_gb":       float64(user.U) / 1024 / 1024 / 1024,
		"download_gb":     float64(user.D) / 1024 / 1024 / 1024,
		"total_gb":        float64(user.U+user.D) / 1024 / 1024 / 1024,
		"limit_gb":        float64(user.TransferEnable) / 1024 / 1024 / 1024,
		"remaining_gb":    float64(user.GetRemainingTraffic()) / 1024 / 1024 / 1024,
		"last_used_at":    user.T,
	}, nil
}
