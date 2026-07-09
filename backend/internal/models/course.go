
package models

import (
	"time"
)

type Course struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Level       string    `gorm:"size:20;not null" json:"level"`
	LevelName   string    `gorm:"size:50" json:"levelName"`
	Thumbnail   string    `gorm:"size:500" json:"thumbnail"`
	SortOrder   int       `gorm:"default:0" json:"sortOrder"`
	IsPublished bool      `gorm:"default:true" json:"isPublished"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
