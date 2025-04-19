package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Service handles YouTube API operations
type Service struct {
	service *youtube.Service
}

// Chapter represents a video chapter timestamp and title
type Chapter struct {
	Time  time.Duration
	Title string
}

// NewService creates a new YouTube service using OAuth2 credentials
func NewService(credentialsFile string) (*Service, error) {
	ctx := context.Background()

	// Check if credentials file exists
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("credentials file not found: %s\nPlease download OAuth credentials from Google Cloud Console", credentialsFile)
	}

	// Read the credentials file
	b, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// Configure the OAuth2 config
	config, err := google.ConfigFromJSON(b, youtube.YoutubeScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file: %v", err)
	}

	// Get the token
	token, err := getTokenFromCache(config)
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
		fmt.Println("Attempting to get a new token from web...")
		token, err = getTokenFromWeb(ctx, config)
		if err != nil {
			return nil, err
		}
		err = saveToken(token)
		if err != nil {
			fmt.Printf("Warning: unable to save token: %v\n", err)
		}
	}

	// Create the YouTube service
	svc, err := youtube.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
	if err != nil {
		return nil, fmt.Errorf("unable to create YouTube service: %v", err)
	}

	return &Service{service: svc}, nil
}

// UpdateVideoChapters updates a YouTube video's description with chapter timestamps
func (s *Service) UpdateVideoChapters(videoID string, chapters []Chapter, preserveDescription bool) error {
	ctx := context.Background()

	// Get the existing video details
	videoCall := s.service.Videos.List([]string{"snippet"}).Id(videoID)
	videoResponse, err := videoCall.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to get video info: %v", err)
	}

	if len(videoResponse.Items) == 0 {
		return fmt.Errorf("video not found with ID: %s", videoID)
	}

	video := videoResponse.Items[0]
	snippet := video.Snippet
	originalDescription := snippet.Description

	// Format chapters into description
	chapterText := formatChapters(chapters)

	var newDescription string
	if preserveDescription {
		// Check if the existing description already has chapters
		existingDesc := originalDescription
		chapterStartIndex := strings.Index(existingDesc, "0:00 ")

		if chapterStartIndex != -1 {
			// Find the start of the chapter list in the description
			lineStart := 0
			for i := chapterStartIndex; i >= 0; i-- {
				if i == 0 || existingDesc[i-1] == '\n' {
					lineStart = i
					break
				}
			}

			// Replace existing chapters with new ones
			beforeChapters := existingDesc[:lineStart]
			newDescription = beforeChapters + chapterText
		} else {
			// Append chapters to the end
			if originalDescription != "" && !strings.HasSuffix(originalDescription, "\n") {
				newDescription = originalDescription + "\n\n" + chapterText
			} else {
				newDescription = originalDescription + "\n" + chapterText
			}
		}
	} else {
		// Just use the chapter text as the new description
		newDescription = chapterText
	}

	// Update video with new description
	snippet.Description = newDescription
	updateCall := s.service.Videos.Update([]string{"snippet"}, video)
	_, err = updateCall.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to update video: %v", err)
	}

	return nil
}

// formatChapters formats the chapters into a string suitable for YouTube description
func formatChapters(chapters []Chapter) string {
	if len(chapters) == 0 {
		return ""
	}

	// Make sure the first chapter starts at 0:00
	hasZeroChapter := false
	for _, ch := range chapters {
		if ch.Time == 0 {
			hasZeroChapter = true
			break
		}
	}

	var builder strings.Builder

	// YouTube requires chapters to start with 0:00
	if !hasZeroChapter {
		// Add introduction line
		builder.WriteString("Chapters:\n")
		// Add 0:00 chapter if it doesn't exist
		builder.WriteString("0:00 Introduction\n")
	} else {
		builder.WriteString("Chapters:\n")
	}

	// Add all chapters
	for _, chapter := range chapters {
		// Format time as MM:SS or HH:MM:SS
		var timeStr string
		totalSeconds := int(chapter.Time.Seconds())
		hours := totalSeconds / 3600
		minutes := (totalSeconds % 3600) / 60
		seconds := totalSeconds % 60

		if hours > 0 {
			timeStr = fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
		} else {
			timeStr = fmt.Sprintf("%d:%02d", minutes, seconds)
		}

		builder.WriteString(fmt.Sprintf("%s %s\n", timeStr, chapter.Title))
	}

	return builder.String()
}

// getTokenFromCache retrieves a token from a local file.
func getTokenFromCache(config *oauth2.Config) (*oauth2.Token, error) {
	tokenFile := getTokenCacheFile()

	f, err := os.Open(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("token cache file not found: %v", err)
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	if err != nil {
		return nil, fmt.Errorf("token cache file is invalid: %v", err)
	}

	// Check if token is expired
	if !token.Valid() {
		return nil, fmt.Errorf("cached token is expired")
	}

	return token, nil
}

// saveToken saves a token to a file
func saveToken(token *oauth2.Token) error {
	tokenFile := getTokenCacheFile()

	// Create directory if it doesn't exist
	dir := filepath.Dir(tokenFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("unable to create token directory: %v", err)
	}

	f, err := os.OpenFile(tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

// getTokenCacheFile returns the file path for the token cache
func getTokenCacheFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".config", "cmgen", "youtube-token.json")
}
