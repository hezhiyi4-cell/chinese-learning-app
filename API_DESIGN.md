
# API 接口设计文档

## 基础信息
- Base URL: `https://api.chineselearning.com/v1`
- 认证方式: JWT Bearer Token

---

## 1. 用户认证接口

### 1.1 注册
- **POST** `/auth/register`
- Request Body:
  ```json
  {
    "email": "user@example.com",
    "password": "password123",
    "nickname": "Learning Chinese"
  }
  ```
- Response:
  ```json
  {
    "success": true,
    "token": "jwt_token_here",
    "user": {
      "id": 1,
      "email": "user@example.com",
      "nickname": "Learning Chinese"
    }
  }
  ```

### 1.2 登录
- **POST** `/auth/login`
- Request Body:
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```
- Response:
  ```json
  {
    "success": true,
    "token": "jwt_token_here",
    "user": {
      "id": 1,
      "email": "user@example.com",
      "nickname": "Learning Chinese",
      "level": 1,
      "experience": 0
    }
  }
  ```

### 1.3 获取用户信息
- **GET** `/auth/me`
- Headers: `Authorization: Bearer &lt;token&gt;`
- Response: User object

---

## 2. 课程接口

### 2.1 获取课程列表
- **GET** `/courses`
- Query Params: `?level=L1`
- Response:
  ```json
  {
    "courses": [
      {
        "id": 1,
        "level": "L1",
        "title": "Greetings and Basics",
        "description": "Learn basic Chinese greetings",
        "is_premium": false
      }
    ]
  }
  ```

### 2.2 获取课程详情
- **GET** `/courses/:id`
- Response: Course object with lessons

### 2.3 获取课时详情
- **GET** `/lessons/:id`
- Response: Lesson object with vocabulary

---

## 3. 学习进度接口

### 3.1 更新学习进度
- **POST** `/lessons/:id/progress`
- Request Body:
  ```json
  {
    "completed": true,
    "score": 95
  }
  ```

### 3.2 获取用户学习进度
- **GET** `/users/me/progress`
- Response: Learning progress summary

---

## 4. AI 接口

### 4.1 发音评测
- **POST** `/ai/evaluate-pronunciation`
- Request: FormData with audio file
- Response:
  ```json
  {
    "transcript": "你好",
    "score": 85,
    "tone_score": 90,
    "errors": [
      {
        "position": 0,
        "error": "Third tone should be lower",
        "suggestion": "Try lowering your pitch"
      }
    ]
  }
  ```

### 4.2 AI 对话
- **POST** `/ai/chat`
- Request Body:
  ```json
  {
    "message": "Hello, how are you?",
    "context": "daily_conversation"
  }
  ```
- Response:
  ```json
  {
    "reply": "你好！我很好，谢谢。",
    "pinyin": "Nǐ hǎo! Wǒ hěn hǎo, xièxie.",
    "translation": "Hello! I'm fine, thank you.",
    "explanation": "Breakdown of the sentence..."
  }
  ```

---

## 5. 游戏化接口

### 5.1 获取排行榜
- **GET** `/leaderboard`
- Query Params: `?type=global`
- Response:
  ```json
  {
    "leaderboard": [
      {
        "rank": 1,
        "user": { "nickname": "User1", "avatar_url": "..." },
        "score": 15000
      }
    ]
  }
  ```

### 5.2 获取用户徽章
- **GET** `/users/me/badges`
- Response: Array of badges

---

## WebSocket 接口

### 实时聊天
- Endpoint: `/ws/chat`
- Events:
  - `join_room`
  - `send_message`
  - `receive_message`
