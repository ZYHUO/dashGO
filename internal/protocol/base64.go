package protocol

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"dashgo/internal/model"
	"dashgo/internal/service"
)

// GenerateBase64Links 生成 Base64 编码的订阅链接
func GenerateBase64Links(servers []service.ServerInfo, user *model.User) string {
	var links []string

	for _, server := range servers {
		link := generateLink(server, user)
		if link != "" {
			links = append(links, link)
		}
	}

	return base64.StdEncoding.EncodeToString([]byte(strings.Join(links, "\n")))
}

func generateLink(server service.ServerInfo, user *model.User) string {
	switch server.Type {
	case model.ServerTypeShadowsocks:
		return generateSSLink(server, user)
	case model.ServerTypeVmess:
		return generateVmessLink(server, user)
	case model.ServerTypeVless:
		return generateVlessLink(server, user)
	case model.ServerTypeTrojan:
		return generateTrojanLink(server, user)
	case model.ServerTypeHysteria:
		return generateHysteriaLink(server, user)
	case model.ServerTypeTuic:
		return generateTuicLink(server, user)
	case model.ServerTypeAnytls:
		return generateAnyTLSLink(server, user)
	case "shadowtls":
		return generateShadowTLSLink(server, user)
	}
	return ""
}

// ss://method:password@host:port#name
func generateSSLink(server service.ServerInfo, user *model.User) string {
	ps := server.ProtocolSettings

	// 获取加密方式，默认 aes-256-gcm
	cipher := "aes-256-gcm"
	if c, ok := ps["cipher"].(string); ok && c != "" {
		cipher = c
	} else if m, ok := ps["method"].(string); ok && m != "" {
		cipher = m
	}

	// 密码：对于 SS2022，使用 server.Password（已包含服务器密钥+用户密钥格式）
	// 对于普通 SS，使用用户 UUID
	password := server.Password
	if password == "" {
		password = user.UUID
	}

	// Base64 encode method:password
	userInfo := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", cipher, password)))

	link := fmt.Sprintf("ss://%s@%s:%s", userInfo, server.Host, server.Port)

	// Add plugin if exists
	if plugin, ok := ps["plugin"].(string); ok && plugin != "" {
		params := url.Values{}
		params.Set("plugin", plugin)
		if opts, ok := ps["plugin_opts"].(string); ok {
			params.Set("plugin-opts", opts)
		}
		link += "?" + params.Encode()
	}

	link += "#" + url.QueryEscape(server.Name)
	return link
}

// vmess://base64(json)
func generateVmessLink(server service.ServerInfo, user *model.User) string {
	ps := server.ProtocolSettings

	vmess := map[string]interface{}{
		"v":    "2",
		"ps":   server.Name,
		"add":  server.Host,
		"port": server.Port,
		"id":   user.UUID,
		"aid":  0,
		"scy":  "auto",
		"net":  "tcp",
		"type": "none",
		"tls":  "",
	}

	if network, ok := ps["network"].(string); ok {
		vmess["net"] = network
	}

	if tls, ok := ps["tls"].(float64); ok && tls > 0 {
		vmess["tls"] = "tls"
	}

	if ns, ok := ps["network_settings"].(map[string]interface{}); ok {
		switch vmess["net"] {
		case "ws":
			if path, ok := ns["path"].(string); ok {
				vmess["path"] = path
			}
			if headers, ok := ns["headers"].(map[string]interface{}); ok {
				if host, ok := headers["Host"].(string); ok {
					vmess["host"] = host
				}
			}
		case "grpc":
			if sn, ok := ns["serviceName"].(string); ok {
				vmess["path"] = sn
			}
		}
	}

	jsonData, _ := json.Marshal(vmess)
	return "vmess://" + base64.StdEncoding.EncodeToString(jsonData)
}

// vless://uuid@host:port?params#name
func generateVlessLink(server service.ServerInfo, user *model.User) string {
	ps := server.ProtocolSettings

	params := url.Values{}
	params.Set("encryption", "none")

	if network, ok := ps["network"].(string); ok {
		params.Set("type", network)
	} else {
		params.Set("type", "tcp")
	}

	if flow, ok := ps["flow"].(string); ok && flow != "" {
		params.Set("flow", flow)
	}

	if tls, ok := ps["tls"].(float64); ok {
		if tls == 2 { // Reality
			params.Set("security", "reality")
			params.Set("fp", "chrome")
			if reality, ok := ps["reality_settings"].(map[string]interface{}); ok {
				if pk, ok := reality["public_key"].(string); ok {
					params.Set("pbk", pk)
				}
				if sid, ok := reality["short_id"].(string); ok {
					params.Set("sid", sid)
				}
				if sn, ok := reality["server_name"].(string); ok {
					params.Set("sni", sn)
				}
			}
		} else if tls > 0 {
			params.Set("security", "tls")
			params.Set("fp", "chrome")
			if tlsSettings, ok := ps["tls_settings"].(map[string]interface{}); ok {
				if sn, ok := tlsSettings["server_name"].(string); ok {
					params.Set("sni", sn)
				}
			}
		}
	} else {
		params.Set("security", "none")
	}

	// Network settings
	if ns, ok := ps["network_settings"].(map[string]interface{}); ok {
		network := params.Get("type")
		switch network {
		case "ws":
			if path, ok := ns["path"].(string); ok {
				params.Set("path", path)
			}
			if headers, ok := ns["headers"].(map[string]interface{}); ok {
				if host, ok := headers["Host"].(string); ok {
					params.Set("host", host)
				}
			}
		case "grpc":
			if sn, ok := ns["serviceName"].(string); ok {
				params.Set("serviceName", sn)
			}
		}
	}

	link := fmt.Sprintf("vless://%s@%s:%s?%s#%s",
		user.UUID, server.Host, server.Port, params.Encode(), url.QueryEscape(server.Name))
	return link
}

