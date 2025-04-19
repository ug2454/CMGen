# Auto Chapter-Mark Generator

An intelligent tool for video creators to automatically generate chapter markers for their videos.

## Features

- Automatic scene detection using FFmpeg
- Chapter marker generation with timestamps
- Web UI for editing and previewing chapters
- YouTube description export
- Optional YouTube API integration

## Prerequisites

- Go 1.20 or higher
- Node.js 18 or higher
- FFmpeg installed on your system
- (Optional) YouTube API credentials

## Project Structure

```
CMGen/
├── cmd/                    # CLI application
│   └── cmgen/             # Main CLI entry point
├── internal/              # Internal packages
│   ├── processor/         # Video processing logic
│   ├── detector/          # Scene detection
│   └── youtube/          # YouTube API integration
├── web/                   # React frontend
├── api/                   # API endpoints
└── pkg/                   # Public packages
```

## Getting Started

1. Clone the repository
2. Install dependencies:
   ```bash
   # Backend
   go mod download

   # Frontend
   cd web
   npm install
   ```

3. Build the project:
   ```bash
   # Backend
   go build ./cmd/cmgen

   # Frontend
   cd web
   npm run build
   ```

4. Run the application:
   ```bash
   ./cmgen video.mp4 --threshold 0.3 --draft chapters.json
   ```

## Development

- Backend: Written in Go, uses FFmpeg for video processing
- Frontend: React application with TypeScript
- API: REST endpoints for communication between frontend and backend

## License

MIT License 