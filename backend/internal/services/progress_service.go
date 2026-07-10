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

type AchievementItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Category    string `json:"category"`
	Unlocked    bool   `json:"unlocked"`
	Progress    int    `json:"progress"`
	Target      int    `json:"target"`
}

type StatsResponse struct {
	TotalLessons     int                  `json:"totalLessons"`
	Completed        int                  `json:"completed"`
	PerfectedLessons int                  `json:"perfectedLessons"`
	CompletedCourses int                  `json:"completedCourses"`
	TotalXP          int                  `json:"totalXP"`
	Rank             string               `json:"rank"`
	RankProgress     int                  `json:"rankProgress"`
	CurrentStreak    int                  `json:"currentStreak"`
	BestStreak       int                  `json:"bestStreak"`
	NextRank         string               `json:"nextRank"`
	CurrentRankMinXP int                  `json:"currentRankMinXp"`
	NextRankMinXP    int                  `json:"nextRankMinXp"`
	XPToNextRank     int                  `json:"xpToNextRank"`
	CourseProgress   []CourseProgressItem `json:"courseProgress"`
	Achievements     []AchievementItem    `json:"achievements"`
}

type LeaderboardEntry struct {
	Position         int    `json:"position"`
	UserID           uint   `json:"userId"`
	Nickname         string `json:"nickname"`
	Rank             string `json:"rank"`
	TotalXP          int    `json:"totalXP"`
	Completed        int    `json:"completed"`
	PerfectedLessons int    `json:"perfectedLessons"`
	BestStreak       int    `json:"bestStreak"`
	IsCurrentUser    bool   `json:"isCurrentUser"`
}

