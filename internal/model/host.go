package model

// Host 主机模型 - 运行 sing-box 的服务器
type Host struct {
	ID            int64   `gorm:"primaryKey;column:id" json:"id"`
	Name          string  `gorm:"column:name" json:"name"`
	Token         string  `gorm:"column:token;size:64;uniqueIndex" json:"token"`
	IP            string  `gorm:"column:ip" json:"ip"`
	AgentPort     int     `gorm:"column:agent_port;default:9999" json:"agent_port"`
	Status        int     `gorm:"column:status;default:0" json:"status"` // 0=离线 1=在线
	LastHeartbeat *int64  `gorm:"column:last_heartbeat" json:"last_heartbeat"`
	SystemInfo    JSONMap `gorm:"column:system_info;type:json" json:"system_info"`
	// SOCKS 出口配置（可选）
	SocksOutbound *string `gorm:"column:socks_outbound;type:text" json:"socks_outbound"` // SOCKS5 代理地址，格式：socks5://user:pass@host:port
	CreatedAt     int64   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     int64   `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Host) TableName() string {
	return "v2_host"
}

// HostStatus 主机状态
const (
	HostStatusOffline = 0
	HostStatusOnline  = 1
)

// ServerNode 节点模型 - 主机上运行的服务
type ServerNode struct {
	ID                int64     `gorm:"primaryKey;column:id" json:"id"`
	HostID            int64     `gorm:"column:host_id;index" json:"host_id"`
	Name              string    `gorm:"column:name" json:"name"`
	Type              string    `gorm:"column:type" json:"type"` // shadowsocks, vless, trojan 等
	ListenPort        int       `gorm:"column:listen_port" json:"listen_port"`
	GroupIDs          JSONArray `gorm:"column:group_ids;type:json" json:"group_ids"`
	Rate              float64   `gorm:"column:rate;default:1" json:"rate"`
	Show              bool      `gorm:"column:show;default:true" json:"show"`
	Sort              *int      `gorm:"column:sort" json:"sort"`
	ProtocolSettings  JSONMap   `gorm:"column:protocol_settings;type:json" json:"protocol_settings"`
	TLSSettings       JSONMap   `gorm:"column:tls_settings;type:json" json:"tls_settings"`
	TransportSettings JSONMap   `gorm:"column:transport_settings;type:json" json:"transport_settings"`
	CreatedAt         int64     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt         int64     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ServerNode) TableName() string {
	return "v2_server_node"
}

// 节点类型
const (
	NodeTypeShadowsocks = "shadowsocks"
	NodeTypeVMess       = "vmess"
	NodeTypeVLESS       = "vless"
	NodeTypeTrojan      = "trojan"
	NodeTypeHysteria2   = "hysteria2"
	NodeTypeTUIC        = "tuic"
	NodeTypeAnyTLS      = "anytls"
	NodeTypeShadowTLS   = "shadowtls"
	NodeTypeNaive       = "naive"
	NodeTypeSocks       = "socks"
	NodeTypeHTTP        = "http"
)

// GetGroupIDsAsInt64 获取 group_ids 为 int64 数组
func (n *ServerNode) GetGroupIDsAsInt64() []int64 {
	result := make([]int64, 0)
	for _, v := range n.GroupIDs {
		switch val := v.(type) {
		case float64:
			result = append(result, int64(val))
		}
	}
	return result
}
