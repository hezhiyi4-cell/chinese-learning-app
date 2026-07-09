
# Flutter 最简单下载安装脚本
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  中文学习App - Flutter 简易安装" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. 配置国内镜像
Write-Host "1. 配置国内镜像..." -ForegroundColor Yellow
$env:PUB_HOSTED_URL = "https://pub.flutter-io.cn"
$env:FLUTTER_STORAGE_BASE_URL = "https://storage.flutter-io.cn"
[Environment]::SetEnvironmentVariable("PUB_HOSTED_URL", $env:PUB_HOSTED_URL, "User")
[Environment]::SetEnvironmentVariable("FLUTTER_STORAGE_BASE_URL", $env:FLUTTER_STORAGE_BASE_URL, "User")
Write-Host "   镜像配置完成！" -ForegroundColor Green

# 2. 检查Flutter是否已安装
Write-Host ""
Write-Host "2. 检查Flutter..." -ForegroundColor Yellow
$flutterDir = "C:\flutter"
if (Test-Path "$flutterDir\bin\flutter.bat") {
    Write-Host "   ✅ Flutter 已安装！" -ForegroundColor Green
    
    # 直接添加到当前Path
    $env:Path = "$flutterDir\bin;" + $env:Path
    
    Write-Host ""
    Write-Host "3. 尝试启动项目..." -ForegroundColor Yellow
    Set-Location c:\Users\user\Desktop\ChineseLearningApp\ai_trae_app
    Write-Host "   a. 安装依赖..." -ForegroundColor Yellow
    &amp; "$flutterDir\bin\flutter.bat" pub get
    Write-Host ""
    Write-Host "   b. 启动应用..." -ForegroundColor Yellow
    &amp; "$flutterDir\bin\flutter.bat" run -d chrome
    exit 0
}

# 3. 下载Flutter
Write-Host "   Flutter 未安装，准备下载..." -ForegroundColor Yellow
Write-Host "   下载地址: https://storage.flutter-io.cn/flutter_infra_release/releases/stable/windows/flutter_windows_3.22.0-stable.zip" -ForegroundColor Gray

$zipPath = "$env:USERPROFILE\Downloads\flutter_sdk.zip"

Write-Host ""
Write-Host "3. 开始下载..." -ForegroundColor Yellow
Write-Host "   这可能需要5-10分钟，请耐心等待..." -ForegroundColor Gray

try {
    Invoke-WebRequest -Uri "https://storage.flutter-io.cn/flutter_infra_release/releases/stable/windows/flutter_windows_3.22.0-stable.zip" -OutFile $zipPath -UseBasicParsing
    Write-Host "   ✅ 下载完成！" -ForegroundColor Green
} catch {
    Write-Host "   ❌ 下载失败！" -ForegroundColor Red
    Write-Host "   请手动下载: https://storage.flutter-io.cn/flutter_infra_release/releases/stable/windows/flutter_windows_3.22.0-stable.zip" -ForegroundColor Yellow
    Write-Host "   解压到: C:\flutter" -ForegroundColor Yellow
    Read-Host "   按任意键退出"
    exit 1
}

# 4. 解压
Write-Host ""
Write-Host "4. 解压文件..." -ForegroundColor Yellow
Write-Host "   这可能需要几分钟..." -ForegroundColor Gray
if (Test-Path $flutterDir) {
    Remove-Item -Path $flutterDir -Recurse -Force -ErrorAction SilentlyContinue
}
Expand-Archive -Path $zipPath -DestinationPath "C:\" -Force

Write-Host "   ✅ 解压完成！" -ForegroundColor Green

# 5. 配置环境变量
Write-Host ""
Write-Host "5. 配置环境变量..." -ForegroundColor Yellow
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $userPath.Contains("$flutterDir\bin")) {
    [Environment]::SetEnvironmentVariable("Path", "$flutterDir\bin;" + $userPath, "User")
    Write-Host "   ✅ 环境变量已配置！" -ForegroundColor Green
} else {
    Write-Host "   环境变量已存在！" -ForegroundColor Green
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  安装完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "⚠️  重要：请执行以下操作" -ForegroundColor Red
Write-Host "  1. 关闭所有 PowerShell/终端窗口" -ForegroundColor Yellow
Write-Host "  2. 重新打开一个新的 PowerShell 窗口" -ForegroundColor Yellow
Write-Host "  3. 运行: simple_install.ps1" -ForegroundColor Yellow
Write-Host ""
Write-Host "或者，你可以直接手动运行：" -ForegroundColor Cyan
Write-Host "  cd c:\Users\user\Desktop\ChineseLearningApp\ai_trae_app" -ForegroundColor Gray
Write-Host "  flutter pub get" -ForegroundColor Gray
Write-Host "  flutter run -d chrome" -ForegroundColor Gray
Write-Host ""
Read-Host "按任意键退出"
