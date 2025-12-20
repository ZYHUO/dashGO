//go:build debug
// +build debug

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
	"strings"
	"sync"
	"syscall"
	"time"
)

// 调试版本的 Agent，增加详细日志输出

var (
	debugPanelURL            string
	debugToken               string
	debugConfigPath          string
	debugSingboxBin          string
	debugTriggerUpdate       bool
	debugAutoUpdate          bool
	debugUpdateCheckInterval int
	debugMode                bool
)

func init() {
	flag.StringVar(&debugPanelURL, "panel", "", "面板地址 (如 https://your-panel.com)")
	flag.StringVar(&debugToken, "token", "", "主机 Token")
	flag.StringVar(&debugConfigPath, "config", "/etc/sing-box/config.json", "sing-box 配置文件路径")
	flag.StringVar(&debugSingboxBin, "singbox", "sing-box", "sing-box 可执行文件路径")
	flag.BoolVar(&debugTriggerUpdate, "update", false, "手动触发更新")
	flag.BoolVar(&debugAutoUpdate, "auto-update", true, "是否启用自动更新检查")
	flag.IntVar(&debugUpdateCheckInterval, "update-check-interval", 3600, "更新检查间隔（秒）")
	flag.BoolVar(&debugMode, "debug", false, "启用调试模式")
}



// 这个函数现在由 runStartupDiagnostic 替代，保留用于兼容性
func checkSystemEnvironment() error {
	fmt.Printf("[INFO] 检查系统环境...\n")
	
	// 检查操作系统
	if debugMode || os.Getenv("DEBUG") != "" {
		fmt.Printf("[DEBUG] 操作系统: %s\n", runtime.GOOS)
		fmt.Printf("[DEBUG] 架构: %s\n", runtime.GOARCH)
		fmt.Printf("[DEBUG] CPU 核心数: %d\n", runtime.NumCPU())
	}
	
	// 检查 sing-box 可执行文件
	singboxPath, err := exec.LookPath(debugSingboxBin)
	if err != nil {
		fmt.Printf("[ERROR] 找不到 sing-box 可执行文件: %v\n", err)
		return fmt.Errorf("sing-box not found: %w", err)
	}
	fmt.Printf("[INFO] 找到 sing-box: %s\n", singboxPath)
	
	// 测试 sing-box 版本
	cmd := exec.Command(debugSingboxBin, "version")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("[ERROR] 无法获取 sing-box 版本: %v\n", err)
		return fmt.Errorf("failed to get sing-box version: %w", err)
	}
	fmt.Printf("[INFO] sing-box 版本: %s\n", string(output))
	
	// 检查配置目录
	configDir := "/etc/sing-box"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		fmt.Printf("[INFO] 创建配置目录: %s\n", configDir)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Printf("[ERROR] 无法创建配置目录: %v\n", err)
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}
	
	// 检查配置文件权限
	if _, err := os.Stat(debugConfigPath); err == nil {
		info, _ := os.Stat(debugConfigPath)
		if debugMode || os.Getenv("DEBUG") != "" {
			fmt.Printf("[DEBUG] 配置文件权限: %v\n", info.Mode())
		}
	}
	
	return nil
}

