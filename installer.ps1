
# Flutter Installer - Project Folder Version
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Flutter Installer" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$ProjectDir = "c:\Users\user\Desktop\ChineseLearningApp"
$FlutterDir = "$ProjectDir\flutter"
$ZipPath = "$ProjectDir\flutter.zip"

# Check if Flutter is already installed
if (Test-Path "$FlutterDir\bin\flutter.bat") {
    Write-Host "Flutter is already installed!" -ForegroundColor Green
    Write-Host ""
    
    # Set environment
    $env:Path = "$FlutterDir\bin;" + $env:Path
    $env:PUB_HOSTED_URL = "https://pub.flutter-io.cn"
    $env:FLUTTER_STORAGE_BASE_URL = "https://storage.flutter-io.cn"
    
    Write-Host "Starting project..." -ForegroundColor Yellow
    Set-Location $ProjectDir\ai_trae_app
    
    Write-Host "Step 1: Install dependencies..." -ForegroundColor Yellow
    Start-Process -FilePath "$FlutterDir\bin\flutter.bat" -ArgumentList "pub", "get" -NoNewWindow -Wait
    
    Write-Host "Step 2: Launch app..." -ForegroundColor Yellow
    Start-Process -FilePath "$FlutterDir\bin\flutter.bat" -ArgumentList "run", "-d", "chrome"
    exit 0
}

Write-Host "Downloading Flutter..." -ForegroundColor Yellow
Write-Host "Please wait, this may take 5-10 minutes..." -ForegroundColor Gray

$ProgressPreference = 'SilentlyContinue'
Invoke-WebRequest -Uri "https://storage.flutter-io.cn/flutter_infra_release/releases/stable/windows/flutter_windows_3.22.0-stable.zip" -OutFile $ZipPath -UseBasicParsing

Write-Host "Download complete!" -ForegroundColor Green
Write-Host ""

Write-Host "Extracting files..." -ForegroundColor Yellow
if (Test-Path $FlutterDir) {
    Remove-Item -Path $FlutterDir -Recurse -Force -ErrorAction SilentlyContinue
}
Expand-Archive -Path $ZipPath -DestinationPath $ProjectDir -Force

# The extracted folder is called "flutter" already, so it should be at $FlutterDir
Write-Host "Extraction complete!" -ForegroundColor Green
Write-Host ""

Write-Host "Setting up environment..." -ForegroundColor Yellow
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $UserPath.Contains("$FlutterDir\bin")) {
    [Environment]::SetEnvironmentVariable("Path", "$FlutterDir\bin;" + $UserPath, "User")
}
[Environment]::SetEnvironmentVariable("PUB_HOSTED_URL", "https://pub.flutter-io.cn", "User")
[Environment]::SetEnvironmentVariable("FLUTTER_STORAGE_BASE_URL", "https://storage.flutter-io.cn", "User")

Write-Host "Environment configured!" -ForegroundColor Green
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Installation Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "IMPORTANT: Close all terminal windows first!" -ForegroundColor Red
Write-Host "Then re-open and run installer.ps1 again!" -ForegroundColor Yellow
Write-Host ""
Write-Host "Or run manually:" -ForegroundColor Cyan
Write-Host "cd c:\Users\user\Desktop\ChineseLearningApp\ai_trae_app" -ForegroundColor Gray
Write-Host "..\flutter\bin\flutter pub get" -ForegroundColor Gray
Write-Host "..\flutter\bin\flutter run -d chrome" -ForegroundColor Gray
Write-Host ""
Read-Host "Press any key to exit"
