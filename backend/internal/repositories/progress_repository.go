
package repositories

import (
	"chinese-learning-app/internal/models"
	"time"

	"gorm.io/gorm"
)

type ProgressRepository struct {
	db *gorm.DB
}

func NewProgressRepository(db *gorm.DB) *ProgressRepository {
	return &ProgressRepository{
		db: db,
	}
}

func (r *ProgressRepository) GetByUserAndLesson(userID, lessonID uint) (*models.UserProgress, error) {
	var progress models.UserProgress
	err := r.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).First(&progress).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &progress, nil
}

func (r *ProgressRepository) GetAllByUser(userID uint) ([]models.UserProgress, error) {
	var progress []models.UserProgress
	err := r.db.Where("user_id = ?", userID).Order("lesson_id ASC").Find(&progress).Error
	return progress, err
}

func (r *ProgressRepository) CreateOrUpdate(progress *models.UserProgress) error {
	existing, err := r.GetByUserAndLesson(progress.UserID, progress.LessonID)
	if err != nil {
		return err
	}

	if existing != nil {
		if progress.Score > existing.Score {
			existing.Score = progress.Score
			existing.Status = progress.Status
			existing.UpdatedAt = time.Now()
			if progress.Status == "completed" || progress.Status == "perfected" {
				now := time.Now()
				existing.CompletedAt = &now
			}
		} else {
			existing.UpdatedAt = time.Now()
		}
		existing.Attempts++
		return r.db.Save(existing).Error
	}

	return r.db.Create(progress).Error
}

func (r *ProgressRepository) GetStats(userID uint) (map[string]interface{}, error) {
	progressList, err := r.GetAllByUser(userID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"totalLessons": 0,
		"completed":    0,
		"totalXP":      0,
	}

	for _, p := range progressList {
		stats["totalLessons"] = stats["totalLessons"].(int) + 1
		if p.Status == "completed" || p.Status == "perfected" {
			stats["completed"] = stats["completed"].(int) + 1
		}
	}

	return stats, nil
}
