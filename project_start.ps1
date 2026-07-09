
# 中文学习App - 一键启动脚本
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  中文学习App - 一键启动" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 配置环境
$env:PUB_HOSTED_URL = "https://pub.flutter-io.cn"
$env:FLUTTER_STORAGE_BASE_URL = "https://storage.flutter-io.cn"

# 确保Flutter在Path中
if (-not $env:Path.Contains("C:\flutter\bin")) {
    $env:Path = "C:\flutter\bin;" + $env:Path
}

Write-Host "1. 检查后端..." -ForegroundColor Yellow
# 检查后端是否正在运行
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 2
    if ($response.StatusCode -eq 200) {
        Write-Host "   ✅ 后端已在运行！" -ForegroundColor Green
    }
} catch {
    Write-Host "   ⚠️  后端未运行，正在启动..." -ForegroundColor Yellow
    # 启动后端（新窗口）
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd c:\Users\user\Desktop\ChineseLearningApp\backend; go run ./cmd/server"
    Start-Sleep -Seconds 5
    Write-Host "   ✅ 后端启动中..." -ForegroundColor Green
}

Write-Host ""
Write-Host "2. 启动Flutter前端..." -ForegroundColor Yellow
Set-Location c:\Users\user\Desktop\ChineseLearningApp\ai_trae_app

Write-Host "   a. 安装依赖..." -ForegroundColor Yellow
&amp; flutter pub get

Write-Host ""
Write-Host "   b. 启动应用（Chrome浏览器）..." -ForegroundColor Yellow
&amp; flutter run -d chrome

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  应用已启动！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
