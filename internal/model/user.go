package model

import "time"

// User 用户模型
type User struct {
	ID                  int64   `gorm:"primaryKey;column:id" json:"id"`
	InviteUserID        *int64  `gorm:"column:invite_user_id" json:"invite_user_id"`
	TelegramID          *int64  `gorm:"column:telegram_id" json:"telegram_id"`
	Email               string  `gorm:"column:email;uniqueIndex;size:64" json:"email"`
	Password            string  `gorm:"column:password;size:64" json:"-"`
	PasswordAlgo        *string `gorm:"column:password_algo;size:10" json:"-"`
	PasswordSalt        *string `gorm:"column:password_salt;size:10" json:"-"`
	Balance             int64   `gorm:"column:balance;default:0" json:"balance"`
	Discount            *int    `gorm:"column:discount" json:"discount"`
	CommissionType      int     `gorm:"column:commission_type;default:0" json:"commission_type"`
	CommissionRate      *int    `gorm:"column:commission_rate" json:"commission_rate"`
	CommissionBalance   int64   `gorm:"column:commission_balance;default:0" json:"commission_balance"`
	T                   int64   `gorm:"column:t;default:0" json:"t"`
	U                   int64   `gorm:"column:u;default:0" json:"u"`
	D                   int64   `gorm:"column:d;default:0" json:"d"`
	TransferEnable      int64   `gorm:"column:transfer_enable;default:0" json:"transfer_enable"`
	Banned              bool    `gorm:"column:banned;default:false" json:"banned"`
	IsAdmin             bool    `gorm:"column:is_admin;default:false" json:"is_admin"`
	IsStaff             bool    `gorm:"column:is_staff;default:false" json:"is_staff"`
	LastLoginAt         *int64  `gorm:"column:last_login_at" json:"last_login_at"`
	LastLoginIP         *int    `gorm:"column:last_login_ip" json:"last_login_ip"`
	UUID                string  `gorm:"column:uuid;size:36" json:"uuid"`
	GroupID             *int64  `gorm:"column:group_id" json:"group_id"`
	PlanID              *int64  `gorm:"column:plan_id" json:"plan_id"`
	SpeedLimit          *int    `gorm:"column:speed_limit" json:"speed_limit"`
	DeviceLimit         *int    `gorm:"column:device_limit" json:"device_limit"`
	RemindExpire        *int    `gorm:"column:remind_expire;default:1" json:"remind_expire"`
	RemindTraffic       *int    `gorm:"column:remind_traffic;default:1" json:"remind_traffic"`
	Token               string  `gorm:"column:token;size:32" json:"token"`
	ExpiredAt           *int64  `gorm:"column:expired_at;default:0" json:"expired_at"`
	Remarks             *string `gorm:"column:remarks;type:text" json:"remarks"`
	LastCheckinAt       *int64  `gorm:"column:last_checkin_at" json:"last_checkin_at"`
	RegisterIP          *string `gorm:"column:register_ip;size:45" json:"register_ip"`
	CreatedAt           int64   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt           int64   `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// 关联
	Plan *Plan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

func (User) TableName() string {
	return "v2_user"
}

// IsActive 检查用户是否活�?
func (u *User) IsActive() bool {
	if u.Banned {
		return false
	}
	if u.ExpiredAt != nil && *u.ExpiredAt > 0 && *u.ExpiredAt < time.Now().Unix() {
		return false
	}
	if u.PlanID == nil {
		return false
	}
	return true
}

// HasTraffic 检查用户是否有剩余流量
func (u *User) HasTraffic() bool {
	return u.U+u.D < u.TransferEnable
}

// GetUsedTraffic 获取已使用流�?
func (u *User) GetUsedTraffic() int64 {
	return u.U + u.D
}

// GetRemainingTraffic 获取剩余流量
func (u *User) GetRemainingTraffic() int64 {
	remaining := u.TransferEnable - u.U - u.D
	if remaining < 0 {
		return 0
	}
	return remaining
}
