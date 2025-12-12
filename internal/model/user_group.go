package model

// UserGroup 用户组模型（核心�?
// 用户组决定用户可以访问哪些节点和购买哪些套餐
// 注意：流量、速度、设备限制等由套餐（Plan）决定，不在用户组中设置
type UserGroup struct {
	ID          int64     `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name" json:"name"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	ServerIDs   JSONArray `gorm:"column:server_ids;type:json" json:"server_ids"` // 该组可访问的节点ID列表
	PlanIDs     JSONArray `gorm:"column:plan_ids;type:json" json:"plan_ids"`     // 该组可购买的套餐ID列表
	Sort        int       `gorm:"column:sort;default:0" json:"sort"`
	CreatedAt   int64     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   int64     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	
	// 保留这些字段用于向后兼容，但标记为废�?
	// Deprecated: 流量应该由套餐决定，不应该在用户组中设置
	DefaultTransferEnable int64 `gorm:"column:default_transfer_enable;default:0" json:"default_transfer_enable,omitempty"`
	// Deprecated: 速度限制应该由套餐决�?
	DefaultSpeedLimit *int `gorm:"column:default_speed_limit" json:"default_speed_limit,omitempty"`
	// Deprecated: 设备限制应该由套餐决�?
	DefaultDeviceLimit *int `gorm:"column:default_device_limit" json:"default_device_limit,omitempty"`
}

func (UserGroup) TableName() string {
	return "v2_user_group"
}

// GetServerIDsAsInt64 获取 server_ids �?int64 数组
func (g *UserGroup) GetServerIDsAsInt64() []int64 {
	result := make([]int64, 0)
	for _, v := range g.ServerIDs {
		switch val := v.(type) {
		case float64:
			result = append(result, int64(val))
		case int64:
			result = append(result, val)
		case int:
			result = append(result, int64(val))
		}
	}
	return result
}

// GetPlanIDsAsInt64 获取 plan_ids �?int64 数组
func (g *UserGroup) GetPlanIDsAsInt64() []int64 {
	result := make([]int64, 0)
	for _, v := range g.PlanIDs {
		switch val := v.(type) {
		case float64:
			result = append(result, int64(val))
		case int64:
			result = append(result, val)
		case int:
			result = append(result, int64(val))
		}
	}
	return result
}

// HasServer 检查该组是否可以访问指定节�?
func (g *UserGroup) HasServer(serverID int64) bool {
	serverIDs := g.GetServerIDsAsInt64()
	for _, id := range serverIDs {
		if id == serverID {
			return true
		}
	}
	return false
}

// HasPlan 检查该组是否可以购买指定套�?
func (g *UserGroup) HasPlan(planID int64) bool {
	planIDs := g.GetPlanIDsAsInt64()
	for _, id := range planIDs {
		if id == planID {
			return true
		}
	}
	return false
}
