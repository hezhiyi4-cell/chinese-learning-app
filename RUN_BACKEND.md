
# 运行后端服务器

## ✅ 编译成功！

后端代码已经成功编译为 `server.exe`

---

## 如何启动服务器

### 方法1：直接运行可执行文件（Windows）
```powershell
cd c:\Users\user\Desktop\ChineseLearningApp\backend
.\server.exe
```

### 方法2：使用 Go 直接运行（开发时推荐）
```powershell
cd c:\Users\user\Desktop\ChineseLearningApp\backend
go run ./cmd/server
```

---

## 测试服务器

服务器启动后，访问以下地址测试：

- **健康检查**：http://localhost:8080/health
- **API v1**：http://localhost:8080/api/v1

---

## 下一步

1. 启动服务器并测试
2. 我们可以继续添加更多功能（用户登录、课程等）
3. 等你准备好 Flutter 后，我们可以开始开发前端
