package model

// Order 订单模型
type Order struct {
	ID                      int64     `gorm:"primaryKey;column:id" json:"id"`
	InviteUserID            *int64    `gorm:"column:invite_user_id" json:"invite_user_id"`
	UserID                  int64     `gorm:"column:user_id;index" json:"user_id"`
	PlanID                  int64     `gorm:"column:plan_id" json:"plan_id"`
	CouponID                *int64    `gorm:"column:coupon_id" json:"coupon_id"`
	PaymentID               *int64    `gorm:"column:payment_id" json:"payment_id"`
	Type                    int       `gorm:"column:type" json:"type"`
	Period                  string    `gorm:"column:period" json:"period"`
	TradeNo                 string    `gorm:"column:trade_no;uniqueIndex;size:36" json:"trade_no"`
	CallbackNo              *string   `gorm:"column:callback_no" json:"callback_no"`
	TotalAmount             int64     `gorm:"column:total_amount" json:"total_amount"`
	HandlingAmount          *int64    `gorm:"column:handling_amount" json:"handling_amount"`
	DiscountAmount          *int64    `gorm:"column:discount_amount" json:"discount_amount"`
	SurplusAmount           *int64    `gorm:"column:surplus_amount" json:"surplus_amount"`
	RefundAmount            *int64    `gorm:"column:refund_amount" json:"refund_amount"`
	BalanceAmount           *int64    `gorm:"column:balance_amount" json:"balance_amount"`
	SurplusOrderIDs         JSONArray `gorm:"column:surplus_order_ids;type:json" json:"surplus_order_ids"`
	Status                  int       `gorm:"column:status;default:0" json:"status"`
	CommissionStatus        int       `gorm:"column:commission_status;default:0" json:"commission_status"`
	CommissionBalance       int64     `gorm:"column:commission_balance;default:0" json:"commission_balance"`
	ActualCommissionBalance *int64    `gorm:"column:actual_commission_balance" json:"actual_commission_balance"`
	PaidAt                  *int64    `gorm:"column:paid_at" json:"paid_at"`
	CreatedAt               int64     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt               int64     `gorm:"column:updated_at;autoUpdateTime;index" json:"updated_at"`
}

func (Order) TableName() string {
	return "v2_order"
}

// 订单状态
const (
	OrderStatusPending    = 0 // 待支付
	OrderStatusProcessing = 1 // 开通中
	OrderStatusCancelled  = 2 // 已取消
	OrderStatusCompleted  = 3 // 已完成
	OrderStatusDiscounted = 4 // 已折抵
)

// 订单类型
const (
	OrderTypeNewPurchase  = 1 // 新购
	OrderTypeRenewal      = 2 // 续费
	OrderTypeUpgrade      = 3 // 升级
	OrderTypeResetTraffic = 4 // 流量重置
)

// Payment 支付方式
type Payment struct {
	ID                 int64    `gorm:"primaryKey;column:id" json:"id"`
	UUID               string   `gorm:"column:uuid;size:32" json:"uuid"`
	Payment            string   `gorm:"column:payment;size:16" json:"payment"`
	Name               string   `gorm:"column:name" json:"name"`
	Icon               *string  `gorm:"column:icon" json:"icon"`
	Config             string   `gorm:"column:config;type:text" json:"config"`
	NotifyDomain       *string  `gorm:"column:notify_domain;size:128" json:"notify_domain"`
	HandlingFeeFixed   *int64   `gorm:"column:handling_fee_fixed" json:"handling_fee_fixed"`
	HandlingFeePercent *float64 `gorm:"column:handling_fee_percent;type:decimal(5,2)" json:"handling_fee_percent"`
	Enable             bool     `gorm:"column:enable;default:false" json:"enable"`
	Sort               *int     `gorm:"column:sort" json:"sort"`
	CreatedAt          int64    `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          int64    `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Payment) TableName() string {
	return "v2_payment"
}

// Coupon 优惠券
type Coupon struct {
	ID               int64   `gorm:"primaryKey;column:id" json:"id"`
	Code             string  `gorm:"column:code" json:"code"`
	Name             string  `gorm:"column:name" json:"name"`
	Type             int     `gorm:"column:type" json:"type"`
	Value            int64   `gorm:"column:value" json:"value"`
	Show             bool    `gorm:"column:show;default:false" json:"show"`
	LimitUse         *int    `gorm:"column:limit_use" json:"limit_use"`
	LimitUseWithUser *int    `gorm:"column:limit_use_with_user" json:"limit_use_with_user"`
	LimitPlanIDs     *string `gorm:"column:limit_plan_ids" json:"limit_plan_ids"`
	LimitPeriod      *string `gorm:"column:limit_period" json:"limit_period"`
	StartedAt        int64   `gorm:"column:started_at" json:"started_at"`
	EndedAt          int64   `gorm:"column:ended_at" json:"ended_at"`
	CreatedAt        int64   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        int64   `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Coupon) TableName() string {
	return "v2_coupon"
}

// 优惠券类型
const (
	CouponTypeAmount  = 1 // 固定金额
	CouponTypePercent = 2 // 百分比折扣
)
