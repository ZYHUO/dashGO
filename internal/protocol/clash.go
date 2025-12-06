package protocol

import (
	"fmt"
	"strings"

	"xboard/internal/model"
	"xboard/internal/service"

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

	// 更新代理组
	for i := range config.ProxyGroups {
		if config.ProxyGroups[i].Name == "Proxy" || config.ProxyGroups[i].Name == "Auto" {
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
		// 获取加密方式，优先使用 method，其次 cipher
		cipher := "aes-256-gcm"
		if m, ok := ps["method"].(string); ok && m != "" {
			cipher = m
		} else if c, ok := ps["cipher"].(string); ok && c != "" {
			cipher = c
		}

		proxy := map[string]interface{}{
			"name":     server.Name,
			"type":     "ss",
			"server":   server.Host,
			"port":     port,
			"cipher":   cipher,
			"password": server.Password,
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
		}
		if tls, ok := ps["tls"].(float64); ok && tls > 0 {
			proxy["tls"] = true
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
		}
		if flow, ok := ps["flow"].(string); ok && flow != "" {
			proxy["flow"] = flow
		}
		if tls, ok := ps["tls"].(float64); ok {
			if tls == 2 { // Reality
				proxy["tls"] = true
				if reality, ok := ps["reality_settings"].(map[string]interface{}); ok {
					proxy["servername"] = reality["server_name"]
					proxy["reality-opts"] = map[string]interface{}{
						"public-key": reality["public_key"],
						"short-id":   reality["short_id"],
					}
				}
			} else if tls > 0 {
				proxy["tls"] = true
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
		}
		if sn, ok := ps["server_name"].(string); ok && sn != "" {
			proxy["sni"] = sn
		}
		if insecure, ok := ps["allow_insecure"].(bool); ok {
			proxy["skip-cert-verify"] = insecure
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
			if obfs, ok := ps["obfs"].(map[string]interface{}); ok {
				if open, ok := obfs["open"].(bool); ok && open {
					proxy["obfs"] = obfs["type"]
					proxy["obfs-password"] = obfs["password"]
				}
			}
		} else {
			proxy["auth-str"] = user.UUID
			if bw, ok := ps["bandwidth"].(map[string]interface{}); ok {
				if up, ok := bw["up"].(float64); ok {
					proxy["up"] = fmt.Sprintf("%d Mbps", int(up))
				}
				if down, ok := bw["down"].(float64); ok {
					proxy["down"] = fmt.Sprintf("%d Mbps", int(down))
				}
			}
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

	case model.ServerTypeTuic:
		proxy := map[string]interface{}{
			"name":               server.Name,
			"type":               "tuic",
			"server":             server.Host,
			"port":               port,
			"uuid":               user.UUID,
			"password":           user.UUID,
			"congestion-controller": "cubic",
			"udp-relay-mode":     "native",
		}

		if cc, ok := ps["congestion_control"].(string); ok {
			proxy["congestion-controller"] = cc
		}
		if tls, ok := ps["tls"].(map[string]interface{}); ok {
			if sn, ok := tls["server_name"].(string); ok {
				proxy["sni"] = sn
			}
		}

		return proxy
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
	// 处理端口范围，取第一个端口
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
			NameServer: []string{"223.5.5.5", "119.29.29.29"},
			Fallback:   []string{"8.8.8.8", "1.1.1.1"},
			FallbackFilter: &ClashFallbackFilter{
				GeoIP:  true,
				IPCidr: []string{"240.0.0.0/4"},
			},
		},
		Proxies: []map[string]interface{}{},
		ProxyGroups: []ClashProxyGroup{
			{
				Name:    "Proxy",
				Type:    "select",
				Proxies: []string{"Auto", "DIRECT"},
			},
			{
				Name:     "Auto",
				Type:     "url-test",
				Proxies:  []string{},
				URL:      "http://www.gstatic.com/generate_204",
				Interval: 300,
			},
		},
		Rules: []string{
			"GEOIP,LAN,DIRECT",
			"GEOIP,CN,DIRECT",
			"MATCH,Proxy",
		},
	}
}
