package protocol

import (
	"fmt"
	"strings"

	"dashgo/internal/model"
	"dashgo/internal/service"
)

// GenerateSurgeConfig 生成 Surge 配置
func GenerateSurgeConfig(servers []service.ServerInfo, user *model.User) string {
	var sb strings.Builder

	// General
	sb.WriteString("[General]\n")
	sb.WriteString("loglevel = notify\n")
	sb.WriteString("dns-server = 223.5.5.5, 119.29.29.29\n")
	sb.WriteString("skip-proxy = 127.0.0.1, 192.168.0.0/16, 10.0.0.0/8, 172.16.0.0/12, 100.64.0.0/10, localhost, *.local\n")
	sb.WriteString("allow-wifi-access = false\n")
	sb.WriteString("external-controller-access = password@0.0.0.0:6170\n")
	sb.WriteString("\n")

	// Proxy
	sb.WriteString("[Proxy]\n")
	sb.WriteString("DIRECT = direct\n")
	proxyNames := []string{}

	for _, server := range servers {
		line := buildSurgeProxy(server, user)
		if line != "" {
			sb.WriteString(line + "\n")
			proxyNames = append(proxyNames, server.Name)
		}
	}
	sb.WriteString("\n")

	// Proxy Group
	sb.WriteString("[Proxy Group]\n")
	sb.WriteString(fmt.Sprintf("Proxy = select, Auto, DIRECT, %s\n", strings.Join(proxyNames, ", ")))
	sb.WriteString(fmt.Sprintf("Auto = url-test, %s, url=http://www.gstatic.com/generate_204, interval=300\n", strings.Join(proxyNames, ", ")))
	sb.WriteString("\n")

	// Rule
	sb.WriteString("[Rule]\n")
	sb.WriteString("GEOIP,CN,DIRECT\n")
	sb.WriteString("FINAL,Proxy\n")

	return sb.String()
}

