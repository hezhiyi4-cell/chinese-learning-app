
package models

import (
	"time"
)

type UserProgress struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"userId"`
	LessonID    uint      `gorm:"not null;index" json:"lessonId"`
	Status      string    `gorm:"size:20;default:'pending'" json:"status"`
	Score       int       `gorm:"default:0" json:"score"`
	Attempts    int       `gorm:"default:0" json:"attempts"`
	CompletedAt *time.Time `json:"completedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
