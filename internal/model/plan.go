package model

// Plan 套餐模型
type Plan struct {
	ID                 int64     `gorm:"primaryKey;column:id" json:"id"`
	GroupID            *int64    `gorm:"column:group_id" json:"group_id"`
	TransferEnable     int64     `gorm:"column:transfer_enable" json:"transfer_enable"`       // 流量配额（字节）
	Name               string    `gorm:"column:name" json:"name"`
	SpeedLimit         *int      `gorm:"column:speed_limit" json:"speed_limit"`               // 速度限制（Mbps�?
	DeviceLimit        *int      `gorm:"column:device_limit" json:"device_limit"`             // 设备数量限制
	Show               bool      `gorm:"column:show;default:false" json:"show"`
	Sell               bool      `gorm:"column:sell;default:true" json:"sell"`
	Renew              bool      `gorm:"column:renew;default:true" json:"renew"`
	Sort               int       `gorm:"column:sort" json:"sort"`
	Content            string    `gorm:"column:content;type:text" json:"content"`
	MonthPrice         *int64    `gorm:"column:month_price" json:"month_price"`
	QuarterPrice       *int64    `gorm:"column:quarter_price" json:"quarter_price"`
	HalfYearPrice      *int64    `gorm:"column:half_year_price" json:"half_year_price"`
	YearPrice          *int64    `gorm:"column:year_price" json:"year_price"`
	TwoYearPrice       *int64    `gorm:"column:two_year_price" json:"two_year_price"`
	ThreeYearPrice     *int64    `gorm:"column:three_year_price" json:"three_year_price"`
	OnetimePrice       *int64    `gorm:"column:onetime_price" json:"onetime_price"`
	ResetPrice         *int64    `gorm:"column:reset_price" json:"reset_price"`
	ResetTrafficMethod *int      `gorm:"column:reset_traffic_method" json:"reset_traffic_method"`
	CapacityLimit      *int      `gorm:"column:capacity_limit" json:"capacity_limit"`         // 最大可售数量（null�?=不限制）
	SoldCount          int       `gorm:"column:sold_count;default:0" json:"sold_count"`       // 已售出数�?
	UpgradeGroupID     *int64    `gorm:"column:upgrade_group_id" json:"upgrade_group_id"`     // 购买后升级到的用户组ID
	CreatedAt          int64     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          int64     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Plan) TableName() string {
	return "v2_plan"
}

// 流量重置方式
const (
	ResetTrafficFollowSystem  = -1 // 跟随系统设置
	ResetTrafficFirstDayMonth = 0  // 每月1�?
	ResetTrafficMonthly       = 1  // 按月重置
	ResetTrafficNever         = 2  // 不重�?
	ResetTrafficFirstDayYear  = 3  // 每年1�?�?
	ResetTrafficYearly        = 4  // 按年重置
)

// 订阅周期
const (
	PeriodMonthly     = "monthly"
	PeriodQuarterly   = "quarterly"
	PeriodHalfYearly  = "half_yearly"
	PeriodYearly      = "yearly"
	PeriodTwoYearly   = "two_yearly"
	PeriodThreeYearly = "three_yearly"
	PeriodOnetime     = "onetime"
	PeriodResetTraffic = "reset_traffic"
)

// GetPeriodDays 获取周期天数
func GetPeriodDays(period string) int {
	switch period {
	case PeriodMonthly:
		return 30
	case PeriodQuarterly:
		return 90
	case PeriodHalfYearly:
		return 180
	case PeriodYearly:
		return 365
	case PeriodTwoYearly:
		return 730
	case PeriodThreeYearly:
		return 1095
	case PeriodOnetime:
		return -1
	default:
		return 0
	}
}

// GetPriceByPeriod 获取指定周期的价�?
func (p *Plan) GetPriceByPeriod(period string) int64 {
	switch period {
	case PeriodMonthly:
		if p.MonthPrice != nil {
			return *p.MonthPrice
		}
	case PeriodQuarterly:
		if p.QuarterPrice != nil {
			return *p.QuarterPrice
		}
	case PeriodHalfYearly:
		if p.HalfYearPrice != nil {
			return *p.HalfYearPrice
		}
	case PeriodYearly:
		if p.YearPrice != nil {
			return *p.YearPrice
		}
	case PeriodTwoYearly:
		if p.TwoYearPrice != nil {
			return *p.TwoYearPrice
		}
	case PeriodThreeYearly:
		if p.ThreeYearPrice != nil {
			return *p.ThreeYearPrice
		}
	case PeriodOnetime:
		if p.OnetimePrice != nil {
			return *p.OnetimePrice
		}
	case PeriodResetTraffic:
		if p.ResetPrice != nil {
			return *p.ResetPrice
		}
	}
	return 0
}

// CanPurchase 检查套餐是否可以购�?
func (p *Plan) CanPurchase() bool {
	// 如果没有设置限制，可以购�?
	if p.CapacityLimit == nil || *p.CapacityLimit <= 0 {
		return true
	}
	
	// 检查是否还有剩�?
	return p.SoldCount < *p.CapacityLimit
}

// GetRemainingCount 获取剩余可售数量
func (p *Plan) GetRemainingCount() int {
	// 如果没有设置限制，返�?-1 表示不限�?
	if p.CapacityLimit == nil || *p.CapacityLimit <= 0 {
		return -1
	}
	
	remaining := *p.CapacityLimit - p.SoldCount
	if remaining < 0 {
		return 0
	}
	return remaining
}
