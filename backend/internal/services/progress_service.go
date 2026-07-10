package services

import (
	"chinese-learning-app/internal/models"
	"chinese-learning-app/internal/repositories"
	"math"
	"time"
)

const (
	lessonCompleteXP = 10
	perfectBonusXP   = 5
	streakBonusXP    = 5
)

type ProgressResponse struct {
	LessonID  uint       `json:"lessonId"`
	Status    string     `json:"status"`
	Score     int        `json:"score"`
	Attempts  int        `json:"attempts"`
	Completed *time.Time `json:"completedAt"`
}

type RewardItem struct {
	Type   string `json:"type"`
	Label  string `json:"label"`
	Points int    `json:"points"`
}

type StatsResponse struct {
	TotalLessons     int                  `json:"totalLessons"`
	Completed        int                  `json:"completed"`
	TotalXP          int                  `json:"totalXP"`
	Rank             string               `json:"rank"`
	RankProgress     int                  `json:"rankProgress"`
	CurrentStreak    int                  `json:"currentStreak"`
	NextRank         string               `json:"nextRank"`
	CurrentRankMinXP int                  `json:"currentRankMinXp"`
	NextRankMinXP    int                  `json:"nextRankMinXp"`
	XPToNextRank     int                  `json:"xpToNextRank"`
	CourseProgress   []CourseProgressItem `json:"courseProgress"`
}

type CourseProgressItem struct {
	CourseID  uint `json:"courseId"`
	Completed int  `json:"completed"`
	Perfected int  `json:"perfected"`
	Total     int  `json:"total"`
	EarnedXP  int  `json:"earnedXp"`
	TotalXP   int  `json:"totalXp"`
}

type ProgressUpdateResponse struct {
	Progress      *models.UserProgress `json:"progress"`
	Rewards       []RewardItem         `json:"rewards"`
	TotalXP       int                  `json:"totalXp"`
	Rank          string               `json:"rank"`
	PreviousRank  string               `json:"previousRank"`
	RankUp        bool                 `json:"rankUp"`
	CurrentStreak int                  `json:"currentStreak"`
	Stats         *StatsResponse       `json:"stats"`
}

type rankTier struct {
	Name  string
	MinXP int
}

var rankTiers = []rankTier{
	{Name: "青铜", MinXP: 0},
	{Name: "白银", MinXP: 60},
	{Name: "黄金", MinXP: 140},
	{Name: "铂金", MinXP: 260},
	{Name: "钻石", MinXP: 420},
	{Name: "大师", MinXP: 620},
}

type ProgressService struct {
	progressRepo *repositories.ProgressRepository
	userRepo     *repositories.UserRepository
	courseRepo   *repositories.CourseRepository
}

