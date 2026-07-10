package database

import (
	"chinese-learning-app/internal/models"
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	var err error
	dbPath := os.Getenv("SQLITE_PATH")
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

	log.Println("Database connected successfully")

	err = DB.AutoMigrate(
		&models.User{},
		&models.Course{},
		&models.Lesson{},
		&models.UserProgress{},
		&models.PaymentOrder{},
		&models.PaymentSubscription{},
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
