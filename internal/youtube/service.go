package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

type Service struct {
	service *youtube.Service
}

type Chapter struct {
	Timestamp float64 `json:"timestamp"`
	Title     string  `json:"title"`
}

func NewService(credentialsPath string) (*Service, error) {
	ctx := context.Background()

	// Read credentials file
	b, err := ioutil.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// Get config from credentials
	config, err := google.ConfigFromJSON(b, youtube.YoutubeScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	// Get token from cache or request new one
	token, err := getTokenFromCache(config)
	if err != nil {
		token, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		saveToken(token)
	}

	// Create YouTube service
	client := config.Client(ctx, token)
	service, err := youtube.New(client)
	if err != nil {
		return nil, fmt.Errorf("unable to create YouTube service: %v", err)
	}

	return &Service{service: service}, nil
}

func (s *Service) UpdateVideoChapters(videoID string, chapters []Chapter) error {
	// Format chapters for YouTube description
	description := formatChaptersForYouTube(chapters)

	// Update video description
	call := s.service.Videos.Update([]string{"snippet"}, &youtube.Video{
		Id: videoID,
		Snippet: &youtube.VideoSnippet{
			Description: description,
		},
	})

	_, err := call.Do()
	return err
}

func formatChaptersForYouTube(chapters []Chapter) string {
	description := "Chapters:\n"
	for _, chapter := range chapters {
		minutes := int(chapter.Timestamp / 60)
		seconds := int(chapter.Timestamp) % 60
		description += fmt.Sprintf("%02d:%02d %s\n", minutes, seconds, chapter.Title)
	}
	return description
}

func getTokenFromCache(config *oauth2.Config) (*oauth2.Token, error) {
	tokenFile := "token.json"
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(token *oauth2.Token) {
	file := "token.json"
	os.MkdirAll(filepath.Dir(file), 0700)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Warning: unable to cache oauth token: %v\n", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
