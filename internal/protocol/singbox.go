package protocol

import (
	"encoding/json"
	"strings"

	"dashgo/internal/model"
	"dashgo/internal/service"
)

// SingBoxConfig sing-box 配置结构
type SingBoxConfig struct {
	Log       *LogConfig       `json:"log,omitempty"`
	DNS       *DNSConfig       `json:"dns,omitempty"`
	Inbounds  []Inbound        `json:"inbounds,omitempty"`
	Outbounds []Outbound       `json:"outbounds"`
	Route     *RouteConfig     `json:"route,omitempty"`
}

type LogConfig struct {
	Level     string `json:"level,omitempty"`
	Timestamp bool   `json:"timestamp,omitempty"`
}

type DNSConfig struct {
	Servers []DNSServer `json:"servers,omitempty"`
}

type DNSServer struct {
	Tag     string `json:"tag,omitempty"`
	Address string `json:"address"`
}

type Inbound struct {
	Type   string `json:"type"`
	Tag    string `json:"tag,omitempty"`
	Listen string `json:"listen,omitempty"`
	Port   int    `json:"listen_port,omitempty"`
}

type Outbound struct {
	Type           string                 `json:"type"`
	Tag            string                 `json:"tag"`
	Server         string                 `json:"server,omitempty"`
	ServerPort     int                    `json:"server_port,omitempty"`
	UUID           string                 `json:"uuid,omitempty"`
	Password       string                 `json:"password,omitempty"`
	Method         string                 `json:"method,omitempty"`
	TLS            *TLSConfig             `json:"tls,omitempty"`
	Transport      map[string]interface{} `json:"transport,omitempty"`
	Flow           string                 `json:"flow,omitempty"`
	Outbounds      []string               `json:"outbounds,omitempty"`
	// Hysteria specific
	UpMbps         int                    `json:"up_mbps,omitempty"`
	DownMbps       int                    `json:"down_mbps,omitempty"`
	Obfs           interface{}            `json:"obfs,omitempty"`
	// TUIC specific
	CongestionControl string               `json:"congestion_control,omitempty"`
	UDPRelayMode      string               `json:"udp_relay_mode,omitempty"`
}

type TLSConfig struct {
	Enabled    bool        `json:"enabled"`
	ServerName string      `json:"server_name,omitempty"`
	Insecure   bool        `json:"insecure,omitempty"`
	ALPN       []string    `json:"alpn,omitempty"`
	UTLS       *UTLSConfig `json:"utls,omitempty"`
	Reality    *Reality    `json:"reality,omitempty"`
}

