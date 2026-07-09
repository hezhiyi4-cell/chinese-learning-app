
# 中文学习App - AI核心功能集成指南

## 🎉 完成状态

**✅ AI 核心功能已完整集成并运行！**

## 已实现功能

### 1. 数据模型
- `PronunciationResult` - 发音评测结果
- `PronunciationError` - 发音错误详情
- `ChatMessage` - 对话消息
- `ChatResponse` - AI 对话响应
- `Correction` - 语言纠错

### 2. AI Service 层
- `AIService` - AI 服务主类
- `SpeechToText()` - 语音转文字（模拟模式）
- `EvaluatePronunciation()` - 发音评测 + 声调纠错
- `ChatWithTutor()` - AI 助教对话
- `GetAvailableScenes()` - 获取场景列表

### 3. Handler 层
- `SpeechToText()` - 语音转文字 API
- `Evaluate()` - 发音评测 API
- `Chat()` - AI 助教对话 API
- `GetScenes()` - 获取场景列表 API

### 4. Tone Utils
- `ToPinyin()` - 汉字转带声调拼音
- `ExtractCharacters()` - 提取汉字
- `CompareTones()` - 声调对比
- `CalculateScore()` - 发音分数计算

### 5. AI 场景支持
- `free_chat` - 自由对话
- `restaurant` - 餐厅场景
- `airport` - 机场场景
- `hotel` - 酒店场景
- `shopping` - 购物场景
- `interview` - 面试场景
- `hospital` - 医院场景

---

## API 接口文档

### 公开接口（无需 Token）

#### 1. 健康检查
```
GET /health
```

#### 2. 语音转文字
```
POST /api/v1/ai/speech-to-text
Content-Type: multipart/form-data

参数：
- audio: 音频文件
```

#### 3. 发音评测
```
POST /api/v1/ai/evaluate
Content-Type: multipart/form-data

参数：
- audio: 音频文件
- expectedText: 用户应该读的文本

响应：
{
  "transcript": "你好吗",
  "score": 85,
  "errors": [
    {
      "position": 0,
      "character": "你",
      "expected": "nǐ",
      "actual": "ní (声调模拟错误)",
      "errorType": "tone"
    }
  ],
  "feedback": "这是模拟模式下的评测结果。配置 OpenAI API Key 后可以获得真实评测！"
}
```

### 需认证接口（需要 Token）

#### 4. 获取场景列表
```
GET /api/v1/ai/scenes
Authorization: Bearer &lt;YOUR_TOKEN&gt;

响应：
{
  "scenes": [
    {"id": "free_chat", "name": "自由对话", ...},
    ...
  ]
}
```

#### 5. AI 助教对话
```
POST /api/v1/ai/chat
Authorization: Bearer &lt;YOUR_TOKEN&gt;
Content-Type: application/json

{
  "message": "你好，我想点一份炒饭",
  "scene": "restaurant",
  "history": [
    {"role": "user", "content": "..."}
  ]
}

响应：
{
  "reply": "好的，我们有炒饭、面条和饺子...",
  "corrections": []
}
```

---

## 当前运行模式

### 模拟模式（默认）
- 不需要 OpenAI API Key
- 可以测试完整的 API 流程
- 返回预设的模拟响应
- 适合开发和测试阶段

### 真实 AI 模式（待扩展）
- 需要配置 OpenAI API Key
- 真实的 Whisper 语音识别
- 真实的 GPT-4o-mini 对话
- 可以在未来扩展

---

## 测试步骤

### 1. 确保服务器正在运行
服务器现在应该正在 http://localhost:8080 上运行！

### 2. 测试公开接口

#### 健康检查
```bash
GET http://localhost:8080/health
```

#### 获取课程
```bash
GET http://localhost:8080/api/v1/courses
```

### 3. 注册/登录获取 Token

#### 注册
```bash
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json

{
  "email": "you@example.com",
  "password": "123456",
  "nickname": "测试用户"
}
```

#### 登录获取 Token
```bash
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "you@example.com",
  "password": "123456"
}

响应：
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {...}
}
```

### 4. 测试 AI 功能

#### 获取场景列表
```bash
GET http://localhost:8080/api/v1/ai/scenes
Authorization: Bearer &lt;PASTE_TOKEN_HERE&gt;
```

#### AI 对话
```bash
POST http://localhost:8080/api/v1/ai/chat
Authorization: Bearer &lt;PASTE_TOKEN_HERE&gt;
Content-Type: application/json

{
  "message": "你好，我想点一份炒饭",
  "scene": "restaurant",
  "history": []
}
```

---

## 下一步（可选）

### 配置真实 OpenAI API（可选）

如果你想使用真实的 OpenAI AI 功能：

1. 获取 OpenAI API Key：
   - 访问 https://platform.openai.com/account/api-keys
   - 创建新的 API Key

2. 在项目根目录创建 `.env` 文件：
```bash
cd C:\Users\user\Desktop\ChineseLearningApp\backend
copy .env.example .env
notepad .env
```

3. 在 `.env` 文件中填入你的 Key：
```
OPENAI_API_KEY=sk-...
```

4. 重新启动服务器

5. 修改 `internal/services/ai_service.go`，接入真实的 OpenAI 接口

---

## 文件结构

```
backend/
├── cmd/server/main.go            # 主程序
├── internal/
│   ├── config/config.go          # 配置管理
│   ├── handlers/
│   │   ├── auth_handler.go       # 认证接口
│   │   ├── course_handler.go     # 课程接口
│   │   └── ai_handler.go         # AI 接口
│   ├── services/
│   │   ├── auth_service.go       # 认证服务
│   │   ├── course_service.go     # 课程服务
│   │   ├── progress_service.go   # 进度服务
│   │   └── ai_service.go         # AI 服务
│   ├── repositories/
│   │   ├── user_repository.go    # 用户仓储
│   │   ├── course_repository.go  # 课程仓储
│   │   └── progress_repository.go# 进度仓储
│   ├── models/
│   │   ├── user.go
│   │   ├── course.go
│   │   ├── lesson.go
│   │   └── user_progress.go
│   ├── middleware/auth.go        # JWT 中间件
│   └── utils/tone_utils.go       # 声调工具
├── test/
│   ├── auth_test.http
│   ├── course_test.http
│   └── ai_test.http
├── .env.example                  # 环境变量模板
├── go.mod
└── go.sum
```

---

## 完整项目功能总结

✅ 用户系统（注册/登录/Token 认证）
✅ 课程系统（L0-L2 级课程 + 课时）
✅ 学习进度追踪
✅ 发音评测（模拟模式）
✅ AI 助教对话（多场景支持）
✅ 声调对比与纠错
✅ 完整的 RESTful API

---

## 准备好开发前端了！

现在后端已经非常完整了！下一步可以开始开发 Flutter 前端界面，或者继续完善后端功能！
