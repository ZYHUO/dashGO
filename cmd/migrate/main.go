package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"xboard/internal/config"
	"xboard/internal/model"
	"xboard/pkg/database"

	"gorm.io/gorm"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	migrationsDir := flag.String("migrations", "migrations", "迁移文件目录")
	action := flag.String("action", "up", "操作: up=执行迁移, status=查看状态, auto=自动迁移模型")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 连接数据库
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 确保迁移记录表存在
	db.AutoMigrate(&Migration{})

	switch *action {
	case "up":
		runMigrations(db, *migrationsDir)
	case "status":
		showStatus(db, *migrationsDir)
	case "auto":
		autoMigrate(db)
	default:
		fmt.Println("用法: migrate -action [up|status|auto]")
		fmt.Println("  up     - 执行 SQL 迁移文件")
		fmt.Println("  status - 查看迁移状态")
		fmt.Println("  auto   - 自动迁移模型结构")
	}
}

// Migration 迁移记录
type Migration struct {
	ID        int64  `gorm:"primaryKey"`
	Name      string `gorm:"size:255;uniqueIndex"`
	ExecutedAt int64 `gorm:"autoCreateTime"`
}

func (Migration) TableName() string {
	return "migrations"
}

// runMigrations 执行迁移
func runMigrations(db *gorm.DB, dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("读取迁移目录失败: %v", err)
	}

	// 过滤并排序 SQL 文件
	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	// 获取已执行的迁移
	var executed []Migration
	db.Find(&executed)
	executedMap := make(map[string]bool)
	for _, m := range executed {
		executedMap[m.Name] = true
	}

	// 执行未执行的迁移
	count := 0
	for _, name := range sqlFiles {
		if executedMap[name] {
			continue
		}

		fmt.Printf("执行迁移: %s ... ", name)

		// 读取 SQL 文件
		content, err := ioutil.ReadFile(filepath.Join(dir, name))
		if err != nil {
			fmt.Printf("失败: %v\n", err)
			continue
		}

		// 分割并执行 SQL 语句
		statements := splitSQL(string(content))
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" || strings.HasPrefix(stmt, "--") {
				continue
			}
			if err := db.Exec(stmt).Error; err != nil {
				// 忽略某些错误（如字段已存在）
				if !strings.Contains(err.Error(), "Duplicate") && 
				   !strings.Contains(err.Error(), "already exists") {
					fmt.Printf("警告: %v\n", err)
				}
			}
		}

		// 记录迁移
		db.Create(&Migration{Name: name})
		fmt.Println("完成")
		count++
	}

	if count == 0 {
		fmt.Println("没有需要执行的迁移")
	} else {
		fmt.Printf("成功执行 %d 个迁移\n", count)
	}
}

// showStatus 显示迁移状态
func showStatus(db *gorm.DB, dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("读取迁移目录失败: %v", err)
	}

	var executed []Migration
	db.Find(&executed)
	executedMap := make(map[string]bool)
	for _, m := range executed {
		executedMap[m.Name] = true
	}

	fmt.Println("迁移状态:")
	fmt.Println("----------------------------------------")
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			status := "[ ] 待执行"
			if executedMap[f.Name()] {
				status = "[✓] 已执行"
			}
			fmt.Printf("%s  %s\n", status, f.Name())
		}
	}
}

// autoMigrate 自动迁移模型
func autoMigrate(db *gorm.DB) {
	fmt.Println("自动迁移模型结构...")

	models := []interface{}{
		&model.User{},
		&model.Plan{},
		&model.Server{},
		&model.Order{},
		&model.Setting{},
		&model.Stat{},
		&model.StatUser{},
		&model.StatServer{},
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
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			fmt.Printf("迁移 %T 失败: %v\n", m, err)
		} else {
			fmt.Printf("迁移 %T 成功\n", m)
		}
	}

	fmt.Println("自动迁移完成")
}

// splitSQL 分割 SQL 语句
func splitSQL(content string) []string {
	var statements []string
	var current strings.Builder
	
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		current.WriteString(line)
		current.WriteString(" ")
		if strings.HasSuffix(line, ";") {
			statements = append(statements, current.String())
			current.Reset()
		}
	}
	
	if current.Len() > 0 {
		statements = append(statements, current.String())
	}
	
	return statements
}
