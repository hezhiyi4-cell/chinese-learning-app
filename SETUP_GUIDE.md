
# 开发工具安装指南

## 一、本地开发环境（推荐先试试这个）

### 1. 安装 Go
- 下载地址：https://golang.org/dl/
- 选择 Windows 版本（go1.21.x.windows-amd64.msi）
- 安装后打开新终端，输入 `go version` 验证

### 2. 安装 Flutter
- 下载地址：https://flutter.dev/docs/get-started/install/windows
- 下载 Flutter SDK zip 包
- 解压到 `C:\flutter`
- 配置环境变量：
  - 将 `C:\flutter\bin` 添加到 PATH
- 打开新终端，输入 `flutter --version` 验证
- 运行 `flutter doctor` 检查环境

### 3. 安装代码编辑器
- VS Code（推荐）：https://code.visualstudio.com/
- 安装 Flutter 扩展

### 4. 数据库（可选，前期可以用云数据库）
- PostgreSQL：https://www.postgresql.org/download/windows/

---

## 二、云端开发环境（备选方案）

### GitHub Codespaces
1. 注册 GitHub 账号
2. 创建项目仓库
3. 使用 Codespaces 在浏览器中开发
4. 免费额度足够前期开发

### GitPod
类似 Codespaces，也有免费额度

---

## 三、先做什么？

**推荐方案**：
1. 我们先继续完善项目文档和设计（不写代码先设计）
2. 你可以慢慢决定是否安装工具或用云端
3. 设计完成后，我们再开始写代码

这样对你的电脑配置没有任何压力！
