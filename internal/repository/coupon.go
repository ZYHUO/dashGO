package repository

import (
	"dashgo/internal/model"

	"gorm.io/gorm"
)

type CouponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

func (r *CouponRepository) Create(coupon *model.Coupon) error {
	return r.db.Create(coupon).Error
}

func (r *CouponRepository) Update(coupon *model.Coupon) error {
	return r.db.Save(coupon).Error
}

func (r *CouponRepository) Delete(id int64) error {
	return r.db.Delete(&model.Coupon{}, id).Error
}

func (r *CouponRepository) FindByID(id int64) (*model.Coupon, error) {
	var coupon model.Coupon
	err := r.db.First(&coupon, id).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *CouponRepository) FindByCode(code string) (*model.Coupon, error) {
	var coupon model.Coupon
	err := r.db.Where("code = ?", code).First(&coupon).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *CouponRepository) GetAll() ([]model.Coupon, error) {
	var coupons []model.Coupon
	err := r.db.Order("created_at DESC").Find(&coupons).Error
	return coupons, err
}

func (r *CouponRepository) GetUsedCount(couponID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Order{}).Where("coupon_id = ? AND status = ?", couponID, model.OrderStatusCompleted).Count(&count).Error
	return count, err
}

func (r *CouponRepository) GetUserUsedCount(couponID, userID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Order{}).Where("coupon_id = ? AND user_id = ? AND status = ?", couponID, userID, model.OrderStatusCompleted).Count(&count).Error
	return count, err
}

func (r *CouponRepository) RecordUsage(couponID, orderID, userID int64) error {
	// 更新订单的优惠券 ID
	return r.db.Model(&model.Order{}).Where("id = ?", orderID).Update("coupon_id", couponID).Error
}

// InviteCodeRepository 邀请码仓库
type InviteCodeRepository struct {
	db *gorm.DB
}

func NewInviteCodeRepository(db *gorm.DB) *InviteCodeRepository {
	return &InviteCodeRepository{db: db}
}

func (r *InviteCodeRepository) Create(code *model.InviteCode) error {
	return r.db.Create(code).Error
}

func (r *InviteCodeRepository) Update(code *model.InviteCode) error {
	return r.db.Save(code).Error
}

func (r *InviteCodeRepository) FindByID(id int64) (*model.InviteCode, error) {
	var code model.InviteCode
	err := r.db.First(&code, id).Error
	if err != nil {
		return nil, err
	}
	return &code, nil
}

func (r *InviteCodeRepository) FindByCode(code string) (*model.InviteCode, error) {
	var inviteCode model.InviteCode
	err := r.db.Where("code = ?", code).First(&inviteCode).Error
	if err != nil {
		return nil, err
	}
	return &inviteCode, nil
}

func (r *InviteCodeRepository) FindByUserID(userID int64) ([]model.InviteCode, error) {
	var codes []model.InviteCode
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&codes).Error
	return codes, err
}

func (r *InviteCodeRepository) IncrementPV(id int64) error {
	return r.db.Model(&model.InviteCode{}).Where("id = ?", id).Update("pv", gorm.Expr("pv + 1")).Error
}

// CommissionLogRepository 佣金记录仓库
type CommissionLogRepository struct {
	db *gorm.DB
}

func NewCommissionLogRepository(db *gorm.DB) *CommissionLogRepository {
	return &CommissionLogRepository{db: db}
}

func (r *CommissionLogRepository) Create(log *model.CommissionLog) error {
	return r.db.Create(log).Error
}

func (r *CommissionLogRepository) FindByInviteUserID(userID int64, page, pageSize int) ([]model.CommissionLog, int64, error) {
	var logs []model.CommissionLog
	var total int64

	r.db.Model(&model.CommissionLog{}).Where("invite_user_id = ?", userID).Count(&total)
	err := r.db.Where("invite_user_id = ?", userID).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

func (r *CommissionLogRepository) SumByInviteUserID(userID int64) (int64, error) {
	var sum int64
	err := r.db.Model(&model.CommissionLog{}).
		Where("invite_user_id = ?", userID).
		Select("COALESCE(SUM(get_amount), 0)").
		Scan(&sum).Error
	return sum, err
}

// NoticeRepository 公告仓库
type NoticeRepository struct {
	db *gorm.DB
}

func NewNoticeRepository(db *gorm.DB) *NoticeRepository {
	return &NoticeRepository{db: db}
}

func (r *NoticeRepository) Create(notice *model.Notice) error {
	return r.db.Create(notice).Error
}

func (r *NoticeRepository) Update(notice *model.Notice) error {
	return r.db.Save(notice).Error
}

func (r *NoticeRepository) Delete(id int64) error {
	return r.db.Delete(&model.Notice{}, id).Error
}

func (r *NoticeRepository) FindByID(id int64) (*model.Notice, error) {
	var notice model.Notice
	err := r.db.First(&notice, id).Error
	if err != nil {
		return nil, err
	}
	return &notice, nil
}

func (r *NoticeRepository) GetAll() ([]model.Notice, error) {
	var notices []model.Notice
	err := r.db.Order("sort ASC, created_at DESC").Find(&notices).Error
	return notices, err
}

func (r *NoticeRepository) GetVisible() ([]model.Notice, error) {
	var notices []model.Notice
	err := r.db.Where("`show` = ?", true).Order("sort ASC, created_at DESC").Find(&notices).Error
	return notices, err
}

// KnowledgeRepository 知识库仓库
type KnowledgeRepository struct {
	db *gorm.DB
}

func NewKnowledgeRepository(db *gorm.DB) *KnowledgeRepository {
	return &KnowledgeRepository{db: db}
}

func (r *KnowledgeRepository) Create(knowledge *model.Knowledge) error {
	return r.db.Create(knowledge).Error
}

func (r *KnowledgeRepository) Update(knowledge *model.Knowledge) error {
	return r.db.Save(knowledge).Error
}

func (r *KnowledgeRepository) Delete(id int64) error {
	return r.db.Delete(&model.Knowledge{}, id).Error
}

func (r *KnowledgeRepository) FindByID(id int64) (*model.Knowledge, error) {
	var knowledge model.Knowledge
	err := r.db.First(&knowledge, id).Error
	if err != nil {
		return nil, err
	}
	return &knowledge, nil
}

func (r *KnowledgeRepository) GetAll() ([]model.Knowledge, error) {
	var items []model.Knowledge
	err := r.db.Order("sort ASC, created_at DESC").Find(&items).Error
	return items, err
}

func (r *KnowledgeRepository) GetVisible(language string) ([]model.Knowledge, error) {
	var items []model.Knowledge
	query := r.db.Where("`show` = ?", true)
	if language != "" {
		query = query.Where("language = ?", language)
	}
	err := query.Order("sort ASC, created_at DESC").Find(&items).Error
	return items, err
}

func (r *KnowledgeRepository) GetByCategory(category, language string) ([]model.Knowledge, error) {
	var items []model.Knowledge
	query := r.db.Where("`show` = ? AND category = ?", true, category)
	if language != "" {
		query = query.Where("language = ?", language)
	}
	err := query.Order("sort ASC").Find(&items).Error
	return items, err
}

func (r *KnowledgeRepository) GetCategories(language string) ([]string, error) {
	var categories []string
	query := r.db.Model(&model.Knowledge{}).Where("`show` = ?", true)
	if language != "" {
		query = query.Where("language = ?", language)
	}
	err := query.Distinct("category").Pluck("category", &categories).Error
	return categories, err
}
