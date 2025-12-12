package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Node     NodeConfig     `yaml:"node"`
	Mail     MailConfig     `yaml:"mail"`
	Telegram TelegramConfig `yaml:"telegram"`
	Admin    AdminConfig    `yaml:"admin"`
}

type AdminConfig struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

type AppConfig struct {
	Name   string `yaml:"name"`
	Mode   string `yaml:"mode"` // debug, release
	Listen string `yaml:"listen"`
}

type DatabaseConfig struct {
	Driver   string `yaml:"driver"` // mysql, sqlite
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpireHour int    `yaml:"expire_hour"`
}

type NodeConfig struct {
	Token        string `yaml:"token"`         // Node communication token
	PushInterval int    `yaml:"push_interval"` // seconds
	PullInterval int    `yaml:"pull_interval"` // seconds
	EnableSync   bool   `yaml:"enable_sync"`   // 是否启用主动节点同步（默�?false，使�?Agent 模式�?
}

type MailConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	FromName   string `yaml:"from_name"`
	FromAddr   string `yaml:"from_addr"`
	Encryption string `yaml:"encryption"` // ssl, tls, none
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.App.Listen == "" {
		cfg.App.Listen = ":8080"
	}
	if cfg.JWT.ExpireHour == 0 {
		cfg.JWT.ExpireHour = 24
	}
	if cfg.Node.PushInterval == 0 {
		cfg.Node.PushInterval = 60
	}
	if cfg.Node.PullInterval == 0 {
		cfg.Node.PullInterval = 60
	}

	return &cfg, nil
}
