package service

import (
	"errors"
	"time"

	"dashgo/internal/model"
	"dashgo/internal/repository"
	"dashgo/pkg/utils"
)

// InviteService 邀请服
type InviteService struct {
	inviteRepo     *repository.InviteCodeRepository
	userRepo       *repository.UserRepository
	commissionRepo *repository.CommissionLogRepository
}

func NewInviteService(
	inviteRepo *repository.InviteCodeRepository,
	userRepo *repository.UserRepository,
	commissionRepo *repository.CommissionLogRepository,
) *InviteService {
	return &InviteService{
		inviteRepo:     inviteRepo,
		userRepo:       userRepo,
		commissionRepo: commissionRepo,
	}
}

// GetUserInviteCodes 获取用户的邀请码
func (s *InviteService) GetUserInviteCodes(userID int64) ([]model.InviteCode, error) {
	return s.inviteRepo.FindByUserID(userID)
}

// GenerateInviteCode 生成邀请码
func (s *InviteService) GenerateInviteCode(userID int64) (*model.InviteCode, error) {
	code := &model.InviteCode{
		UserID:    userID,
		Code:      utils.GenerateToken(8),
		Status:    false,
		PV:        0,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	if err := s.inviteRepo.Create(code); err != nil {
		return nil, err
	}

	return code, nil
}

// ValidateInviteCode 验证邀请码
func (s *InviteService) ValidateInviteCode(code string) (*model.InviteCode, error) {
	inviteCode, err := s.inviteRepo.FindByCode(code)
	if err != nil {
		return nil, errors.New("invalid invite code")
	}

	if inviteCode.Status {
		return nil, errors.New("invite code already used")
	}

	// 增加访问次数
	s.inviteRepo.IncrementPV(inviteCode.ID)

	return inviteCode, nil
}

// UseInviteCode 使用邀请码
func (s *InviteService) UseInviteCode(code string, newUserID int64) error {
	inviteCode, err := s.inviteRepo.FindByCode(code)
	if err != nil {
		return errors.New("invalid invite code")
	}

	// 更新新用户的邀请人
	newUser, err := s.userRepo.FindByID(newUserID)
	if err != nil {
		return err
	}

	newUser.InviteUserID = &inviteCode.UserID
	if err := s.userRepo.Update(newUser); err != nil {
		return err
	}

	// 标记邀请码已使用
	inviteCode.Status = true
	return s.inviteRepo.Update(inviteCode)
}

// CalculateCommission 计算佣金
func (s *InviteService) CalculateCommission(order *model.Order) (int64, error) {
	user, err := s.userRepo.FindByID(order.UserID)
	if err != nil {
		return 0, err
	}

	if user.InviteUserID == nil {
		return 0, nil
	}

	inviter, err := s.userRepo.FindByID(*user.InviteUserID)
	if err != nil {
		return 0, nil
	}

	// 计算佣金
	var commission int64
	switch inviter.CommissionType {
	case 0: // 系统默认
		// 默认 10%
		commission = order.TotalAmount * 10 / 100
	case 1: // 按周期
		commission = order.TotalAmount * 10 / 100
	case 2: // 按订单
		if inviter.CommissionRate != nil {
			commission = order.TotalAmount * int64(*inviter.CommissionRate) / 100
		}
	}

	return commission, nil
}

// RecordCommission 记录佣金
func (s *InviteService) RecordCommission(order *model.Order, commission int64) error {
	user, err := s.userRepo.FindByID(order.UserID)
	if err != nil {
		return err
	}

	if user.InviteUserID == nil {
		return nil
	}

	// 创建佣金记录
	log := &model.CommissionLog{
		InviteUserID: *user.InviteUserID,
		UserID:       order.UserID,
		TradeNo:      order.TradeNo,
		OrderAmount:  order.TotalAmount,
		GetAmount:    commission,
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
	}

	if err := s.commissionRepo.Create(log); err != nil {
		return err
	}

	// 更新邀请人佣金余额
	inviter, _ := s.userRepo.FindByID(*user.InviteUserID)
	if inviter != nil {
		inviter.CommissionBalance += commission
		s.userRepo.Update(inviter)
	}

	return nil
}

// GetCommissionLogs 获取佣金记录
func (s *InviteService) GetCommissionLogs(userID int64, page, pageSize int) ([]model.CommissionLog, int64, error) {
	return s.commissionRepo.FindByInviteUserID(userID, page, pageSize)
}

// WithdrawCommission 提现佣金
func (s *InviteService) WithdrawCommission(userID int64, amount int64) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if user.CommissionBalance < amount {
		return errors.New("insufficient commission balance")
	}

	// 转入余额
	user.CommissionBalance -= amount
	user.Balance += amount

	return s.userRepo.Update(user)
}

// GetInviteStats 获取邀请统计
func (s *InviteService) GetInviteStats(userID int64) (map[string]interface{}, error) {
	// 获取邀请人数
	invitedCount, _ := s.userRepo.CountByInviteUserID(userID)

	// 获取佣金统计
	totalCommission, _ := s.commissionRepo.SumByInviteUserID(userID)

	// 获取用户佣金余额
	user, _ := s.userRepo.FindByID(userID)
	var commissionBalance int64
	if user != nil {
		commissionBalance = user.CommissionBalance
	}

	return map[string]interface{}{
		"invited_count":      invitedCount,
		"total_commission":   totalCommission,
		"commission_balance": commissionBalance,
	}, nil
}
