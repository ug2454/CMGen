# CMGen Build Script for Windows
Write-Host "=== CMGen Build Script ===" -ForegroundColor Cyan
Write-Host "Building Go backend and React frontend..." -ForegroundColor Cyan

# Check for required tools
try { 
    $null = & go version 
} 
catch { 
    Write-Host "Error: Go is required but not installed." -ForegroundColor Red
    exit 1 
}

try { 
    $null = & node --version 
} 
catch { 
    Write-Host "Error: Node.js is required but not installed." -ForegroundColor Red
    exit 1 
}

try { 
    $null = & npm --version 
} 
catch { 
    Write-Host "Error: npm is required but not installed." -ForegroundColor Red
    exit 1 
}

try { 
    $null = & ffmpeg -version 
} 
catch { 
    Write-Host "Warning: FFmpeg is required for running the application." -ForegroundColor Yellow
}

# Create necessary directories
New-Item -ItemType Directory -Force -Path bin | Out-Null
New-Item -ItemType Directory -Force -Path web\build | Out-Null

# Build Go backend
Write-Host "Building Go backend..." -ForegroundColor Cyan
& go build -ldflags="-s -w" -o bin/cmgen.exe ./cmd/cmgen

# Build React frontend
Write-Host "Building React frontend..." -ForegroundColor Cyan
Push-Location web
& npm ci --quiet
& npm run build
Pop-Location

# Copy web build to where the Go server expects it
Write-Host "Copying web build to static directory..." -ForegroundColor Cyan
New-Item -ItemType Directory -Force -Path static | Out-Null
Copy-Item -Path web\build\* -Destination static -Recurse -Force

# Create example chapters.json for testing
if (-not (Test-Path "chapters.json")) {
    Write-Host "Creating example chapters.json file..." -ForegroundColor Cyan
    
    # Create sample chapter data
    $chapter1 = @{
        "timestamp" = 0
        "title" = "Introduction"
    }
    $chapter2 = @{
        "timestamp" = 120
        "title" = "Chapter 1"
    }
    $chapter3 = @{
        "timestamp" = 300
        "title" = "Chapter 2"
    }
    
    $chapterArray = @($chapter1, $chapter2, $chapter3)
    $jsonContent = $chapterArray | ConvertTo-Json
    
    Set-Content -Path "chapters.json" -Value $jsonContent
    Write-Host "[+] Example chapters.json created" -ForegroundColor Green
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "[+] Build completed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Usage examples:" -ForegroundColor Cyan
Write-Host "1. Process a video:          .\bin\cmgen.exe video.mp4 --threshold 0.3"
Write-Host "2. Start web interface:      .\bin\cmgen.exe --web"
Write-Host "3. Use draft chapters:       .\bin\cmgen.exe video.mp4 --draft chapters.json"
Write-Host "4. Upload to YouTube:        .\bin\cmgen.exe youtube VIDEO_ID chapters.json"
Write-Host ""
Write-Host "For more options, run:       .\bin\cmgen.exe --help"
Write-Host "========================================" -ForegroundColor Cyan 