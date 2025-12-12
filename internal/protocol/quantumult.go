package protocol

import (
	"fmt"
	"strings"

	"dashgo/internal/model"
	"dashgo/internal/service"

	"gopkg.in/yaml.v3"
)

// GenerateQuantumultXConfig 生成 Quantumult X 配置
func GenerateQuantumultXConfig(servers []service.ServerInfo, user *model.User) string {
	var lines []string

	for _, server := range servers {
		line := buildQuantumultXProxy(server, user)
		if line != "" {
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

func buildQuantumultXProxy(server service.ServerInfo, user *model.User) string {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	switch server.Type {
	case model.ServerTypeShadowsocks:
		// 获取加密方式，默�?aes-256-gcm
		cipher := "aes-256-gcm"
		if c, ok := ps["cipher"].(string); ok && c != "" {
			cipher = c
		} else if m, ok := ps["method"].(string); ok && m != "" {
			cipher = m
		}
		// 密码
		password := server.Password
		if password == "" {
			password = user.UUID
		}
		// shadowsocks=example.com:443, method=chacha20-ietf-poly1305, password=pwd, obfs=wss, obfs-host=example.com, obfs-uri=/path, fast-open=false, udp-relay=false, tag=节点�?
		line := fmt.Sprintf("shadowsocks=%s:%d, method=%s, password=%s",
			server.Host, port, cipher, password)

		// 插件
		if plugin, ok := ps["plugin"].(string); ok && plugin != "" {
			if ns, ok := ps["plugin_opts"].(string); ok {
				if strings.Contains(ns, "obfs=http") {
					line += ", obfs=http"
				} else if strings.Contains(ns, "obfs=tls") {
					line += ", obfs=tls"
				}
				// 提取 host
				if strings.Contains(ns, "obfs-host=") {
					parts := strings.Split(ns, "obfs-host=")
					if len(parts) > 1 {
						host := strings.Split(parts[1], ";")[0]
						line += fmt.Sprintf(", obfs-host=%s", host)
					}
				}
			}
		}

		line += ", fast-open=false, udp-relay=true"
		line += fmt.Sprintf(", tag=%s", server.Name)
		return line

	case model.ServerTypeVmess:
		// vmess=example.com:443, method=chacha20-ietf-poly1305, password=uuid, obfs=wss, obfs-host=example.com, obfs-uri=/path, tls-verification=true, fast-open=false, udp-relay=false, tag=节点�?
		line := fmt.Sprintf("vmess=%s:%d, method=chacha20-poly1305, password=%s",
			server.Host, port, user.UUID)

		if network, ok := ps["network"].(string); ok {
			switch network {
			case "ws":
				if tls, ok := ps["tls"].(float64); ok && tls > 0 {
					line += ", obfs=wss"
				} else {
					line += ", obfs=ws"
				}
				if ns, ok := ps["network_settings"].(map[string]interface{}); ok {
					if path, ok := ns["path"].(string); ok {
						line += fmt.Sprintf(", obfs-uri=%s", path)
					}
					if headers, ok := ns["headers"].(map[string]interface{}); ok {
						if host, ok := headers["Host"].(string); ok {
							line += fmt.Sprintf(", obfs-host=%s", host)
						}
					}
				}
			case "tcp":
				if tls, ok := ps["tls"].(float64); ok && tls > 0 {
					line += ", obfs=over-tls"
				}
			}
		}

		if tls, ok := ps["tls"].(float64); ok && tls > 0 {
			line += ", tls-verification=false"
			if tlsSettings, ok := ps["tls_settings"].(map[string]interface{}); ok {
				if sn, ok := tlsSettings["server_name"].(string); ok {
					line += fmt.Sprintf(", tls-host=%s", sn)
				}
			}
		}

		line += ", fast-open=false, udp-relay=true"
		line += fmt.Sprintf(", tag=%s", server.Name)
		return line

	case model.ServerTypeTrojan:
		// trojan=example.com:443, password=pwd, over-tls=true, tls-verification=true, fast-open=false, udp-relay=false, tag=节点�?
		line := fmt.Sprintf("trojan=%s:%d, password=%s, over-tls=true",
			server.Host, port, user.UUID)

		if insecure, ok := ps["allow_insecure"].(bool); ok && insecure {
			line += ", tls-verification=false"
		} else {
			line += ", tls-verification=true"
		}

		if sn, ok := ps["server_name"].(string); ok && sn != "" {
			line += fmt.Sprintf(", tls-host=%s", sn)
		}

		line += ", fast-open=false, udp-relay=true"
		line += fmt.Sprintf(", tag=%s", server.Name)
		return line

	case model.ServerTypeHysteria:
		version := 2
		if v, ok := ps["version"].(float64); ok {
			version = int(v)
		}

		if version == 2 {
			// hysteria2=example.com:443, password=pwd, download-bandwidth=100, tag=节点�?
			line := fmt.Sprintf("hysteria2=%s:%d, password=%s",
				server.Host, port, user.UUID)

			if bw, ok := ps["bandwidth"].(map[string]interface{}); ok {
				if down, ok := bw["down"].(float64); ok {
					line += fmt.Sprintf(", download-bandwidth=%d", int(down))
				}
			}

			if tls, ok := ps["tls"].(map[string]interface{}); ok {
				if sn, ok := tls["server_name"].(string); ok {
					line += fmt.Sprintf(", sni=%s", sn)
				}
				if insecure, ok := tls["allow_insecure"].(bool); ok && insecure {
					line += ", skip-cert-verify=true"
				}
			}

			line += fmt.Sprintf(", tag=%s", server.Name)
			return line
		}

	case model.ServerTypeTuic:
		// tuic=example.com:443, password=pwd, uuid=uuid, tag=节点�?
		line := fmt.Sprintf("tuic=%s:%d, password=%s, uuid=%s",
			server.Host, port, user.UUID, user.UUID)

		if cc, ok := ps["congestion_control"].(string); ok {
			line += fmt.Sprintf(", congestion-control=%s", cc)
		}

		if tls, ok := ps["tls"].(map[string]interface{}); ok {
			if sn, ok := tls["server_name"].(string); ok {
				line += fmt.Sprintf(", sni=%s", sn)
			}
			if insecure, ok := tls["allow_insecure"].(bool); ok && insecure {
				line += ", skip-cert-verify=true"
			}
		}

		line += fmt.Sprintf(", tag=%s", server.Name)
		return line

	case model.ServerTypeVless:
		// vless=example.com:443, method=none, password=uuid, tag=节点�?
		line := fmt.Sprintf("vless=%s:%d, method=none, password=%s",
			server.Host, port, user.UUID)

		if tls, ok := ps["tls"].(float64); ok {
			if tls == 2 { // Reality
				line += ", obfs=over-tls, tls-verification=false"
				if reality, ok := ps["reality_settings"].(map[string]interface{}); ok {
					if sn, ok := reality["server_name"].(string); ok {
						line += fmt.Sprintf(", tls-host=%s", sn)
					}
				}
			} else if tls > 0 {
				line += ", obfs=over-tls, tls-verification=true"
				if tlsSettings, ok := ps["tls_settings"].(map[string]interface{}); ok {
					if sn, ok := tlsSettings["server_name"].(string); ok {
						line += fmt.Sprintf(", tls-host=%s", sn)
					}
				}
			}
		}

		if network, ok := ps["network"].(string); ok {
			switch network {
			case "ws":
				if tls, ok := ps["tls"].(float64); ok && tls > 0 {
					line += ", obfs=wss"
				} else {
					line += ", obfs=ws"
				}
				if ns, ok := ps["network_settings"].(map[string]interface{}); ok {
					if path, ok := ns["path"].(string); ok {
						line += fmt.Sprintf(", obfs-uri=%s", path)
					}
					if headers, ok := ns["headers"].(map[string]interface{}); ok {
						if host, ok := headers["Host"].(string); ok {
							line += fmt.Sprintf(", obfs-host=%s", host)
						}
					}
				}
			}
		}

		line += ", fast-open=false, udp-relay=true"
		line += fmt.Sprintf(", tag=%s", server.Name)
		return line

	case model.ServerTypeAnytls:
		// anytls=example.com:443, password=pwd, tag=节点�?
		line := fmt.Sprintf("anytls=%s:%d, password=%s",
			server.Host, port, user.UUID)

		if tls, ok := ps["tls"].(map[string]interface{}); ok {
			if sn, ok := tls["server_name"].(string); ok {
				line += fmt.Sprintf(", sni=%s", sn)
			}
		}

		line += fmt.Sprintf(", tag=%s", server.Name)
		return line
	}

	return ""
}

// GenerateLoonConfig 生成 Loon 配置
func GenerateLoonConfig(servers []service.ServerInfo, user *model.User) string {
	var lines []string

	for _, server := range servers {
		line := buildLoonProxy(server, user)
		if line != "" {
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

func buildLoonProxy(server service.ServerInfo, user *model.User) string {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	switch server.Type {
	case model.ServerTypeShadowsocks:
		// 获取加密方式，默�?aes-256-gcm
		cipher := "aes-256-gcm"
		if c, ok := ps["cipher"].(string); ok && c != "" {
			cipher = c
		} else if m, ok := ps["method"].(string); ok && m != "" {
			cipher = m
		}
		// 密码
		password := server.Password
		if password == "" {
			password = user.UUID
		}
		// 节点�?= Shadowsocks,服务器地址,端口,加密方式,密码
		line := fmt.Sprintf("%s = Shadowsocks,%s,%d,%s,\"%s\"",
			server.Name, server.Host, port, cipher, password)
		return line

	case model.ServerTypeVmess:
		// 节点�?= vmess,服务器地址,端口,加密方式,UUID,transport
		line := fmt.Sprintf("%s = vmess,%s,%d,auto,\"%s\"",
			server.Name, server.Host, port, user.UUID)

		if network, ok := ps["network"].(string); ok {
			switch network {
			case "ws":
				line += ",transport=ws"
				if ns, ok := ps["network_settings"].(map[string]interface{}); ok {
					if path, ok := ns["path"].(string); ok {
						line += fmt.Sprintf(",path=%s", path)
					}
					if headers, ok := ns["headers"].(map[string]interface{}); ok {
						if host, ok := headers["Host"].(string); ok {
							line += fmt.Sprintf(",host=%s", host)
						}
					}
				}
			}
		}

		if tls, ok := ps["tls"].(float64); ok && tls > 0 {
			line += ",over-tls=true"
		}

		return line

	case model.ServerTypeTrojan:
		// 节点�?= trojan,服务器地址,端口,密码
		line := fmt.Sprintf("%s = trojan,%s,%d,\"%s\"",
			server.Name, server.Host, port, user.UUID)

		if sn, ok := ps["server_name"].(string); ok && sn != "" {
			line += fmt.Sprintf(",tls-name=%s", sn)
		}

		if insecure, ok := ps["allow_insecure"].(bool); ok && insecure {
			line += ",skip-cert-verify=true"
		}

		return line

	case model.ServerTypeHysteria:
		version := 2
		if v, ok := ps["version"].(float64); ok {
			version = int(v)
		}

		if version == 2 {
			// 节点�?= Hysteria2,服务器地址,端口,密码
			line := fmt.Sprintf("%s = Hysteria2,%s,%d,\"%s\"",
				server.Name, server.Host, port, user.UUID)

			if tls, ok := ps["tls"].(map[string]interface{}); ok {
				if sn, ok := tls["server_name"].(string); ok {
					line += fmt.Sprintf(",tls-name=%s", sn)
				}
				if insecure, ok := tls["allow_insecure"].(bool); ok && insecure {
					line += ",skip-cert-verify=true"
				}
			}

			if bw, ok := ps["bandwidth"].(map[string]interface{}); ok {
				if down, ok := bw["down"].(float64); ok {
					line += fmt.Sprintf(",download=%d", int(down))
				}
			}

			return line
		}

	case model.ServerTypeTuic:
		// 节点�?= tuic,服务器地址,端口,uuid,密码
		line := fmt.Sprintf("%s = tuic,%s,%d,\"%s\",\"%s\"",
			server.Name, server.Host, port, user.UUID, user.UUID)

		if tls, ok := ps["tls"].(map[string]interface{}); ok {
			if sn, ok := tls["server_name"].(string); ok {
				line += fmt.Sprintf(",tls-name=%s", sn)
			}
		}

		return line

	case model.ServerTypeVless:
		// 节点�?= vless,服务器地址,端口,uuid
		line := fmt.Sprintf("%s = vless,%s,%d,\"%s\"",
			server.Name, server.Host, port, user.UUID)

		if flow, ok := ps["flow"].(string); ok && flow != "" {
			line += fmt.Sprintf(",flow=%s", flow)
		}

		if tls, ok := ps["tls"].(float64); ok {
			if tls == 2 { // Reality
				line += ",over-tls=true"
				if reality, ok := ps["reality_settings"].(map[string]interface{}); ok {
					if sn, ok := reality["server_name"].(string); ok {
						line += fmt.Sprintf(",tls-name=%s", sn)
					}
					if pk, ok := reality["public_key"].(string); ok {
						line += fmt.Sprintf(",reality-public-key=%s", pk)
					}
					if sid, ok := reality["short_id"].(string); ok && sid != "" {
						line += fmt.Sprintf(",reality-short-id=%s", sid)
					}
				}
			} else if tls > 0 {
				line += ",over-tls=true"
				if tlsSettings, ok := ps["tls_settings"].(map[string]interface{}); ok {
					if sn, ok := tlsSettings["server_name"].(string); ok {
						line += fmt.Sprintf(",tls-name=%s", sn)
					}
				}
			}
		}

		if network, ok := ps["network"].(string); ok {
			switch network {
			case "ws":
				line += ",transport=ws"
				if ns, ok := ps["network_settings"].(map[string]interface{}); ok {
					if path, ok := ns["path"].(string); ok {
						line += fmt.Sprintf(",path=%s", path)
					}
					if headers, ok := ns["headers"].(map[string]interface{}); ok {
						if host, ok := headers["Host"].(string); ok {
							line += fmt.Sprintf(",host=%s", host)
						}
					}
				}
			case "grpc":
				line += ",transport=grpc"
				if ns, ok := ps["network_settings"].(map[string]interface{}); ok {
					if sn, ok := ns["serviceName"].(string); ok {
						line += fmt.Sprintf(",grpc-service-name=%s", sn)
					}
				}
			}
		}

		return line

	case model.ServerTypeAnytls:
		// 节点�?= anytls,服务器地址,端口,密码
		line := fmt.Sprintf("%s = anytls,%s,%d,\"%s\"",
			server.Name, server.Host, port, user.UUID)

		if tls, ok := ps["tls"].(map[string]interface{}); ok {
			if sn, ok := tls["server_name"].(string); ok {
				line += fmt.Sprintf(",tls-name=%s", sn)
			}
		}

		return line
	}

	return ""
}

// GenerateShadowrocketConfig 生成 Shadowrocket 配置 (Base64 URI)
func GenerateShadowrocketConfig(servers []service.ServerInfo, user *model.User) string {
	// Shadowrocket 使用标准 URI 格式，与 Base64Links 相同
	return GenerateBase64Links(servers, user)
}

// GenerateClashMetaConfig 生成 Clash Meta (mihomo) 配置
func GenerateClashMetaConfig(servers []service.ServerInfo, user *model.User) string {
	// Clash Meta 完全兼容 Clash 配置，但支持更多协议
	config := getDefaultClashMetaConfig()

	proxyNames := []string{}
	for _, server := range servers {
		proxy := buildClashMetaProxy(server, user)
		if proxy != nil {
			config.Proxies = append(config.Proxies, proxy)
			proxyNames = append(proxyNames, server.Name)
		}
	}

	// 更新代理�?
	for i := range config.ProxyGroups {
		if config.ProxyGroups[i].Name == "Proxy" || config.ProxyGroups[i].Name == "Auto" {
			config.ProxyGroups[i].Proxies = append(config.ProxyGroups[i].Proxies, proxyNames...)
		}
	}

	data, _ := yaml.Marshal(config)
	return string(data)
}

func buildClashMetaProxy(server service.ServerInfo, user *model.User) map[string]interface{} {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	// 基础配置�?Clash 相同
	proxy := buildClashProxy(server, user)
	if proxy == nil {
		return nil
	}

	// Clash Meta 特有协议支持
	switch server.Type {
	case model.ServerTypeVless:
		proxy["type"] = "vless"
		proxy["uuid"] = user.UUID
		proxy["server"] = server.Host
		proxy["port"] = port

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
					proxy["client-fingerprint"] = "chrome"
				}
			} else if tls > 0 {
				proxy["tls"] = true
			}
		}

		if network, ok := ps["network"].(string); ok {
			proxy["network"] = network
			addClashTransportOpts(proxy, network, ps)
		}

	case model.ServerTypeAnytls:
		// Clash Meta 支持 AnyTLS
		proxy["type"] = "anytls"
		proxy["server"] = server.Host
		proxy["port"] = port
		proxy["password"] = user.UUID

		if tls, ok := ps["tls"].(map[string]interface{}); ok {
			if sn, ok := tls["server_name"].(string); ok {
				proxy["sni"] = sn
			}
			if insecure, ok := tls["allow_insecure"].(bool); ok {
				proxy["skip-cert-verify"] = insecure
			}
		}

		if paddingScheme, ok := ps["padding_scheme"].([]interface{}); ok {
			proxy["padding-scheme"] = paddingScheme
		}
	}

	return proxy
}

func getDefaultClashMetaConfig() *ClashConfig {
	config := getDefaultClashConfig()
	// Clash Meta 特有配置
	return config
}


