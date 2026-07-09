
# 数据库设计文档

## 1. 用户相关表

### users 表
| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | BIGINT | 用户ID | PRIMARY KEY, AUTO INCREMENT |
| email | VARCHAR(255) | 邮箱 | UNIQUE, NOT NULL |
| password_hash | VARCHAR(255) | 密码哈希 | NOT NULL |
| nickname | VARCHAR(100) | 昵称 | |
| level | INT | 等级 | DEFAULT 1 |
| experience | INT | 经验值 | DEFAULT 0 |
| avatar_url | VARCHAR(500) | 头像URL | |
| preferred_language | VARCHAR(10) | 偏好语言 | DEFAULT 'en' |
| created_at | TIMESTAMP | 创建时间 | DEFAULT NOW() |
| updated_at | TIMESTAMP | 更新时间 | DEFAULT NOW() |

### user_profiles 表
| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | BIGINT | 配置ID | PRIMARY KEY, AUTO INCREMENT |
| user_id | BIGINT | 用户ID | FOREIGN KEY |
| bio | TEXT | 个人简介 | |
| country | VARCHAR(100) | 国家 | |
| learning_goal | VARCHAR(255) | 学习目标 | |
| daily_goal_minutes | INT | 每日目标分钟 | DEFAULT 30 |

---

## 2. 课程相关表

### courses 表
| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | BIGINT | 课程ID | PRIMARY KEY, AUTO INCREMENT |
| level | VARCHAR(20) | 等级 | L0/L1/L2/L3/L4 |
| title | VARCHAR(255) | 标题 | NOT NULL |
| description | TEXT | 描述 | |
| thumbnail_url | VARCHAR(500) | 缩略图 | |
| order_index | INT | 排序 | |
| is_premium | BOOLEAN | 是否付费 | DEFAULT FALSE |

### lessons 表
| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | BIGINT | 课时ID | PRIMARY KEY, AUTO INCREMENT |
| course_id | BIGINT | 课程ID | FOREIGN KEY |
| title | VARCHAR(255) | 标题 | NOT NULL |
| content | TEXT | 内容 | |
| video_url | VARCHAR(500) | 视频URL | |
| order_index | INT | 排序 | |

### vocabulary 表
| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | BIGINT | 词汇ID | PRIMARY KEY, AUTO INCREMENT |
| lesson_id | BIGINT | 课时ID | FOREIGN KEY |
| chinese_word | VARCHAR(100) | 中文词 | NOT NULL |
| pinyin | VARCHAR(100) | 拼音 | |
| translation | VARCHAR(255) | 翻译 | |
| audio_url | VARCHAR(500) | 音频URL | |
| image_url | VARCHAR(500) | 图片URL | |

---

## 3. 学习进度表

### user_lesson_progress 表
| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | BIGINT | 进度ID | PRIMARY KEY, AUTO INCREMENT |
| user_id | BIGINT | 用户ID | FOREIGN KEY |
| lesson_id | BIGINT | 课时ID | FOREIGN KEY |
| completed | BOOLEAN | 是否完成 | DEFAULT FALSE |
| score | INT | 得分 | |
| started_at | TIMESTAMP | 开始时间 | |
| completed_at | TIMESTAMP | 完成时间 | |

### user_vocabulary 表
| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | BIGINT | ID | PRIMARY KEY, AUTO INCREMENT |
| user_id | BIGINT | 用户ID | FOREIGN KEY |
| vocabulary_id | BIGINT | 词汇ID | FOREIGN KEY |
| mastery_level | INT | 掌握程度 | 0-5 |
| review_count | INT | 复习次数 | DEFAULT 0 |
| last_reviewed | TIMESTAMP | 最后复习 | |

---

## 4. 游戏化相关表

### badges 表
| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | BIGINT | 徽章ID | PRIMARY KEY, AUTO INCREMENT |
| name | VARCHAR(100) | 名称 | NOT NULL |
| description | TEXT | 描述 | |
| icon_url | VARCHAR(500) | 图标 | |
| condition_type | VARCHAR(50) | 条件类型 | |
| condition_value | INT | 条件值 | |

### user_badges 表
| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | BIGINT | ID | PRIMARY KEY, AUTO INCREMENT |
| user_id | BIGINT | 用户ID | FOREIGN KEY |
| badge_id | BIGINT | 徽章ID | FOREIGN KEY |
| earned_at | TIMESTAMP | 获得时间 | DEFAULT NOW() |

### leaderboard 表（Redis Sorted Set）
- 使用 Redis 存储排行榜

---

## 5. 社群相关表

### posts 表
| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | BIGINT | 帖子ID | PRIMARY KEY, AUTO INCREMENT |
| user_id | BIGINT | 用户ID | FOREIGN KEY |
| content | TEXT | 内容 | NOT NULL |
| created_at | TIMESTAMP | 创建时间 | DEFAULT NOW() |

### comments 表
| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| id | BIGINT | 评论ID | PRIMARY KEY, AUTO INCREMENT |
| post_id | BIGINT | 帖子ID | FOREIGN KEY |
| user_id | BIGINT | 用户ID | FOREIGN KEY |
| content | TEXT | 内容 | NOT NULL |
| created_at | TIMESTAMP | 创建时间 | DEFAULT NOW() |