// trojan://password@host:port?params#name
func generateTrojanLink(server service.ServerInfo, user *model.User) string {
	ps := server.ProtocolSettings

	params := url.Values{}

	if sn, ok := ps["server_name"].(string); ok && sn != "" {
		params.Set("sni", sn)
	}

	if network, ok := ps["network"].(string); ok && network != "" {
		params.Set("type", network)
		if ns, ok := ps["network_settings"].(map[string]interface{}); ok {
			switch network {
			case "ws":
				if path, ok := ns["path"].(string); ok {
					params.Set("path", path)
				}
			case "grpc":
				if sn, ok := ns["serviceName"].(string); ok {
					params.Set("serviceName", sn)
				}
			}
		}
	}

	link := fmt.Sprintf("trojan://%s@%s:%s", user.UUID, server.Host, server.Port)
	if len(params) > 0 {
		link += "?" + params.Encode()
	}
	link += "#" + url.QueryEscape(server.Name)
	return link
}

// hysteria2://password@host:port?params#name
func generateHysteriaLink(server service.ServerInfo, user *model.User) string {
	ps := server.ProtocolSettings

	version := 2
	if v, ok := ps["version"].(float64); ok {
		version = int(v)
	}

	params := url.Values{}

	if tls, ok := ps["tls"].(map[string]interface{}); ok {
		if sn, ok := tls["server_name"].(string); ok {
			params.Set("sni", sn)
		}
		if insecure, ok := tls["allow_insecure"].(bool); ok && insecure {
			params.Set("insecure", "1")
		}
	}

	if obfs, ok := ps["obfs"].(map[string]interface{}); ok {
		if open, ok := obfs["open"].(bool); ok && open {
			params.Set("obfs", obfs["type"].(string))
			params.Set("obfs-password", obfs["password"].(string))
		}
	}

	var scheme string
	if version == 2 {
		scheme = "hysteria2"
	} else {
		scheme = "hysteria"
		if bw, ok := ps["bandwidth"].(map[string]interface{}); ok {
			if up, ok := bw["up"].(float64); ok {
				params.Set("upmbps", fmt.Sprintf("%d", int(up)))
			}
			if down, ok := bw["down"].(float64); ok {
				params.Set("downmbps", fmt.Sprintf("%d", int(down)))
			}
		}
	}

	link := fmt.Sprintf("%s://%s@%s:%s", scheme, user.UUID, server.Host, server.Port)
	if len(params) > 0 {
		link += "?" + params.Encode()
	}
	link += "#" + url.QueryEscape(server.Name)
	return link
}

// tuic://uuid:password@host:port?params#name
func generateTuicLink(server service.ServerInfo, user *model.User) string {
	ps := server.ProtocolSettings

	params := url.Values{}
	params.Set("congestion_control", "cubic")
	params.Set("udp_relay_mode", "native")

	if cc, ok := ps["congestion_control"].(string); ok {
		params.Set("congestion_control", cc)
	}
	if urm, ok := ps["udp_relay_mode"].(string); ok {
		params.Set("udp_relay_mode", urm)
	}

	if tls, ok := ps["tls"].(map[string]interface{}); ok {
		if sn, ok := tls["server_name"].(string); ok {
			params.Set("sni", sn)
		}
	}

	link := fmt.Sprintf("tuic://%s:%s@%s:%s?%s#%s",
		user.UUID, user.UUID, server.Host, server.Port, params.Encode(), url.QueryEscape(server.Name))
	return link
}

// anytls://password@host:port?params#name
func generateAnyTLSLink(server service.ServerInfo, user *model.User) string {
	ps := server.ProtocolSettings

	params := url.Values{}

	if tls, ok := ps["tls"].(map[string]interface{}); ok {
		if sn, ok := tls["server_name"].(string); ok && sn != "" {
			params.Set("sni", sn)
		}
		if insecure, ok := tls["allow_insecure"].(bool); ok && insecure {
			params.Set("insecure", "1")
		}
	}

	link := fmt.Sprintf("anytls://%s@%s:%s", user.UUID, server.Host, server.Port)
	if len(params) > 0 {
		link += "?" + params.Encode()
	}
	link += "#" + url.QueryEscape(server.Name)
	return link
}

// ShadowTLS 使用 ss:// 格式，带 shadow-tls 插件
func generateShadowTLSLink(server service.ServerInfo, user *model.User) string {
	ps := server.ProtocolSettings

	// 获取加密方式
	cipher := "2022-blake3-aes-128-gcm"
	if method, ok := ps["detour_method"].(string); ok && method != "" {
		cipher = method
	}

	// 握手服务器
	handshakeServer := "addons.mozilla.org"
	if hs, ok := ps["handshake_server"].(string); ok && hs != "" {
		handshakeServer = hs
	}

	// Base64 encode method:password
	userInfo := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", cipher, user.UUID)))

	params := url.Values{}
	params.Set("plugin", "shadow-tls")
	params.Set("plugin-opts", fmt.Sprintf("host=%s;password=%s;version=3", handshakeServer, user.UUID))

	link := fmt.Sprintf("ss://%s@%s:%s?%s#%s",
		userInfo, server.Host, server.Port, params.Encode(), url.QueryEscape(server.Name))
	return link
}
