package database

import (
	"fmt"
	"os"
	"path/filepath"

	"dashgo/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
		dialector = mysql.Open(dsn)
	case "sqlite":
		// Ensure the directory exists for SQLite database file
		if err := ensureSQLiteDir(cfg.Database); err != nil {
			return nil, fmt.Errorf("failed to create SQLite directory: %w", err)
		}
		dialector = sqlite.Open(cfg.Database)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), // 只打印警告和错误，不打印普通SQL
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// ensureSQLiteDir ensures the directory for SQLite database file exists
func ensureSQLiteDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "." || dir == "" {
		return nil // Current directory, no need to create
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Create directory with 0755 permissions
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
