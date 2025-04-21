#!/bin/bash
set -e

echo "=== CMGen Build Script ==="
echo "Building Go backend and React frontend..."

# Check for required tools
command -v go >/dev/null 2>&1 || { echo "Error: Go is required but not installed."; exit 1; }
command -v node >/dev/null 2>&1 || { echo "Error: Node.js is required but not installed."; exit 1; }
command -v npm >/dev/null 2>&1 || { echo "Error: npm is required but not installed."; exit 1; }
command -v ffmpeg >/dev/null 2>&1 || { echo "Warning: FFmpeg is required for running the application."; }

# Create necessary directories
mkdir -p bin
mkdir -p web/build

# Build Go backend
echo "Building Go backend..."
go build -ldflags="-s -w" -o bin/cmgen ./cmd/cmgen

# Make binary executable
chmod +x bin/cmgen

# Build React frontend
echo "Building React frontend..."
cd web
npm ci --quiet
npm run build
cd ..

# Copy web build to where the Go server expects it
echo "Copying web build to static directory..."
mkdir -p static
cp -r web/build/* static/

# Create example chapters.json for testing
if [ ! -f "chapters.json" ]; then
  echo "Creating example chapters.json file..."
  cat > chapters.json << EOL
[
  {
    "timestamp": 0,
    "title": "Introduction"
  },
  {
    "timestamp": 120,
    "title": "Chapter 1"
  },
  {
    "timestamp": 300,
    "title": "Chapter 2"
  }
]
EOL
  echo "✓ Example chapters.json created"
fi

echo "========================================"
echo "✓ Build completed successfully!"
echo ""
echo "Usage examples:"
echo "1. Process a video:          ./bin/cmgen video.mp4 --threshold 0.3"
echo "2. Start web interface:      ./bin/cmgen --web"
echo "3. Use draft chapters:       ./bin/cmgen video.mp4 --draft chapters.json"
echo "4. Upload to YouTube:        ./bin/cmgen youtube VIDEO_ID chapters.json"
echo ""
echo "For more options, run:       ./bin/cmgen --help"
echo "========================================" 