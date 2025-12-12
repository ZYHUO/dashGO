package service

import (
	"encoding/json"
	"strconv"
	"time"

	"dashgo/internal/repository"
	"dashgo/pkg/cache"
)

type SettingService struct {
	settingRepo *repository.SettingRepository
	cache       *cache.Client
}

func NewSettingService(settingRepo *repository.SettingRepository, cache *cache.Client) *SettingService {
	return &SettingService{
		settingRepo: settingRepo,
		cache:       cache,
	}
}

// Get 获取设置
func (s *SettingService) Get(key string) (string, error) {
	// 先从缓存获取
	cacheKey := "setting:" + key
	if val, err := s.cache.Get(cacheKey); err == nil {
		return val, nil
	}

	// 从数据库获取
	val, err := s.settingRepo.Get(key)
	if err != nil {
		return "", err
	}

	// 写入缓存
	s.cache.Set(cacheKey, val, time.Hour)
	return val, nil
}

// Set 设置
func (s *SettingService) Set(key, value string) error {
	if err := s.settingRepo.Set(key, value); err != nil {
		return err
	}

	// 更新缓存
	cacheKey := "setting:" + key
	return s.cache.Set(cacheKey, value, time.Hour)
}

// GetInt 获取整数设置
func (s *SettingService) GetInt(key string, defaultVal int) int {
	val, err := s.Get(key)
	if err != nil {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}

// GetBool 获取布尔设置
func (s *SettingService) GetBool(key string, defaultVal bool) bool {
	val, err := s.Get(key)
	if err != nil {
		return defaultVal
	}
	return val == "1" || val == "true"
}

// GetAll 获取所有设�?
func (s *SettingService) GetAll() (map[string]string, error) {
	return s.settingRepo.GetAll()
}

// 常用设置 key
const (
	SettingAppName            = "app_name"
	SettingAppURL             = "app_url"
	SettingServerPushInterval = "server_push_interval"
	SettingServerPullInterval = "server_pull_interval"
	SettingSubscribeURL       = "subscribe_url"

	// 站点设置
	SettingSiteName        = "site_name"
	SettingSiteLogo        = "site_logo"
	SettingSiteDescription = "site_description"
	SettingSiteKeywords    = "site_keywords"
	SettingSiteTheme       = "site_theme"
	SettingSitePrimaryColor = "site_primary_color"
	SettingSiteFavicon     = "site_favicon"
	SettingSiteFooter      = "site_footer"
	SettingSiteTOS         = "site_tos"
	SettingSitePrivacy     = "site_privacy"

	// 注册设置
	SettingRegisterEnable       = "register_enable"
	SettingRegisterInviteOnly   = "register_invite_only"
	SettingRegisterTrial        = "register_trial"
	SettingRegisterTrialDays    = "register_trial_days"
	SettingRegisterTrialTraffic = "register_trial_traffic"
	SettingRegisterIPLimit      = "register_ip_limit"

	// 邮件设置
	SettingMailEnable = "mail_enable"
	SettingMailVerify = "mail_verify"

	// Telegram 设置
	SettingTelegramEnable   = "telegram_enable"
	SettingTelegramBotToken = "telegram_bot_token"
	SettingTelegramChatID   = "telegram_chat_id"

	// 支付设置
	SettingPaymentCurrency = "payment_currency"
	SettingPaymentSymbol   = "payment_symbol"
)

// SiteSettings 站点设置结构
type SiteSettings struct {
	Name         string `json:"name"`
	Logo         string `json:"logo"`
	Description  string `json:"description"`
	Keywords     string `json:"keywords"`
	Theme        string `json:"theme"`
	PrimaryColor string `json:"primary_color"`
	Favicon      string `json:"favicon"`
	Footer       string `json:"footer"`
	TOS          string `json:"tos"`
	Privacy      string `json:"privacy"`
	Currency     string `json:"currency"`
	CurrencySymbol string `json:"currency_symbol"`
}

// GetSiteSettings 获取站点设置
func (s *SettingService) GetSiteSettings() (*SiteSettings, error) {
	// 尝试从缓存获�?
	var settings SiteSettings
	cacheKey := cache.KeySiteSettings
	if val, err := s.cache.Get(cacheKey); err == nil {
		if err := json.Unmarshal([]byte(val), &settings); err == nil {
			return &settings, nil
		}
	}

	// 从数据库获取
	settings = SiteSettings{
		Name:         s.GetString(SettingSiteName, "XBoard"),
		Logo:         s.GetString(SettingSiteLogo, ""),
		Description:  s.GetString(SettingSiteDescription, ""),
		Keywords:     s.GetString(SettingSiteKeywords, ""),
		Theme:        s.GetString(SettingSiteTheme, "default"),
		PrimaryColor: s.GetString(SettingSitePrimaryColor, "#6366f1"),
		Favicon:      s.GetString(SettingSiteFavicon, ""),
		Footer:       s.GetString(SettingSiteFooter, ""),
		TOS:          s.GetString(SettingSiteTOS, ""),
		Privacy:      s.GetString(SettingSitePrivacy, ""),
		Currency:     s.GetString(SettingPaymentCurrency, "CNY"),
		CurrencySymbol: s.GetString(SettingPaymentSymbol, "¥"),
	}

	// 写入缓存
	data, _ := json.Marshal(settings)
	s.cache.Set(cacheKey, string(data), time.Hour)

	return &settings, nil
}

// SetSiteSettings 设置站点设置
func (s *SettingService) SetSiteSettings(settings *SiteSettings) error {
	pairs := map[string]string{
		SettingSiteName:        settings.Name,
		SettingSiteLogo:        settings.Logo,
		SettingSiteDescription: settings.Description,
		SettingSiteKeywords:    settings.Keywords,
		SettingSiteTheme:       settings.Theme,
		SettingSitePrimaryColor: settings.PrimaryColor,
		SettingSiteFavicon:     settings.Favicon,
		SettingSiteFooter:      settings.Footer,
		SettingSiteTOS:         settings.TOS,
		SettingSitePrivacy:     settings.Privacy,
		SettingPaymentCurrency: settings.Currency,
		SettingPaymentSymbol:   settings.CurrencySymbol,
	}

	for key, value := range pairs {
		if err := s.Set(key, value); err != nil {
			return err
		}
	}

	// 清除缓存
	s.cache.Del(cache.KeySiteSettings)
	return nil
}

// GetString 获取字符串设�?
func (s *SettingService) GetString(key, defaultVal string) string {
	val, err := s.Get(key)
	if err != nil || val == "" {
		return defaultVal
	}
	return val
}

// SetMultiple 批量设置
func (s *SettingService) SetMultiple(settings map[string]string) error {
	for key, value := range settings {
		if err := s.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}

// GetPublicSettings 获取公开设置（前端使用）
func (s *SettingService) GetPublicSettings() map[string]interface{} {
	site, _ := s.GetSiteSettings()
	return map[string]interface{}{
		"site_name":        site.Name,
		"site_logo":        site.Logo,
		"site_description": site.Description,
		"site_theme":       site.Theme,
		"primary_color":    site.PrimaryColor,
		"favicon":          site.Favicon,
		"footer":           site.Footer,
		"currency":         site.Currency,
		"currency_symbol":  site.CurrencySymbol,
		"register_enable":  s.GetBool(SettingRegisterEnable, true),
		"invite_only":      s.GetBool(SettingRegisterInviteOnly, false),
		"telegram_enable":  s.GetBool(SettingTelegramEnable, false),
		"mail_verify":      s.GetBool(SettingMailVerify, false),
	}
}
