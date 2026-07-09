
package models

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Nickname     string    `gorm:"size:50" json:"nickname"`
	Role         string    `gorm:"size:20;default:'user'" json:"role"`
	Rank         string    `gorm:"default:'青铜'" json:"rank"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
