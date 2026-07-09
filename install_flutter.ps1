
# Flutter 一键安装脚本
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  中文学习App - Flutter 安装脚本" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 配置国内镜像
Write-Host "1. 配置国内镜像..." -ForegroundColor Yellow
$env:PUB_HOSTED_URL = "https://pub.flutter-io.cn"
$env:FLUTTER_STORAGE_BASE_URL = "https://storage.flutter-io.cn"
Write-Host "   镜像已配置" -ForegroundColor Green

# 检查Flutter是否已安装
Write-Host ""
Write-Host "2. 检查Flutter..." -ForegroundColor Yellow
$flutterPath = "C:\flutter\bin\flutter.bat"
if (Test-Path $flutterPath) {
    Write-Host "   Flutter 已安装！" -ForegroundColor Green
    &amp; $flutterPath --version
} else {
    Write-Host "   Flutter 未安装，开始下载..." -ForegroundColor Yellow
    
    # 下载Flutter
    $flutterUrl = "https://storage.googleapis.com/flutter_infra_release/releases/stable/windows/flutter_windows_3.22.0-stable.zip"
    $zipPath = "$env:USERPROFILE\Downloads\flutter.zip"
    
    Write-Host "   下载中，请稍候..." -ForegroundColor Yellow
    Invoke-WebRequest -Uri $flutterUrl -OutFile $zipPath
    
    Write-Host "   解压到 C:\flutter..." -ForegroundColor Yellow
    Expand-Archive -Path $zipPath -DestinationPath "C:\" -Force
    
    Write-Host "   Flutter 下载完成！" -ForegroundColor Green
}

Write-Host ""
Write-Host "3. 配置环境变量..." -ForegroundColor Yellow
# 临时设置Path
$env:Path = "C:\flutter\bin;" + $env:Path

# 检查是否需要永久配置
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $currentPath.Contains("C:\flutter\bin")) {
    Write-Host "   正在添加 Flutter 到用户环境变量..." -ForegroundColor Yellow
    [Environment]::SetEnvironmentVariable("Path", "C:\flutter\bin;" + $currentPath, "User")
    Write-Host "   环境变量已配置" -ForegroundColor Green
    Write-Host ""
    Write-Host "⚠️  请关闭当前终端，重新打开一个新的终端，然后继续！" -ForegroundColor Red
    Write-Host ""
} else {
    Write-Host "   Flutter 已在 Path 中" -ForegroundColor Green
}

# 配置镜像（用户环境变量）
[Environment]::SetEnvironmentVariable("PUB_HOSTED_URL", "https://pub.flutter-io.cn", "User")
[Environment]::SetEnvironmentVariable("FLUTTER_STORAGE_BASE_URL", "https://storage.flutter-io.cn", "User")

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  安装完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "接下来：" -ForegroundColor Yellow
Write-Host " 1. 请关闭所有终端窗口，重新打开一个新的PowerShell"
Write-Host " 2. 运行: cd c:\Users\user\Desktop\ChineseLearningApp\ai_trae_app"
Write-Host " 3. 运行: flutter pub get"
Write-Host " 4. 运行: flutter run -d chrome"
Write-Host ""
Write-Host "或者，可以直接运行 project_start.ps1 来一键启动项目！" -ForegroundColor Cyan
Write-Host ""
