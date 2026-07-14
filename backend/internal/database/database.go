package database

import (
	"chinese-learning-app/internal/config"
	"chinese-learning-app/internal/models"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) error {
	var err error
	driver := "sqlite"
	if cfg != nil && strings.TrimSpace(cfg.DBDriver) != "" {
		driver = strings.ToLower(strings.TrimSpace(cfg.DBDriver))
	}
	if cfg != nil && strings.TrimSpace(cfg.DatabaseURL) != "" {
		driver = "postgres"
	}

	switch driver {
	case "postgres":
		dsn := ""
		if cfg != nil && strings.TrimSpace(cfg.DatabaseURL) != "" {
			dsn = cfg.DatabaseURL
		} else if cfg != nil {
			dsn = "host=" + cfg.DBHost +
				" port=" + cfg.DBPort +
				" user=" + cfg.DBUser +
				" password=" + cfg.DBPassword +
				" dbname=" + cfg.DBName +
				" sslmode=disable TimeZone=UTC"
		}
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return err
		}
	default:
		dbPath := os.Getenv("SQLITE_PATH")
		if cfg != nil && strings.TrimSpace(cfg.SQLitePath) != "" {
			dbPath = cfg.SQLitePath
		}
		if dbPath == "" {
			dbPath = "chinese_learning.db"
		}
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil && filepath.Dir(dbPath) != "." {
			return err
		}
		DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			return err
		}
	}

	log.Println("Database connected successfully")

	err = DB.AutoMigrate(
		&models.User{},
		&models.Course{},
		&models.Lesson{},
		&models.UserProgress{},
		&models.PaymentOrder{},
		&models.PaymentSubscription{},
		&models.ToneBattleQuestion{},
		&models.ToneBattleMatch{},
	)
	if err != nil {
		return err
	}

	log.Println("Database migrated successfully")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}
