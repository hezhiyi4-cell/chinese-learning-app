package repositories

import (
	"chinese-learning-app/internal/models"

	"gorm.io/gorm"
)

type ToneBattleRepository struct {
	db *gorm.DB
}

func NewToneBattleRepository(db *gorm.DB) *ToneBattleRepository {
	return &ToneBattleRepository{db: db}
}

func (r *ToneBattleRepository) CountQuestions() (int64, error) {
	var count int64
	err := r.db.Model(&models.ToneBattleQuestion{}).Count(&count).Error
	return count, err
}

func (r *ToneBattleRepository) CreateQuestions(questions []models.ToneBattleQuestion) error {
	if len(questions) == 0 {
		return nil
	}
	return r.db.Create(&questions).Error
}

func (r *ToneBattleRepository) ListQuestions(limit int) ([]models.ToneBattleQuestion, error) {
	var questions []models.ToneBattleQuestion
	query := r.db.Model(&models.ToneBattleQuestion{}).Order("tone ASC").Order("syllable ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&questions).Error
	return questions, err
}

func (r *ToneBattleRepository) RandomQuestion(excludeIDs []uint) (*models.ToneBattleQuestion, error) {
	var question models.ToneBattleQuestion
	query := r.db.Model(&models.ToneBattleQuestion{})
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	err := query.Order("RANDOM()").First(&question).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &question, nil
}

func (r *ToneBattleRepository) CreateMatch(match *models.ToneBattleMatch) error {
	return r.db.Create(match).Error
}

func (r *ToneBattleRepository) UpdateMatch(match *models.ToneBattleMatch) error {
	return r.db.Save(match).Error
}
