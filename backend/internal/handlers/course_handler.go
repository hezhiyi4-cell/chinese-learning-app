package handlers

import (
	"net/http"
	"strconv"

	"chinese-learning-app/internal/models"
	"chinese-learning-app/internal/services"

	"github.com/gin-gonic/gin"
)

type CourseHandler struct {
	courseService   *services.CourseService
	progressService *services.ProgressService
}

func NewCourseHandler(courseService *services.CourseService, progressService *services.ProgressService) *CourseHandler {
	return &CourseHandler{
		courseService:   courseService,
		progressService: progressService,
	}
}

func (h *CourseHandler) GetCourses(c *gin.Context) {
	level := c.Query("level")
	courses, err := h.courseService.GetCourseList(level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch courses"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"courses": courses})
}

func (h *CourseHandler) GetCoursesForAdmin(c *gin.Context) {
	level := c.Query("level")
	courses, err := h.courseService.GetCourseListForAdmin(level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch courses"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"courses": courses})
}

func (h *CourseHandler) GetCourseDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	detail, err := h.courseService.GetCourseDetail(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch course"})
		return
	}
	if detail == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *CourseHandler) GetLesson(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	lesson, err := h.courseService.GetLessonContent(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch lesson"})
		return
	}
	if lesson == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lesson not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"lesson": lesson})
}

func (h *CourseHandler) GetProgress(c *gin.Context) {
	userID, _ := c.Get("userId")
	progress, err := h.progressService.GetUserProgress(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch progress"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"progress": progress})
}

type UpdateProgressRequest struct {
	Score int `json:"score" binding:"required"`
}

type CreateCourseRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Level       string `json:"level" binding:"required"`
	LevelName   string `json:"levelName"`
	Thumbnail   string `json:"thumbnail"`
	SortOrder   int    `json:"sortOrder"`
	IsPublished bool   `json:"isPublished"`
}

type CreateLessonRequest struct {
	Title     string `json:"title" binding:"required"`
	Type      string `json:"type" binding:"required"`
	Content   string `json:"content"`
	AudioURL  string `json:"audioUrl"`
	SortOrder int    `json:"sortOrder"`
	IsFree    bool   `json:"isFree"`
	XpReward  int    `json:"xpReward"`
}

type UpdateLessonRequest struct {
	CourseID  uint   `json:"courseId"`
	Title     string `json:"title" binding:"required"`
	Type      string `json:"type" binding:"required"`
	Content   string `json:"content"`
	AudioURL  string `json:"audioUrl"`
	SortOrder int    `json:"sortOrder"`
	IsFree    bool   `json:"isFree"`
	XpReward  int    `json:"xpReward"`
}

type UpdatePublishRequest struct {
	IsPublished bool `json:"isPublished"`
}

type ReorderLessonsRequest struct {
	CourseID  uint   `json:"courseId" binding:"required"`
	LessonIDs []uint `json:"lessonIds" binding:"required"`
}

type ReorderCoursesRequest struct {
	CourseIDs []uint `json:"courseIds" binding:"required"`
}

func (h *CourseHandler) CreateCourse(c *gin.Context) {
	var req CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	courseInput := servicesCourseFromRequest(req)
	course, err := h.courseService.CreateCourse(courseInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create course"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"course": course})
}

func (h *CourseHandler) UpdateCourse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	var req CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	courseInput := servicesCourseFromRequest(req)
	course, err := h.courseService.UpdateCourse(uint(id), courseInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update course"})
		return
	}
	if course == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"course": course})
}

func (h *CourseHandler) DeleteCourse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	if err := h.courseService.DeleteCourse(uint(id)); err != nil {
		if err.Error() == "course not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete course"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course deleted successfully"})
}

func (h *CourseHandler) UpdateCoursePublishStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	var req UpdatePublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	course, err := h.courseService.UpdateCoursePublishStatus(uint(id), req.IsPublished)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update course publish status"})
		return
	}
	if course == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"course": course})
}

