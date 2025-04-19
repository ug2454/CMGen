# Auto Chapter-Mark Generator

Automatically generate and edit chapter markers for your videos, with optional YouTube integration.

## Features

- Detect scenes in videos using FFmpeg
- Edit chapter titles and timestamps
- Export chapters to JSON
- Upload chapters directly to YouTube videos

## Setup

### Prerequisites

- Go 1.20+
- Node.js 18+
- FFmpeg

### Installation

1. Clone the repository
```bash
git clone https://github.com/yourusername/cmgen.git
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

**Important**: When running the application for the first time, you'll be prompted to authenticate. The application will:
1. Open a browser window for you to log in with your Google account
2. Ask you to grant permission to manage your YouTube videos
3. Provide an authorization code to paste back into the application
4. After authentication, a token will be saved for future use

## Usage

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

3. Open your browser to `http://localhost:3000`

### Building for Production

1. Build the frontend
```bash
cd web
npm run build
cd ..
```

2. Build the backend
```bash
go build -o cmgen ./cmd/cmgen
```

3. Run the application
```bash
./cmgen --web
```

## Docker

```bash
docker build -t cmgen .
docker run -p 3000:3000 cmgen
```

## Troubleshooting

### YouTube Integration Issues

1. **"Missing redirect URL" error**
   - Make sure you selected "Desktop application" (not "Web application") when creating OAuth credentials
   - If you must use Web application credentials, add `http://localhost:8080/callback` as an authorized redirect URI

2. **Authentication failures**
   - Delete the `token.json` file and try again
   - Ensure you're using the correct Google account with YouTube access
   - Check that you've enabled the YouTube Data API v3 for your project

3. **Permission errors**
   - Ensure the YouTube account you're authenticating with has permission to edit the video
   - If you're trying to modify someone else's video, you won't have permission

## License

MIT 