package services

import (
	"testing"
	"time"

	"chinese-learning-app/internal/models"
	"chinese-learning-app/internal/repositories"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newProgressServiceForTest(t *testing.T) (*ProgressService, *repositories.UserRepository, *repositories.CourseRepository, *repositories.ProgressRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Course{}, &models.Lesson{}, &models.UserProgress{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	courseRepo := repositories.NewCourseRepository(db)
	progressRepo := repositories.NewProgressRepository(db)
	service := NewProgressService(progressRepo, userRepo, courseRepo)
	return service, userRepo, courseRepo, progressRepo
}

func createCourseWithLessons(t *testing.T, courseRepo *repositories.CourseRepository, course *models.Course, lessonTitles ...string) *models.Course {
	t.Helper()

	if err := courseRepo.Create(course); err != nil {
		t.Fatalf("create course: %v", err)
	}
	for index, title := range lessonTitles {
		lesson := &models.Lesson{
			CourseID:  course.ID,
			Title:     title,
			Type:      "dialogue",
			Content:   title,
			SortOrder: index + 1,
			XpReward:  10,
		}
		if err := courseRepo.CreateLesson(lesson); err != nil {
			t.Fatalf("create lesson: %v", err)
		}
	}
	return course
}

func createLessonProgress(t *testing.T, progressRepo *repositories.ProgressRepository, userID, lessonID uint, status string, score int) {
	t.Helper()

	now := time.Now()
	progress := &models.UserProgress{
		UserID:      userID,
		LessonID:    lessonID,
		Status:      status,
		Score:       score,
		Attempts:    1,
		CompletedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := progressRepo.CreateOrUpdate(progress); err != nil {
		t.Fatalf("create progress: %v", err)
	}
}

func TestGetUserStatsProvidesContinueLearning(t *testing.T) {
	service, userRepo, courseRepo, progressRepo := newProgressServiceForTest(t)

	user := &models.User{Email: "learner@example.com", PasswordHash: "hash", Nickname: "学习者"}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	course := createCourseWithLessons(t, courseRepo, &models.Course{
		Title:       "L0 日常入门",
		Description: "从打招呼开始",
		Level:       "L0",
		LevelName:   "L0",
		SortOrder:   1,
		IsPublished: true,
	}, "第一课", "第二课", "第三课")

	lessons, err := courseRepo.GetLessons(course.ID)
	if err != nil {
		t.Fatalf("get lessons: %v", err)
	}
	createLessonProgress(t, progressRepo, user.ID, lessons[0].ID, "completed", 80)

	stats, err := service.GetUserStats(user.ID)
	if err != nil {
		t.Fatalf("get user stats: %v", err)
	}
	if stats.ContinueLearning == nil {
		t.Fatal("expected continue learning item")
	}
	if stats.ContinueLearning.CourseID != course.ID {
		t.Fatalf("expected continue learning course %d, got %d", course.ID, stats.ContinueLearning.CourseID)
	}
	if stats.ContinueLearning.LessonOrder != 2 {
		t.Fatalf("expected continue learning lesson order 2, got %d", stats.ContinueLearning.LessonOrder)
	}
	if stats.ContinueLearning.LessonTitle != "第二课" {
		t.Fatalf("expected continue learning lesson title 第二课, got %q", stats.ContinueLearning.LessonTitle)
	}
}

func TestGetUserStatsRecommendsL1AfterCompletingL0(t *testing.T) {
	service, userRepo, courseRepo, progressRepo := newProgressServiceForTest(t)

	user := &models.User{Email: "advanced@example.com", PasswordHash: "hash", Nickname: "进阶学习者"}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	l0Course := createCourseWithLessons(t, courseRepo, &models.Course{
		Title:       "L0 生存中文",
		Description: "先学会最常用表达",
		Level:       "L0",
		LevelName:   "L0",
		SortOrder:   1,
		IsPublished: true,
	}, "L0 第一课", "L0 第二课")
	l1Course := createCourseWithLessons(t, courseRepo, &models.Course{
		Title:       "L1 场景对话",
		Description: "进入更完整的交流训练",
		Level:       "L1",
		LevelName:   "L1",
		SortOrder:   2,
		IsPublished: true,
	}, "L1 第一课", "L1 第二课")

	l0Lessons, err := courseRepo.GetLessons(l0Course.ID)
	if err != nil {
		t.Fatalf("get l0 lessons: %v", err)
	}
	for _, lesson := range l0Lessons {
		createLessonProgress(t, progressRepo, user.ID, lesson.ID, "perfected", 100)
	}

	stats, err := service.GetUserStats(user.ID)
	if err != nil {
		t.Fatalf("get user stats: %v", err)
	}
	if stats.RecommendedCourse == nil {
		t.Fatal("expected recommended course")
	}
	if stats.RecommendedCourse.CourseID != l1Course.ID {
		t.Fatalf("expected recommended course %d, got %d", l1Course.ID, stats.RecommendedCourse.CourseID)
	}
	if stats.RecommendedCourse.Level != "L1" {
		t.Fatalf("expected recommended level L1, got %q", stats.RecommendedCourse.Level)
	}
	if stats.RecommendedCourse.FirstLessonTitle != "L1 第一课" {
		t.Fatalf("expected recommended first lesson title L1 第一课, got %q", stats.RecommendedCourse.FirstLessonTitle)
	}
	if stats.ContinueLearning == nil || stats.ContinueLearning.CourseID != l1Course.ID {
		t.Fatal("expected continue learning to move to the first lesson of the recommended L1 course")
	}
}
