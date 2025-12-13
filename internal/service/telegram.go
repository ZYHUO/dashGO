package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"dashgo/internal/config"
	"dashgo/internal/model"
	"dashgo/internal/repository"
)

// TelegramService Telegram Bot 服务
type TelegramService struct {
	botToken    string
	chatID      string
	httpClient  *http.Client
	userRepo    *repository.UserRepository
	settingRepo *repository.SettingRepository
}

func NewTelegramService(cfg config.TelegramConfig) *TelegramService {
	return &TelegramService{
		botToken:   cfg.BotToken,
		chatID:     cfg.ChatID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetRepositories 设置仓库依赖
func (s *TelegramService) SetRepositories(userRepo *repository.UserRepository, settingRepo *repository.SettingRepository) {
	s.userRepo = userRepo
	s.settingRepo = settingRepo
}

// GetBotToken 获取 Bot Token
func (s *TelegramService) GetBotToken() string {
	return s.botToken
}

// TelegramUpdate Telegram 更新
type TelegramUpdate struct {
	UpdateID      int64                  `json:"update_id"`
	Message       *TelegramMessage       `json:"message"`
	CallbackQuery *TelegramCallbackQuery `json:"callback_query"`
}

// TelegramMessage Telegram 消息
type TelegramMessage struct {
	MessageID int64         `json:"message_id"`
	From      *TelegramUser `json:"from"`
	Chat      *TelegramChat `json:"chat"`
	Text      string        `json:"text"`
	Date      int64         `json:"date"`
}

// TelegramUser Telegram 用户
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// TelegramChat Telegram 聊天
type TelegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// TelegramCallbackQuery 回调查询
type TelegramCallbackQuery struct {
	ID      string           `json:"id"`
	From    *TelegramUser    `json:"from"`
	Message *TelegramMessage `json:"message"`
	Data    string           `json:"data"`
}

// InlineKeyboard 内联键盘
type InlineKeyboard struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// InlineKeyboardButton 内联键盘按钮
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}


// SendMessage 发送消和
func (s *TelegramService) SendMessage(chatID int64, text string, parseMode string) error {
	if s.botToken == "" {
		return fmt.Errorf("telegram bot not configured")
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.botToken)
	params := url.Values{}
	params.Set("chat_id", fmt.Sprintf("%d", chatID))
	params.Set("text", text)
	if parseMode != "" {
		params.Set("parse_mode", parseMode)
	}
	resp, err := s.httpClient.PostForm(apiURL, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// SendMessageWithKeyboard 发送带键盘的消和
func (s *TelegramService) SendMessageWithKeyboard(chatID int64, text string, keyboard *InlineKeyboard) error {
	if s.botToken == "" {
		return fmt.Errorf("telegram bot not configured")
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.botToken)
	data := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}
	body, _ := json.Marshal(data)
	resp, err := s.httpClient.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// AnswerCallbackQuery 回答回调查询
func (s *TelegramService) AnswerCallbackQuery(queryID string, text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", s.botToken)
	params := url.Values{}
	params.Set("callback_query_id", queryID)
	if text != "" {
		params.Set("text", text)
	}
	resp, err := s.httpClient.PostForm(apiURL, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// SendMarkdown 发和Markdown 消息
func (s *TelegramService) SendMarkdown(chatID int64, text string) error {
	return s.SendMessage(chatID, text, "Markdown")
}

// HandleUpdate 处理 Telegram 更新
func (s *TelegramService) HandleUpdate(update *TelegramUpdate) error {
	if update.CallbackQuery != nil {
		return s.handleCallback(update.CallbackQuery)
	}
	if update.Message == nil {
		return nil
	}
	msg := update.Message
	text := strings.TrimSpace(msg.Text)
	if strings.HasPrefix(text, "/") {
		return s.handleCommand(msg)
	}
	return nil
}

func (s *TelegramService) handleCallback(query *TelegramCallbackQuery) error {
	s.AnswerCallbackQuery(query.ID, "")
	parts := strings.Split(query.Data, ":")
	if len(parts) < 1 {
		return nil
	}
	switch parts[0] {
	case "unbind":
		return s.doUnbind(query.From.ID, query.Message.Chat.ID)
	case "refresh":
		return s.cmdInfo(&TelegramMessage{From: query.From, Chat: query.Message.Chat})
	}
	return nil
}

func (s *TelegramService) handleCommand(msg *TelegramMessage) error {
	parts := strings.Fields(msg.Text)
	if len(parts) == 0 {
		return nil
	}
	cmd := strings.ToLower(strings.Split(parts[0], "@")[0])
	switch cmd {
	case "/start":
		return s.cmdStart(msg)
	case "/help":
		return s.cmdHelp(msg)
	case "/bind":
		if len(parts) < 2 {
			return s.SendMarkdown(msg.Chat.ID, "和请提供邮箱：`/bind your@email.com`")
		}
		return s.cmdBind(msg, parts[1])
	case "/unbind":
		return s.cmdUnbind(msg)
	case "/info", "/me":
		return s.cmdInfo(msg)
	case "/traffic":
		return s.cmdTraffic(msg)
	case "/subscribe", "/sub":
		return s.cmdSubscribe(msg)
	case "/checkin":
		return s.cmdCheckin(msg)
	default:
		return s.SendMessage(msg.Chat.ID, "和未知命令，输和/help 查看帮助", "")
	}
}

func (s *TelegramService) cmdStart(msg *TelegramMessage) error {
	siteName := s.getSiteName()
	text := fmt.Sprintf("🎉 *欢迎使用 %s Bot*\n\n/bind <邮箱> - 绑定账户\n/info - 查看账户\n/traffic - 流量使用\n/subscribe - 订阅链接\n/checkin - 每日签到\n/help - 帮助", siteName)
	return s.SendMarkdown(msg.Chat.ID, text)
}

func (s *TelegramService) cmdHelp(msg *TelegramMessage) error {
	text := "📖 *帮助*\n\n/bind <邮箱> - 绑定账户\n/unbind - 解绑账户\n/info - 账户信息\n/traffic - 流量使用\n/subscribe - 订阅链接\n/checkin - 每日签到"
	return s.SendMarkdown(msg.Chat.ID, text)
}

func (s *TelegramService) cmdBind(msg *TelegramMessage, email string) error {
	existingUser, _ := s.userRepo.FindByTelegramID(msg.From.ID)
	if existingUser != nil {
		return s.SendMarkdown(msg.Chat.ID, fmt.Sprintf("⚠️ 已绑定：`%s`\n使用 /unbind 解绑", existingUser.Email))
	}
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return s.SendMarkdown(msg.Chat.ID, "和未找到该邮箱账户")
	}
	if user.TelegramID != nil && *user.TelegramID != 0 {
		return s.SendMarkdown(msg.Chat.ID, "和该账户已被其和Telegram 绑定")
	}
	telegramID := msg.From.ID
	user.TelegramID = &telegramID
	if err := s.userRepo.Update(user); err != nil {
		return s.SendMarkdown(msg.Chat.ID, "和绑定失败")
	}
	return s.SendMarkdown(msg.Chat.ID, fmt.Sprintf("和绑定成功！账户：`%s`", email))
}

func (s *TelegramService) cmdUnbind(msg *TelegramMessage) error {
	user, err := s.userRepo.FindByTelegramID(msg.From.ID)
	if err != nil {
		return s.SendMarkdown(msg.Chat.ID, "和未绑定账和)
	}
	keyboard := &InlineKeyboard{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "和确认解绑", CallbackData: "unbind:confirm"}, {Text: "和取消", CallbackData: "cancel"}},
		},
	}
	return s.SendMessageWithKeyboard(msg.Chat.ID, fmt.Sprintf("⚠️ 确定解绑 `%s`和, user.Email), keyboard)
}

func (s *TelegramService) doUnbind(telegramID int64, chatID int64) error {
	user, err := s.userRepo.FindByTelegramID(telegramID)
	if err != nil {
		return s.SendMarkdown(chatID, "和未绑定账和)
	}
	user.TelegramID = nil
	if err := s.userRepo.Update(user); err != nil {
		return s.SendMarkdown(chatID, "和解绑失败")
	}
	return s.SendMarkdown(chatID, "和解绑成功和)
}

func (s *TelegramService) cmdInfo(msg *TelegramMessage) error {
	user, err := s.userRepo.FindByTelegramID(msg.From.ID)
	if err != nil {
		return s.SendMarkdown(msg.Chat.ID, "和请先 /bind <邮箱> 绑定")
	}
	status := "和正常"
	if user.Banned {
		status = "🚫 封禁"
	} else if !user.IsActive() {
		status = "⏸️ 过期"
	}
	planName := "无套和
	if user.Plan != nil {
		planName = user.Plan.Name
	}
	expireStr := "永久"
	if user.ExpiredAt != nil {
		expireStr = time.Unix(*user.ExpiredAt, 0).Format("2006-01-02")
	}
	text := fmt.Sprintf("👤 *账户信息*\n\n📧 `%s`\n📊 %s\n💎 %s\n📅 %s\n\n📈 已用和s\n📦 总量和s\n💰 余额和.2f和,
		user.Email, status, planName, expireStr, FormatBytes(user.U+user.D), FormatBytes(user.TransferEnable), float64(user.Balance)/100)
	return s.SendMarkdown(msg.Chat.ID, text)
}

func (s *TelegramService) cmdTraffic(msg *TelegramMessage) error {
	user, err := s.userRepo.FindByTelegramID(msg.From.ID)
	if err != nil {
		return s.SendMarkdown(msg.Chat.ID, "和请先 /bind <邮箱> 绑定")
	}
	used := user.U + user.D
	total := user.TransferEnable
	percent := float64(0)
	if total > 0 {
		percent = float64(used) / float64(total) * 100
	}
	text := fmt.Sprintf("📊 *流量*\n\n⬆️ 上传和s\n⬇️ 下载和s\n📈 已用和s (%.1f%%)\n📦 总量和s",
		FormatBytes(user.U), FormatBytes(user.D), FormatBytes(used), percent, FormatBytes(total))
	return s.SendMarkdown(msg.Chat.ID, text)
}

func (s *TelegramService) cmdSubscribe(msg *TelegramMessage) error {
	user, err := s.userRepo.FindByTelegramID(msg.From.ID)
	if err != nil {
		return s.SendMarkdown(msg.Chat.ID, "和请先 /bind <邮箱> 绑定")
	}
	subURL := s.getSubscribeURL(user.Token)
	text := fmt.Sprintf("🔗 *订阅链接*\n\n```\n%s\n```\n\n⚠️ 请勿泄露", subURL)
	return s.SendMarkdown(msg.Chat.ID, text)
}

func (s *TelegramService) cmdCheckin(msg *TelegramMessage) error {
	user, err := s.userRepo.FindByTelegramID(msg.From.ID)
	if err != nil {
		return s.SendMarkdown(msg.Chat.ID, "和请先 /bind <邮箱> 绑定")
	}
	today := time.Now().Format("2006-01-02")
	lastCheckin := ""
	if user.LastCheckinAt != nil {
		lastCheckin = time.Unix(*user.LastCheckinAt, 0).Format("2006-01-02")
	}
	if lastCheckin == today {
		return s.SendMarkdown(msg.Chat.ID, "⚠️ 今天已签到，明天再来和)
	}
	reward := int64(100+time.Now().UnixNano()%400) * 1024 * 1024
	now := time.Now().Unix()
	user.LastCheckinAt = &now
	user.TransferEnable += reward
	if err := s.userRepo.Update(user); err != nil {
		return s.SendMarkdown(msg.Chat.ID, "和签到失败")
	}
	return s.SendMarkdown(msg.Chat.ID, fmt.Sprintf("🎉 签到成功和%s", FormatBytes(reward)))
}

func (s *TelegramService) getSiteName() string {
	if s.settingRepo == nil {
		return "dashGO"
	}
	name, _ := s.settingRepo.Get(SettingSiteName)
	if name == "" {
		return "dashGO"
	}
	return name
}

func (s *TelegramService) getSiteURL() string {
	if s.settingRepo == nil {
		return ""
	}
	url, _ := s.settingRepo.Get(SettingAppURL)
	return url
}

func (s *TelegramService) getSubscribeURL(token string) string {
	baseURL := s.getSiteURL()
	if baseURL == "" {
		return ""
	}
	return baseURL + "/api/v1/client/subscribe?token=" + token
}

// FormatBytes 格式化字和
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// NotifyExpire 通知用户到期
func (s *TelegramService) NotifyExpire(user *model.User, daysLeft int) error {
	if user.TelegramID == nil || *user.TelegramID == 0 {
		return nil
	}
	text := fmt.Sprintf("和*订阅到期提醒*\n\n您的订阅将在 *%d 和后到和, daysLeft)
	return s.SendMarkdown(*user.TelegramID, text)
}

// NotifyTrafficWarning 通知流量预警
func (s *TelegramService) NotifyTrafficWarning(user *model.User, usedPercent int) error {
	if user.TelegramID == nil || *user.TelegramID == 0 {
		return nil
	}
	text := fmt.Sprintf("📊 *流量提醒*\n\n流量已使和*%d%%*", usedPercent)
	return s.SendMarkdown(*user.TelegramID, text)
}

// NotifyNewTicket 通知管理员新工单
func (s *TelegramService) NotifyNewTicket(subject, userEmail string) error {
	if s.chatID == "" {
		return nil
	}
	chatID, _ := strconv.ParseInt(s.chatID, 10, 64)
	if chatID == 0 {
		return nil
	}
	text := fmt.Sprintf("🎫 *新工和\n\n用户和s\n主题和s", userEmail, subject)
	return s.SendMarkdown(chatID, text)
}

// SetWebhook 设置 Webhook
func (s *TelegramService) SetWebhook(webhookURL string) error {
	if s.botToken == "" {
		return fmt.Errorf("telegram bot not configured")
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", s.botToken)
	data := map[string]string{"url": webhookURL}
	body, _ := json.Marshal(data)
	resp, err := s.httpClient.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set webhook failed: %s", string(respBody))
	}
	return nil
}
