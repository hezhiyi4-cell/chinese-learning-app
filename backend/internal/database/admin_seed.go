package database

import (
	"chinese-learning-app/internal/models"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func EnsureDefaultAdmin(db *gorm.DB, email, password string) error {
	var user models.User
	err := db.Where("email = ?", email).First(&user).Error
	if err == nil {
		if user.Role != "admin" {
			user.Role = "admin"
			user.UpdatedAt = time.Now()
			return db.Save(&user).Error
		}
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	admin := &models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Nickname:     "管理员",
		Role:         "admin",
		Rank:         "管理员",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return db.Create(admin).Error
}
