package repository

import (
	"dashgo/internal/model"

	"gorm.io/gorm"
)

// TicketRepository 工单仓库
type TicketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(ticket *model.Ticket) error {
	return r.db.Create(ticket).Error
}

func (r *TicketRepository) Update(ticket *model.Ticket) error {
	return r.db.Save(ticket).Error
}

func (r *TicketRepository) Delete(id int64) error {
	return r.db.Delete(&model.Ticket{}, id).Error
}

func (r *TicketRepository) FindByID(id int64) (*model.Ticket, error) {
	var ticket model.Ticket
	err := r.db.First(&ticket, id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *TicketRepository) FindByUserID(userID int64, page, pageSize int) ([]model.Ticket, int64, error) {
	var tickets []model.Ticket
	var total int64

	r.db.Model(&model.Ticket{}).Where("user_id = ?", userID).Count(&total)
	err := r.db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&tickets).Error

	return tickets, total, err
}

func (r *TicketRepository) FindAll(status *int, page, pageSize int) ([]model.Ticket, int64, error) {
	var tickets []model.Ticket
	var total int64

	query := r.db.Model(&model.Ticket{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	query.Count(&total)
	err := query.Order("updated_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&tickets).Error

	return tickets, total, err
}

func (r *TicketRepository) CountPending() (int64, error) {
	var count int64
	err := r.db.Model(&model.Ticket{}).
		Where("status = ? AND reply_status = ?", 0, 0).
		Count(&count).Error
	return count, err
}

// TicketMessageRepository 工单消息仓库
type TicketMessageRepository struct {
	db *gorm.DB
}

func NewTicketMessageRepository(db *gorm.DB) *TicketMessageRepository {
	return &TicketMessageRepository{db: db}
}

func (r *TicketMessageRepository) Create(message *model.TicketMessage) error {
	return r.db.Create(message).Error
}

func (r *TicketMessageRepository) FindByTicketID(ticketID int64) ([]model.TicketMessage, error) {
	var messages []model.TicketMessage
	err := r.db.Where("ticket_id = ?", ticketID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

func (r *TicketMessageRepository) Delete(id int64) error {
	return r.db.Delete(&model.TicketMessage{}, id).Error
}
