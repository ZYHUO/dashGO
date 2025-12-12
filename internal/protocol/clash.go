package protocol

import (
	"fmt"
	"strings"

	"dashgo/internal/model"
	"dashgo/internal/service"

	"gopkg.in/yaml.v3"
)

// ClashConfig Clash 配置结构
type ClashConfig struct {
	Port               int                      `yaml:"port,omitempty"`
	SocksPort          int                      `yaml:"socks-port,omitempty"`
	AllowLan           bool                     `yaml:"allow-lan"`
	Mode               string                   `yaml:"mode"`
	LogLevel           string                   `yaml:"log-level"`
	ExternalController string                   `yaml:"external-controller,omitempty"`
	DNS                *ClashDNS                `yaml:"dns,omitempty"`
	Proxies            []map[string]interface{} `yaml:"proxies"`
	ProxyGroups        []ClashProxyGroup        `yaml:"proxy-groups"`
	Rules              []string                 `yaml:"rules"`
}

type ClashDNS struct {
	Enable       bool     `yaml:"enable"`
	IPv6         bool     `yaml:"ipv6"`
	NameServer   []string `yaml:"nameserver"`
	Fallback     []string `yaml:"fallback,omitempty"`
	FallbackFilter *ClashFallbackFilter `yaml:"fallback-filter,omitempty"`
}

type ClashFallbackFilter struct {
	GeoIP  bool     `yaml:"geoip"`
	IPCidr []string `yaml:"ipcidr,omitempty"`
}

type ClashProxyGroup struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Proxies  []string `yaml:"proxies"`
	URL      string   `yaml:"url,omitempty"`
	Interval int      `yaml:"interval,omitempty"`
}

// GenerateClashConfig 生成 Clash 配置
func GenerateClashConfig(servers []service.ServerInfo, user *model.User) string {
	config := getDefaultClashConfig()
	
	proxyNames := []string{}
	for _, server := range servers {
		proxy := buildClashProxy(server, user)
		if proxy != nil {
			config.Proxies = append(config.Proxies, proxy)
			proxyNames = append(proxyNames, server.Name)
		}
	}

	// 更新代理组，将节点添加到需要的组中
	for i := range config.ProxyGroups {
		groupName := config.ProxyGroups[i].Name
		switch groupName {
		case "🚀 节点选择":
			// 节点选择组：添加所有节�?
			config.ProxyGroups[i].Proxies = append(config.ProxyGroups[i].Proxies, proxyNames...)
		case "♻️ 自动选择", "🔯 故障转移", "🔮 负载均衡":
			// 自动选择/故障转移/负载均衡：添加所有节�?
			config.ProxyGroups[i].Proxies = proxyNames
		case "📲 电报消息", "🤖 OpenAI", "📹 YouTube", "🎬 Netflix", "🍎 苹果服务", "🎮 游戏平台", "🐟 漏网之鱼":
			// 其他分组：添加所有节点到末尾
			config.ProxyGroups[i].Proxies = append(config.ProxyGroups[i].Proxies, proxyNames...)
		}
	}

	data, _ := yaml.Marshal(config)
	return string(data)
}

