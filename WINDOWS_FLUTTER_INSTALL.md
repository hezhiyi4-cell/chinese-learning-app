
# Windows Flutter 安装完整指南

## 方法一：使用已经下载的 Flutter（最快）

### 重要提示：你的电脑里已经有 Flutter SDK 了！！！

位置：`c:\Users\user\Desktop\ChineseLearningApp\flutter`

### 步骤1：安装 Git（必需）

如果 Git 还没安装，请先安装：
- 下载地址：https://github.com/git-for-windows/git/releases/download/v2.55.0.windows.2/Git-2.55.0.windows.1-64-bit.exe
- 或者用命令（winget）：
  ```powershell
  winget install --id Git.Git -e --accept-package-agreements --accept-source-agreements
  ```
安装时一直选默认即可！

### 步骤2：把 Flutter 移动到 C:\flutter

把 `c:\Users\user\Desktop\ChineseLearningApp\flutter` 复制到 `C:\flutter`

### 步骤3：配置环境变量

1. 按 `Win + X`，选择"系统"
2. 选择"高级系统设置"
3. 点击"环境变量"按钮
4. 在"用户变量"区域：
   - 找到 `Path`，点击"编辑"
   - 点击"新建"
   - 输入：`C:\flutter\bin`
   - 点击"确定"

5. 同时在"用户变量"里添加：
   - 点击"新建"
   - 变量名：`PUB_HOSTED_URL`
   - 变量值：`https://pub.flutter-io.cn`
   - 点击"确定"
   
   再点击"新建"
   - 变量名：`FLUTTER_STORAGE_BASE_URL`
   - 变量值：`https://storage.flutter-io.cn`
   - 点击"确定"

### 步骤4：关闭所有终端，重新打开 PowerShell

### 步骤5：验证安装

在新的 PowerShell 窗口运行：
```powershell
flutter --version
```

如果看到版本号，说明成功！

### 步骤6：运行 Flutter Doctor

```powershell
flutter doctor
```

这会告诉你还有什么需要做的（比如 Android Studio 等，但对于 Web 开发不需要）

---

## 方法二：重新下载（如果方法一不行）

### 下载地址

Flutter Stable 版本（最新）：
https://storage.flutter-io.cn/flutter_infra_release/releases/stable/windows/flutter_windows_3.22.0-stable.zip

### 解压位置

把 zip 解压到：`C:\flutter`

确保路径是：`C:\flutter\bin\flutter.bat`

### 然后按照上面的方法一继续！

---

## 启动项目

Flutter 安装好后，运行：

```powershell
# 进入项目目录
cd c:\Users\user\Desktop\ChineseLearningApp\ai_trae_app

# 安装依赖
flutter pub get

# 运行 Web 版本（Chrome）
flutter run -d chrome
```

---

## 常见问题

### flutter doctor 提示缺少 Android Studio
- 对于 Web 和 Windows 开发，不需要 Android Studio
- 如果要开发 Android App，再去下载安装

### flutter pub get 下载慢
- 确保你配置了镜像（PUB_HOSTED_URL 和 FLUTTER_STORAGE_BASE_URL）
- 或者使用 VPN

### 端口被占用
- 如果提示端口被占用，运行其他 Chrome 实例：
  ```powershell
  flutter run -d chrome --web-port 8081
  ```

---

## 最终验证

打开浏览器，访问：
- 后端API：http://localhost:8080/api/v1/courses
- 前端（Flutter）：http://localhost:（端口，运行后会显示）

完成！
