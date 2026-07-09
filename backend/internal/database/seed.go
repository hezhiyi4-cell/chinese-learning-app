
package database

import (
	"chinese-learning-app/internal/models"
	"chinese-learning-app/internal/repositories"
	"time"
)

func SeedInitialData(courseRepo *repositories.CourseRepository) error {
	count, err := courseRepo.CountCourses()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// === L0 零基础课程 ===
	l0Course := &models.Course{
		Title:       "零基础入门 - 拼音与声调",
		Description: "从零开始学习中文拼音和声调，为中文学习打下坚实基础",
		Level:       "L0",
		LevelName:   "零基础",
		SortOrder:   1,
		IsPublished: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := courseRepo.Create(l0Course); err != nil {
		return err
	}

	// L0 课时
	l0Lessons := []*models.Lesson{
		{
			CourseID:  l0Course.ID,
			Title:     "第1课：拼音声母表",
			Type:      "pronunciation",
			Content:   `{"introduction": "学习中文声母的基础发音", "content": "b, p, m, f, d, t, n, l, g, k, h, j, q, x, zh, ch, sh, r, z, c, s"}`,
			IsFree:    true,
			XpReward:  10,
			SortOrder: 1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			CourseID:  l0Course.ID,
			Title:     "第2课：拼音韵母表",
			Type:      "pronunciation",
			Content:   `{"introduction": "学习中文韵母的标准发音", "content": "a, o, e, i, u, ü, ai, ei, ui, ao, ou, iu, ie, üe, er, an, en, in, un, ün, ang, eng, ing, ong"}`,
			IsFree:    true,
			XpReward:  10,
			SortOrder: 2,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			CourseID:  l0Course.ID,
			Title:     "第3课：声调入门（一、二声）",
			Type:      "tone",
			Content:   `{"introduction": "学习第一声和第二声", "content": "一声平，二声扬"}`,
			IsFree:    true,
			XpReward:  15,
			SortOrder: 3,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			CourseID:  l0Course.ID,
			Title:     "第4课：声调进阶（三、四声、轻声）",
			Type:      "tone",
			Content:   `{"introduction": "学习第三声、第四声和轻声", "content": "三声拐弯，四声降"}`,
			IsFree:    false,
			XpReward:  15,
			SortOrder: 4,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	for _, lesson := range l0Lessons {
		if err := courseRepo.CreateLesson(lesson); err != nil {
			return err
		}
	}

	// === L1 入门课程 ===
	l1Course := &models.Course{
		Title:       "日常中文入门",
		Description: "学习最常用的中文词汇和基础对话",
		Level:       "L1",
		LevelName:   "入门",
		SortOrder:   2,
		IsPublished: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := courseRepo.Create(l1Course); err != nil {
		return err
	}

	// L1 课时
	l1Lessons := []*models.Lesson{
		{
			CourseID:  l1Course.ID,
			Title:     "第1课：数字1-10",
			Type:      "vocab",
			Content:   `{"introduction": "学习中文数字1-10的写法和读法", "content": "一、二、三、四、五、六、七、八、九、十"}`,
			IsFree:    true,
			XpReward:  15,
			SortOrder: 1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			CourseID:  l1Course.ID,
			Title:     "第2课：基本问候语",
			Type:      "dialogue",
			Content:   `{"introduction": "学习基本的问候和礼貌用语", "content": "你好！谢谢！再见！"}`,
			IsFree:    true,
			XpReward:  20,
			SortOrder: 2,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			CourseID:  l1Course.ID,
			Title:     "第3课：自我介绍",
			Type:      "dialogue",
			Content:   `{"introduction": "学习如何介绍自己", "content": "我叫... 很高兴认识你！"}`,
			IsFree:    false,
			XpReward:  20,
			SortOrder: 3,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	for _, lesson := range l1Lessons {
		if err := courseRepo.CreateLesson(lesson); err != nil {
			return err
		}
	}

	// === L2 初级课程 ===
	l2Course := &models.Course{
		Title:       "初级日常会话",
		Description: "学习常见生活场景的中文表达",
		Level:       "L2",
		LevelName:   "初级",
		SortOrder:   3,
		IsPublished: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := courseRepo.Create(l2Course); err != nil {
		return err
	}

	// L2 课时
	l2Lessons := []*models.Lesson{
		{
			CourseID:  l2Course.ID,
			Title:     "第1课：日常购物",
			Type:      "dialogue",
			Content:   `{"introduction": "学习超市购物和询问价格", "content": "这个多少钱？我要买..."}`,
			IsFree:    true,
			XpReward:  25,
			SortOrder: 1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			CourseID:  l2Course.ID,
			Title:     "第2课：餐厅点餐",
			Type:      "dialogue",
			Content:   `{"introduction": "学习在餐厅点餐的常用语", "content": "我要一份... 这个很好吃！"}`,
			IsFree:    false,
			XpReward:  25,
			SortOrder: 2,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			CourseID:  l2Course.ID,
			Title:     "第3课：问路与交通",
			Type:      "dialogue",
			Content:   `{"introduction": "学习问路和乘坐交通工具", "content": "请问...在哪里？"}`,
			IsFree:    false,
			XpReward:  30,
			SortOrder: 3,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	for _, lesson := range l2Lessons {
		if err := courseRepo.CreateLesson(lesson); err != nil {
			return err
		}
	}

	return nil
}