func buildClashProxy(server service.ServerInfo, user *model.User) map[string]interface{} {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	switch server.Type {
	case model.ServerTypeShadowsocks:
		// 获取加密方式，优先使�?method，其�?cipher
		cipher := "aes-256-gcm"
		if m, ok := ps["method"].(string); ok && m != "" {
			cipher = m
		} else if c, ok := ps["cipher"].(string); ok && c != "" {
			cipher = c
		}

		// 密码：对�?SS2022，使�?server.Password（已包含服务器密�?用户密钥格式�?
		// 对于普�?SS，使用用�?UUID
		password := server.Password
		if password == "" {
			password = user.UUID
		}

		proxy := map[string]interface{}{
			"name":     server.Name,
			"type":     "ss",
			"server":   server.Host,
			"port":     port,
			"cipher":   cipher,
			"password": password,
			"udp":      true,
		}
		if plugin, ok := ps["plugin"].(string); ok && plugin != "" {
			proxy["plugin"] = plugin
			if opts, ok := ps["plugin_opts"].(string); ok {
				proxy["plugin-opts"] = parsePluginOpts(opts)
			}
		}
		return proxy

	case model.ServerTypeVmess:
		proxy := map[string]interface{}{
			"name":     server.Name,
			"type":     "vmess",
			"server":   server.Host,
			"port":     port,
			"uuid":     user.UUID,
			"alterId":  0,
			"cipher":   "auto",
			"udp":      true,
		}
		if tls, ok := ps["tls"].(float64); ok && tls > 0 {
			proxy["tls"] = true
			if tlsSettings, ok := ps["tls_settings"].(map[string]interface{}); ok {
				if sn, ok := tlsSettings["server_name"].(string); ok {
					proxy["servername"] = sn
				}
				if insecure, ok := tlsSettings["allow_insecure"].(bool); ok {
					proxy["skip-cert-verify"] = insecure
				}
			}
		}
		if network, ok := ps["network"].(string); ok {
			proxy["network"] = network
			addClashTransportOpts(proxy, network, ps)
		}
		return proxy

	case model.ServerTypeVless:
		// Clash Meta 支持 VLESS
		proxy := map[string]interface{}{
			"name":   server.Name,
			"type":   "vless",
			"server": server.Host,
			"port":   port,
			"uuid":   user.UUID,
			"udp":    true,
		}
		if flow, ok := ps["flow"].(string); ok && flow != "" {
			proxy["flow"] = flow
		}
		if tls, ok := ps["tls"].(float64); ok {
			if tls == 2 { // Reality
				proxy["tls"] = true
				proxy["client-fingerprint"] = "chrome"
				if reality, ok := ps["reality_settings"].(map[string]interface{}); ok {
					proxy["servername"] = reality["server_name"]
					proxy["reality-opts"] = map[string]interface{}{
						"public-key": reality["public_key"],
						"short-id":   reality["short_id"],
					}
				}
			} else if tls > 0 {
				proxy["tls"] = true
				proxy["client-fingerprint"] = "chrome"
				if tlsSettings, ok := ps["tls_settings"].(map[string]interface{}); ok {
					if sn, ok := tlsSettings["server_name"].(string); ok {
						proxy["servername"] = sn
					}
					if insecure, ok := tlsSettings["allow_insecure"].(bool); ok {
						proxy["skip-cert-verify"] = insecure
					}
				}
			}
		}
		if network, ok := ps["network"].(string); ok {
			proxy["network"] = network
			addClashTransportOpts(proxy, network, ps)
		}
		return proxy

	case model.ServerTypeTrojan:
		proxy := map[string]interface{}{
			"name":     server.Name,
			"type":     "trojan",
			"server":   server.Host,
			"port":     port,
			"password": user.UUID,
			"udp":      true,
		}
		if sn, ok := ps["server_name"].(string); ok && sn != "" {
			proxy["sni"] = sn
		}
		if insecure, ok := ps["allow_insecure"].(bool); ok {
			proxy["skip-cert-verify"] = insecure
		}
		// TLS settings
		if tlsSettings, ok := ps["tls_settings"].(map[string]interface{}); ok {
			if sn, ok := tlsSettings["server_name"].(string); ok {
				proxy["sni"] = sn
			}
			if insecure, ok := tlsSettings["allow_insecure"].(bool); ok {
				proxy["skip-cert-verify"] = insecure
			}
		}
		if network, ok := ps["network"].(string); ok && network != "" {
			proxy["network"] = network
			addClashTransportOpts(proxy, network, ps)
		}
		return proxy

	case model.ServerTypeHysteria:
		version := 2
		if v, ok := ps["version"].(float64); ok {
			version = int(v)
		}

		var proxyType string
		if version == 2 {
			proxyType = "hysteria2"
		} else {
			proxyType = "hysteria"
		}

		proxy := map[string]interface{}{
			"name":   server.Name,
			"type":   proxyType,
			"server": server.Host,
			"port":   port,
		}

		if version == 2 {
			proxy["password"] = user.UUID
			// Hysteria2 obfs
			if obfs, ok := ps["obfs"].(map[string]interface{}); ok {
				if open, ok := obfs["open"].(bool); ok && open {
					proxy["obfs"] = obfs["type"]
					proxy["obfs-password"] = obfs["password"]
				}
			}
		} else {
			// Hysteria1
			proxy["auth-str"] = user.UUID
			proxy["protocol"] = "udp"
			if bw, ok := ps["bandwidth"].(map[string]interface{}); ok {
				if up, ok := bw["up"].(float64); ok {
					proxy["up"] = fmt.Sprintf("%d Mbps", int(up))
				}
				if down, ok := bw["down"].(float64); ok {
					proxy["down"] = fmt.Sprintf("%d Mbps", int(down))
				}
			}
			if obfs, ok := ps["obfs"].(map[string]interface{}); ok {
				if pw, ok := obfs["password"].(string); ok && pw != "" {
					proxy["obfs"] = pw
				}
			}
		}

		// TLS settings
		if tls, ok := ps["tls"].(map[string]interface{}); ok {
			if sn, ok := tls["server_name"].(string); ok {
				proxy["sni"] = sn
			}
			if insecure, ok := tls["allow_insecure"].(bool); ok {
				proxy["skip-cert-verify"] = insecure
			}
		}

		return proxy

	case model.ServerTypeTuic:
		proxy := map[string]interface{}{
			"name":                  server.Name,
			"type":                  "tuic",
			"server":                server.Host,
			"port":                  port,
			"uuid":                  user.UUID,
			"password":              user.UUID,
			"congestion-controller": "cubic",
			"udp-relay-mode":        "native",
			"reduce-rtt":            true,
			"alpn":                  []string{"h3"},
		}

		if cc, ok := ps["congestion_control"].(string); ok {
			proxy["congestion-controller"] = cc
		}
		if urm, ok := ps["udp_relay_mode"].(string); ok {
			proxy["udp-relay-mode"] = urm
		}
		if tls, ok := ps["tls"].(map[string]interface{}); ok {
			if sn, ok := tls["server_name"].(string); ok {
				proxy["sni"] = sn
			}
			if insecure, ok := tls["allow_insecure"].(bool); ok {
				proxy["skip-cert-verify"] = insecure
			}
		}

		return proxy

	case model.ServerTypeAnytls:
		proxy := map[string]interface{}{
			"name":               server.Name,
			"type":               "anytls",
			"server":             server.Host,
			"port":               port,
			"password":           user.UUID,
			"client-fingerprint": "chrome",
			"udp":                true,
		}
		if tls, ok := ps["tls"].(map[string]interface{}); ok {
			if sn, ok := tls["server_name"].(string); ok {
				proxy["sni"] = sn
			}
			if insecure, ok := tls["allow_insecure"].(bool); ok {
				proxy["skip-cert-verify"] = insecure
			}
		}
		return proxy

	case "shadowtls":
		// ShadowTLS �?Clash Meta 中需要配�?SS 使用
		proxy := map[string]interface{}{
			"name":               server.Name,
			"type":               "ss",
			"server":             server.Host,
			"port":               port,
			"cipher":             "2022-blake3-aes-128-gcm",
			"password":           user.UUID,
			"client-fingerprint": "chrome",
			"plugin":             "shadow-tls",
			"plugin-opts": map[string]interface{}{
				"host":     "addons.mozilla.org",
				"password": user.UUID,
				"version":  3,
			},
		}
		if hs, ok := ps["handshake_server"].(string); ok && hs != "" {
			proxy["plugin-opts"].(map[string]interface{})["host"] = hs
		}
		if method, ok := ps["detour_method"].(string); ok && method != "" {
			proxy["cipher"] = method
		}
		return proxy

	case "naive":
		// NaiveProxy �?Clash Meta 中不直接支持，返�?nil
		return nil
	}

	return nil
}