type LeaderboardResponse struct {
	Scope           string             `json:"scope"`
	Title           string             `json:"title"`
	Note            string             `json:"note"`
	CurrentUserRank int                `json:"currentUserRank"`
	Entries         []LeaderboardEntry `json:"entries"`
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
	Progress        *models.UserProgress `json:"progress"`
	Rewards         []RewardItem         `json:"rewards"`
	NewAchievements []AchievementItem    `json:"newAchievements"`
	TotalXP         int                  `json:"totalXp"`
	Rank            string               `json:"rank"`
	PreviousRank    string               `json:"previousRank"`
	RankUp          bool                 `json:"rankUp"`
	CurrentStreak   int                  `json:"currentStreak"`
	Stats           *StatsResponse       `json:"stats"`
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
	updatedProgressList, err := s.progressRepo.GetAllByUser(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	var previousStats *StatsResponse
	if user != nil {
		previousUser := *user
		previousUser.TotalXP = maxInt(previousUser.TotalXP, deriveXPFromProgressList(priorProgressList))
		if previousUser.BestStreak < previousUser.CurrentStreak {
			previousUser.BestStreak = previousUser.CurrentStreak
		}
		previousStats, err = s.buildStatsFromSnapshot(priorProgressList, &previousUser)
		if err != nil {
			return nil, err
		}
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
		if user.BestStreak < user.CurrentStreak {
			user.BestStreak = user.CurrentStreak
		}
		totalXP = user.TotalXP
		currentStreak = user.CurrentStreak
		previousRank = currentRankName(user.TotalXP, user.Rank)

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

			if user.CurrentStreak > user.BestStreak {
				user.BestStreak = user.CurrentStreak
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

	stats, err := s.buildStatsFromSnapshot(updatedProgressList, user)
	if err != nil {
		return nil, err
	}
	if stats != nil {
		totalXP = stats.TotalXP
		currentStreak = stats.CurrentStreak
	}

	return &ProgressUpdateResponse{
		Progress:        savedProgress,
		Rewards:         rewards,
		NewAchievements: diffUnlockedAchievements(previousStats, stats),
		TotalXP:         totalXP,
		Rank:            stats.Rank,
		PreviousRank:    previousRank,
		RankUp:          rankUp,
		CurrentStreak:   currentStreak,
		Stats:           stats,
	}, nil
}

func (s *ProgressService) GetUserStats(userID uint) (*StatsResponse, error) {
	progressList, err := s.progressRepo.GetAllByUser(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	return s.buildStatsFromSnapshot(progressList, user)
}

func (s *ProgressService) GetLeaderboard(userID uint, scope string) (*LeaderboardResponse, error) {
	users, err := s.userRepo.ListNonAdminUsers()
	if err != nil {
		return nil, err
	}

	entries := make([]LeaderboardEntry, 0, len(users))
	currentUserRank := 0

	for _, user := range users {
		progressList, err := s.progressRepo.GetAllByUser(user.ID)
		if err != nil {
			return nil, err
		}

		stats, err := s.buildStatsFromSnapshot(progressList, &user)
		if err != nil {
			return nil, err
		}

		entry := LeaderboardEntry{
			UserID:           user.ID,
			Nickname:         displayName(user),
			Rank:             stats.Rank,
			TotalXP:          stats.TotalXP,
			Completed:        stats.Completed,
			PerfectedLessons: stats.PerfectedLessons,
			BestStreak:       stats.BestStreak,
			IsCurrentUser:    user.ID == userID,
		}
		entries = append(entries, entry)
	}

	sortLeaderboard(entries)
	for i := range entries {
		entries[i].Position = i + 1
		if entries[i].UserID == userID {
			currentUserRank = entries[i].Position
		}
	}

	normalizedScope := scope
	if normalizedScope != "friends" {
		normalizedScope = "global"
	}

	response := &LeaderboardResponse{
		Scope:           normalizedScope,
		CurrentUserRank: currentUserRank,
	}

	if normalizedScope == "friends" {
		response.Title = "好友榜"
		response.Note = "当前版本暂未接入真实好友关系，这里先展示与你积分接近的学习搭子榜。"
		response.Entries = selectNearbyEntries(entries, userID, 5)
		return response, nil
	}

	response.Title = "全球榜"
	response.Note = "按积分、最佳连击和学习完成度综合排序。"
	response.Entries = takeEntries(entries, 10)
	return response, nil
}

func (s *ProgressService) buildStatsFromSnapshot(progressList []models.UserProgress, user *models.User) (*StatsResponse, error) {
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
	perfectedLessons := 0
	completedCourses := 0
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
				perfectedLessons++
				item.EarnedXP += perfectBonusXP
			}
		}

		if item.Total > 0 && item.Completed == item.Total {
			completedCourses++
		}
		courseProgress = append(courseProgress, item)
	}

	derivedXP := deriveXPFromProgressList(progressList)
	totalXP := 0
	currentStreak := 0
	bestStreak := 0
	rank := "青铜"
	if user != nil {
		totalXP = maxInt(user.TotalXP, derivedXP)
		currentStreak = user.CurrentStreak
		bestStreak = maxInt(user.BestStreak, user.CurrentStreak)
		rank = currentRankName(totalXP, user.Rank)
	} else {
		totalXP = derivedXP
	}

	currentRankMinXP, nextRank, nextRankMinXP, rankProgress, xpToNextRank := rankProgressMeta(totalXP)

	return &StatsResponse{
		TotalLessons:     totalLessons,
		Completed:        completed,
		PerfectedLessons: perfectedLessons,
		CompletedCourses: completedCourses,
		TotalXP:          totalXP,
		Rank:             rank,
		RankProgress:     rankProgress,
		CurrentStreak:    currentStreak,
		BestStreak:       bestStreak,
		NextRank:         nextRank,
		CurrentRankMinXP: currentRankMinXP,
		NextRankMinXP:    nextRankMinXP,
		XPToNextRank:     xpToNextRank,
		CourseProgress:   courseProgress,
		Achievements:     buildAchievements(totalXP, completed, perfectedLessons, completedCourses, bestStreak),
	}, nil
}

func buildAchievements(totalXP, completedLessons, perfectedLessons, completedCourses, bestStreak int) []AchievementItem {
	makeAchievement := func(id, name, description, icon, category string, progress, target int) AchievementItem {
		current := maxInt(0, progress)
		if current > target {
			current = target
		}
		return AchievementItem{
			ID:          id,
			Name:        name,
			Description: description,
			Icon:        icon,
			Category:    category,
			Unlocked:    progress >= target,
			Progress:    current,
			Target:      target,
		}
	}

	return []AchievementItem{
		makeAchievement("first_lesson", "初出茅庐", "完成第一节课，正式开启中文学习之旅。", "🌱", "学习", completedLessons, 1),
		makeAchievement("perfect_3", "精准表达", "完成 3 次完美通关，发音与表达都很稳定。", "🎯", "学习", perfectedLessons, 3),
		makeAchievement("course_1", "整门拿下", "完整学完 1 门课程，建立系统学习节奏。", "📘", "学习", completedCourses, 1),
		makeAchievement("streak_3", "三日不辍", "连续打卡 3 天，保持学习热度。", "🔥", "打卡", bestStreak, 3),
		makeAchievement("streak_7", "一周连胜", "连续打卡 7 天，形成稳定习惯。", "🏅", "打卡", bestStreak, 7),
		makeAchievement("xp_120", "经验达人", "累计获得 120 积分，迈入更成熟的阶段。", "💎", "成长", totalXP, 120),
	}
}

func diffUnlockedAchievements(before, after *StatsResponse) []AchievementItem {
	if after == nil {
		return nil
	}

	beforeUnlocked := map[string]bool{}
	if before != nil {
		for _, item := range before.Achievements {
			beforeUnlocked[item.ID] = item.Unlocked
		}
	}

	var result []AchievementItem
	for _, item := range after.Achievements {
		if item.Unlocked && !beforeUnlocked[item.ID] {
			result = append(result, item)
		}
	}
	return result
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

func displayName(user models.User) string {
	if user.Nickname != "" {
		return user.Nickname
	}
	if user.Email != "" {
		return user.Email
	}
	return "学习者"
}

func sortLeaderboard(entries []LeaderboardEntry) {
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if shouldSwap(entries[i], entries[j]) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}

func shouldSwap(left, right LeaderboardEntry) bool {
	if right.TotalXP != left.TotalXP {
		return right.TotalXP > left.TotalXP
	}
	if right.BestStreak != left.BestStreak {
		return right.BestStreak > left.BestStreak
	}
	if right.Completed != left.Completed {
		return right.Completed > left.Completed
	}
	if right.PerfectedLessons != left.PerfectedLessons {
		return right.PerfectedLessons > left.PerfectedLessons
	}
	return right.UserID < left.UserID
}

func takeEntries(entries []LeaderboardEntry, limit int) []LeaderboardEntry {
	if len(entries) <= limit {
		return entries
	}
	return entries[:limit]
}

func selectNearbyEntries(entries []LeaderboardEntry, userID uint, window int) []LeaderboardEntry {
	if len(entries) <= window {
		return entries
	}

	currentIndex := -1
	for i, entry := range entries {
		if entry.UserID == userID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return takeEntries(entries, window)
	}

	half := window / 2
	start := currentIndex - half
	if start < 0 {
		start = 0
	}
	end := start + window
	if end > len(entries) {
		end = len(entries)
		start = maxInt(0, end-window)
	}
	return entries[start:end]
}
