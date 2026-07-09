
package services

import (
	"chinese-learning-app/internal/models"
	"chinese-learning-app/internal/repositories"
	"time"
)

type ProgressResponse struct {
	LessonID  uint      `json:"lessonId"`
	Status    string    `json:"status"`
	Score     int       `json:"score"`
	Attempts  int       `json:"attempts"`
	Completed *time.Time `json:"completedAt"`
}

type StatsResponse struct {
	TotalLessons int    `json:"totalLessons"`
	Completed    int    `json:"completed"`
	TotalXP      int    `json:"totalXP"`
	Rank         string `json:"rank"`
	RankProgress int    `json:"rankProgress"`
	CourseProgress []CourseProgressItem `json:"courseProgress"`
}

type CourseProgressItem struct {
	CourseID  uint `json:"courseId"`
	Completed int  `json:"completed"`
	Total     int  `json:"total"`
}

type ProgressService struct {
	progressRepo *repositories.ProgressRepository
	userRepo     *repositories.UserRepository
	courseRepo   *repositories.CourseRepository
}

func NewProgressService(progressRepo *repositories.ProgressRepository, userRepo *repositories.UserRepository, courseRepo *repositories.CourseRepository) *ProgressService {
	return &ProgressService{
		progressRepo: progressRepo,
		userRepo: userRepo,
		courseRepo: courseRepo,
	}
}

func (s *ProgressService) GetUserProgress(userID uint) ([]ProgressResponse, error) {
	progressList, err := s.progressRepo.GetAllByUser(userID)
	if err != nil {
		return nil, err
	}

	var result []ProgressResponse
	for _, p := range progressList {
		result = append(result, ProgressResponse{
			LessonID:  p.LessonID,
			Status:    p.Status,
			Score:     p.Score,
			Attempts:  p.Attempts,
			Completed: p.CompletedAt,
		})
	}
	return result, nil
}

func (s *ProgressService) UpdateProgress(userID, lessonID uint, score int) (*models.UserProgress, error) {
	status := "in_progress"
	if score >= 90 {
		status = "perfected"
	} else if score >= 70 {
		status = "completed"
	}

	progress := &models.UserProgress{
		UserID:    userID,
		LessonID:  lessonID,
		Status:    status,
		Score:     score,
		Attempts:  1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if status == "completed" || status == "perfected" {
		now := time.Now()
		progress.CompletedAt = &now
	}

	err := s.progressRepo.CreateOrUpdate(progress)
	if err != nil {
		return nil, err
	}

	return progress, nil
}

func (s *ProgressService) GetUserStats(userID uint) (*StatsResponse, error) {
	progressList, err := s.progressRepo.GetAllByUser(userID)
	if err != nil {
		return nil, err
	}

	statusByLessonID := map[uint]string{}
	for _, p := range progressList {
		statusByLessonID[p.LessonID] = p.Status
	}

	courses, err := s.courseRepo.GetAll("")
	if err != nil {
		return nil, err
	}

	totalLessons := 0
	completed := 0
	var courseProgress []CourseProgressItem

	for _, course := range courses {
		lessons, err := s.courseRepo.GetLessons(course.ID)
		if err != nil {
			return nil, err
		}
		total := len(lessons)
		totalLessons += total

		c := 0
		for _, lesson := range lessons {
			st := statusByLessonID[lesson.ID]
			if st == "completed" || st == "perfected" {
				c++
			}
		}
		completed += c
		courseProgress = append(courseProgress, CourseProgressItem{
			CourseID:  course.ID,
			Completed: c,
			Total:     total,
		})
	}

	user, _ := s.userRepo.FindByID(userID)
	rank := "青铜"
	if user != nil {
		rank = user.Rank
	}

	return &StatsResponse{
		TotalLessons: totalLessons,
		Completed:    completed,
		TotalXP:      completed * 10,
		Rank:         rank,
		RankProgress: (completed % 10) * 10,
		CourseProgress: courseProgress,
	}, nil
}