func NewProgressService(progressRepo *repositories.ProgressRepository, userRepo *repositories.UserRepository, courseRepo *repositories.CourseRepository) *ProgressService {
	return &ProgressService{
		progressRepo: progressRepo,
		userRepo:     userRepo,
		courseRepo:   courseRepo,
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

func (s *ProgressService) UpdateProgress(userID, lessonID uint, score int) (*ProgressUpdateResponse, error) {
	existing, err := s.progressRepo.GetByUserAndLesson(userID, lessonID)
	if err != nil {
		return nil, err
	}
	priorProgressList, err := s.progressRepo.GetAllByUser(userID)
	if err != nil {
		return nil, err
	}

	oldCompleted := isCompletedStatus(existingStatus(existing))
	oldPerfected := existingStatus(existing) == "perfected"
	status := scoreToStatus(score)

	progress := &models.UserProgress{
		UserID:    userID,
		LessonID:  lessonID,
		Status:    status,
		Score:     score,
		Attempts:  1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if isCompletedStatus(status) {
		now := time.Now()
		progress.CompletedAt = &now
	}

	if err := s.progressRepo.CreateOrUpdate(progress); err != nil {
		return nil, err
	}

	savedProgress, err := s.progressRepo.GetByUserAndLesson(userID, lessonID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	rewards := make([]RewardItem, 0, 3)
	if isCompletedStatus(status) && !oldCompleted {
		rewards = append(rewards, RewardItem{Type: "lesson_complete", Label: "完成课时", Points: lessonCompleteXP})
	}
	if status == "perfected" && !oldPerfected {
		rewards = append(rewards, RewardItem{Type: "perfect_bonus", Label: "完美通关", Points: perfectBonusXP})
	}

	previousRank := "青铜"
	totalXP := 0
	currentStreak := 0
	rankUp := false

	if user != nil {
		baselineXP := deriveXPFromProgressList(priorProgressList)
		if user.TotalXP < baselineXP {
			user.TotalXP = baselineXP
		}
		totalXP = user.TotalXP
		currentStreak = user.CurrentStreak
		previousRank = currentRankName(user.TotalXP, user.Rank)

		// 仅在本次学习有实质进展时结算打卡与积分。
		if len(rewards) > 0 {
			now := time.Now()
			today := startOfDay(now)

			if user.LastCheckInAt == nil {
				user.CurrentStreak = 1
				user.LastCheckInAt = &today
			} else {
				lastDay := startOfDay(*user.LastCheckInAt)
				dayGap := int(today.Sub(lastDay).Hours() / 24)
				switch {
				case dayGap <= 0:
				case dayGap == 1:
					if user.CurrentStreak < 1 {
						user.CurrentStreak = 1
					}
					user.CurrentStreak++
					rewards = append(rewards, RewardItem{Type: "streak_bonus", Label: "连续打卡", Points: streakBonusXP})
					user.LastCheckInAt = &today
				default:
					user.CurrentStreak = 1
					user.LastCheckInAt = &today
				}
			}

			gainedXP := sumRewardPoints(rewards)
			if gainedXP > 0 {
				user.TotalXP += gainedXP
			}

			newRank := currentRankName(user.TotalXP, user.Rank)
			rankUp = previousRank != newRank
			user.Rank = newRank
			user.UpdatedAt = now
			if err := s.userRepo.Update(user); err != nil {
				return nil, err
			}

			totalXP = user.TotalXP
			currentStreak = user.CurrentStreak
		}
	}

	stats, err := s.GetUserStats(userID)
	if err != nil {
		return nil, err
	}
	if stats != nil {
		totalXP = stats.TotalXP
		currentStreak = stats.CurrentStreak
	}

	return &ProgressUpdateResponse{
		Progress:      savedProgress,
		Rewards:       rewards,
		TotalXP:       totalXP,
		Rank:          stats.Rank,
		PreviousRank:  previousRank,
		RankUp:        rankUp,
		CurrentStreak: currentStreak,
		Stats:         stats,
	}, nil
}

func (s *ProgressService) GetUserStats(userID uint) (*StatsResponse, error) {
	progressList, err := s.progressRepo.GetAllByUser(userID)
	if err != nil {
		return nil, err
	}

	progressByLessonID := map[uint]models.UserProgress{}
	for _, p := range progressList {
		progressByLessonID[p.LessonID] = p
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

		item := CourseProgressItem{
			CourseID: course.ID,
			Total:    len(lessons),
			TotalXP:  len(lessons) * (lessonCompleteXP + perfectBonusXP),
		}
		totalLessons += len(lessons)

		for _, lesson := range lessons {
			p, ok := progressByLessonID[lesson.ID]
			if !ok {
				continue
			}
			if isCompletedStatus(p.Status) {
				item.Completed++
				completed++
				item.EarnedXP += lessonCompleteXP
			}
			if p.Status == "perfected" {
				item.Perfected++
				item.EarnedXP += perfectBonusXP
			}
		}

		courseProgress = append(courseProgress, item)
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	derivedXP := deriveXPFromProgressList(progressList)
	totalXP := 0
	currentStreak := 0
	rank := "青铜"
	if user != nil {
		if user.TotalXP < derivedXP {
			user.TotalXP = derivedXP
		}
		totalXP = user.TotalXP
		currentStreak = user.CurrentStreak
		rank = currentRankName(user.TotalXP, user.Rank)
	}

	currentRankMinXP, nextRank, nextRankMinXP, rankProgress, xpToNextRank := rankProgressMeta(totalXP)

	return &StatsResponse{
		TotalLessons:     totalLessons,
		Completed:        completed,
		TotalXP:          totalXP,
		Rank:             rank,
		RankProgress:     rankProgress,
		CurrentStreak:    currentStreak,
		NextRank:         nextRank,
		CurrentRankMinXP: currentRankMinXP,
		NextRankMinXP:    nextRankMinXP,
		XPToNextRank:     xpToNextRank,
		CourseProgress:   courseProgress,
	}, nil
}

func scoreToStatus(score int) string {
	switch {
	case score >= 90:
		return "perfected"
	case score >= 70:
		return "completed"
	default:
		return "in_progress"
	}
}

func isCompletedStatus(status string) bool {
	return status == "completed" || status == "perfected"
}

func existingStatus(progress *models.UserProgress) string {
	if progress == nil {
		return ""
	}
	return progress.Status
}

func sumRewardPoints(rewards []RewardItem) int {
	total := 0
	for _, reward := range rewards {
		total += reward.Points
	}
	return total
}

func startOfDay(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func currentRankName(totalXP int, fallback string) string {
	rank := fallback
	for _, tier := range rankTiers {
		if totalXP >= tier.MinXP {
			rank = tier.Name
		}
	}
	if rank == "" {
		return "青铜"
	}
	return rank
}

func rankProgressMeta(totalXP int) (currentRankMinXP int, nextRank string, nextRankMinXP int, rankProgress int, xpToNextRank int) {
	currentIndex := 0
	for i, tier := range rankTiers {
		if totalXP >= tier.MinXP {
			currentIndex = i
		}
	}

	current := rankTiers[currentIndex]
	if currentIndex == len(rankTiers)-1 {
		return current.MinXP, current.Name, current.MinXP, 100, 0
	}

	next := rankTiers[currentIndex+1]
	span := next.MinXP - current.MinXP
	progress := 0
	if span > 0 {
		progress = int(math.Round(float64(totalXP-current.MinXP) / float64(span) * 100))
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	return current.MinXP, next.Name, next.MinXP, progress, maxInt(0, next.MinXP-totalXP)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func deriveXPFromProgressList(progressList []models.UserProgress) int {
	total := 0
	for _, p := range progressList {
		if isCompletedStatus(p.Status) {
			total += lessonCompleteXP
		}
		if p.Status == "perfected" {
			total += perfectBonusXP
		}
	}
	return total
}
