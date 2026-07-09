
# 最简单的Flutter下载脚本
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Flutter 下载器" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$flutterDir = "C:\flutter"
$zipPath = "$env:USERPROFILE\Downloads\flutter.zip"

# 检查Flutter是否已安装
if (Test-Path "$flutterDir\bin\flutter.bat") {
    Write-Host "✅ Flutter 已安装！" -ForegroundColor Green
    Write-Host ""
    
    # 测试运行
    $env:Path = "$flutterDir\bin;" + $env:Path
    $env:PUB_HOSTED_URL = "https://pub.flutter-io.cn"
    $env:FLUTTER_STORAGE_BASE_URL = "https://storage.flutter-io.cn"
    
    Write-Host "正在启动项目..." -ForegroundColor Yellow
    Set-Location c:\Users\user\Desktop\ChineseLearningApp\ai_trae_app
    
    Write-Host "1. 安装依赖..." -ForegroundColor Yellow
    Start-Process -FilePath "$flutterDir\bin\flutter.bat" -ArgumentList "pub", "get" -NoNewWindow -Wait
    
    Write-Host "2. 启动应用..." -ForegroundColor Yellow
    Start-Process -FilePath "$flutterDir\bin\flutter.bat" -ArgumentList "run", "-d", "chrome"
    exit 0
}

Write-Host "📥 Flutter 未安装，准备下载..." -ForegroundColor Yellow

# 下载
Write-Host "下载地址: https://storage.flutter-io.cn/flutter_infra_release/releases/stable/windows/flutter_windows_3.22.0-stable.zip" -ForegroundColor Gray
Write-Host ""
Write-Host "正在下载..." -ForegroundColor Yellow
Write-Host "请等待，这可能需要5-10分钟..." -ForegroundColor Gray

$ProgressPreference = 'SilentlyContinue'
Invoke-WebRequest -Uri "https://storage.flutter-io.cn/flutter_infra_release/releases/stable/windows/flutter_windows_3.22.0-stable.zip" -OutFile $zipPath -UseBasicParsing

Write-Host "✅ 下载完成！" -ForegroundColor Green
Write-Host ""

# 解压
Write-Host "📦 正在解压到 C:\flutter..." -ForegroundColor Yellow
if (Test-Path $flutterDir) {
    Remove-Item -Path $flutterDir -Recurse -Force -ErrorAction SilentlyContinue
}
Expand-Archive -Path $zipPath -DestinationPath "C:\" -Force

Write-Host "✅ 解压完成！" -ForegroundColor Green
Write-Host ""

# 配置环境变量
Write-Host "⚙️  正在配置环境变量..." -ForegroundColor Yellow
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $userPath.Contains("$flutterDir\bin")) {
    [Environment]::SetEnvironmentVariable("Path", "$flutterDir\bin;" + $userPath, "User")
}
[Environment]::SetEnvironmentVariable("PUB_HOSTED_URL", "https://pub.flutter-io.cn", "User")
[Environment]::SetEnvironmentVariable("FLUTTER_STORAGE_BASE_URL", "https://storage.flutter-io.cn", "User")

Write-Host "✅ 环境变量配置完成！" -ForegroundColor Green
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  安装完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "⚠️  请先关闭所有终端窗口！" -ForegroundColor Red
Write-Host "然后重新打开，再运行 download_flutter.ps1" -ForegroundColor Yellow
Write-Host ""
Write-Host "或者手动运行：" -ForegroundColor Cyan
Write-Host "cd c:\Users\user\Desktop\ChineseLearningApp\ai_trae_app" -ForegroundColor Gray
Write-Host "flutter pub get" -ForegroundColor Gray
Write-Host "flutter run -d chrome" -ForegroundColor Gray
Write-Host ""
Read-Host "按任意键退出"
