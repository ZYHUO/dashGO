package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

var (
	panelURL            string
	token               string
	configPath          string
	singboxBin          string
	triggerUpdate       bool
	autoUpdate          bool
	updateCheckInterval int
)

func init() {
	flag.StringVar(&panelURL, "panel", "", "面板地址 (�? https://your-panel.com)")
	flag.StringVar(&token, "token", "", "主机 Token")
	flag.StringVar(&configPath, "config", "/etc/sing-box/config.json", "sing-box 配置文件路径")
	flag.StringVar(&singboxBin, "singbox", "sing-box", "sing-box 可执行文件路�?)
	flag.BoolVar(&triggerUpdate, "update", false, "手动触发更新")
	flag.BoolVar(&autoUpdate, "auto-update", true, "是否启用自动更新检�?)
	flag.IntVar(&updateCheckInterval, "update-check-interval", 3600, "更新检查间隔（秒）")
}

type AgentConfig struct {
	SingBoxConfig map[string]interface{} `json:"singbox_config"`
	Nodes         []NodeConfig           `json:"nodes"`
}

type NodeConfig struct {
	ID    int64                    `json:"id"`
	Type  string                   `json:"type"`
	Port  int                      `json:"port"`
	Tag   string                   `json:"tag"`
	Users []map[string]interface{} `json:"users"`
}

type Agent struct {
	panelURL            string
	token               string
	configPath          string
	singboxBin          string
	singboxCmd          *exec.Cmd
	lastConfig          string
	httpClient          *http.Client
	userVersions        map[int64]int64        // 节点用户版本缓存
	userHashes          map[int64]string       // 节点用户哈希缓存
	lastTraffic         map[string]TrafficData // 上次流量数据，用于计算增�?
	nodeConfigs         []NodeConfig           // 当前节点配置
	clashAPIPort        int                    // Clash API 端口
	portUserMap         map[int][]string       // 端口到用户的映射（用于单端口多用户场景）
	versionManager      *VersionManager        // 版本管理�?
	updateChecker       *UpdateChecker         // 更新检查器
	updateNotifier      *UpdateNotifier        // 更新通知�?
	updatePending       *UpdateInfo            // 待处理的更新信息
	manualUpdate        bool                   // 是否手动触发更新
	autoUpdate          bool                   // 是否启用自动更新检�?
	updateCheckInterval time.Duration          // 更新检查间�?
	updateMutex         sync.Mutex             // 更新互斥�?
	updating            bool                   // 是否正在更新
}

// TrafficData 流量数据
type TrafficData struct {
	Upload   int64
	Download int64
}

func NewAgent(manualUpdate bool, autoUpdate bool, updateCheckInterval int) *Agent {
	versionManager := NewVersionManager(Version)
	updateChecker := NewUpdateChecker(panelURL, token, versionManager)
	updateNotifier := NewUpdateNotifier(panelURL, token)
	
	return &Agent{
		panelURL:            panelURL,
		token:               token,
		configPath:          configPath,
		singboxBin:          singboxBin,
		httpClient:          &http.Client{Timeout: 30 * time.Second},
		userVersions:        make(map[int64]int64),
		userHashes:          make(map[int64]string),
		lastTraffic:         make(map[string]TrafficData),
		portUserMap:         make(map[int][]string),
		clashAPIPort:        9090,
		versionManager:      versionManager,
		updateChecker:       updateChecker,
		updateNotifier:      updateNotifier,
		manualUpdate:        manualUpdate,
		autoUpdate:          autoUpdate,
		updateCheckInterval: time.Duration(updateCheckInterval) * time.Second,
		updating:            false,
	}
}

// getNodeUsers 获取节点用户（支持增量同步）
// nodeType: "server" �?"node"
func (a *Agent) getNodeUsers(nodeID int64, nodeType string) ([]map[string]interface{}, bool, error) {
	hash := a.userHashes[nodeID]

	url := fmt.Sprintf("/users?node_id=%d&type=%s&hash=%s", nodeID, nodeType, hash)
	result, err := a.apiRequest("GET", url, nil)
	if err != nil {
		return nil, false, err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("invalid response")
	}

	hasChange, _ := data["has_change"].(bool)
	if !hasChange {
		return nil, false, nil
	}

	// 更新哈希
	if h, ok := data["hash"].(string); ok {
		a.userHashes[nodeID] = h
	}

	users, ok := data["users"].([]interface{})
	if !ok {
		return nil, true, nil
	}

	result_users := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		if user, ok := u.(map[string]interface{}); ok {
			result_users = append(result_users, user)
		}
	}

	return result_users, true, nil
}

func (a *Agent) apiRequest(method, path string, body interface{}) (map[string]interface{}, error) {
	url := a.panelURL + "/api/v1/agent" + path
	
	var reqBody io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", a.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		if errMsg, ok := result["error"].(string); ok {
			return nil, fmt.Errorf(errMsg)
		}
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return result, nil
}

