package model

// StatUser 用户流量统计
type StatUser struct {
	ID         int64   `gorm:"primaryKey;column:id" json:"id"`
	UserID     int64   `gorm:"column:user_id;index" json:"user_id"`
	ServerRate float64 `gorm:"column:server_rate;type:decimal(10,2)" json:"server_rate"`
	U          int64   `gorm:"column:u" json:"u"`
	D          int64   `gorm:"column:d" json:"d"`
	RecordType string  `gorm:"column:record_type;size:2" json:"record_type"`
	RecordAt   int64   `gorm:"column:record_at;index" json:"record_at"`
	CreatedAt  int64   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  int64   `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (StatUser) TableName() string {
	return "v2_stat_user"
}

// StatServer 节点流量统计
type StatServer struct {
	ID         int64  `gorm:"primaryKey;column:id" json:"id"`
	ServerID   int64  `gorm:"column:server_id;index" json:"server_id"`
	ServerType string `gorm:"column:server_type;size:11" json:"server_type"`
	U          int64  `gorm:"column:u" json:"u"`
	D          int64  `gorm:"column:d" json:"d"`
	RecordType string `gorm:"column:record_type;size:1" json:"record_type"`
	RecordAt   int64  `gorm:"column:record_at;index" json:"record_at"`
	CreatedAt  int64  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  int64  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (StatServer) TableName() string {
	return "v2_stat_server"
}

// Stat 全局统计
type Stat struct {
	ID                int64  `gorm:"primaryKey;column:id" json:"id"`
	RecordAt          int64  `gorm:"column:record_at;uniqueIndex" json:"record_at"`
	RecordType        string `gorm:"column:record_type;size:1" json:"record_type"`
	OrderCount        int    `gorm:"column:order_count" json:"order_count"`
	OrderTotal        int64  `gorm:"column:order_total" json:"order_total"`
	CommissionCount   int    `gorm:"column:commission_count" json:"commission_count"`
	CommissionTotal   int64  `gorm:"column:commission_total" json:"commission_total"`
	PaidCount         int    `gorm:"column:paid_count" json:"paid_count"`
	PaidTotal         int64  `gorm:"column:paid_total" json:"paid_total"`
	RegisterCount     int    `gorm:"column:register_count" json:"register_count"`
	InviteCount       int    `gorm:"column:invite_count" json:"invite_count"`
	TransferUsedTotal string `gorm:"column:transfer_used_total;size:32" json:"transfer_used_total"`
	CreatedAt         int64  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt         int64  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Stat) TableName() string {
	return "v2_stat"
}

// ServerLog 节点日志（流量记录）
type ServerLog struct {
	ID        int64   `gorm:"primaryKey;column:id" json:"id"`
	UserID    int64   `gorm:"column:user_id;index" json:"user_id"`
	ServerID  int64   `gorm:"column:server_id;index" json:"server_id"`
	U         int64   `gorm:"column:u" json:"u"`
	D         int64   `gorm:"column:d" json:"d"`
	Rate      float64 `gorm:"column:rate;type:decimal(10,2)" json:"rate"`
	Method    string  `gorm:"column:method;size:32" json:"method"`
	CreatedAt int64   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt int64   `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ServerLog) TableName() string {
	return "v2_server_log"
}
