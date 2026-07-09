# 部署指南（Render 后端 + Vercel 前端）

本指南用于把当前项目的：
- 后端（Go + Gin）部署到 Render
- 前端（静态 `web.html`）部署到 Vercel

最终实现：用户通过 Vercel 域名访问网页，网页通过 HTTPS 调用 Render 的后端 API。

## 0. 重要限制（先读）

### Render 免费版文件不会持久化

本项目当前使用：
- SQLite 数据库文件（默认 `chinese_learning.db`）
- 本地上传目录（默认 `./uploads`，封面上传会存这里）

在 Render 免费 Web Service 上，这些本地文件在以下情况下可能丢失：
- 服务空闲休眠后再唤醒
- 服务重启
- 重新部署

如果你要长期稳定保存数据和封面：
- 需要升级到可挂载 Persistent Disk 的 Render 付费计划，或
- 后续迁移到云数据库（如 Postgres）+ 对象存储（如 S3/Cloudflare R2）

### AI 功能可选

如果你不配置 `OPENAI_API_KEY`，AI 相关接口可能不可用；课程/登录/后台/封面等功能不受影响。

## 1. 推送代码到 GitHub（必做）

在本机项目目录执行：

```powershell
cd C:\Users\user\Desktop\ChineseLearningApp
git status -sb
git remote -v
git push -u origin main
```

如果 `git push` 卡住或报错：
- 通常是 GitHub 登录授权没有完成（系统会弹出授权窗口或打开浏览器授权页）
- 或者使用了错误的仓库地址/账号无权限

## 2. 部署后端到 Render（推荐 Blueprint）

项目根目录已经包含 `render.yaml`，可以用 Blueprint 一键创建服务。

### 2.1 创建服务

1. 打开 Render Dashboard
2. New + → Blueprint
3. 选择你的 GitHub 仓库 `hezhiyi4-cell/chinese-learning-app`
4. 一路确认创建

### 2.2 Render 环境变量（默认已在 render.yaml 配好）

关键环境变量：
- `JWT_SECRET`（建议随机强密码）
- `DEFAULT_ADMIN_EMAIL` / `DEFAULT_ADMIN_PASSWORD`
- `SQLITE_PATH`（默认 `chinese_learning.db`）
- `UPLOAD_DIR`（默认 `./uploads`）
- `OPENAI_API_KEY`（可选）

### 2.3 验证后端

部署成功后，Render 会给一个 URL，例如：

```
https://YOUR_BACKEND.onrender.com
```

验证：
- `GET https://YOUR_BACKEND.onrender.com/health`
- `GET https://YOUR_BACKEND.onrender.com/api/v1/courses`

## 3. 配置前端指向线上后端

前端不再写死 `localhost`，而是通过根目录的 `config.js` 配置 API 地址。

修改文件：
- `config.js`

把：

```js
apiBase: "http://localhost:8080/api/v1"
```

改成：

```js
apiBase: "https://YOUR_BACKEND.onrender.com/api/v1"
```

然后提交并推送：

```powershell
cd C:\Users\user\Desktop\ChineseLearningApp
git add .
git commit -m "chore: set production api base"
git push
```

## 4. 部署前端到 Vercel

项目根目录已经包含 `vercel.json`，会把根路径 `/` 重写到 `/web.html`。

### 4.1 创建 Vercel 项目

1. 打开 Vercel Dashboard
2. Add New → Project
3. 导入 GitHub 仓库 `hezhiyi4-cell/chinese-learning-app`
4. Framework 选 `Other`
5. Build Command / Output Directory 留空
6. Deploy

### 4.2 线上验证

拿到 Vercel 域名后依次验证：
- 首页课程能加载
- 注册/登录可用
- 课程详情可用
- 管理员后台可用（默认账号见环境变量）
- 课程封面显示/上传可用

## 5. 常见问题

### 5.1 页面能打开但登录/课程加载失败

优先检查：
- `config.js` 的 `apiBase` 是否已改成 Render 域名
- Render 后端是否已成功部署并可访问 `/health`

### 5.2 首次访问后端很慢

Render 免费版空闲会休眠，首次唤醒会慢一些，属于正常现象。

### 5.3 上传封面成功后过段时间消失

这是 Render 免费版临时文件系统导致的预期行为。
