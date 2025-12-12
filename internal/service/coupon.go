package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"dashgo/internal/model"
	"dashgo/internal/repository"
)

// CouponService 优惠券服�?
type CouponService struct {
	couponRepo *repository.CouponRepository
	orderRepo  *repository.OrderRepository
}

func NewCouponService(couponRepo *repository.CouponRepository, orderRepo *repository.OrderRepository) *CouponService {
	return &CouponService{
		couponRepo: couponRepo,
		orderRepo:  orderRepo,
	}
}

// CheckCoupon 检查优惠券是否可用
func (s *CouponService) CheckCoupon(code string, planID int64, period string, userID int64) (*model.Coupon, int64, error) {
	coupon, err := s.couponRepo.FindByCode(code)
	if err != nil {
		return nil, 0, errors.New("coupon not found")
	}

	// 检查时�?
	now := time.Now().Unix()
	if coupon.StartedAt > now {
		return nil, 0, errors.New("coupon not started")
	}
	if coupon.EndedAt < now {
		return nil, 0, errors.New("coupon expired")
	}

	// 检查使用次�?
	if coupon.LimitUse != nil && *coupon.LimitUse > 0 {
		usedCount, _ := s.couponRepo.GetUsedCount(coupon.ID)
		if usedCount >= int64(*coupon.LimitUse) {
			return nil, 0, errors.New("coupon usage limit reached")
		}
	}

	// 检查用户使用次�?
	if coupon.LimitUseWithUser != nil && *coupon.LimitUseWithUser > 0 {
		userUsedCount, _ := s.couponRepo.GetUserUsedCount(coupon.ID, userID)
		if userUsedCount >= int64(*coupon.LimitUseWithUser) {
			return nil, 0, errors.New("you have reached the usage limit for this coupon")
		}
	}

	// 检查套餐限�?
	if coupon.LimitPlanIDs != nil && *coupon.LimitPlanIDs != "" {
		planIDs := strings.Split(*coupon.LimitPlanIDs, ",")
		planIDStr := strconv.FormatInt(planID, 10)
		found := false
		for _, pid := range planIDs {
			if strings.TrimSpace(pid) == planIDStr {
				found = true
				break
			}
		}
		if !found {
			return nil, 0, errors.New("coupon not applicable to this plan")
		}
	}

	// 检查周期限�?
	if coupon.LimitPeriod != nil && *coupon.LimitPeriod != "" {
		periods := strings.Split(*coupon.LimitPeriod, ",")
		found := false
		for _, p := range periods {
			if p == period {
				found = true
				break
			}
		}
		if !found {
			return nil, 0, errors.New("coupon not applicable to this period")
		}
	}

	return coupon, coupon.Value, nil
}

// CalculateDiscount 计算折扣金额
func (s *CouponService) CalculateDiscount(coupon *model.Coupon, amount int64) int64 {
	switch coupon.Type {
	case model.CouponTypeAmount:
		// 固定金额
		if coupon.Value >= amount {
			return amount
		}
		return coupon.Value
	case model.CouponTypePercent:
		// 百分比折�?
		discount := amount * coupon.Value / 100
		return discount
	}
	return 0
}

// UseCoupon 使用优惠�?
func (s *CouponService) UseCoupon(couponID, orderID, userID int64) error {
	return s.couponRepo.RecordUsage(couponID, orderID, userID)
}

// GetAll 获取所有优惠券
func (s *CouponService) GetAll() ([]model.Coupon, error) {
	return s.couponRepo.GetAll()
}

// GetByID 根据 ID 获取优惠�?
func (s *CouponService) GetByID(id int64) (*model.Coupon, error) {
	return s.couponRepo.FindByID(id)
}

// Create 创建优惠�?
func (s *CouponService) Create(coupon *model.Coupon) error {
	return s.couponRepo.Create(coupon)
}

// Update 更新优惠�?
func (s *CouponService) Update(coupon *model.Coupon) error {
	return s.couponRepo.Update(coupon)
}

// Delete 删除优惠�?
func (s *CouponService) Delete(id int64) error {
	return s.couponRepo.Delete(id)
}

// GenerateCodes 批量生成优惠券码
func (s *CouponService) GenerateCodes(coupon *model.Coupon, count int) ([]string, error) {
	codes := make([]string, 0, count)
	for i := 0; i < count; i++ {
		code := generateRandomCode(8)
		newCoupon := *coupon
		newCoupon.ID = 0
		newCoupon.Code = code
		if err := s.couponRepo.Create(&newCoupon); err != nil {
			continue
		}
		codes = append(codes, code)
	}
	return codes, nil
}

// generateRandomCode 生成随机�?
func generateRandomCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(time.Nanosecond)
	}
	return string(result)
}
