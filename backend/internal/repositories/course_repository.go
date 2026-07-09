
package repositories

import (
	"chinese-learning-app/internal/models"

	"gorm.io/gorm"
)

type CourseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) *CourseRepository {
	return &CourseRepository{
		db: db,
	}
}

func (r *CourseRepository) GetAll(level string) ([]models.Course, error) {
	var result []models.Course
	query := r.db.Where("is_published = ?", true)
	if level != "" {
		query = query.Where("level = ?", level)
	}
	err := query.Order("sort_order ASC").Order("id ASC").Find(&result).Error
	return result, err
}

func (r *CourseRepository) GetAllForAdmin(level string) ([]models.Course, error) {
	var result []models.Course
	query := r.db.Model(&models.Course{})
	if level != "" {
		query = query.Where("level = ?", level)
	}
	err := query.Order("sort_order ASC").Order("id ASC").Find(&result).Error
	return result, err
}

func (r *CourseRepository) GetByID(id uint) (*models.Course, error) {
	var course models.Course
	err := r.db.First(&course, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) GetLessons(courseID uint) ([]models.Lesson, error) {
	var lessons []models.Lesson
	err := r.db.Where("course_id = ?", courseID).Order("sort_order ASC").Order("id ASC").Find(&lessons).Error
	return lessons, err
}

func (r *CourseRepository) Create(course *models.Course) error {
	return r.db.Create(course).Error
}

func (r *CourseRepository) Update(course *models.Course) error {
	return r.db.Save(course).Error
}

func (r *CourseRepository) UpdatePublishStatus(id uint, isPublished bool) (*models.Course, error) {
	var course models.Course
	if err := r.db.First(&course, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	course.IsPublished = isPublished
	if err := r.db.Save(&course).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var lessons []models.Lesson
		if err := tx.Where("course_id = ?", id).Find(&lessons).Error; err != nil {
			return err
		}

		if len(lessons) > 0 {
			lessonIDs := make([]uint, 0, len(lessons))
			for _, lesson := range lessons {
				lessonIDs = append(lessonIDs, lesson.ID)
			}
			if err := tx.Where("lesson_id IN ?", lessonIDs).Delete(&models.UserProgress{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("course_id = ?", id).Delete(&models.Lesson{}).Error; err != nil {
			return err
		}

		return tx.Delete(&models.Course{}, id).Error
	})
}

func (r *CourseRepository) CreateLesson(lesson *models.Lesson) error {
	return r.db.Create(lesson).Error
}

func (r *CourseRepository) GetLessonByID(id uint) (*models.Lesson, error) {
	var lesson models.Lesson
	err := r.db.First(&lesson, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &lesson, nil
}

func (r *CourseRepository) UpdateLesson(lesson *models.Lesson) error {
	return r.db.Save(lesson).Error
}

func (r *CourseRepository) DeleteLesson(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("lesson_id = ?", id).Delete(&models.UserProgress{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Lesson{}, id).Error
	})
}

func (r *CourseRepository) ReorderLessons(courseID uint, lessonIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var lessons []models.Lesson
		if err := tx.Where("course_id = ?", courseID).Find(&lessons).Error; err != nil {
			return err
		}
		if len(lessons) == 0 {
			return nil
		}

		lessonMap := make(map[uint]models.Lesson, len(lessons))
		for _, lesson := range lessons {
			lessonMap[lesson.ID] = lesson
		}
		if len(lessonMap) != len(lessonIDs) {
			return gorm.ErrRecordNotFound
		}

		for index, lessonID := range lessonIDs {
			lesson, ok := lessonMap[lessonID]
			if !ok {
				return gorm.ErrRecordNotFound
			}
			if err := tx.Model(&models.Lesson{}).
				Where("id = ?", lesson.ID).
				Updates(map[string]interface{}{"sort_order": index + 1}).
				Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *CourseRepository) ReorderCourses(courseIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var courses []models.Course
		if err := tx.Find(&courses).Error; err != nil {
			return err
		}
		if len(courses) == 0 {
			return nil
		}

		courseMap := make(map[uint]models.Course, len(courses))
		for _, course := range courses {
			courseMap[course.ID] = course
		}
		if len(courseMap) != len(courseIDs) {
			return gorm.ErrRecordNotFound
		}

		for index, courseID := range courseIDs {
			course, ok := courseMap[courseID]
			if !ok {
				return gorm.ErrRecordNotFound
			}
			if err := tx.Model(&models.Course{}).
				Where("id = ?", course.ID).
				Updates(map[string]interface{}{"sort_order": index + 1}).
				Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *CourseRepository) CountCourses() (int64, error) {
	var count int64
	err := r.db.Model(&models.Course{}).Count(&count).Error
	return count, err
}