func buildSurgeProxy(server service.ServerInfo, user *model.User) string {
	ps := server.ProtocolSettings
	port := parsePort(server.Port)

	switch server.Type {
	case model.ServerTypeShadowsocks:
		// 获取加密方式，默告aes-256-gcm
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
		// Surge 支持告SS 加密方式
		line := fmt.Sprintf("%s = ss, %s, %d, encrypt-method=%s, password=%s",
			server.Name, server.Host, port, cipher, password)

		// 插件支持
		if plugin, ok := ps["plugin"].(string); ok && plugin != "" {
			if plugin == "obfs-local" || plugin == "simple-obfs" {
				if opts, ok := ps["plugin_opts"].(string); ok {
					// 解析 obfs 参数
					if strings.Contains(opts, "obfs=http") {
						line += ", obfs=http"
						if strings.Contains(opts, "obfs-host=") {
							// 提取 host
							parts := strings.Split(opts, "obfs-host=")
							if len(parts) > 1 {
								host := strings.Split(parts[1], ";")[0]
								line += fmt.Sprintf(", obfs-host=%s", host)
							}
						}
					} else if strings.Contains(opts, "obfs=tls") {
						line += ", obfs=tls"
					}
				}
			}
		}
		return line

	case model.ServerTypeVmess:
		line := fmt.Sprintf("%s = vmess, %s, %d, username=%s",
			server.Name, server.Host, port, user.UUID)

		if tls, ok := ps["tls"].(float64); ok && tls > 0 {
			line += ", tls=true"
			if tlsSettings, ok := ps["tls_settings"].(map[string]interface{}); ok {
				if sn, ok := tlsSettings["server_name"].(string); ok {
					line += fmt.Sprintf(", sni=%s", sn)
				}
			}
		}

		if network, ok := ps["network"].(string); ok {
			switch network {
			case "ws":
				line += ", ws=true"
				if ns, ok := ps["network_settings"].(map[string]interface{}); ok {
					if path, ok := ns["path"].(string); ok {
						line += fmt.Sprintf(", ws-path=%s", path)
					}
					if headers, ok := ns["headers"].(map[string]interface{}); ok {
						if host, ok := headers["Host"].(string); ok {
							line += fmt.Sprintf(", ws-headers=Host:%s", host)
						}
					}
				}
			}
		}
		return line

	case model.ServerTypeTrojan:
		line := fmt.Sprintf("%s = trojan, %s, %d, password=%s",
			server.Name, server.Host, port, user.UUID)

		if sn, ok := ps["server_name"].(string); ok && sn != "" {
			line += fmt.Sprintf(", sni=%s", sn)
		}
		if insecure, ok := ps["allow_insecure"].(bool); ok && insecure {
			line += ", skip-cert-verify=true"
		}
		return line

	case model.ServerTypeHysteria:
		version := 2
		if v, ok := ps["version"].(float64); ok {
			version = int(v)
		}

		if version == 2 {
			line := fmt.Sprintf("%s = hysteria2, %s, %d, password=%s",
				server.Name, server.Host, port, user.UUID)

			if tls, ok := ps["tls"].(map[string]interface{}); ok {
				if sn, ok := tls["server_name"].(string); ok {
					line += fmt.Sprintf(", sni=%s", sn)
				}
				if insecure, ok := tls["allow_insecure"].(bool); ok && insecure {
					line += ", skip-cert-verify=true"
				}
			}

			if bw, ok := ps["bandwidth"].(map[string]interface{}); ok {
				if down, ok := bw["down"].(float64); ok {
					line += fmt.Sprintf(", download-bandwidth=%d", int(down))
				}
			}
			return line
		}

	case model.ServerTypeTuic:
		line := fmt.Sprintf("%s = tuic, %s, %d, token=%s",
			server.Name, server.Host, port, user.UUID)

		if tls, ok := ps["tls"].(map[string]interface{}); ok {
			if sn, ok := tls["server_name"].(string); ok {
				line += fmt.Sprintf(", sni=%s", sn)
			}
		}
		return line

	case model.ServerTypeAnytls:
		// Surge 5.8+ 支持 AnyTLS
		line := fmt.Sprintf("%s = anytls, %s, %d, password=%s",
			server.Name, server.Host, port, user.UUID)

		if tls, ok := ps["tls"].(map[string]interface{}); ok {
			if sn, ok := tls["server_name"].(string); ok {
				line += fmt.Sprintf(", sni=%s", sn)
			}
		}
		return line

	case model.ServerTypeVless:
		// Surge 5+ 支持 VLESS
		line := fmt.Sprintf("%s = vless, %s, %d, username=%s",
			server.Name, server.Host, port, user.UUID)

		if flow, ok := ps["flow"].(string); ok && flow != "" {
			line += fmt.Sprintf(", flow=%s", flow)
		}

		if tls, ok := ps["tls"].(float64); ok {
			if tls == 2 { // Reality
				line += ", tls=true"
				if reality, ok := ps["reality_settings"].(map[string]interface{}); ok {
					if sn, ok := reality["server_name"].(string); ok {
						line += fmt.Sprintf(", sni=%s", sn)
					}
					if pk, ok := reality["public_key"].(string); ok {
						line += fmt.Sprintf(", reality-public-key=%s", pk)
					}
					if sid, ok := reality["short_id"].(string); ok && sid != "" {
						line += fmt.Sprintf(", reality-short-id=%s", sid)
					}
				}
			} else if tls > 0 {
				line += ", tls=true"
				if tlsSettings, ok := ps["tls_settings"].(map[string]interface{}); ok {
					if sn, ok := tlsSettings["server_name"].(string); ok {
						line += fmt.Sprintf(", sni=%s", sn)
					}
				}
			}
		}

		if network, ok := ps["network"].(string); ok {
			switch network {
			case "ws":
				line += ", ws=true"
				if ns, ok := ps["network_settings"].(map[string]interface{}); ok {
					if path, ok := ns["path"].(string); ok {
						line += fmt.Sprintf(", ws-path=%s", path)
					}
					if headers, ok := ns["headers"].(map[string]interface{}); ok {
						if host, ok := headers["Host"].(string); ok {
							line += fmt.Sprintf(", ws-headers=Host:%s", host)
						}
					}
				}
			case "grpc":
				line += ", grpc=true"
				if ns, ok := ps["network_settings"].(map[string]interface{}); ok {
					if sn, ok := ns["serviceName"].(string); ok {
						line += fmt.Sprintf(", grpc-service-name=%s", sn)
					}
				}
			}
		}
		return line

	case "shadowtls":
		// ShadowTLS 告Surge 中使告SS + shadow-tls 插件
		cipher := "2022-blake3-aes-128-gcm"
		if method, ok := ps["detour_method"].(string); ok && method != "" {
			cipher = method
		}
		handshakeServer := "addons.mozilla.org"
		if hs, ok := ps["handshake_server"].(string); ok && hs != "" {
			handshakeServer = hs
		}

		line := fmt.Sprintf("%s = ss, %s, %d, encrypt-method=%s, password=%s, shadow-tls-password=%s, shadow-tls-sni=%s, shadow-tls-version=3",
			server.Name, server.Host, port, cipher, user.UUID, user.UUID, handshakeServer)
		return line
	}

	return ""
}

// GenerateSurfboardConfig 生成 Surfboard 配置 (类似 Surge)
func GenerateSurfboardConfig(servers []service.ServerInfo, user *model.User) string {
	// Surfboard 配置格式告Surge 类似
	return GenerateSurgeConfig(servers, user)
}
