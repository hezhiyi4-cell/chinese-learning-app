
# 外国人学中文 - 多端游戏化学习APP

## 项目简介
这是一个面向外国人的中文学习平台，结合游戏化学习体验，支持课程学习、学习进度、管理后台、封面上传，并提供一个可直接运行的 `web.html` 演示前端。

## 项目架构
- **前端（演示版）**: 纯 HTML/CSS/JS（`web.html` + `config.js`）
- **后端**: Golang + Gin
- **数据库**: SQLite（GORM，纯 Go 驱动）
- **文件上传**: 本地 `backend/uploads/`（课程封面）
- **AI（可选）**: OpenAI（未配置 Key 时可使用本地模拟模式）

## 部署
- Render 后端 + Vercel 前端：见 [DEPLOY_RENDER_VERCEL.md](file:///c:/Users/user/Desktop/ChineseLearningApp/DEPLOY_RENDER_VERCEL.md)

## 开发阶段
- [ ] 阶段0：项目初始化与基础架构
- [ ] 阶段1：用户与课程核心系统
- [ ] 阶段2：AI核心能力集成
- [ ] 阶段3：游戏化与奖励系统
- [ ] 阶段4：多端UI与实时功能
- [ ] 阶段5：测试、部署与上线