func addClashTransportOpts(proxy map[string]interface{}, network string, ps model.JSONMap) {
	ns, _ := ps["network_settings"].(map[string]interface{})

	switch network {
	case "ws":
		wsOpts := map[string]interface{}{}
		if path, ok := ns["path"].(string); ok {
			wsOpts["path"] = path
		}
		if headers, ok := ns["headers"].(map[string]interface{}); ok {
			wsOpts["headers"] = headers
		}
		if len(wsOpts) > 0 {
			proxy["ws-opts"] = wsOpts
		}

	case "grpc":
		grpcOpts := map[string]interface{}{}
		if sn, ok := ns["serviceName"].(string); ok {
			grpcOpts["grpc-service-name"] = sn
		}
		if len(grpcOpts) > 0 {
			proxy["grpc-opts"] = grpcOpts
		}

	case "h2":
		h2Opts := map[string]interface{}{}
		if host, ok := ns["host"].([]interface{}); ok {
			h2Opts["host"] = host
		}
		if path, ok := ns["path"].(string); ok {
			h2Opts["path"] = path
		}
		if len(h2Opts) > 0 {
			proxy["h2-opts"] = h2Opts
		}
	}
}

func parsePluginOpts(opts string) map[string]interface{} {
	result := make(map[string]interface{})
	parts := strings.Split(opts, ";")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

// parsePort 解析端口字符串，返回整数端口
func parsePort(portStr string) int {
	// 处理端口范围，取第一个端�?
	if strings.Contains(portStr, "-") {
		parts := strings.Split(portStr, "-")
		portStr = parts[0]
	}
	port := 0
	fmt.Sscanf(portStr, "%d", &port)
	return port
}

func getDefaultClashConfig() *ClashConfig {
	return &ClashConfig{
		Port:      7890,
		SocksPort: 7891,
		AllowLan:  false,
		Mode:      "rule",
		LogLevel:  "info",
		DNS: &ClashDNS{
			Enable:     true,
			IPv6:       false,
			NameServer: []string{"223.5.5.5", "119.29.29.29", "https://doh.pub/dns-query"},
			Fallback:   []string{"8.8.8.8", "1.1.1.1", "https://dns.google/dns-query"},
			FallbackFilter: &ClashFallbackFilter{
				GeoIP:  true,
				IPCidr: []string{"240.0.0.0/4", "0.0.0.0/32"},
			},
		},
		Proxies: []map[string]interface{}{},
		ProxyGroups: []ClashProxyGroup{
			{
				Name:    "🚀 节点选择",
				Type:    "select",
				Proxies: []string{"♻️ 自动选择", "🔯 故障转移", "🔮 负载均衡", "DIRECT"},
			},
			{
				Name:     "♻️ 自动选择",
				Type:     "url-test",
				Proxies:  []string{},
				URL:      "http://www.gstatic.com/generate_204",
				Interval: 300,
			},
			{
				Name:     "🔯 故障转移",
				Type:     "fallback",
				Proxies:  []string{},
				URL:      "http://www.gstatic.com/generate_204",
				Interval: 300,
			},
			{
				Name:    "🔮 负载均衡",
				Type:    "load-balance",
				Proxies: []string{},
				URL:     "http://www.gstatic.com/generate_204",
			},
			{
				Name:    "📲 电报消息",
				Type:    "select",
				Proxies: []string{"🚀 节点选择", "♻️ 自动选择", "DIRECT"},
			},
			{
				Name:    "🤖 OpenAI",
				Type:    "select",
				Proxies: []string{"🚀 节点选择", "♻️ 自动选择"},
			},
			{
				Name:    "📹 YouTube",
				Type:    "select",
				Proxies: []string{"🚀 节点选择", "♻️ 自动选择", "DIRECT"},
			},
			{
				Name:    "🎬 Netflix",
				Type:    "select",
				Proxies: []string{"🚀 节点选择", "♻️ 自动选择", "DIRECT"},
			},
			{
				Name:    "🍎 苹果服务",
				Type:    "select",
				Proxies: []string{"DIRECT", "🚀 节点选择"},
			},
			{
				Name:    "🎮 游戏平台",
				Type:    "select",
				Proxies: []string{"DIRECT", "🚀 节点选择"},
			},
			{
				Name:    "🐟 漏网之鱼",
				Type:    "select",
				Proxies: []string{"🚀 节点选择", "♻️ 自动选择", "DIRECT"},
			},
		},
		Rules: []string{
			// 本地/局域网
			"DOMAIN-SUFFIX,local,DIRECT",
			"IP-CIDR,127.0.0.0/8,DIRECT",
			"IP-CIDR,172.16.0.0/12,DIRECT",
			"IP-CIDR,192.168.0.0/16,DIRECT",
			"IP-CIDR,10.0.0.0/8,DIRECT",
			"IP-CIDR,100.64.0.0/10,DIRECT",
			// OpenAI
			"DOMAIN-SUFFIX,openai.com,🤖 OpenAI",
			"DOMAIN-SUFFIX,ai.com,🤖 OpenAI",
			"DOMAIN-SUFFIX,anthropic.com,🤖 OpenAI",
			"DOMAIN-SUFFIX,claude.ai,🤖 OpenAI",
			"DOMAIN-KEYWORD,openai,🤖 OpenAI",
			// Telegram
			"DOMAIN-SUFFIX,telegram.org,📲 电报消息",
			"DOMAIN-SUFFIX,t.me,📲 电报消息",
			"DOMAIN-SUFFIX,tg.dev,📲 电报消息",
			"IP-CIDR,91.108.0.0/16,📲 电报消息",
			"IP-CIDR,109.239.140.0/24,📲 电报消息",
			"IP-CIDR,149.154.160.0/20,📲 电报消息",
			// YouTube
			"DOMAIN-SUFFIX,youtube.com,📹 YouTube",
			"DOMAIN-SUFFIX,googlevideo.com,📹 YouTube",
			"DOMAIN-SUFFIX,ytimg.com,📹 YouTube",
			"DOMAIN-SUFFIX,yt.be,📹 YouTube",
			// Netflix
			"DOMAIN-SUFFIX,netflix.com,🎬 Netflix",
			"DOMAIN-SUFFIX,netflix.net,🎬 Netflix",
			"DOMAIN-SUFFIX,nflximg.com,🎬 Netflix",
			"DOMAIN-SUFFIX,nflximg.net,🎬 Netflix",
			"DOMAIN-SUFFIX,nflxvideo.net,🎬 Netflix",
			// Apple
			"DOMAIN-SUFFIX,apple.com,🍎 苹果服务",
			"DOMAIN-SUFFIX,icloud.com,🍎 苹果服务",
			"DOMAIN-SUFFIX,icloud-content.com,🍎 苹果服务",
			"DOMAIN-SUFFIX,mzstatic.com,🍎 苹果服务",
			// 游戏
			"DOMAIN-SUFFIX,steam.com,🎮 游戏平台",
			"DOMAIN-SUFFIX,steampowered.com,🎮 游戏平台",
			"DOMAIN-SUFFIX,steamcommunity.com,🎮 游戏平台",
			"DOMAIN-SUFFIX,epicgames.com,🎮 游戏平台",
			"DOMAIN-SUFFIX,ea.com,🎮 游戏平台",
			// Google
			"DOMAIN-SUFFIX,google.com,🚀 节点选择",
			"DOMAIN-SUFFIX,googleapis.com,🚀 节点选择",
			"DOMAIN-SUFFIX,gstatic.com,🚀 节点选择",
			"DOMAIN-SUFFIX,gmail.com,🚀 节点选择",
			// GitHub
			"DOMAIN-SUFFIX,github.com,🚀 节点选择",
			"DOMAIN-SUFFIX,githubusercontent.com,🚀 节点选择",
			"DOMAIN-SUFFIX,githubassets.com,🚀 节点选择",
			// Twitter
			"DOMAIN-SUFFIX,twitter.com,🚀 节点选择",
			"DOMAIN-SUFFIX,x.com,🚀 节点选择",
			"DOMAIN-SUFFIX,twimg.com,🚀 节点选择",
			// Facebook
			"DOMAIN-SUFFIX,facebook.com,🚀 节点选择",
			"DOMAIN-SUFFIX,fb.com,🚀 节点选择",
			"DOMAIN-SUFFIX,instagram.com,🚀 节点选择",
			"DOMAIN-SUFFIX,whatsapp.com,🚀 节点选择",
			// 国内直连
			"DOMAIN-SUFFIX,cn,DIRECT",
			"DOMAIN-SUFFIX,baidu.com,DIRECT",
			"DOMAIN-SUFFIX,qq.com,DIRECT",
			"DOMAIN-SUFFIX,weixin.com,DIRECT",
			"DOMAIN-SUFFIX,taobao.com,DIRECT",
			"DOMAIN-SUFFIX,jd.com,DIRECT",
			"DOMAIN-SUFFIX,bilibili.com,DIRECT",
			"DOMAIN-SUFFIX,163.com,DIRECT",
			"DOMAIN-SUFFIX,126.com,DIRECT",
			"DOMAIN-SUFFIX,sina.com,DIRECT",
			"DOMAIN-SUFFIX,weibo.com,DIRECT",
			"DOMAIN-SUFFIX,zhihu.com,DIRECT",
			"DOMAIN-SUFFIX,douyin.com,DIRECT",
			"DOMAIN-SUFFIX,tiktok.com,🚀 节点选择",
			// GeoIP
			"GEOIP,LAN,DIRECT",
			"GEOIP,CN,DIRECT",
			// 兜底
			"MATCH,🐟 漏网之鱼",
		},
	}
}