func (a *Agent) sendHeartbeat() error {
	systemInfo := map[string]interface{}{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"cpus":    runtime.NumCPU(),
		"version": a.versionManager.GetCurrentVersion(),
	}

	result, err := a.apiRequest("POST", "/heartbeat", map[string]interface{}{
		"system_info": systemInfo,
	})
	
	// 检查心跳响应中是否包含版本信息
	if err == nil && result != nil {
		if data, ok := result["data"].(map[string]interface{}); ok {
			if versionInfo, ok := data["version_info"].(map[string]interface{}); ok {
				// 将版本信息转换为 UpdateInfo
				updateInfo := &UpdateInfo{}
				if latestVersion, ok := versionInfo["latest_version"].(string); ok {
					updateInfo.LatestVersion = latestVersion
				}
				if downloadURL, ok := versionInfo["download_url"].(string); ok {
					updateInfo.DownloadURL = downloadURL
				}
				if sha256, ok := versionInfo["sha256"].(string); ok {
					updateInfo.SHA256 = sha256
				}
				if fileSize, ok := versionInfo["file_size"].(float64); ok {
					updateInfo.FileSize = int64(fileSize)
				}
				if strategy, ok := versionInfo["strategy"].(string); ok {
					updateInfo.Strategy = strategy
				}
				if releaseNotes, ok := versionInfo["release_notes"].(string); ok {
					updateInfo.ReleaseNotes = releaseNotes
				}
				
				// 如果有版本信息，检查是否需要更�?
				if updateInfo.LatestVersion != "" {
					a.handleUpdateInfo(updateInfo)
				}
			}
		}
	}
	
	return err
}

// checkForUpdates 检查更�?
func (a *Agent) checkForUpdates() error {
	currentVersion := a.versionManager.GetCurrentVersion()
	
	updateInfo, err := a.updateChecker.CheckUpdate(currentVersion)
	if err != nil {
		return fmt.Errorf("failed to check update: %w", err)
	}
	
	a.handleUpdateInfo(updateInfo)
	return nil
}