func (h *CourseHandler) UploadCourseThumbnail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "thumbnail file is required"})
		return
	}

	course, imageURL, err := h.courseService.UploadCourseThumbnail(uint(id), fileHeader)
	if err != nil {
		switch err.Error() {
		case "course not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		case "thumbnail file is required", "thumbnail file is empty", "thumbnail size cannot exceed 2MB", "only JPG and PNG images are supported":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload thumbnail"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Thumbnail uploaded successfully",
		"imageUrl": imageURL,
		"course":   course,
	})
}

func (h *CourseHandler) DeleteCourseThumbnail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	course, err := h.courseService.DeleteCourseThumbnail(uint(id))
	if err != nil {
		if err.Error() == "course not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete thumbnail"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thumbnail deleted successfully",
		"course":  course,
	})
}

func (h *CourseHandler) CreateLesson(c *gin.Context) {
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	var req CreateLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lessonInput := servicesLessonFromCreateRequest(req)
	lesson, err := h.courseService.CreateLesson(uint(courseID), lessonInput)
	if err != nil {
		if err.Error() == "course not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create lesson"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"lesson": lesson})
}

func (h *CourseHandler) UpdateLesson(c *gin.Context) {
	lessonID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	var req UpdateLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lessonInput := servicesLessonFromUpdateRequest(req)
	lesson, err := h.courseService.UpdateLesson(uint(lessonID), lessonInput)
	if err != nil {
		if err.Error() == "course not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lesson"})
		return
	}
	if lesson == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lesson not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lesson": lesson})
}

func (h *CourseHandler) DeleteLesson(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	if err := h.courseService.DeleteLesson(uint(id)); err != nil {
		if err.Error() == "lesson not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Lesson not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete lesson"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lesson deleted successfully"})
}

func (h *CourseHandler) ReorderLessons(c *gin.Context) {
	var req ReorderLessonsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.LessonIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lessonIds cannot be empty"})
		return
	}

	if err := h.courseService.ReorderLessons(req.CourseID, req.LessonIDs); err != nil {
		if err.Error() == "course not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reorder lessons"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lessons reordered successfully"})
}

func (h *CourseHandler) ReorderCourses(c *gin.Context) {
	var req ReorderCoursesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.CourseIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "courseIds cannot be empty"})
		return
	}

	if err := h.courseService.ReorderCourses(req.CourseIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reorder courses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Courses reordered successfully"})
}

func (h *CourseHandler) UpdateProgress(c *gin.Context) {
	userID, _ := c.Get("userId")
	lessonIDStr := c.Param("lessonId")
	lessonID, err := strconv.ParseUint(lessonIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	var req UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.progressService.UpdateProgress(userID.(uint), uint(lessonID), req.Score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update progress"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CourseHandler) GetStats(c *gin.Context) {
	userID, _ := c.Get("userId")
	stats, err := h.progressService.GetUserStats(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func (h *CourseHandler) GetLeaderboard(c *gin.Context) {
	userID, _ := c.Get("userId")
	scope := c.DefaultQuery("scope", "global")

	leaderboard, err := h.progressService.GetLeaderboard(userID.(uint), scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"leaderboard": leaderboard})
}

func servicesCourseFromRequest(req CreateCourseRequest) *models.Course {
	return &models.Course{
		Title:       req.Title,
		Description: req.Description,
		Level:       req.Level,
		LevelName:   req.LevelName,
		Thumbnail:   req.Thumbnail,
		SortOrder:   req.SortOrder,
		IsPublished: req.IsPublished,
	}
}

func servicesLessonFromCreateRequest(req CreateLessonRequest) *models.Lesson {
	return &models.Lesson{
		Title:     req.Title,
		Type:      req.Type,
		Content:   req.Content,
		AudioURL:  req.AudioURL,
		SortOrder: req.SortOrder,
		IsFree:    req.IsFree,
		XpReward:  req.XpReward,
	}
}

func servicesLessonFromUpdateRequest(req UpdateLessonRequest) *models.Lesson {
	return &models.Lesson{
		CourseID:  req.CourseID,
		Title:     req.Title,
		Type:      req.Type,
		Content:   req.Content,
		AudioURL:  req.AudioURL,
		SortOrder: req.SortOrder,
		IsFree:    req.IsFree,
		XpReward:  req.XpReward,
	}
}
