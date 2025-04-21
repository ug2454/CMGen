# Auto Chapter-Mark Generator (CMGen)

Automatically generate and edit chapter markers for your videos, with optional YouTube integration.

## Features

- Advanced scene detection using both visual and audio cues
- Intelligent chapter point selection based on content analysis
- Edit chapter titles and timestamps with a modern web UI
- Export chapters to JSON format
- Upload chapters directly to YouTube videos
- Docker support for easy deployment

## Setup

### Prerequisites

- Go 1.20+
- Node.js 18+
- FFmpeg

### Quick Start

#### Using Build Scripts

**On Linux/macOS:**
```bash
chmod +x build.sh
./build.sh
```

**On Windows:**
```
build.bat
```

This will build both the frontend and backend, and create an example `chapters.json` file.

### Manual Installation

1. Clone the repository
```bash
git clone https://github.com/ug2454/cmgen.git
cd cmgen
```

2. Install Go dependencies
```bash
go mod tidy
```

3. Install Node.js dependencies
```bash
cd web
npm install
cd ..
```

4. Build the application
```bash
cd web
npm run build
cd ..
go build -o cmgen ./cmd/cmgen
```

### YouTube API Integration

To use the YouTube integration feature:

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project
3. Enable the YouTube Data API v3:
   - Navigate to "APIs & Services" > "Library"
   - Search for "YouTube Data API v3"
   - Click "Enable"
4. Create OAuth credentials:
   - Go to "APIs & Services" > "Credentials"
   - Click "Create Credentials" > "OAuth client ID"
   - Select "Desktop application" (not "Web application")
   - Enter a name for your client (e.g., "CMGen")
   - Click "Create"
5. Download the credentials JSON file
6. Rename it to `credentials.json` and place it in the project root directory

**Important**: 
- Keep your `credentials.json` file private and never commit it to version control
- The file is already added to `.gitignore` to help prevent accidental commits
- When running in Docker, mount the credentials file as a volume rather than building it into the image

When running the application for the first time, you'll be prompted to authenticate. The application will:
1. Open a browser window for you to log in with your Google account
2. Ask you to grant permission to manage your YouTube videos
3. Provide an authorization code to paste back into the application
4. After authentication, a token will be saved in `~/.config/cmgen/youtube-token.json` for future use
   - This token file is also excluded from version control

## Usage

### Common Use Cases

#### Process a Video File
```bash
./cmgen video.mp4 --threshold 0.3 --min-gap 30 --min-duration 15
```

#### Start the Web UI
```bash
./cmgen --web
```

#### Use a Draft Chapters File
```bash
./cmgen video.mp4 --draft chapters.json
```

#### Upload to YouTube
```bash
./cmgen youtube VIDEO_ID chapters.json
```

### One-liner Examples

```bash
# Process video with medium sensitivity and immediately edit in the web UI
./cmgen video.mp4 --threshold 0.3 --draft chapters.json && cd web && npm start

# Process video with custom parameters
./cmgen video.mp4 --threshold 0.4 --min-gap 20 --min-duration 10 --max-scenes 15
```

### Running in Development Mode

1. Start the backend
```bash
go run ./cmd/cmgen --web
```

2. Start the frontend
```bash
cd web
npm start
```

3. Open your browser to `http://localhost:8080`

### Docker

#### Using Docker Compose (recommended)

Make sure you have the following directory structure:
```
.
├── docker-compose.yml
├── credentials.json (for YouTube integration)
├── videos/
└── chapters/
```

Then run:
```bash
docker-compose up -d
```

#### Using Docker directly

Build the image:
```bash
docker build -t cmgen .
```

Run the container:
```bash
docker run -p 8080:8080 -v /path/to/videos:/videos -v $(pwd)/credentials.json:/app/credentials.json -v $HOME/.config/cmgen:/app/.config/cmgen cmgen --web
```


**Security note**: The above command mounts both your credentials and OAuth token directory as volumes, keeping them out of the Docker image itself. This approach:
- Keeps your credentials private
- Allows reuse of your authenticated token
- Prevents your sensitive data from being built into shareable images

## Advanced Options

### Scene Detection Parameters

- `--threshold` (`-t`): Sensitivity for visual scene detection (0.1-1.0, default: 0.2)
- `--min-gap` (`-g`): Minimum gap between scenes in seconds (default: 10)
- `--min-duration` (`-d`): Minimum scene duration in seconds (default: 5)
- `--max-scenes` (`-m`): Maximum number of scenes to detect (default: 30)

### YouTube Options

- `--preserve` (`-p`): Preserve existing video description when adding chapters (default: true)

## Troubleshooting

### YouTube Integration Issues

- **Error: "redirect_uri_mismatch"**: Make sure you're using Desktop Application credentials (not Web Application)
- **Error: "unable to read authorization code"**: You need to complete the OAuth flow locally first, then mount the token directory when running in Docker
- **Authentication fails**: Ensure you have added your email as a test user in the OAuth consent screen

### FFmpeg Issues

- **"ffmpeg not found"**: Ensure FFmpeg is properly installed and in your PATH

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT 

## Images
![image](https://github.com/user-attachments/assets/bdcc904f-9569-4b4e-a2ac-3a992459ebfe)
![image](https://github.com/user-attachments/assets/e643fb62-3f6a-4f85-8116-c3b8dfac34c2)
![image](https://github.com/user-attachments/assets/1425a642-1368-42fb-91d2-c5c311e0b7d7)

