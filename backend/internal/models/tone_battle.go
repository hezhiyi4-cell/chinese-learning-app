package models

import "time"

type ToneBattleQuestion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Hanzi     string    `gorm:"size:32;not null" json:"hanzi"`
	Syllable  string    `gorm:"size:32;index;not null" json:"syllable"`
	Pinyin    string    `gorm:"size:32;not null" json:"pinyin"`
	Tone      int       `gorm:"index;not null" json:"tone"`
	AudioPath string    `gorm:"size:255" json:"audioPath"`
	Category  string    `gorm:"size:32;default:'foundation'" json:"category"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ToneBattleMatch struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	RoomID         string     `gorm:"size:64;uniqueIndex;not null" json:"roomId"`
	PlayerOneID    uint       `gorm:"index;not null" json:"playerOneId"`
	PlayerTwoID    uint       `gorm:"index;not null" json:"playerTwoId"`
	PlayerOneScore int        `gorm:"default:0" json:"playerOneScore"`
	PlayerTwoScore int        `gorm:"default:0" json:"playerTwoScore"`
	WinnerID       *uint      `gorm:"index" json:"winnerId"`
	Status         string     `gorm:"size:20;default:'pending'" json:"status"`
	RoundsPlayed   int        `gorm:"default:0" json:"roundsPlayed"`
	StartedAt      *time.Time `json:"startedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}
