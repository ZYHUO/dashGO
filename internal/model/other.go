package model

// Setting 系统设置
type Setting struct {
	ID        int64  `gorm:"primaryKey;column:id" json:"id"`
	Key       string `gorm:"column:key;uniqueIndex;size:255" json:"key"`
	Value     string `gorm:"column:value;type:text" json:"value"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt int64  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Setting) TableName() string {
	return "v2_settings"
}

// Notice 公告
type Notice struct {
	ID        int64   `gorm:"primaryKey;column:id" json:"id"`
	Title     string  `gorm:"column:title" json:"title"`
	Content   string  `gorm:"column:content;type:text" json:"content"`
	Show      bool    `gorm:"column:show;default:false" json:"show"`
	ImgURL    *string `gorm:"column:img_url" json:"img_url"`
	Tags      *string `gorm:"column:tags" json:"tags"`
	Sort      *int    `gorm:"column:sort" json:"sort"`
	CreatedAt int64   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt int64   `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Notice) TableName() string {
	return "v2_notice"
}

// Ticket 工单
type Ticket struct {
	ID          int64  `gorm:"primaryKey;column:id" json:"id"`
	UserID      int64  `gorm:"column:user_id;index" json:"user_id"`
	Subject     string `gorm:"column:subject" json:"subject"`
	Level       int    `gorm:"column:level" json:"level"`
	Status      int    `gorm:"column:status;default:0" json:"status"`
	ReplyStatus int    `gorm:"column:reply_status;default:1" json:"reply_status"`
	CreatedAt   int64  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Ticket) TableName() string {
	return "v2_ticket"
}

// TicketMessage 工单消息
type TicketMessage struct {
	ID        int64  `gorm:"primaryKey;column:id" json:"id"`
	UserID    int64  `gorm:"column:user_id" json:"user_id"`
	TicketID  int64  `gorm:"column:ticket_id;index" json:"ticket_id"`
	Message   string `gorm:"column:message;type:text" json:"message"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt int64  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (TicketMessage) TableName() string {
	return "v2_ticket_message"
}

// Knowledge 知识库
type Knowledge struct {
	ID        int64   `gorm:"primaryKey;column:id" json:"id"`
	Language  string  `gorm:"column:language;size:5" json:"language"`
	Category  string  `gorm:"column:category" json:"category"`
	Title     string  `gorm:"column:title" json:"title"`
	Body      string  `gorm:"column:body;type:text" json:"body"`
	Sort      *int    `gorm:"column:sort" json:"sort"`
	Show      bool    `gorm:"column:show;default:false" json:"show"`
	CreatedAt int64   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt int64   `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Knowledge) TableName() string {
	return "v2_knowledge"
}

// InviteCode 邀请码
type InviteCode struct {
	ID        int64  `gorm:"primaryKey;column:id" json:"id"`
	UserID    int64  `gorm:"column:user_id;index" json:"user_id"`
	Code      string `gorm:"column:code;size:32" json:"code"`
	Status    bool   `gorm:"column:status;default:false" json:"status"`
	PV        int    `gorm:"column:pv;default:0" json:"pv"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt int64  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (InviteCode) TableName() string {
	return "v2_invite_code"
}

// CommissionLog 佣金记录
type CommissionLog struct {
	ID           int64  `gorm:"primaryKey;column:id" json:"id"`
	InviteUserID int64  `gorm:"column:invite_user_id" json:"invite_user_id"`
	UserID       int64  `gorm:"column:user_id" json:"user_id"`
	TradeNo      string `gorm:"column:trade_no;size:36" json:"trade_no"`
	OrderAmount  int64  `gorm:"column:order_amount" json:"order_amount"`
	GetAmount    int64  `gorm:"column:get_amount" json:"get_amount"`
	CreatedAt    int64  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    int64  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (CommissionLog) TableName() string {
	return "v2_commission_log"
}
