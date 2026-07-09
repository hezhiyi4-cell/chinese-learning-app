
# 📱 Flutter 前端 - 安装和启动指南

## ⚠️ 当前状态

**Flutter SDK 尚未安装** - 需要先安装才能运行前端项目

---

## 📥 第一步：安装 Flutter SDK

### Windows 系统安装步骤

1. **下载 Flutter SDK**
   - 访问：https://flutter.dev/docs/get-started/install/windows
   - 下载最新的 Stable 版本（zip 压缩包）
   - 解压到一个不含中文和空格的目录，例如：
     ```
     C:\flutter
     ```

2. **配置环境变量**
   - 右键点击"此电脑" → "属性" → "高级系统设置" → "环境变量"
   - 在"用户变量"中找到"Path"，点击"编辑"
   - 添加 Flutter 的 bin 目录：
     ```
     C:\flutter\bin
     ```

3. **验证安装**
   - 打开新的 PowerShell 或命令提示符窗口
   - 运行：
     ```bash
     flutter --version
     ```
   - 运行 Flutter 医生检查环境：
     ```bash
     flutter doctor
     ```

---

## 🚀 第二步：启动 Flutter 项目

### Flutter 安装完成后，按以下步骤启动：

#### 1️⃣ 安装依赖
```bash
cd c:\Users\user\Desktop\ChineseLearningApp\ai_trae_app
flutter pub get
```

#### 2️⃣ 确保后端正在运行
后端应该已经在 http://localhost:8080 运行中！

#### 3️⃣ 启动应用（选择以下任一方式）

**方式 A：在 Chrome 浏览器中运行（推荐）**
```bash
flutter run -d chrome
```

**方式 B：在 Windows 桌面运行**
```bash
flutter run -d windows
```

**方式 C：查看可用设备**
```bash
flutter devices
```

---

## 📱 应用功能说明

Flutter 应用已实现以下功能：

### 页面流程
1. **启动页** → 2秒后自动跳转
2. **注册/登录页面** → 用户认证
3. **首页** → 显示课程列表
4. **课程详情页** → 显示课时
5. **课时学习页** → 学习内容

---

## 🛠️ 常见问题解决

### 问题 1：`flutter` 命令找不到
**原因**：环境变量未配置
**解决**：确保 `C:\flutter\bin` 已添加到 Path 中，并打开新的终端窗口

### 问题 2：`flutter pub get` 失败
**原因**：网络问题
**解决**：配置国内镜像，或使用：
```bash
flutter pub get --no-offline
```

### 问题 3：连接后端失败
**原因**：后端未启动
**解决**：先启动后端服务器：
```bash
cd c:\Users\user\Desktop\ChineseLearningApp\backend
go run ./cmd/server
```

---

## 📚 项目结构详解

```
ai_trae_app/
├── lib/
│   ├── main.dart                          # App 入口
│   ├── core/
│   │   ├── constants/
│   │   │   └── api_constants.dart        # API 地址常量
│   │   ├── network/
│   │   │   └── dio_client.dart           # Dio 网络请求（自动 JWT）
│   │   └── router/
│   │       └── app_router.dart           # GoRouter 路由配置
│   ├── data/
│   │   ├── models/
│   │   │   ├── user_model.dart
│   │   │   └── course_model.dart
│   │   └── services/
│   │       ├── auth_service.dart
│   │       └── course_service.dart
│   └── features/
│       ├── splash/
│       │   └── splash_screen.dart
│       ├── auth/
│       │   ├── providers/
│       │   │   └── auth_provider.dart
│       │   └── screens/
│       │       ├── login_screen.dart
│       │       └── register_screen.dart
│       ├── home/
│       │   └── home_screen.dart
│       └── courses/
│           ├── providers/
│           │   └── course_provider.dart
│           └── screens/
│               ├── course_detail_screen.dart
│               └── lesson_screen.dart
```

---

## 🎯 下一步

当 Flutter 安装完成后，你就可以：
1. 运行 `flutter pub get` 安装依赖
2. 启动应用并测试完整流程

---

## 💡 提示

- Flutter 安装过程中可能需要一些时间，请耐心等待
- 确保后端服务器正在运行
- 如遇到问题，可以运行 `flutter doctor` 检查环境
