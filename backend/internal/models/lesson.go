
package models

import (
	"time"
)

type Lesson struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CourseID  uint      `gorm:"not null;index" json:"courseId"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Type      string    `gorm:"size:20;not null" json:"type"`
	Content   string    `gorm:"type:text" json:"content"`
	AudioURL  string    `gorm:"size:500" json:"audioUrl"`
	SortOrder int       `gorm:"default:0" json:"sortOrder"`
	IsFree    bool      `gorm:"default:false" json:"isFree"`
	XpReward  int       `gorm:"default:10" json:"xpReward"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