// 安全的 HTTP 请求
func safeAPIRequest(method, url string, body interface{}, timeout time.Duration) (map[string]interface{}, error) {
	if debugMode || os.Getenv("DEBUG") != "" {
		fmt.Printf("[DEBUG] API 请求: %s %s\n", method, url)
	}
	
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
		if debugMode || os.Getenv("DEBUG") != "" {
			fmt.Printf("[DEBUG] 请求体大小: %d bytes\n", len(data))
		}
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", debugToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("XBoard-Agent/%s (%s/%s)", Version, runtime.GOOS, runtime.GOARCH))

	client := &http.Client{Timeout: timeout}
	
	if debugMode || os.Getenv("DEBUG") != "" {
		fmt.Printf("[DEBUG] 发送请求到: %s\n", req.URL.String())
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] 请求失败: %v\n", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if debugMode || os.Getenv("DEBUG") != "" {
		fmt.Printf("[DEBUG] 响应状态: %d\n", resp.StatusCode)
	}
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	if debugMode || os.Getenv("DEBUG") != "" {
		fmt.Printf("[DEBUG] 响应体大小: %d bytes\n", len(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Printf("[ERROR] 响应解析失败: %v\n", err)
		if debugMode || os.Getenv("DEBUG") != "" {
			fmt.Printf("[DEBUG] 响应内容: %s\n", string(respBody))
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != 200 {
		if errMsg, ok := result["error"].(string); ok {
			return nil, fmt.Errorf("API error: %s", errMsg)
		}
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return result, nil
}

// 安全的 sing-box 启动
func safeSingboxStart(configPath, singboxBin string) (*exec.Cmd, error) {
	fmt.Printf("[INFO] 启动 sing-box...\n")
	if debugMode || os.Getenv("DEBUG") != "" {
		fmt.Printf("[DEBUG] 配置文件: %s\n", configPath)
		fmt.Printf("[DEBUG] 可执行文件: %s\n", singboxBin)
	}
	
	// 检查配置文件
	if _, err := os.Stat(configPath); err != nil {
		return nil, fmt.Errorf("config file not found: %w", err)
	}
	
	// 验证配置文件
	checkCmd := exec.Command(singboxBin, "check", "-c", configPath)
	if output, err := checkCmd.CombinedOutput(); err != nil {
		fmt.Printf("[ERROR] 配置文件验证失败: %v\n", err)
		fmt.Printf("[ERROR] 验证输出: %s\n", string(output))
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	if debugMode || os.Getenv("DEBUG") != "" {
		fmt.Printf("[DEBUG] 配置文件验证通过\n")
	}
	
	// 启动 sing-box
	cmd := exec.Command(singboxBin, "run", "-c", configPath)
	
	// 创建日志文件
	logFile, err := os.OpenFile("/tmp/sing-box.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		if debugMode || os.Getenv("DEBUG") != "" {
			fmt.Printf("[DEBUG] 无法创建日志文件: %v\n", err)
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		if debugMode || os.Getenv("DEBUG") != "" {
			fmt.Printf("[DEBUG] sing-box 日志输出到: /tmp/sing-box.log\n")
		}
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("[ERROR] 启动 sing-box 失败: %v\n", err)
		return nil, fmt.Errorf("failed to start sing-box: %w", err)
	}

	fmt.Printf("[INFO] sing-box 已启动，PID: %d\n", cmd.Process.Pid)
	
	// 等待一小段时间确保启动成功
	time.Sleep(2 * time.Second)
	
	// 检查进程是否还在运行
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		return nil, fmt.Errorf("sing-box exited immediately with code: %d", cmd.ProcessState.ExitCode())
	}
	
	return cmd, nil
}

// 安全的进程停止
func safeSingboxStop(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	
	fmt.Printf("[INFO] 停止 sing-box (PID: %d)\n", cmd.Process.Pid)
	
	// 发送 SIGTERM 信号
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		fmt.Printf("[ERROR] 发送 SIGTERM 失败: %v\n", err)
		// 强制杀死进程
		cmd.Process.Kill()
	}
	
	// 等待进程退出
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	
	select {
	case err := <-done:
		if err != nil {
			if debugMode || os.Getenv("DEBUG") != "" {
				fmt.Printf("[DEBUG] sing-box 退出: %v\n", err)
			}
		} else {
			fmt.Printf("[INFO] sing-box 已正常退出\n")
		}
	case <-time.After(10 * time.Second):
		fmt.Printf("[ERROR] sing-box 未在10秒内退出，强制杀死\n")
		cmd.Process.Kill()
		<-done
	}
}

// 增强的 Agent 结构
type DebugAgent struct {
	panelURL      string
	token         string
	configPath    string
	singboxBin    string
	singboxCmd    *exec.Cmd
	lastConfig    string
	httpClient    *http.Client
	mutex         sync.Mutex
	
	// 新增的调试组件
	logger        *DebugLogger
	systemChecker *AlpineSystemChecker
	errorHandler  *AlpineErrorHandler
	diagnosticTool *DiagnosticTool
}

func NewDebugAgent() *DebugAgent {
	// 确定日志级别
	logLevel := LogLevelInfo
	if debugMode || os.Getenv("DEBUG") != "" {
		logLevel = LogLevelDebug
	}
	if os.Getenv("TRACE") != "" {
		logLevel = LogLevelTrace
	}
	
	// 创建调试日志记录器
	logger := NewDebugLogger(logLevel, true)
	
	// 创建系统检查器
	systemChecker := NewAlpineSystemChecker(logger)
	
	// 创建错误处理器
	errorHandler := NewAlpineErrorHandler(logger, systemChecker)
	
	// 创建诊断工具
	diagnosticTool := NewDiagnosticTool(logger, systemChecker)
	
	agent := &DebugAgent{
		panelURL:       debugPanelURL,
		token:          debugToken,
		configPath:     debugConfigPath,
		singboxBin:     debugSingboxBin,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		logger:         logger,
		systemChecker:  systemChecker,
		errorHandler:   errorHandler,
		diagnosticTool: diagnosticTool,
	}
	
	return agent
}

func (a *DebugAgent) apiRequest(method, path string, body interface{}) (map[string]interface{}, error) {
	url := a.panelURL + "/api/v1/agent" + path
	
	// 记录 API 请求
	a.logger.LogAPIRequest(method, url, body)
	
	result, err := safeAPIRequest(method, url, body, 30*time.Second)
	
	// 记录 API 响应
	if err == nil {
		a.logger.LogAPIResponse(200, result)
	} else {
		a.logger.LogAPIResponse(0, map[string]interface{}{"error": err.Error()})
		// 使用错误处理器处理 API 错误
		alpineErr := a.errorHandler.HandleError(err, fmt.Sprintf("API请求: %s %s", method, path))
		if alpineErr != nil {
			a.logger.LogError(alpineErr, "API请求")
		}
	}
	
	return result, err
}

// runStartupDiagnostic 运行启动诊断
func (a *DebugAgent) runStartupDiagnostic() error {
	a.logger.Info("=== 开始启动诊断 ===")
	
	// 运行快速检查
	result, err := a.diagnosticTool.RunQuickCheck()
	if err != nil {
		return fmt.Errorf("快速检查失败: %w", err)
	}
	
	// 检查是否有严重错误
	hasError := false
	for _, err := range result.Errors {
		if err.Category == ErrorCategoryDependency && strings.Contains(err.Message, "sing-box") {
			hasError = true
			a.logger.Error("严重错误: %s", err.Message)
			if err.Suggestion != "" {
				a.logger.Info("建议: %s", err.Suggestion)
			}
		}
	}
	
	if hasError {
		return fmt.Errorf("启动诊断发现严重错误")
	}
	
	// 记录系统信息
	if result.SystemInfo != nil {
		a.logger.LogSystemInfo(*result.SystemInfo)
	}
	
	a.logger.Info("=== 启动诊断完成 ===")
	return nil
}

// runDiagnosticScript 运行诊断脚本
func (a *DebugAgent) runDiagnosticScript() {
	a.logger.Info("运行诊断脚本...")
	
	// 执行 shell 脚本
	output, err := a.diagnosticTool.ExecuteShellScript()
	if err != nil {
		a.logger.Error("诊断脚本执行失败: %v", err)
	}
	
	// 输出脚本结果
	fmt.Println(output)
	
	// 运行完整诊断
	a.logger.Info("\n运行完整诊断...")
	result, err := a.diagnosticTool.RunFullDiagnostic()
	if err != nil {
		a.logger.Error("完整诊断失败: %v", err)
		os.Exit(1)
	}
	
	// 生成并输出报告
	report := a.diagnosticTool.GenerateReport(result)
	fmt.Println("\n" + report)
	
	// 根据状态设置退出码
	if len(result.Errors) > 0 {
		os.Exit(1)
	}
}

func (a *DebugAgent) sendHeartbeat() error {
	a.logger.Debug("发送心跳...")
	
	systemInfo := map[string]interface{}{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"cpus":    runtime.NumCPU(),
		"version": Version,
	}

	result, err := a.apiRequest("POST", "/heartbeat", map[string]interface{}{
		"system_info": systemInfo,
	})
	
	if err != nil {
		a.logger.Error("心跳失败: %v", err)
		return err
	}
	
	a.logger.Debug("心跳成功: %+v", result)
	return nil
}

func (a *DebugAgent) getConfig() (map[string]interface{}, error) {
	a.logger.Debug("获取配置...")
	
	result, err := a.apiRequest("GET", "/config", nil)
	if err != nil {
		return nil, err
	}

	data, ok := result["data"]
	if !ok {
		return nil, fmt.Errorf("invalid response: missing data field")
	}
	
	config, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response: data is not an object")
	}
	
	a.logger.Debug("获取到配置，大小: %d 字段", len(config))
	return config, nil
}

func (a *DebugAgent) updateConfig(config map[string]interface{}) (bool, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	a.logger.Debug("更新配置...")
	
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		alpineErr := a.errorHandler.HandleError(err, "配置序列化")
		return false, alpineErr
	}
	
	configStr := string(configJSON)
	if configStr == a.lastConfig {
		a.logger.Debug("配置未变化，跳过更新")
		return false, nil
	}

	// 备份旧配置
	if _, err := os.Stat(a.configPath); err == nil {
		backupPath := a.configPath + ".backup"
		if err := exec.Command("cp", a.configPath, backupPath).Run(); err != nil {
			a.logger.Debug("备份配置失败: %v", err)
		} else {
			a.logger.Debug("配置已备份到: %s", backupPath)
		}
	}

	// 写入新配置
	if err := os.WriteFile(a.configPath, configJSON, 0644); err != nil {
		alpineErr := a.errorHandler.HandleError(err, "配置文件写入")
		return false, alpineErr
	}

	a.lastConfig = configStr
	a.logger.Info("配置已更新")
	return true, nil
}

func (a *DebugAgent) startSingbox() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	// 停止旧进程
	if a.singboxCmd != nil {
		safeSingboxStop(a.singboxCmd)
		a.singboxCmd = nil
	}

	// 启动新进程
	cmd, err := safeSingboxStart(a.configPath, a.singboxBin)
	if err != nil {
		alpineErr := a.errorHandler.HandleError(err, "sing-box启动")
		return alpineErr
	}
	
	a.singboxCmd = cmd
	return nil
}

