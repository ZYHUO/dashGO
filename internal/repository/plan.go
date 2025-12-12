package repository

import (
	"dashgo/internal/model"

	"gorm.io/gorm"
)

type PlanRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

func (r *PlanRepository) Create(plan *model.Plan) error {
	return r.db.Create(plan).Error
}

func (r *PlanRepository) Update(plan *model.Plan) error {
	return r.db.Save(plan).Error
}

func (r *PlanRepository) Delete(id int64) error {
	return r.db.Delete(&model.Plan{}, id).Error
}

func (r *PlanRepository) FindByID(id int64) (*model.Plan, error) {
	var plan model.Plan
	err := r.db.First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *PlanRepository) GetAll() ([]model.Plan, error) {
	var plans []model.Plan
	err := r.db.Order("sort ASC").Find(&plans).Error
	return plans, err
}

func (r *PlanRepository) GetAvailable() ([]model.Plan, error) {
	var plans []model.Plan
	err := r.db.Where("`show` = ? AND sell = ?", true, true).Order("sort ASC").Find(&plans).Error
	return plans, err
}

func (r *PlanRepository) List(page, pageSize int) ([]model.Plan, int64, error) {
	var plans []model.Plan
	var total int64

	r.db.Model(&model.Plan{}).Count(&total)
	err := r.db.Order("sort ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&plans).Error
	return plans, total, err
}

// IncrementSoldCount 增加已售数量（原子操作）
func (r *PlanRepository) IncrementSoldCount(planID int64) error {
	return r.db.Model(&model.Plan{}).Where("id = ?", planID).
		UpdateColumn("sold_count", gorm.Expr("sold_count + ?", 1)).Error
}

// DecrementSoldCount 减少已售数量（原子操作）
func (r *PlanRepository) DecrementSoldCount(planID int64) error {
	return r.db.Model(&model.Plan{}).Where("id = ? AND sold_count > 0", planID).
		UpdateColumn("sold_count", gorm.Expr("sold_count - ?", 1)).Error
}
