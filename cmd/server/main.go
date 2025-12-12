package main

import (
	"flag"
	"log"
	"time"

	"dashgo/internal/config"
	"dashgo/internal/handler"
	"dashgo/internal/model"
	"dashgo/internal/repository"
	"dashgo/internal/service"
	"dashgo/pkg/cache"
	"dashgo/pkg/database"
	"dashgo/pkg/utils"

	"github.com/gin-gonic/gin"
)

// initAdminUser 初始化管理员用户
func initAdminUser(userRepo *repository.UserRepository, email, password string) error {
	// 检查是否已存在
	existing, _ := userRepo.FindByEmail(email)
	if existing != nil {
		// 已存在，确保是管理员
		if !existing.IsAdmin {
			existing.IsAdmin = true
			return userRepo.Update(existing)
		}
		return nil
	}

	// 创建新管理员
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	admin := &model.User{
		Email:     email,
		Password:  hashedPassword,
		UUID:      utils.GenerateUUID(),
		Token:     utils.GenerateToken(32),
		IsAdmin:   true,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	return userRepo.Create(admin)
}

func main() {
	configPath := flag.String("config", "configs/config.yaml", "config file path")
	flag.Parse()

	// Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(
		&model.User{},
		&model.Plan{},
		&model.Server{},
		&model.Order{},
		&model.Setting{},
		&model.Stat{},
		&model.StatUser{},
		&model.StatServer{},
		&model.ServerLog{},
		&model.Ticket{},
		&model.TicketMessage{},
		&model.Payment{},
		&model.Coupon{},
		&model.InviteCode{},
		&model.CommissionLog{},
		&model.Notice{},
		&model.Knowledge{},
		&model.Host{},
		&model.ServerNode{},
		&model.ServerGroup{},
		&model.UserGroup{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize cache
	cacheClient, err := cache.New(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect redis: %v", err)
	}

	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Initialize services
	services := service.NewServices(repos, cacheClient, cfg)

	// Initialize admin user if configured
	if cfg.Admin.Email != "" && cfg.Admin.Password != "" {
		if err := initAdminUser(repos.User, cfg.Admin.Email, cfg.Admin.Password); err != nil {
			log.Printf("Warning: Failed to init admin user: %v", err)
		} else {
			log.Printf("Admin user initialized: %s", cfg.Admin.Email)
		}
	}

	// Start scheduler
	go services.Scheduler.Start()
	log.Println("Scheduler started")

	// Start node sync service
	go services.NodeSync.StartSyncLoop()
	log.Println("Node sync service started")

	// Initialize HTTP server
	if cfg.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	handler.RegisterRoutes(r, services, cfg)

	log.Printf("Server starting on %s", cfg.App.Listen)
	if err := r.Run(cfg.App.Listen); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
