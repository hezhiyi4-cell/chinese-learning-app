package services

import (
	"chinese-learning-app/internal/models"
	"chinese-learning-app/internal/repositories"
	"math"
	"sort"
	"strconv"
	"strings"
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

type ContinueLearningItem struct {
	CourseID     uint   `json:"courseId"`
	CourseTitle  string `json:"courseTitle"`
	LessonID     uint   `json:"lessonId"`
	LessonTitle  string `json:"lessonTitle"`
	LessonOrder  int    `json:"lessonOrder"`
	Level        string `json:"level"`
	LevelName    string `json:"levelName"`
	Completed    int    `json:"completed"`
	TotalLessons int    `json:"totalLessons"`
}

type RecommendedCourseItem struct {
	CourseID         uint   `json:"courseId"`
	CourseTitle      string `json:"courseTitle"`
	Level            string `json:"level"`
	LevelName        string `json:"levelName"`
	FirstLessonID    uint   `json:"firstLessonId"`
	FirstLessonTitle string `json:"firstLessonTitle"`
	Reason           string `json:"reason"`
}

type StatsResponse struct {
	TotalLessons      int                    `json:"totalLessons"`
	Completed         int                    `json:"completed"`
	PerfectedLessons  int                    `json:"perfectedLessons"`
	CompletedCourses  int                    `json:"completedCourses"`
	TotalXP           int                    `json:"totalXP"`
	Rank              string                 `json:"rank"`
	RankProgress      int                    `json:"rankProgress"`
	CurrentStreak     int                    `json:"currentStreak"`
	BestStreak        int                    `json:"bestStreak"`
	NextRank          string                 `json:"nextRank"`
	CurrentRankMinXP  int                    `json:"currentRankMinXp"`
	NextRankMinXP     int                    `json:"nextRankMinXp"`
	XPToNextRank      int                    `json:"xpToNextRank"`
	CourseProgress    []CourseProgressItem   `json:"courseProgress"`
	ContinueLearning  *ContinueLearningItem  `json:"continueLearning"`
	RecommendedCourse *RecommendedCourseItem `json:"recommendedCourse"`
	Achievements      []AchievementItem      `json:"achievements"`
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

type courseSnapshot struct {
	course               models.Course
	lessons              []models.Lesson
	completed            int
	total                int
	firstIncompleteIndex int
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

		if _, err := s.syncUserRank(user, updatedProgressList); err != nil {
			return nil, err
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

	if _, err := s.syncUserRank(user, progressList); err != nil {
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
	snapshots := make([]courseSnapshot, 0, len(courses))
	lessonIndexByID := map[uint]struct {
		snapshotIndex int
		lessonIndex   int
	}{}

	for courseIndex, course := range courses {
		lessons, err := s.courseRepo.GetLessons(course.ID)
		if err != nil {
			return nil, err
		}

		item := CourseProgressItem{
			CourseID: course.ID,
			Total:    len(lessons),
			TotalXP:  len(lessons) * (lessonCompleteXP + perfectBonusXP),
		}
		snapshot := courseSnapshot{
			course:               course,
			lessons:              lessons,
			total:                len(lessons),
			firstIncompleteIndex: -1,
		}
		totalLessons += len(lessons)

		for index, lesson := range lessons {
			lessonIndexByID[lesson.ID] = struct {
				snapshotIndex int
				lessonIndex   int
			}{snapshotIndex: courseIndex, lessonIndex: index}

			p, ok := progressByLessonID[lesson.ID]
			if !ok || !isCompletedStatus(p.Status) {
				if snapshot.firstIncompleteIndex == -1 {
					snapshot.firstIncompleteIndex = index
				}
				if !ok {
					continue
				}
			}
			if !ok {
				continue
			}
			if isCompletedStatus(p.Status) {
				item.Completed++
				completed++
				item.EarnedXP += lessonCompleteXP
				snapshot.completed++
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
		snapshots = append(snapshots, snapshot)
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
	continueLearning := deriveContinueLearning(progressList, progressByLessonID, snapshots, lessonIndexByID)
	recommendedCourse := deriveRecommendedCourse(snapshots)

	return &StatsResponse{
		TotalLessons:      totalLessons,
		Completed:         completed,
		PerfectedLessons:  perfectedLessons,
		CompletedCourses:  completedCourses,
		TotalXP:           totalXP,
		Rank:              rank,
		RankProgress:      rankProgress,
		CurrentStreak:     currentStreak,
		BestStreak:        bestStreak,
		NextRank:          nextRank,
		CurrentRankMinXP:  currentRankMinXP,
		NextRankMinXP:     nextRankMinXP,
		XPToNextRank:      xpToNextRank,
		CourseProgress:    courseProgress,
		ContinueLearning:  continueLearning,
		RecommendedCourse: recommendedCourse,
		Achievements:      buildAchievements(totalXP, completed, perfectedLessons, completedCourses, bestStreak),
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

func (s *ProgressService) syncUserRank(user *models.User, progressList []models.UserProgress) (string, error) {
	if user == nil {
		return currentRankName(deriveXPFromProgressList(progressList), "青铜"), nil
	}

	derivedXP := deriveXPFromProgressList(progressList)
	totalXP := maxInt(user.TotalXP, derivedXP)
	newRank := currentRankName(totalXP, user.Rank)
	shouldUpdate := false

	if user.TotalXP != totalXP {
		user.TotalXP = totalXP
		shouldUpdate = true
	}
	if user.Rank != newRank {
		user.Rank = newRank
		shouldUpdate = true
	}

	if shouldUpdate {
		user.UpdatedAt = time.Now()
		if err := s.userRepo.Update(user); err != nil {
			return "", err
		}
	}

	return newRank, nil
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

func deriveContinueLearning(
	progressList []models.UserProgress,
	progressByLessonID map[uint]models.UserProgress,
	snapshots []courseSnapshot,
	lessonIndexByID map[uint]struct {
		snapshotIndex int
		lessonIndex   int
	},
) *ContinueLearningItem {
	lastProgress, ok := latestProgress(progressList)
	if ok {
		loc, found := lessonIndexByID[lastProgress.LessonID]
		if found && loc.snapshotIndex >= 0 && loc.snapshotIndex < len(snapshots) {
			snapshot := snapshots[loc.snapshotIndex]

			if !isCompletedStatus(lastProgress.Status) {
				return buildContinueLearningItem(snapshot, loc.lessonIndex)
			}

			for nextIndex := loc.lessonIndex + 1; nextIndex < len(snapshot.lessons); nextIndex++ {
				lesson := snapshot.lessons[nextIndex]
				p, hasProgress := progressByLessonID[lesson.ID]
				if !hasProgress || !isCompletedStatus(p.Status) {
					return buildContinueLearningItem(snapshot, nextIndex)
				}
			}

			if snapshot.firstIncompleteIndex >= 0 && snapshot.firstIncompleteIndex < len(snapshot.lessons) {
				return buildContinueLearningItem(snapshot, snapshot.firstIncompleteIndex)
			}

			for snapshotIndex := loc.snapshotIndex + 1; snapshotIndex < len(snapshots); snapshotIndex++ {
				nextSnapshot := snapshots[snapshotIndex]
				if nextSnapshot.total == 0 || nextSnapshot.completed >= nextSnapshot.total {
					continue
				}
				nextLessonIndex := nextSnapshot.firstIncompleteIndex
				if nextLessonIndex < 0 {
					nextLessonIndex = 0
				}
				return buildContinueLearningItem(nextSnapshot, nextLessonIndex)
			}

			for snapshotIndex := 0; snapshotIndex < loc.snapshotIndex; snapshotIndex++ {
				nextSnapshot := snapshots[snapshotIndex]
				if nextSnapshot.total == 0 || nextSnapshot.completed >= nextSnapshot.total {
					continue
				}
				nextLessonIndex := nextSnapshot.firstIncompleteIndex
				if nextLessonIndex < 0 {
					nextLessonIndex = 0
				}
				return buildContinueLearningItem(nextSnapshot, nextLessonIndex)
			}
		}
	}

	for _, snapshot := range snapshots {
		if snapshot.total == 0 || snapshot.completed >= snapshot.total {
			continue
		}
		lessonIndex := snapshot.firstIncompleteIndex
		if lessonIndex < 0 {
			lessonIndex = 0
		}
		return buildContinueLearningItem(snapshot, lessonIndex)
	}

	return nil
}

func buildContinueLearningItem(snapshot courseSnapshot, lessonIndex int) *ContinueLearningItem {
	if lessonIndex < 0 || lessonIndex >= len(snapshot.lessons) {
		return nil
	}
	lesson := snapshot.lessons[lessonIndex]
	return &ContinueLearningItem{
		CourseID:     snapshot.course.ID,
		CourseTitle:  snapshot.course.Title,
		LessonID:     lesson.ID,
		LessonTitle:  lesson.Title,
		LessonOrder:  lessonIndex + 1,
		Level:        snapshot.course.Level,
		LevelName:    snapshot.course.LevelName,
		Completed:    snapshot.completed,
		TotalLessons: snapshot.total,
	}
}

func deriveRecommendedCourse(snapshots []courseSnapshot) *RecommendedCourseItem {
	type levelSummary struct {
		label            string
		totalCourses     int
		completedCourses int
		firstIncomplete  *RecommendedCourseItem
	}

	orderedLevels := make([]string, 0, len(snapshots))
	levelSummaries := map[string]*levelSummary{}

	for _, snapshot := range snapshots {
		if snapshot.total == 0 {
			continue
		}

		levelKey := snapshot.course.Level
		summary, ok := levelSummaries[levelKey]
		if !ok {
			summary = &levelSummary{label: displayCourseLevel(snapshot.course)}
			levelSummaries[levelKey] = summary
			orderedLevels = append(orderedLevels, levelKey)
		}

		summary.totalCourses++
		if snapshot.completed == snapshot.total {
			summary.completedCourses++
		} else if summary.firstIncomplete == nil {
			summary.firstIncomplete = buildRecommendedCourseItem(snapshot)
		}
	}

	sort.SliceStable(orderedLevels, func(i, j int) bool {
		return courseLevelRank(orderedLevels[i]) < courseLevelRank(orderedLevels[j])
	})

	for index, levelKey := range orderedLevels {
		summary := levelSummaries[levelKey]
		if summary == nil || summary.totalCourses == 0 || summary.completedCourses != summary.totalCourses {
			continue
		}
		for nextIndex := index + 1; nextIndex < len(orderedLevels); nextIndex++ {
			nextSummary := levelSummaries[orderedLevels[nextIndex]]
			if nextSummary == nil || nextSummary.firstIncomplete == nil {
				continue
			}
			item := *nextSummary.firstIncomplete
			item.Reason = summary.label + " 已完成，推荐进入 " + nextSummary.label
			return &item
		}
	}

	for _, snapshot := range snapshots {
		if snapshot.total == 0 || snapshot.completed == snapshot.total {
			continue
		}
		item := buildRecommendedCourseItem(snapshot)
		if item != nil {
			item.Reason = "推荐继续当前学习路径，按顺序学效果更好。"
		}
		return item
	}

	return nil
}

func buildRecommendedCourseItem(snapshot courseSnapshot) *RecommendedCourseItem {
	if len(snapshot.lessons) == 0 {
		return nil
	}
	firstLesson := snapshot.lessons[0]
	return &RecommendedCourseItem{
		CourseID:         snapshot.course.ID,
		CourseTitle:      snapshot.course.Title,
		Level:            snapshot.course.Level,
		LevelName:        snapshot.course.LevelName,
		FirstLessonID:    firstLesson.ID,
		FirstLessonTitle: firstLesson.Title,
	}
}

func displayCourseLevel(course models.Course) string {
	if course.LevelName != "" {
		return course.LevelName
	}
	if course.Level != "" {
		return course.Level
	}
	return "当前阶段"
}

func latestProgress(progressList []models.UserProgress) (models.UserProgress, bool) {
	if len(progressList) == 0 {
		return models.UserProgress{}, false
	}
	best := progressList[0]
	bestTime := progressTimestamp(best)
	for _, item := range progressList[1:] {
		current := progressTimestamp(item)
		if current.After(bestTime) {
			best = item
			bestTime = current
		}
	}
	return best, true
}

func progressTimestamp(progress models.UserProgress) time.Time {
	best := progress.UpdatedAt
	if progress.CompletedAt != nil && progress.CompletedAt.After(best) {
		best = *progress.CompletedAt
	}
	return best
}

func courseLevelRank(level string) int {
	value := strings.TrimSpace(level)
	if value == "" {
		return 1_000_000
	}
	if strings.HasPrefix(strings.ToUpper(value), "L") {
		number, err := strconv.Atoi(strings.TrimPrefix(strings.ToUpper(value), "L"))
		if err == nil {
			return number
		}
	}
	return 1_000_000
}
