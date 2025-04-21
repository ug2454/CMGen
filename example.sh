#!/bin/bash
set -e

# This script demonstrates common usage patterns for CMGen

# Check if cmgen executable exists
if [ ! -f "./cmgen" ]; then
    echo "Error: cmgen executable not found. Run 'build.sh' first."
    exit 1
fi

# Check if FFmpeg is installed
if ! command -v ffmpeg &> /dev/null; then
    echo "Error: FFmpeg is required but not installed."
    exit 1
fi

# Example 1: Basic scene detection
echo "===== Example 1: Basic Scene Detection ====="
echo "Command: ./cmgen video.mp4"
echo "Description: This will detect scenes using default parameters."
echo "Press Enter to continue or Ctrl+C to exit"
read

# Example 2: Custom threshold for more/fewer chapters
echo "===== Example 2: Custom Detection Threshold ====="
echo "Command: ./cmgen video.mp4 --threshold 0.4"
echo "Description: Higher threshold = fewer scenes detected (less sensitive)"
echo "Command: ./cmgen video.mp4 --threshold 0.1"
echo "Description: Lower threshold = more scenes detected (more sensitive)"
echo "Press Enter to continue or Ctrl+C to exit"
read

# Example 3: Adjusting minimum gap between chapters
echo "===== Example 3: Minimum Gap Between Chapters ====="
echo "Command: ./cmgen video.mp4 --min-gap 30"
echo "Description: Ensures at least 30 seconds between chapters"
echo "Press Enter to continue or Ctrl+C to exit"
read

# Example 4: Using a draft file
echo "===== Example 4: Using a Draft File ====="
echo "Command: ./cmgen video.mp4 --draft chapters.json"
echo "Description: Uses existing chapters.json as a starting point"
echo "Press Enter to continue or Ctrl+C to exit"
read

# Example 5: Web UI
echo "===== Example 5: Web UI ====="
echo "Command: ./cmgen --web"
echo "Description: Starts the web UI on http://localhost:8080"
echo "Press Enter to continue or Ctrl+C to exit"
read

# Example 6: YouTube upload
echo "===== Example 6: YouTube Upload ====="
echo "Command: ./cmgen youtube VIDEO_ID chapters.json"
echo "Description: Uploads chapters to a YouTube video"
echo "Note: Requires credentials.json file and authentication"
echo "Press Enter to continue or Ctrl+C to exit"
read

# Example 7: One-liner for processing and editing
echo "===== Example 7: Process + Edit One-liner ====="
echo "Command: ./cmgen video.mp4 --threshold 0.3 --draft chapters.json && cd web && npm start"
echo "Description: Process video, save chapters, and immediately open editor"

echo -e "\nTo run any of these examples, replace 'video.mp4' with your actual video file path." 