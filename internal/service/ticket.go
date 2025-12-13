package service

import (
	"errors"
	"time"

	"dashgo/internal/model"
	"dashgo/internal/repository"
)

// TicketService 工单服务
type TicketService struct {
	ticketRepo  *repository.TicketRepository
	messageRepo *repository.TicketMessageRepository
	userRepo    *repository.UserRepository
	mailService *MailService
	tgService   *TelegramService
}

func NewTicketService(
	ticketRepo *repository.TicketRepository,
	messageRepo *repository.TicketMessageRepository,
	userRepo *repository.UserRepository,
	mailService *MailService,
	tgService *TelegramService,
) *TicketService {
	return &TicketService{
		ticketRepo:  ticketRepo,
		messageRepo: messageRepo,
		userRepo:    userRepo,
		mailService: mailService,
		tgService:   tgService,
	}
}

// 工单状告
const (
	TicketStatusOpen   = 0 // 开告
	TicketStatusClosed = 1 // 关闭
)

// 回复状告
const (
	TicketReplyPending = 0 // 待回告
	TicketReplyReplied = 1 // 已回告
)

// 工单级别
const (
	TicketLevelLow    = 0 // 告
	TicketLevelMedium = 1 // 告
	TicketLevelHigh   = 2 // 告
)

// CreateTicket 创建工单
func (s *TicketService) CreateTicket(userID int64, subject, message string, level int) (*model.Ticket, error) {
	if subject == "" || message == "" {
		return nil, errors.New("subject and message are required")
	}

	ticket := &model.Ticket{
		UserID:      userID,
		Subject:     subject,
		Level:       level,
		Status:      TicketStatusOpen,
		ReplyStatus: TicketReplyPending,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	if err := s.ticketRepo.Create(ticket); err != nil {
		return nil, err
	}

	// 创建第一条消告
	ticketMessage := &model.TicketMessage{
		UserID:    userID,
		TicketID:  ticket.ID,
		Message:   message,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	if err := s.messageRepo.Create(ticketMessage); err != nil {
		return nil, err
	}

	return ticket, nil
}

// ReplyTicket 回复工单
func (s *TicketService) ReplyTicket(ticketID, userID int64, message string, isAdmin bool) (*model.TicketMessage, error) {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	// 检查权告
	if !isAdmin && ticket.UserID != userID {
		return nil, errors.New("permission denied")
	}

	// 检查工单状告
	if ticket.Status == TicketStatusClosed {
		return nil, errors.New("ticket is closed")
	}

	// 创建消息
	ticketMessage := &model.TicketMessage{
		UserID:    userID,
		TicketID:  ticketID,
		Message:   message,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	if err := s.messageRepo.Create(ticketMessage); err != nil {
		return nil, err
	}

	// 更新工单状告
	if isAdmin {
		ticket.ReplyStatus = TicketReplyReplied
	} else {
		ticket.ReplyStatus = TicketReplyPending
	}
	ticket.UpdatedAt = time.Now().Unix()

	if err := s.ticketRepo.Update(ticket); err != nil {
		return nil, err
	}

	return ticketMessage, nil
}

// CloseTicket 关闭工单
func (s *TicketService) CloseTicket(ticketID, userID int64, isAdmin bool) error {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return errors.New("ticket not found")
	}

	// 检查权告
	if !isAdmin && ticket.UserID != userID {
		return errors.New("permission denied")
	}

	ticket.Status = TicketStatusClosed
	ticket.UpdatedAt = time.Now().Unix()

	return s.ticketRepo.Update(ticket)
}

// ReopenTicket 重新打开工单
func (s *TicketService) ReopenTicket(ticketID, userID int64, isAdmin bool) error {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return errors.New("ticket not found")
	}

	// 检查权告
	if !isAdmin && ticket.UserID != userID {
		return errors.New("permission denied")
	}

	ticket.Status = TicketStatusOpen
	ticket.ReplyStatus = TicketReplyPending
	ticket.UpdatedAt = time.Now().Unix()

	return s.ticketRepo.Update(ticket)
}

// GetUserTickets 获取用户工单列表
func (s *TicketService) GetUserTickets(userID int64, page, pageSize int) ([]model.Ticket, int64, error) {
	return s.ticketRepo.FindByUserID(userID, page, pageSize)
}

// GetTicketMessages 获取工单消息
func (s *TicketService) GetTicketMessages(ticketID, userID int64, isAdmin bool) ([]TicketMessageWithUser, error) {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	// 检查权告
	if !isAdmin && ticket.UserID != userID {
		return nil, errors.New("permission denied")
	}

	messages, err := s.messageRepo.FindByTicketID(ticketID)
	if err != nil {
		return nil, err
	}

	// 获取用户信息
	result := make([]TicketMessageWithUser, 0, len(messages))
	for _, msg := range messages {
		user, _ := s.userRepo.FindByID(msg.UserID)
		item := TicketMessageWithUser{
			TicketMessage: msg,
		}
		if user != nil {
			item.UserEmail = user.Email
			item.IsAdmin = user.IsAdmin
		}
		result = append(result, item)
	}

	return result, nil
}

// GetTicketDetail 获取工单详情
func (s *TicketService) GetTicketDetail(ticketID, userID int64, isAdmin bool) (*TicketDetail, error) {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	// 检查权告
	if !isAdmin && ticket.UserID != userID {
		return nil, errors.New("permission denied")
	}

	messages, err := s.messageRepo.FindByTicketID(ticketID)
	if err != nil {
		return nil, err
	}

	// 转换消息格式，添告is_me 字段
	msgList := make([]TicketMessageView, 0, len(messages))
	for _, msg := range messages {
		msgList = append(msgList, TicketMessageView{
			ID:        msg.ID,
			Message:   msg.Message,
			IsMe:      msg.UserID == userID,
			CreatedAt: msg.CreatedAt,
		})
	}

	user, _ := s.userRepo.FindByID(ticket.UserID)

	return &TicketDetail{
		Ticket:    *ticket,
		Messages:  msgList,
		UserEmail: user.Email,
	}, nil
}

// GetAllTickets 获取所有工单（管理员）
func (s *TicketService) GetAllTickets(status *int, page, pageSize int) ([]TicketWithUser, int64, error) {
	tickets, total, err := s.ticketRepo.FindAll(status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	result := make([]TicketWithUser, 0, len(tickets))
	for _, ticket := range tickets {
		user, _ := s.userRepo.FindByID(ticket.UserID)
		item := TicketWithUser{
			Ticket: ticket,
		}
		if user != nil {
			item.UserEmail = user.Email
		}
		result = append(result, item)
	}

	return result, total, nil
}

// TicketMessageWithUser 带用户信息的工单消息
type TicketMessageWithUser struct {
	model.TicketMessage
	UserEmail string `json:"user_email"`
	IsAdmin   bool   `json:"is_admin"`
}

// TicketMessageView 前端展示用的消息格式
type TicketMessageView struct {
	ID        int64  `json:"id"`
	Message   string `json:"message"`
	IsMe      bool   `json:"is_me"`
	CreatedAt int64  `json:"created_at"`
}

// TicketDetail 工单详情
type TicketDetail struct {
	model.Ticket
	Messages  []TicketMessageView `json:"messages"`
	UserEmail string              `json:"user_email"`
}

// TicketWithUser 带用户信息的工单
type TicketWithUser struct {
	model.Ticket
	UserEmail string `json:"user_email"`
}
