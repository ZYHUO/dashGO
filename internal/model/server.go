package model

import (
	"database/sql/driver"
	"encoding/json"
)

// Server 节点模型
type Server struct {
	ID               int64     `gorm:"primaryKey;column:id" json:"id"`
	Type             string    `gorm:"column:type" json:"type"`
	Code             *string   `gorm:"column:code" json:"code"`
	ParentID         *int64    `gorm:"column:parent_id" json:"parent_id"`
	HostID           *int64    `gorm:"column:host_id;index" json:"host_id"` // 绑定的主机ID，用于自动部�?
	GroupIDs         JSONArray `gorm:"column:group_ids;type:json" json:"group_ids"`
	RouteIDs         JSONArray `gorm:"column:route_ids;type:json" json:"route_ids"`
	Name             string    `gorm:"column:name" json:"name"`
	Rate             float64   `gorm:"column:rate" json:"rate"`
	Tags             JSONArray `gorm:"column:tags;type:json" json:"tags"`
	Host             string    `gorm:"column:host" json:"host"`
	Port             string    `gorm:"column:port" json:"port"`
	ServerPort       int       `gorm:"column:server_port" json:"server_port"`
	ProtocolSettings JSONMap   `gorm:"column:protocol_settings;type:json" json:"protocol_settings"`
	Show             bool      `gorm:"column:show;default:false" json:"show"`
	Sort             *int      `gorm:"column:sort" json:"sort"`
	RateTimeEnable   bool      `gorm:"column:rate_time_enable;default:false" json:"rate_time_enable"`
	RateTimeRanges   JSONArray `gorm:"column:rate_time_ranges;type:json" json:"rate_time_ranges"`
	CreatedAt        int64     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        int64     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Server) TableName() string {
	return "v2_server"
}

// Server types
const (
	ServerTypeHysteria    = "hysteria"
	ServerTypeVless       = "vless"
	ServerTypeTrojan      = "trojan"
	ServerTypeVmess       = "vmess"
	ServerTypeTuic        = "tuic"
	ServerTypeShadowsocks = "shadowsocks"
	ServerTypeAnytls      = "anytls"
	ServerTypeSocks       = "socks"
	ServerTypeNaive       = "naive"
	ServerTypeHTTP        = "http"
	ServerTypeMieru       = "mieru"
)

// JSONArray 用于存储 JSON 数组
type JSONArray []interface{}

func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			*j = nil
			return nil
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

// JSONMap 用于存储 JSON 对象
type JSONMap map[string]interface{}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			*j = nil
			return nil
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

// GetGroupIDsAsInt64 获取 group_ids �?int64 数组
func (s *Server) GetGroupIDsAsInt64() []int64 {
	result := make([]int64, 0)
	for _, v := range s.GroupIDs {
		switch val := v.(type) {
		case float64:
			result = append(result, int64(val))
		case string:
			// 尝试解析字符�?
		}
	}
	return result
}

// ServerGroup 节点�?
type ServerGroup struct {
	ID        int64  `gorm:"primaryKey;column:id" json:"id"`
	Name      string `gorm:"column:name" json:"name"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt int64  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ServerGroup) TableName() string {
	return "v2_server_group"
}

// ServerRoute 路由规则
type ServerRoute struct {
	ID          int64   `gorm:"primaryKey;column:id" json:"id"`
	Remarks     string  `gorm:"column:remarks" json:"remarks"`
	Match       string  `gorm:"column:match;type:text" json:"match"`
	Action      string  `gorm:"column:action;size:11" json:"action"`
	ActionValue *string `gorm:"column:action_value" json:"action_value"`
	CreatedAt   int64   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   int64   `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ServerRoute) TableName() string {
	return "v2_server_route"
}