type UTLSConfig struct {
	Enabled     bool   `json:"enabled"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

type Reality struct {
	Enabled   bool   `json:"enabled"`
	PublicKey string `json:"public_key,omitempty"`
	ShortID   string `json:"short_id,omitempty"`
}

type RouteConfig struct {
	Rules        []RouteRule `json:"rules,omitempty"`
	Final        string      `json:"final,omitempty"`
	AutoDetect   bool        `json:"auto_detect_interface,omitempty"`
}

type RouteRule struct {
	Protocol []string `json:"protocol,omitempty"`
	Outbound string   `json:"outbound"`
}

// GenerateSingBoxConfig 生成 sing-box 配置
func GenerateSingBoxConfig(servers []service.ServerInfo, user *model.User) map[string]interface{} {
	config := getDefaultSingBoxConfig()
	
	outbounds := config["outbounds"].([]interface{})
	proxyTags := []string{}

	for _, server := range servers {
		outbound := buildSingBoxOutbound(server, user)
		if outbound != nil {
			outbounds = append(outbounds, outbound)
			proxyTags = append(proxyTags, server.Name)
		}
	}

	// 更新各个分组�?outbounds
	for i, ob := range outbounds {
		if m, ok := ob.(map[string]interface{}); ok {
			tag, _ := m["tag"].(string)
			outType, _ := m["type"].(string)
			
			switch tag {
			case "🚀 节点选择":
				// 节点选择：添加所有节�?
				if existing, ok := m["outbounds"].([]string); ok {
					m["outbounds"] = append(existing, proxyTags...)
				}
			case "♻️ 自动选择", "🔯 故障转移":
				// 自动选择/故障转移：只包含节点
				m["outbounds"] = proxyTags
			case "📲 电报消息", "🤖 OpenAI", "📹 YouTube", "🎬 Netflix", "🍎 苹果服务", "🐟 漏网之鱼":
				// 其他分组：添加所有节�?
				if existing, ok := m["outbounds"].([]string); ok {
					m["outbounds"] = append(existing, proxyTags...)
				}
			default:
				// 其他 selector/urltest 类型
				if outType == "selector" || outType == "urltest" {
					if existing, ok := m["outbounds"].([]string); ok {
						m["outbounds"] = append(existing, proxyTags...)
					} else {
						m["outbounds"] = proxyTags
					}
				}
			}
			outbounds[i] = m
		}
	}

	config["outbounds"] = outbounds
	return config
}

func buildSingBoxOutbound(server service.ServerInfo, user *model.User) map[string]interface{} {
	ps := server.ProtocolSettings

	switch server.Type {
	case model.ServerTypeShadowsocks:
		return buildShadowsocks(server, user)
	case model.ServerTypeVmess:
		return buildVmess(server, user)
	case model.ServerTypeVless:
		return buildVless(server, user)
	case model.ServerTypeTrojan:
		return buildTrojan(server, user)
	case model.ServerTypeHysteria:
		return buildHysteria(server, user)
	case model.ServerTypeTuic:
		return buildTuic(server, user)
	case model.ServerTypeAnytls:
		return buildAnyTLS(server, user)
	case model.ServerTypeSocks:
		return buildSocks(server, user)
	case model.ServerTypeHTTP:
		return buildHTTP(server, user)
	case "shadowtls":
		return buildShadowTLS(server, user)
	case "naive":
		return buildNaive(server, user)
	}

	_ = ps
	return nil
}

func buildShadowTLS(server service.ServerInfo, user *model.User) map[string]interface{} {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	// ShadowTLS 需要配�?Shadowsocks 使用
	out := map[string]interface{}{
		"type":        "shadowtls",
		"tag":         server.Name,
		"server":      server.Host,
		"server_port": port,
		"version":     3,
		"password":    user.UUID,
		"tls": map[string]interface{}{
			"enabled":     true,
			"server_name": "addons.mozilla.org",
			"utls": map[string]interface{}{
				"enabled":     true,
				"fingerprint": "chrome",
			},
		},
	}

	// 握手服务�?
	if hs, ok := ps["handshake_server"].(string); ok && hs != "" {
		out["tls"].(map[string]interface{})["server_name"] = hs
	}

	return out
}

func buildNaive(server service.ServerInfo, user *model.User) map[string]interface{} {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	out := map[string]interface{}{
		"type":        "naive",
		"tag":         server.Name,
		"server":      server.Host,
		"server_port": port,
		"username":    user.UUID[:8],
		"password":    user.UUID,
		"tls": map[string]interface{}{
			"enabled": true,
		},
	}

	if tls, ok := ps["tls"].(map[string]interface{}); ok {
		if sn, ok := tls["server_name"].(string); ok && sn != "" {
			out["tls"].(map[string]interface{})["server_name"] = sn
		}
	}

	return out
}

func buildAnyTLS(server service.ServerInfo, user *model.User) map[string]interface{} {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	out := map[string]interface{}{
		"type":        "anytls",
		"tag":         server.Name,
		"server":      server.Host,
		"server_port": port,
		"password":    user.UUID,
		"tls": map[string]interface{}{
			"enabled": true,
		},
	}

	if tls, ok := ps["tls"].(map[string]interface{}); ok {
		if sn, ok := tls["server_name"].(string); ok {
			out["tls"].(map[string]interface{})["server_name"] = sn
		}
		if insecure, ok := tls["allow_insecure"].(bool); ok {
			out["tls"].(map[string]interface{})["insecure"] = insecure
		}
	}

	// Padding scheme
	if paddingScheme, ok := ps["padding_scheme"].([]interface{}); ok {
		out["padding_scheme"] = paddingScheme
	}

	return out
}

func buildSocks(server service.ServerInfo, user *model.User) map[string]interface{} {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	out := map[string]interface{}{
		"type":        "socks",
		"tag":         server.Name,
		"server":      server.Host,
		"server_port": port,
		"version":     "5",
		"username":    user.UUID,
		"password":    user.UUID,
	}

	if udpOverTcp, ok := ps["udp_over_tcp"].(bool); ok && udpOverTcp {
		out["udp_over_tcp"] = true
	}

	return out
}

func buildHTTP(server service.ServerInfo, user *model.User) map[string]interface{} {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	out := map[string]interface{}{
		"type":        "http",
		"tag":         server.Name,
		"server":      server.Host,
		"server_port": port,
		"username":    user.UUID,
		"password":    user.UUID,
	}

	if path, ok := ps["path"].(string); ok {
		out["path"] = path
	}

	if headers, ok := ps["headers"].(map[string]interface{}); ok {
		out["headers"] = headers
	}

	if tls, ok := ps["tls"].(float64); ok && tls > 0 {
		tlsConfig := map[string]interface{}{
			"enabled": true,
		}
		if tlsSettings, ok := ps["tls_settings"].(map[string]interface{}); ok {
			if insecure, ok := tlsSettings["allow_insecure"].(bool); ok {
				tlsConfig["insecure"] = insecure
			}
			if sn, ok := tlsSettings["server_name"].(string); ok {
				tlsConfig["server_name"] = sn
			}
		}
		out["tls"] = tlsConfig
	}

	return out
}

func buildShadowsocks(server service.ServerInfo, user *model.User) map[string]interface{} {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	// 获取加密方式，优先使�?cipher，其�?method，默�?aes-256-gcm
	cipher := "aes-256-gcm"
	if c, ok := ps["cipher"].(string); ok && c != "" {
		cipher = c
	} else if m, ok := ps["method"].(string); ok && m != "" {
		cipher = m
	}

	// 密码：对�?SS2022，使�?server.Password（已包含服务器密�?用户密钥格式�?
	// 对于普�?SS，使用用�?UUID
	password := server.Password
	if password == "" {
		password = user.UUID
	}

	out := map[string]interface{}{
		"type":        "shadowsocks",
		"tag":         server.Name,
		"server":      server.Host,
		"server_port": port,
		"method":      cipher,
		"password":    password,
	}

	if plugin, ok := ps["plugin"].(string); ok && plugin != "" {
		out["plugin"] = plugin
		if opts, ok := ps["plugin_opts"].(string); ok {
			out["plugin_opts"] = opts
		}
	}

	return out
}

func buildVmess(server service.ServerInfo, user *model.User) map[string]interface{} {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	out := map[string]interface{}{
		"type":        "vmess",
		"tag":         server.Name,
		"server":      server.Host,
		"server_port": port,
		"uuid":        user.UUID,
		"security":    "auto",
		"alter_id":    0,
	}

	// TLS
	if tls, ok := ps["tls"].(float64); ok && tls > 0 {
		tlsConfig := map[string]interface{}{
			"enabled": true,
		}
		if tlsSettings, ok := ps["tls_settings"].(map[string]interface{}); ok {
			if sn, ok := tlsSettings["server_name"].(string); ok {
				tlsConfig["server_name"] = sn
			}
			if insecure, ok := tlsSettings["allow_insecure"].(bool); ok {
				tlsConfig["insecure"] = insecure
			}
		}
		out["tls"] = tlsConfig
	}

	// Transport
	if network, ok := ps["network"].(string); ok {
		transport := buildTransport(network, ps)
		if transport != nil {
			out["transport"] = transport
		}
	}

	return out
}

func buildVless(server service.ServerInfo, user *model.User) map[string]interface{} {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	out := map[string]interface{}{
		"type":            "vless",
		"tag":             server.Name,
		"server":          server.Host,
		"server_port":     port,
		"uuid":            user.UUID,
		"packet_encoding": "xudp",
	}

	if flow, ok := ps["flow"].(string); ok && flow != "" {
		out["flow"] = flow
	}

	// TLS
	if tls, ok := ps["tls"].(float64); ok && tls > 0 {
		tlsConfig := map[string]interface{}{
			"enabled": true,
			"utls": map[string]interface{}{
				"enabled":     true,
				"fingerprint": "chrome",
			},
		}

		if tls == 2 { // Reality
			if reality, ok := ps["reality_settings"].(map[string]interface{}); ok {
				tlsConfig["server_name"] = reality["server_name"]
				tlsConfig["reality"] = map[string]interface{}{
					"enabled":    true,
					"public_key": reality["public_key"],
					"short_id":   reality["short_id"],
				}
			}
		} else {
			if tlsSettings, ok := ps["tls_settings"].(map[string]interface{}); ok {
				if sn, ok := tlsSettings["server_name"].(string); ok {
					tlsConfig["server_name"] = sn
				}
			}
		}
		out["tls"] = tlsConfig
	}

	// Transport
	if network, ok := ps["network"].(string); ok {
		transport := buildTransport(network, ps)
		if transport != nil {
			out["transport"] = transport
		}
	}

	return out
}

func buildTrojan(server service.ServerInfo, user *model.User) map[string]interface{} {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	out := map[string]interface{}{
		"type":        "trojan",
		"tag":         server.Name,
		"server":      server.Host,
		"server_port": port,
		"password":    user.UUID,
		"tls": map[string]interface{}{
			"enabled": true,
		},
	}

	if sn, ok := ps["server_name"].(string); ok && sn != "" {
		out["tls"].(map[string]interface{})["server_name"] = sn
	}
	if insecure, ok := ps["allow_insecure"].(bool); ok {
		out["tls"].(map[string]interface{})["insecure"] = insecure
	}

	// Transport
	if network, ok := ps["network"].(string); ok {
		transport := buildTransport(network, ps)
		if transport != nil {
			out["transport"] = transport
		}
	}

	return out
}

func buildHysteria(server service.ServerInfo, user *model.User) map[string]interface{} {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	version := 2
	if v, ok := ps["version"].(float64); ok {
		version = int(v)
	}

	var outType string
	if version == 2 {
		outType = "hysteria2"
	} else {
		outType = "hysteria"
	}

	out := map[string]interface{}{
		"type":        outType,
		"tag":         server.Name,
		"server":      server.Host,
		"server_port": port,
		"tls": map[string]interface{}{
			"enabled": true,
		},
	}

	// Bandwidth
	if bw, ok := ps["bandwidth"].(map[string]interface{}); ok {
		if up, ok := bw["up"].(float64); ok {
			out["up_mbps"] = int(up)
		}
		if down, ok := bw["down"].(float64); ok {
			out["down_mbps"] = int(down)
		}
	}

	// TLS settings
	if tls, ok := ps["tls"].(map[string]interface{}); ok {
		if sn, ok := tls["server_name"].(string); ok {
			out["tls"].(map[string]interface{})["server_name"] = sn
		}
		if insecure, ok := tls["allow_insecure"].(bool); ok {
			out["tls"].(map[string]interface{})["insecure"] = insecure
		}
	}

	// Password
	if version == 2 {
		out["password"] = user.UUID
		// Obfs for hysteria2
		if obfs, ok := ps["obfs"].(map[string]interface{}); ok {
			if open, ok := obfs["open"].(bool); ok && open {
				out["obfs"] = map[string]interface{}{
					"type":     obfs["type"],
					"password": obfs["password"],
				}
			}
		}
	} else {
		out["auth_str"] = user.UUID
		if obfs, ok := ps["obfs"].(map[string]interface{}); ok {
			if pw, ok := obfs["password"].(string); ok {
				out["obfs"] = pw
			}
		}
	}

	return out
}

func buildTuic(server service.ServerInfo, user *model.User) map[string]interface{} {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	out := map[string]interface{}{
		"type":                "tuic",
		"tag":                 server.Name,
		"server":              server.Host,
		"server_port":         port,
		"uuid":                user.UUID,
		"password":            user.UUID,
		"congestion_control":  "cubic",
		"udp_relay_mode":      "native",
		"zero_rtt_handshake":  true,
		"heartbeat":           "10s",
		"tls": map[string]interface{}{
			"enabled": true,
			"alpn":    []string{"h3"},
		},
	}

	if cc, ok := ps["congestion_control"].(string); ok {
		out["congestion_control"] = cc
	}
	if urm, ok := ps["udp_relay_mode"].(string); ok {
		out["udp_relay_mode"] = urm
	}

	if tls, ok := ps["tls"].(map[string]interface{}); ok {
		if sn, ok := tls["server_name"].(string); ok {
			out["tls"].(map[string]interface{})["server_name"] = sn
		}
	}

	return out
}

func buildTransport(network string, ps model.JSONMap) map[string]interface{} {
	ns, _ := ps["network_settings"].(map[string]interface{})

	switch network {
	case "ws":
		transport := map[string]interface{}{
			"type":                    "ws",
			"max_early_data":          2048,
			"early_data_header_name":  "Sec-WebSocket-Protocol",
		}
		if path, ok := ns["path"].(string); ok {
			transport["path"] = path
		}
		if headers, ok := ns["headers"].(map[string]interface{}); ok {
			if host, ok := headers["Host"].(string); ok {
				transport["headers"] = map[string]string{"Host": host}
			}
		}
		return transport

	case "grpc":
		transport := map[string]interface{}{
			"type": "grpc",
		}
		if sn, ok := ns["serviceName"].(string); ok {
			transport["service_name"] = sn
		}
		return transport

	case "tcp":
		if header, ok := ns["header"].(map[string]interface{}); ok {
			if headerType, ok := header["type"].(string); ok && headerType == "http" {
				return map[string]interface{}{
					"type": "http",
					"path": "/",
				}
			}
		}
	}

	return nil
}

func getDefaultSingBoxConfig() map[string]interface{} {
	return map[string]interface{}{
		"log": map[string]interface{}{
			"level":     "info",
			"timestamp": true,
		},
		"dns": map[string]interface{}{
			"servers": []map[string]interface{}{
				{"tag": "google", "address": "https://dns.google/dns-query", "detour": "🚀 节点选择"},
				{"tag": "cloudflare", "address": "https://cloudflare-dns.com/dns-query", "detour": "🚀 节点选择"},
				{"tag": "alidns", "address": "https://dns.alidns.com/dns-query", "detour": "direct"},
				{"tag": "local", "address": "223.5.5.5", "detour": "direct"},
			},
			"rules": []map[string]interface{}{
				{"domain_suffix": []string{".cn"}, "server": "local"},
				{"geosite": "cn", "server": "local"},
			},
			"final": "google",
		},
		"outbounds": []interface{}{
			map[string]interface{}{
				"type":      "selector",
				"tag":       "🚀 节点选择",
				"outbounds": []string{"♻️ 自动选择", "🔯 故障转移", "direct"},
			},
			map[string]interface{}{
				"type":      "urltest",
				"tag":       "♻️ 自动选择",
				"outbounds": []string{},
				"url":       "https://www.gstatic.com/generate_204",
				"interval":  "5m",
				"tolerance": 50,
			},
			map[string]interface{}{
				"type":      "urltest",
				"tag":       "🔯 故障转移",
				"outbounds": []string{},
				"url":       "https://www.gstatic.com/generate_204",
				"interval":  "5m",
			},
			map[string]interface{}{
				"type":      "selector",
				"tag":       "📲 电报消息",
				"outbounds": []string{"🚀 节点选择", "♻️ 自动选择", "direct"},
			},
			map[string]interface{}{
				"type":      "selector",
				"tag":       "🤖 OpenAI",
				"outbounds": []string{"🚀 节点选择", "♻️ 自动选择"},
			},
			map[string]interface{}{
				"type":      "selector",
				"tag":       "📹 YouTube",
				"outbounds": []string{"🚀 节点选择", "♻️ 自动选择", "direct"},
			},
			map[string]interface{}{
				"type":      "selector",
				"tag":       "🎬 Netflix",
				"outbounds": []string{"🚀 节点选择", "♻️ 自动选择", "direct"},
			},
			map[string]interface{}{
				"type":      "selector",
				"tag":       "🍎 苹果服务",
				"outbounds": []string{"direct", "🚀 节点选择"},
			},
			map[string]interface{}{
				"type":      "selector",
				"tag":       "🐟 漏网之鱼",
				"outbounds": []string{"🚀 节点选择", "♻️ 自动选择", "direct"},
			},
			map[string]interface{}{"type": "direct", "tag": "direct"},
			map[string]interface{}{"type": "block", "tag": "block"},
			map[string]interface{}{"type": "dns", "tag": "dns-out"},
		},
		"route": map[string]interface{}{
			"rules": []map[string]interface{}{
				{"protocol": []string{"dns"}, "outbound": "dns-out"},
				{"ip_is_private": true, "outbound": "direct"},
				// OpenAI
				{"domain_suffix": []string{"openai.com", "ai.com", "anthropic.com", "claude.ai"}, "outbound": "🤖 OpenAI"},
				{"domain_keyword": []string{"openai"}, "outbound": "🤖 OpenAI"},
				// Telegram
				{"domain_suffix": []string{"telegram.org", "t.me", "tg.dev"}, "outbound": "📲 电报消息"},
				{"ip_cidr": []string{"91.108.0.0/16", "109.239.140.0/24", "149.154.160.0/20"}, "outbound": "📲 电报消息"},
				// YouTube
				{"domain_suffix": []string{"youtube.com", "googlevideo.com", "ytimg.com", "yt.be"}, "outbound": "📹 YouTube"},
				// Netflix
				{"domain_suffix": []string{"netflix.com", "netflix.net", "nflximg.com", "nflximg.net", "nflxvideo.net"}, "outbound": "🎬 Netflix"},
				// Apple
				{"domain_suffix": []string{"apple.com", "icloud.com", "icloud-content.com", "mzstatic.com"}, "outbound": "🍎 苹果服务"},
				// Google
				{"domain_suffix": []string{"google.com", "googleapis.com", "gstatic.com", "gmail.com"}, "outbound": "🚀 节点选择"},
				// GitHub
				{"domain_suffix": []string{"github.com", "githubusercontent.com", "githubassets.com"}, "outbound": "🚀 节点选择"},
				// Twitter
				{"domain_suffix": []string{"twitter.com", "x.com", "twimg.com"}, "outbound": "🚀 节点选择"},
				// 国内直连
				{"geosite": "cn", "outbound": "direct"},
				{"geoip": "cn", "outbound": "direct"},
			},
			"final":                 "🐟 漏网之鱼",
			"auto_detect_interface": true,
		},
	}
}

// ToJSON 转换�?JSON 字符�?
func ToJSON(config map[string]interface{}) string {
	data, _ := json.MarshalIndent(config, "", "  ")
	return string(data)
}

// 辅助函数
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func containsString(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
