
package services

import (
	"chinese-learning-app/internal/models"
	"chinese-learning-app/internal/repositories"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CourseDetailResponse struct {
	Course  models.Course   `json:"course"`
	Lessons []models.Lesson `json:"lessons"`
}

type CourseService struct {
	courseRepo *repositories.CourseRepository
}

const maxCourseThumbnailSize int64 = 2 * 1024 * 1024

func NewCourseService(courseRepo *repositories.CourseRepository) *CourseService {
	return &CourseService{courseRepo: courseRepo}
}

func (s *CourseService) GetCourseList(level string) ([]models.Course, error) {
	return s.courseRepo.GetAll(level)
}

func (s *CourseService) GetCourseListForAdmin(level string) ([]models.Course, error) {
	return s.courseRepo.GetAllForAdmin(level)
}

func (s *CourseService) GetCourseDetail(courseID uint) (*CourseDetailResponse, error) {
	course, err := s.courseRepo.GetByID(courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, nil
	}

	lessons, err := s.courseRepo.GetLessons(courseID)
	if err != nil {
		return nil, err
	}

	return &CourseDetailResponse{
		Course:  *course,
		Lessons: lessons,
	}, nil
}

func (s *CourseService) GetLessonContent(lessonID uint) (*models.Lesson, error) {
	return s.courseRepo.GetLessonByID(lessonID)
}

func (s *CourseService) CreateCourse(course *models.Course) (*models.Course, error) {
	now := time.Now()
	course.CreatedAt = now
	course.UpdatedAt = now
	if err := s.courseRepo.Create(course); err != nil {
		return nil, err
	}
	return course, nil
}

func (s *CourseService) UpdateCourse(courseID uint, input *models.Course) (*models.Course, error) {
	existing, err := s.courseRepo.GetByID(courseID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	existing.Title = input.Title
	existing.Description = input.Description
	existing.Level = input.Level
	existing.LevelName = input.LevelName
	existing.Thumbnail = input.Thumbnail
	existing.SortOrder = input.SortOrder
	existing.IsPublished = input.IsPublished
	existing.UpdatedAt = time.Now()

	if err := s.courseRepo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *CourseService) UpdateCoursePublishStatus(courseID uint, isPublished bool) (*models.Course, error) {
	existing, err := s.courseRepo.GetByID(courseID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}
	existing.IsPublished = isPublished
	existing.UpdatedAt = time.Now()
	if err := s.courseRepo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *CourseService) DeleteCourse(courseID uint) error {
	existing, err := s.courseRepo.GetByID(courseID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("course not found")
	}
	if err := s.courseRepo.Delete(courseID); err != nil {
		return err
	}
	removeLocalCourseThumbnail(existing.Thumbnail)
	return nil
}

func (s *CourseService) CreateLesson(courseID uint, lesson *models.Lesson) (*models.Lesson, error) {
	course, err := s.courseRepo.GetByID(courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}

	now := time.Now()
	lesson.CourseID = courseID
	lesson.CreatedAt = now
	lesson.UpdatedAt = now
	if err := s.courseRepo.CreateLesson(lesson); err != nil {
		return nil, err
	}
	return lesson, nil
}

func (s *CourseService) UpdateLesson(lessonID uint, input *models.Lesson) (*models.Lesson, error) {
	existing, err := s.courseRepo.GetLessonByID(lessonID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	if input.CourseID != 0 {
		course, err := s.courseRepo.GetByID(input.CourseID)
		if err != nil {
			return nil, err
		}
		if course == nil {
			return nil, errors.New("course not found")
		}
		existing.CourseID = input.CourseID
	}

	existing.Title = input.Title
	existing.Type = input.Type
	existing.Content = input.Content
	existing.AudioURL = input.AudioURL
	existing.SortOrder = input.SortOrder
	existing.IsFree = input.IsFree
	existing.XpReward = input.XpReward
	existing.UpdatedAt = time.Now()

	if err := s.courseRepo.UpdateLesson(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *CourseService) DeleteLesson(lessonID uint) error {
	existing, err := s.courseRepo.GetLessonByID(lessonID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("lesson not found")
	}
	return s.courseRepo.DeleteLesson(lessonID)
}

func (s *CourseService) ReorderLessons(courseID uint, lessonIDs []uint) error {
	course, err := s.courseRepo.GetByID(courseID)
	if err != nil {
		return err
	}
	if course == nil {
		return errors.New("course not found")
	}
	return s.courseRepo.ReorderLessons(courseID, lessonIDs)
}

func (s *CourseService) ReorderCourses(courseIDs []uint) error {
	return s.courseRepo.ReorderCourses(courseIDs)
}

func (s *CourseService) UploadCourseThumbnail(courseID uint, fileHeader *multipart.FileHeader) (*models.Course, string, error) {
	existing, err := s.courseRepo.GetByID(courseID)
	if err != nil {
		return nil, "", err
	}
	if existing == nil {
		return nil, "", errors.New("course not found")
	}
	if fileHeader == nil {
		return nil, "", errors.New("thumbnail file is required")
	}
	if fileHeader.Size <= 0 {
		return nil, "", errors.New("thumbnail file is empty")
	}
	if fileHeader.Size > maxCourseThumbnailSize {
		return nil, "", errors.New("thumbnail size cannot exceed 2MB")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, "", err
	}

	mimeType := http.DetectContentType(buffer[:n])
	ext, ok := thumbnailExtensionByMime(mimeType)
	if !ok {
		return nil, "", errors.New("only JPG and PNG images are supported")
	}

	uploadDir := resolveUploadDir()
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return nil, "", err
	}

	fileName := fmt.Sprintf("course_%d_%d%s", courseID, time.Now().UnixNano(), ext)
	destinationPath := filepath.Join(uploadDir, fileName)
	if err := copyUploadedFile(fileHeader, destinationPath); err != nil {
		return nil, "", err
	}

	oldThumbnail := existing.Thumbnail
	existing.Thumbnail = "/uploads/" + fileName
	existing.UpdatedAt = time.Now()
	if err := s.courseRepo.Update(existing); err != nil {
		_ = os.Remove(destinationPath)
		return nil, "", err
	}

	if oldThumbnail != "" && oldThumbnail != existing.Thumbnail {
		removeLocalCourseThumbnail(oldThumbnail)
	}

	return existing, existing.Thumbnail, nil
}

func (s *CourseService) DeleteCourseThumbnail(courseID uint) (*models.Course, error) {
	existing, err := s.courseRepo.GetByID(courseID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("course not found")
	}
	if existing.Thumbnail == "" {
		return existing, nil
	}

	oldThumbnail := existing.Thumbnail
	existing.Thumbnail = ""
	existing.UpdatedAt = time.Now()
	if err := s.courseRepo.Update(existing); err != nil {
		return nil, err
	}

	removeLocalCourseThumbnail(oldThumbnail)
	return existing, nil
}

func thumbnailExtensionByMime(mimeType string) (string, bool) {
	switch mimeType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	default:
		return "", false
	}
}

func copyUploadedFile(fileHeader *multipart.FileHeader, destinationPath string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func removeLocalCourseThumbnail(thumbnail string) {
	if !strings.HasPrefix(thumbnail, "/uploads/") {
		return
	}

	fileName := strings.TrimPrefix(thumbnail, "/uploads/")
	if fileName == "" {
		return
	}
	_ = os.Remove(filepath.Join(resolveUploadDir(), fileName))
}

func resolveUploadDir() string {
	uploadDir := os.Getenv("UPLOAD_DIR")
	if strings.TrimSpace(uploadDir) == "" {
		uploadDir = filepath.Join(".", "uploads")
	}
	return uploadDir
}