func (a *DebugAgent) Run() {
	// 设置 panic 恢复
	defer a.errorHandler.RecoverFromPanic()
	
	a.logger.Info("XBoard Agent Debug 版本 %s 启动", Version)
	a.logger.Info("面板: %s", a.panelURL)
	a.logger.Info("调试模式: %v", debugMode || os.Getenv("DEBUG") != "")
	
	// 运行启动诊断
	if err := a.runStartupDiagnostic(); err != nil {
		a.logger.Error("启动诊断失败: %v", err)
		os.Exit(1)
	}
	
	// 首次获取配置
	config, err := a.getConfig()
	if err != nil {
		a.logger.Error("获取配置失败: %v", err)
		os.Exit(1)
	}

	if _, err := a.updateConfig(config); err != nil {
		a.logger.Error("更新配置失败: %v", err)
		os.Exit(1)
	}

	if err := a.startSingbox(); err != nil {
		a.logger.Error("启动 sing-box 失败: %v", err)
		os.Exit(1)
	}

	// 发送首次心跳
	if err := a.sendHeartbeat(); err != nil {
		a.logger.Error("首次心跳失败: %v", err)
	} else {
		a.logger.Info("已连接到面板")
	}

	// 启动定时任务
	heartbeatTicker := time.NewTicker(30 * time.Second)
	configTicker := time.NewTicker(60 * time.Second)
	diagnosticTicker := time.NewTicker(10 * time.Minute) // 每10分钟运行一次诊断

	// 信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	a.logger.Info("Agent 运行中...")
	
	for {
		select {
		case <-heartbeatTicker.C:
			if err := a.sendHeartbeat(); err != nil {
				a.logger.Error("心跳失败: %v", err)
			} else {
				a.logger.Debug("心跳成功")
			}

		case <-configTicker.C:
			config, err := a.getConfig()
			if err != nil {
				a.logger.Error("获取配置失败: %v", err)
				continue
			}

			updated, err := a.updateConfig(config)
			if err != nil {
				a.logger.Error("更新配置失败: %v", err)
				continue
			}

			if updated {
				a.logger.Info("配置已更新，重启 sing-box...")
				if err := a.startSingbox(); err != nil {
					a.logger.Error("重启失败: %v", err)
				}
			}

		case <-diagnosticTicker.C:
			// 定期运行快速诊断
			a.logger.Debug("运行定期诊断...")
			if result, err := a.diagnosticTool.RunQuickCheck(); err == nil {
				if len(result.Errors) > 0 {
					a.logger.Warn("诊断发现 %d 个问题", len(result.Errors))
					for _, err := range result.Errors {
						a.logger.Warn("问题: %s", err.Message)
					}
				}
			}

		case sig := <-sigChan:
			a.logger.Info("收到信号 %v，正在退出...", sig)
			heartbeatTicker.Stop()
			configTicker.Stop()
			diagnosticTicker.Stop()
			
			if a.singboxCmd != nil {
				safeSingboxStop(a.singboxCmd)
			}
			
			// 关闭日志记录器
			if a.logger != nil {
				a.logger.Close()
			}
			
			a.logger.Info("Agent 已退出")
			return
		}
	}
}

func main() {
	flag.Parse()

	if debugPanelURL == "" || debugToken == "" {
		fmt.Println("用法: xboard-agent-debug -panel <面板地址> -token <主机Token>")
		fmt.Println()
		fmt.Println("参数:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("调试选项:")
		fmt.Println("  -debug          启用调试模式")
		fmt.Println("  DEBUG=1         通过环境变量启用调试")
		fmt.Println("  TRACE=1         启用跟踪模式")
		fmt.Println()
		fmt.Println("诊断模式:")
		fmt.Println("  xboard-agent-debug diagnose  运行完整诊断")
		os.Exit(1)
	}

	agent := NewDebugAgent()
	
	// 如果指定了运行诊断脚本
	if len(os.Args) > 1 && os.Args[1] == "diagnose" {
		agent.runDiagnosticScript()
		return
	}
	
	agent.Run()
}