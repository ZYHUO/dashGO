package repository

import (
	"xboard/internal/model"

	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

func (r *OrderRepository) Delete(id int64) error {
	return r.db.Delete(&model.Order{}, id).Error
}

func (r *OrderRepository) FindByID(id int64) (*model.Order, error) {
	var order model.Order
	err := r.db.First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) FindByTradeNo(tradeNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.Where("trade_no = ?", tradeNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) FindByUserID(userID int64) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) FindPendingByUserID(userID int64) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Where("user_id = ? AND status = ?", userID, model.OrderStatusPending).Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) List(page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	r.db.Model(&model.Order{}).Count(&total)
	err := r.db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

func (r *OrderRepository) ListByStatus(status int, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	r.db.Model(&model.Order{}).Where("status = ?", status).Count(&total)
	err := r.db.Where("status = ?", status).Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

// Count 统计订单总数
func (r *OrderRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Order{}).Count(&count).Error
	return count, err
}

// GetTodayStats 获取今日订单统计
func (r *OrderRepository) GetTodayStats() (int64, int64, error) {
	var count int64
	var total int64

	today := getCurrentTimestamp() - (getCurrentTimestamp() % 86400)
	r.db.Model(&model.Order{}).
		Where("created_at >= ?", today).
		Where("status = ?", model.OrderStatusCompleted).
		Count(&count)

	r.db.Model(&model.Order{}).
		Where("created_at >= ?", today).
		Where("status = ?", model.OrderStatusCompleted).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total)

	return count, total, nil
}

// GetMonthStats 获取本月订单统计
func (r *OrderRepository) GetMonthStats() (int64, int64, error) {
	var count int64
	var total int64

	// 本月第一天
	now := getCurrentTimestamp()
	monthStart := now - (now % (86400 * 30)) // 简化处理

	r.db.Model(&model.Order{}).
		Where("created_at >= ?", monthStart).
		Where("status = ?", model.OrderStatusCompleted).
		Count(&count)

	r.db.Model(&model.Order{}).
		Where("created_at >= ?", monthStart).
		Where("status = ?", model.OrderStatusCompleted).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total)

	return count, total, nil
}

// FindAll 查询所有订单（支持状态筛选和分页）
func (r *OrderRepository) FindAll(status *int, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	query.Count(&total)
	err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

// Payment Repository
type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) FindByID(id int64) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.First(&payment, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) FindByUUID(uuid string) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.Where("uuid = ?", uuid).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) GetEnabled() ([]model.Payment, error) {
	var payments []model.Payment
	err := r.db.Where("enable = ?", true).Order("sort ASC").Find(&payments).Error
	return payments, err
}

// CancelExpiredOrders 取消过期订单
func (r *OrderRepository) CancelExpiredOrders(expireSeconds int64) (int64, error) {
	threshold := getCurrentTimestamp() - expireSeconds
	result := r.db.Model(&model.Order{}).
		Where("status = ?", model.OrderStatusPending).
		Where("created_at < ?", threshold).
		Update("status", model.OrderStatusCancelled)
	return result.RowsAffected, result.Error
}

// GetDailyStats 获取指定日期的订单统计
func (r *OrderRepository) GetDailyStats(startTime, endTime int64) (int64, int64, error) {
	var count int64
	var total int64

	r.db.Model(&model.Order{}).
		Where("created_at >= ?", startTime).
		Where("created_at < ?", endTime).
		Where("status = ?", model.OrderStatusCompleted).
		Count(&count)

	r.db.Model(&model.Order{}).
		Where("created_at >= ?", startTime).
		Where("created_at < ?", endTime).
		Where("status = ?", model.OrderStatusCompleted).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total)

	return count, total, nil
}
