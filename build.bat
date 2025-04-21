@echo off
echo ========================================
echo   CMGen Build Script (Windows)
echo ========================================

:: Check required tools
where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Error: go is required but not installed.
    exit /b 1
)

where npm >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Error: npm is required but not installed.
    exit /b 1
)

where ffmpeg >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Error: ffmpeg is required but not installed.
    exit /b 1
)

echo ✓ Required tools detected
echo.

:: Build frontend
echo Building React frontend...
cd web
call npm install
call npm run build
cd ..
echo ✓ Frontend built successfully
echo.

:: Build backend
echo Building Go backend...
go mod tidy
go build -o cmgen.exe ./cmd/cmgen
echo ✓ Backend built successfully
echo.

:: Create example chapters.json for testing
if not exist chapters.json (
  echo Creating example chapters.json file...
  echo [> chapters.json
  echo   {>> chapters.json
  echo     "timestamp": 0,>> chapters.json
  echo     "title": "Introduction">> chapters.json
  echo   },>> chapters.json
  echo   {>> chapters.json
  echo     "timestamp": 120,>> chapters.json
  echo     "title": "Chapter 1">> chapters.json
  echo   },>> chapters.json
  echo   {>> chapters.json
  echo     "timestamp": 300,>> chapters.json
  echo     "title": "Chapter 2">> chapters.json
  echo   }>> chapters.json
  echo ]>> chapters.json
  echo ✓ Example chapters.json created
  echo.
)

echo ========================================
echo ✓ Build completed successfully!
echo.
echo Usage examples:
echo 1. Process a video:          .\cmgen.exe video.mp4 --threshold 0.3
echo 2. Start web interface:      .\cmgen.exe --web
echo 3. Use draft chapters:       .\cmgen.exe video.mp4 --draft chapters.json
echo 4. Upload to YouTube:        .\cmgen.exe youtube VIDEO_ID chapters.json
echo.
echo For more options, run:       .\cmgen.exe --help
echo ======================================== 