// handleUpdateInfo 处理更新信息
func (a *Agent) handleUpdateInfo(updateInfo *UpdateInfo) {
	if updateInfo == nil || updateInfo.LatestVersion == "" {
		return
	}
	
	shouldUpdate, err := a.updateChecker.ShouldUpdate(updateInfo)
	if err != nil {
		updateErr := NewUpdateError("版本比较失败", err)
		HandleError(updateErr)
		return
	}
	
	if !shouldUpdate {
		// 版本相同或当前版本更新，无需更新
		return
	}
	
	// 检测到新版�?
	fmt.Printf("🔔 检测到新版�? %s (当前版本: %s)\n", 
		updateInfo.LatestVersion, 
		a.versionManager.GetCurrentVersion())
	
	if updateInfo.ReleaseNotes != "" {
		fmt.Printf("📝 更新说明: %s\n", updateInfo.ReleaseNotes)
	}
	
	// 根据更新策略决定是否自动更新
	if updateInfo.Strategy == "auto" {
		fmt.Println("🚀 自动更新策略已启用，准备更新...")
		if err := a.performUpdate(updateInfo); err != nil {
			// 错误已在 performUpdate 中处理和记录
			fmt.Printf("�?自动更新失败: %v\n", err)
		}
	} else {
		// 手动更新策略
		fmt.Println("ℹ️  手动更新策略已启用，等待手动触发更新")
		fmt.Printf("   下载地址: %s\n", updateInfo.DownloadURL)
		fmt.Println("   使用 -update 参数重启 Agent 以执行更�?)
		
		// 保存待处理的更新信息
		a.updatePending = updateInfo
		
		// 如果是手动触发更新，立即执行
		if a.manualUpdate {
			fmt.Println("🚀 手动触发更新...")
			if err := a.performUpdate(updateInfo); err != nil {
				// 错误已在 performUpdate 中处理和记录
				fmt.Printf("�?手动更新失败: %v\n", err)
			}
		}
	}
}

// performUpdate 执行更新流程
func (a *Agent) performUpdate(updateInfo *UpdateInfo) error {
	// 使用互斥锁防止并发更�?
	a.updateMutex.Lock()
	defer a.updateMutex.Unlock()
	
	if a.updating {
		err := fmt.Errorf("更新已在进行�?)
		HandleError(err)
		return err
	}
	
	a.updating = true
	defer func() { a.updating = false }()
	
	currentVersion := a.versionManager.GetCurrentVersion()
	targetVersion := updateInfo.LatestVersion
	
	fmt.Printf("🚀 开始更新流�? %s -> %s\n", currentVersion, targetVersion)
	fmt.Println("📥 开始下载新版本...")
	
	// 创建更新�?
	updater, err := NewUpdater()
	if err != nil {
		updateErr := NewUpdateError("创建更新器失�?, err)
		HandleError(updateErr)
		a.updateNotifier.NotifyFailure(currentVersion, targetVersion, updateErr)
		return updateErr
	}
	
	// 创建下载�?
	downloader := NewDownloader()
	
	// 下载新版本到临时文件
	newPath := updater.GetNewPath()
	fmt.Printf("   下载�? %s\n", newPath)
	
	if err := downloader.DownloadWithRetry(updateInfo.DownloadURL, newPath); err != nil {
		updateErr := NewNetworkError("下载失败", err)
		HandleError(updateErr)
		a.updateNotifier.NotifyFailure(currentVersion, targetVersion, updateErr)
		return updateErr
	}
	
	fmt.Println("�?下载完成")
	
	// 验证文件
	fmt.Println("🔍 验证文件完整�?..")
	verifier := NewFileVerifier()
	
	if err := verifier.VerifyAll(newPath, updateInfo.FileSize, updateInfo.SHA256); err != nil {
		// 验证失败，清理下载的文件
		updater.CleanupNew()
		updateErr := NewVerificationError("文件验证失败", err)
		HandleError(updateErr)
		a.updateNotifier.NotifyFailure(currentVersion, targetVersion, updateErr)
		return updateErr
	}
	
	fmt.Println("�?文件验证通过")
	
	// 备份当前版本
	fmt.Println("💾 备份当前版本...")
	if err := updater.Backup(); err != nil {
		updater.CleanupNew()
		updateErr := NewFileError("备份失败", err)
		HandleError(updateErr)
		a.updateNotifier.NotifyFailure(currentVersion, targetVersion, updateErr)
		return updateErr
	}
	
	fmt.Println("�?备份完成")
	
	// 替换可执行文�?
	fmt.Println("🔄 替换可执行文�?..")
	if err := updater.Replace(); err != nil {
		// 替换失败，尝试回�?
		fmt.Println("�?替换失败，正在回�?..")
		if rollbackErr := updater.Rollback(); rollbackErr != nil {
			updateErr := NewUpdateError("替换失败且回滚失�?, err)
			HandleError(updateErr)
			a.updateNotifier.NotifyFailure(currentVersion, targetVersion, updateErr)
			return updateErr
		}
		fmt.Println("�?已回滚到原版�?)
		updateErr := NewUpdateError("替换失败", err)
		HandleError(updateErr)
		a.updateNotifier.NotifyRollback(currentVersion, targetVersion, updateErr)
		return updateErr
	}
	
	fmt.Println("�?替换完成")
	
	// 注意：sing-box 进程继续运行，不需要停�?
	fmt.Println("ℹ️  sing-box 服务继续运行�?..")
	
	// 发送更新成功通知（在重启前发送，因为重启会退出进程）
	fmt.Println("📤 发送更新成功通知...")
	if err := a.updateNotifier.NotifySuccess(currentVersion, targetVersion); err != nil {
		// 通知失败不影响更新流�?
		fmt.Printf("�?发送成功通知失败: %v\n", err)
	}
	
	// 重启 Agent（新进程会接�?sing-box 管理�?
	fmt.Println("🔄 重启 Agent...")
	fmt.Printf("�?更新成功！正在启动新版本 %s\n", targetVersion)
	
	// 重启会导致当前进程退�?
	if err := updater.Restart(); err != nil {
		// 重启失败，回�?
		fmt.Println("�?重启失败，正在回�?..")
		if rollbackErr := updater.Rollback(); rollbackErr != nil {
			updateErr := NewUpdateError("重启失败且回滚失�?, err)
			HandleError(updateErr)
			a.updateNotifier.NotifyFailure(currentVersion, targetVersion, updateErr)
			return updateErr
		}
		fmt.Println("�?已回滚到原版�?)
		updateErr := NewUpdateError("重启失败", err)
		HandleError(updateErr)
		a.updateNotifier.NotifyRollback(currentVersion, targetVersion, updateErr)
		return updateErr
	}
	
	return nil
}

func (a *Agent) getConfig() (*AgentConfig, error) {
	result, err := a.apiRequest("GET", "/config", nil)
	if err != nil {
		return nil, err
	}

	data, ok := result["data"]
	if !ok {
		return nil, fmt.Errorf("invalid response")
	}

	configData, _ := json.Marshal(data)
	var config AgentConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (a *Agent) updateConfig(config *AgentConfig) (bool, error) {
	// 保存节点配置用于流量上报
	a.nodeConfigs = config.Nodes

	// 构建端口到用户的映射
	a.portUserMap = make(map[int][]string)
	for _, node := range config.Nodes {
		users := make([]string, 0, len(node.Users))
		for _, user := range node.Users {
			if name, ok := user["name"].(string); ok {
				users = append(users, name)
			}
		}
		a.portUserMap[node.Port] = users
	}

	// 注入用户�?inbounds
	singboxConfig := config.SingBoxConfig
	hasUserChange := false

	if inbounds, ok := singboxConfig["inbounds"].([]interface{}); ok {
		for i, inbound := range inbounds {
			if ib, ok := inbound.(map[string]interface{}); ok {
				tag, _ := ib["tag"].(string)
				// 找到对应的节点配�?
				for _, node := range config.Nodes {
					if node.Tag == tag {
						// 直接使用配置中的用户（已经是正确格式�?
						// 不再单独调用用户接口，因�?GetAgentConfig 已经返回了正确格式的用户
						if len(node.Users) > 0 {
							ib["users"] = node.Users
							hasUserChange = true
						}
						inbounds[i] = ib
						break
					}
				}
			}
		}
		singboxConfig["inbounds"] = inbounds
	}

	// 添加 experimental 配置用于流量统计
	if _, ok := singboxConfig["experimental"]; !ok {
		singboxConfig["experimental"] = map[string]interface{}{}
	}
	experimental := singboxConfig["experimental"].(map[string]interface{})
	
	// 添加 Clash API 用于获取连接信息
	if _, ok := experimental["clash_api"]; !ok {
		experimental["clash_api"] = map[string]interface{}{
			"external_controller": fmt.Sprintf("127.0.0.1:%d", a.clashAPIPort),
		}
	}
	singboxConfig["experimental"] = experimental

	configJSON, _ := json.MarshalIndent(singboxConfig, "", "  ")
	configStr := string(configJSON)

	if configStr == a.lastConfig && !hasUserChange {
		return false, nil
	}

	// 写入配置文件
	if err := os.WriteFile(a.configPath, configJSON, 0644); err != nil {
		return false, err
	}

	a.lastConfig = configStr
	return true, nil
}

func (a *Agent) startSingbox() error {
	a.stopSingbox()

	a.singboxCmd = exec.Command(a.singboxBin, "run", "-c", a.configPath)
	a.singboxCmd.Stdout = os.Stdout
	a.singboxCmd.Stderr = os.Stderr

	if err := a.singboxCmd.Start(); err != nil {
		return err
	}

	fmt.Println("�?sing-box 已启�?)
	return nil
}

func (a *Agent) stopSingbox() {
	if a.singboxCmd != nil && a.singboxCmd.Process != nil {
		a.singboxCmd.Process.Signal(syscall.SIGTERM)
		a.singboxCmd.Wait()
		fmt.Println("�?sing-box 已停�?)
	}
}

// ConnectionTraffic 连接流量记录
type ConnectionTraffic struct {
	Upload   int64
	Download int64
}

// getTrafficFromClashAPI �?Clash API 获取流量统计
// 通过跟踪每个连接的流量变化来计算用户流量
func (a *Agent) getTrafficFromClashAPI() (map[string]TrafficData, error) {
	url := fmt.Sprintf("http://127.0.0.1:%d/connections", a.clashAPIPort)
	resp, err := a.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 使用 map 解析以支持不同版本的 sing-box
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 按用户聚合当前连接的流量
	traffic := make(map[string]TrafficData)
	
	connections, ok := result["connections"].([]interface{})
	if !ok {
		return traffic, nil
	}

	for _, c := range connections {
		conn, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		upload, _ := conn["upload"].(float64)
		download, _ := conn["download"].(float64)

		// 获取用户名，尝试多种字段
		var user string
		if metadata, ok := conn["metadata"].(map[string]interface{}); ok {
			// 尝试不同的字段名
			if u, ok := metadata["inboundUser"].(string); ok && u != "" {
				user = u
			} else if u, ok := metadata["user"].(string); ok && u != "" {
				user = u
			} else if u, ok := metadata["inbound_user"].(string); ok && u != "" {
				user = u
			}
		}

		if user == "" {
			continue
		}

		data := traffic[user]
		data.Upload += int64(upload)
		data.Download += int64(download)
		traffic[user] = data
	}

	return traffic, nil
}

// reportTraffic 上报流量到面�?
// 策略：优先尝试用户级流量，失败则使用端口流量平均分配
func (a *Agent) reportTraffic() error {
	// 方案1：尝试从 Clash API 获取用户级流�?
	traffic, err := a.getTrafficFromClashAPI()
	if err == nil && len(traffic) > 0 {
		return a.reportUserTraffic(traffic)
	}

	// 方案2：使用端口流量平均分配（备用方案�?
	// 这种方式不够精确，但至少能统计总流�?
	return a.reportTrafficByPort()
}

// reportUserTraffic 上报用户级流量（精确统计�?
func (a *Agent) reportUserTraffic(traffic map[string]TrafficData) error {
	fmt.Printf("📊 获取�?%d 个用户的流量数据\n", len(traffic))

	// 计算增量流量
	trafficReport := make([]map[string]interface{}, 0)
	for user, data := range traffic {
		last := a.lastTraffic[user]
		uploadDelta := data.Upload - last.Upload
		downloadDelta := data.Download - last.Download

		// 只上报有增量的用�?
		if uploadDelta > 0 || downloadDelta > 0 {
			trafficReport = append(trafficReport, map[string]interface{}{
				"username": user,
				"upload":   uploadDelta,
				"download": downloadDelta,
			})
			fmt.Printf("  用户 %s: �?.2f MB �?.2f MB\n", user, float64(uploadDelta)/1024/1024, float64(downloadDelta)/1024/1024)
		}
		a.lastTraffic[user] = data
	}

	if len(trafficReport) == 0 {
		return nil // 没有流量变化
	}

	// 构建上报数据
	nodes := make([]map[string]interface{}, 0)
	for _, node := range a.nodeConfigs {
		nodes = append(nodes, map[string]interface{}{
			"id":    node.ID,
			"users": trafficReport,
		})
	}

	_, err := a.apiRequest("POST", "/traffic", map[string]interface{}{
		"nodes": nodes,
	})
	if err != nil {
		fmt.Printf("�?流量上报失败: %v\n", err)
	} else {
		fmt.Printf("�?已上�?%d 个用户的流量\n", len(trafficReport))
	}
	return err
}

// reportTrafficByPort 通过端口流量平均分配给用户（备用方案�?
// 注意：这种方式不够精确，但至少能统计总流�?
func (a *Agent) reportTrafficByPort() error {
	// 尝试�?Clash API 获取总流�?
	url := fmt.Sprintf("http://127.0.0.1:%d/traffic", a.clashAPIPort)
	resp, err := a.httpClient.Get(url)
	if err != nil {
		// Clash API 完全不可用，跳过本次上报
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Up   int64 `json:"up"`
		Down int64 `json:"down"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	// 如果没有流量，直接返�?
	if result.Up == 0 && result.Down == 0 {
		return nil
	}

	// 计算增量
	lastTotal := a.lastTraffic["__total__"]
	uploadDelta := result.Up - lastTotal.Upload
	downloadDelta := result.Down - lastTotal.Download

	if uploadDelta <= 0 && downloadDelta <= 0 {
		return nil
	}

	a.lastTraffic["__total__"] = TrafficData{
		Upload:   result.Up,
		Download: result.Down,
	}

	fmt.Printf("📊 总流量（平均分配模式�? �?.2f MB �?.2f MB\n", float64(uploadDelta)/1024/1024, float64(downloadDelta)/1024/1024)

	// 统计所有用户数
	totalUsers := 0
	for _, node := range a.nodeConfigs {
		totalUsers += len(a.portUserMap[node.Port])
	}

	if totalUsers == 0 {
		return nil
	}

	// 为每个节点的所有用户平均分配流�?
	nodes := make([]map[string]interface{}, 0)
	for _, node := range a.nodeConfigs {
		users := a.portUserMap[node.Port]
		if len(users) == 0 {
			continue
		}

		// 按节点用户数比例分配流量
		nodeRatio := float64(len(users)) / float64(totalUsers)
		nodeUpload := int64(float64(uploadDelta) * nodeRatio)
		nodeDownload := int64(float64(downloadDelta) * nodeRatio)

		// 再平均分配给该节点的用户
		avgUpload := nodeUpload / int64(len(users))
		avgDownload := nodeDownload / int64(len(users))

		trafficReport := make([]map[string]interface{}, 0, len(users))
		for _, user := range users {
			trafficReport = append(trafficReport, map[string]interface{}{
				"username": user,
				"upload":   avgUpload,
				"download": avgDownload,
			})
		}

		nodes = append(nodes, map[string]interface{}{
			"id":    node.ID,
			"users": trafficReport,
		})

		fmt.Printf("  节点 %d: �?%d 个用户分配流量（平均 �?.2f MB �?.2f MB/人）\n", 
			node.ID, len(users), 
			float64(avgUpload)/1024/1024, 
			float64(avgDownload)/1024/1024)
	}

	if len(nodes) == 0 {
		return nil
	}

	_, err = a.apiRequest("POST", "/traffic", map[string]interface{}{
		"nodes": nodes,
	})
	if err != nil {
		fmt.Printf("�?流量上报失败: %v\n", err)
	} else {
		fmt.Printf("�?已上报流量（平均分配模式）\n")
	}
	return err
}

func (a *Agent) Run() {
	// 启动时记录当前版�?
	currentVersion := a.versionManager.GetCurrentVersion()
	fmt.Printf("XBoard Agent %s\n", currentVersion)
	fmt.Printf("面板: %s\n", a.panelURL)
	
	// 显示更新配置
	if a.autoUpdate {
		fmt.Printf("自动更新: 已启�?(检查间�? %v)\n", a.updateCheckInterval)
	} else {
		fmt.Println("自动更新: 已禁�?)
	}
	
	fmt.Println("正在连接...")

	// 首次获取配置并启�?
	config, err := a.getConfig()
	if err != nil {
		fmt.Printf("�?获取配置失败: %v\n", err)
		os.Exit(1)
	}

	if _, err := a.updateConfig(config); err != nil {
		fmt.Printf("�?更新配置失败: %v\n", err)
		os.Exit(1)
	}

	if err := a.startSingbox(); err != nil {
		fmt.Printf("�?启动 sing-box 失败: %v\n", err)
		os.Exit(1)
	}

	// 发送首次心跳（包含版本信息�?
	if err := a.sendHeartbeat(); err != nil {
		fmt.Printf("�?心跳发送失�? %v\n", err)
	} else {
		fmt.Println("�?已连接到面板")
	}

	// 启动定时任务
	heartbeatTicker := time.NewTicker(30 * time.Second)
	configTicker := time.NewTicker(60 * time.Second)
	trafficTicker := time.NewTicker(60 * time.Second) // 每分钟上报流�?
	
	// 添加定期检查更新的 ticker（可配置间隔�?
	var updateCheckTicker *time.Ticker
	if a.autoUpdate && a.updateCheckInterval > 0 {
		updateCheckTicker = time.NewTicker(a.updateCheckInterval)
		defer updateCheckTicker.Stop()
	}

	// 信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-heartbeatTicker.C:
			if err := a.sendHeartbeat(); err != nil {
				fmt.Printf("�?心跳失败: %v\n", err)
			}

		case <-trafficTicker.C:
			if err := a.reportTraffic(); err != nil {
				// 流量上报失败不打印错误，可能�?sing-box 还没启动完成
			}

		case <-configTicker.C:
			config, err := a.getConfig()
			if err != nil {
				fmt.Printf("�?获取配置失败: %v\n", err)
				continue
			}

			updated, err := a.updateConfig(config)
			if err != nil {
				fmt.Printf("�?更新配置失败: %v\n", err)
				continue
			}

			if updated {
				fmt.Println("配置已更新，重启 sing-box...")
				if err := a.startSingbox(); err != nil {
					fmt.Printf("�?重启失败: %v\n", err)
				}
			}

		case <-func() <-chan time.Time {
			if updateCheckTicker != nil {
				return updateCheckTicker.C
			}
			// 返回一个永远不会触发的 channel
			return make(<-chan time.Time)
		}():
			// 定期检查更�?
			if err := a.checkForUpdates(); err != nil {
				fmt.Printf("�?检查更新失�? %v\n", err)
			}

		case sig := <-sigChan:
			fmt.Printf("\n收到信号 %v，正在退�?..\n", sig)
			heartbeatTicker.Stop()
			configTicker.Stop()
			trafficTicker.Stop()
			if updateCheckTicker != nil {
				updateCheckTicker.Stop()
			}
			a.stopSingbox()
			return
		}
	}
}

func main() {
	flag.Parse()

	if panelURL == "" || token == "" {
		fmt.Println("用法: xboard-agent -panel <面板地址> -token <主机Token>")
		fmt.Println()
		fmt.Println("参数:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	agent := NewAgent(triggerUpdate, autoUpdate, updateCheckInterval)
	agent.Run()
}
