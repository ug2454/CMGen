package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"cmgen/internal/detector"
	"cmgen/internal/youtube"

	"github.com/spf13/cobra"
)

type Chapter struct {
	Timestamp float64 `json:"timestamp"`
	Title     string  `json:"title"`
}

type YouTubeRequest struct {
	VideoID  string    `json:"videoId"`
	Chapters []Chapter `json:"chapters"`
}

var chapters []Chapter

func main() {
	var threshold float64
	var minGap int
	var minDuration int
	var maxScenes int
	var preserveDesc bool
	var webMode bool

	var rootCmd = &cobra.Command{
		Use:   "cmgen [video_file]",
		Short: "CMGen - Auto Chapter-Mark Generator",
		Long:  "Automatically generate chapter markers for videos",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if webMode {
				startWebServer()
				return
			}

			// Get video file path from command-line arguments
			if len(args) < 1 {
				fmt.Println("Usage: cmgen [options] <video_file>")
				cmd.Help()
				os.Exit(1)
			}

			videoPath := args[0]
			if _, err := os.Stat(videoPath); os.IsNotExist(err) {
				log.Fatalf("Video file not found: %s", videoPath)
			}

			// Create scene detector with specified parameters
			detector := detector.NewSceneDetector(threshold, float64(minGap), float64(minDuration), maxScenes)

			// Detect scenes
			fmt.Printf("Processing video: %s\n", videoPath)
			scenes, err := detector.DetectScenes(videoPath)
			if err != nil {
				log.Fatalf("Error detecting scenes: %v", err)
			}

			// Convert scenes to chapters
			chapters = make([]Chapter, len(scenes))
			for i, scene := range scenes {
				chapters[i] = Chapter{
					Timestamp: scene.Timestamp,
					Title:     fmt.Sprintf("Chapter %d", i+1),
				}
			}

			// Write chapters to JSON file
			outputFile := "chapters.json"
			if err := writeChaptersToFile(chapters, outputFile); err != nil {
				log.Fatalf("Error writing chapters to file: %v", err)
			}

			fmt.Printf("Detected %d scenes and wrote to %s\n", len(chapters), outputFile)

			// Play notification sound
			playNotificationSound()
		},
	}

	// Add flags to the root command
	rootCmd.Flags().Float64VarP(&threshold, "threshold", "t", 0.2, "Threshold for scene detection (0.1-1.0)")
	rootCmd.Flags().IntVarP(&minGap, "min-gap", "g", 10, "Minimum gap between scenes in seconds")
	rootCmd.Flags().IntVarP(&minDuration, "min-duration", "d", 5, "Minimum scene duration in seconds")
	rootCmd.Flags().IntVarP(&maxScenes, "max-scenes", "m", 30, "Maximum number of scenes to detect")
	rootCmd.Flags().BoolVarP(&webMode, "web", "w", false, "Start web UI server")

	// Add YouTube command
	var ytCmd = &cobra.Command{
		Use:   "youtube [video_id] [chapters_file]",
		Short: "Upload chapter markers to YouTube",
		Long:  "Upload chapter markers from a JSON file to a YouTube video description",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			videoID := args[0]
			chaptersFile := args[1]

			// Read chapters from JSON file
			data, err := ioutil.ReadFile(chaptersFile)
			if err != nil {
				log.Fatalf("Error reading chapters file: %v", err)
			}

			var chapters []struct {
				Time  float64 `json:"time"`
				Title string  `json:"title"`
			}

			if err := json.Unmarshal(data, &chapters); err != nil {
				log.Fatalf("Error parsing chapters JSON: %v", err)
			}

			// Convert to YouTube chapters
			ytChapters := make([]youtube.Chapter, len(chapters))
			for i, ch := range chapters {
				ytChapters[i] = youtube.Chapter{
					Time:  time.Duration(ch.Time * float64(time.Second)),
					Title: ch.Title,
				}
			}

			// Create YouTube service
			svc, err := youtube.NewService("credentials.json")
			if err != nil {
				log.Fatalf("Error creating YouTube service: %v", err)
			}

			// Update video with chapters
			if err := svc.UpdateVideoChapters(videoID, ytChapters, true); err != nil {
				log.Fatalf("Error updating YouTube video: %v", err)
			}

			fmt.Printf("Successfully updated YouTube video %s with %d chapters\n", videoID, len(chapters))
		},
	}

	ytCmd.Flags().BoolVarP(&preserveDesc, "preserve", "p", true, "Preserve existing video description")
	rootCmd.AddCommand(ytCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func writeChaptersToFile(chapters []Chapter, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(chapters)
}

func startWebServer() {
	// Set up CORS middleware
	corsMiddleware := func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			handler.ServeHTTP(w, r)
		})
	}

	// API endpoints
	http.HandleFunc("/api/chapters", handleChapters)
	http.HandleFunc("/api/detect", handleDetect)
	http.HandleFunc("/api/export", handleExport)
	http.HandleFunc("/api/youtube", handleYouTube)

	// Serve static files
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/build/index.html")
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/build/static"))))

	// Wrap all handlers with CORS middleware
	handler := corsMiddleware(http.DefaultServeMux)

	fmt.Println("Starting web server...")
	if err := http.ListenAndServe(":3000", handler); err != nil {
		log.Fatalf("Error starting web server: %v", err)
	}
}

func handleChapters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Return current chapters
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chapters)

	case http.MethodPost:
		// Update chapters
		var newChapters []Chapter
		if err := json.NewDecoder(r.Body).Decode(&newChapters); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		chapters = newChapters
		if err := writeChaptersToFile(chapters, "chapters.json"); err != nil {
			http.Error(w, "Failed to save chapters", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "No video file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save uploaded file
	tempFile, err := os.CreateTemp("", "upload-*.mp4")
	if err != nil {
		http.Error(w, "Failed to save video", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		http.Error(w, "Failed to save video", http.StatusInternalServerError)
		return
	}

	// Get parameters
	threshold := r.FormValue("threshold")
	minGap := r.FormValue("minGap")
	minDuration := r.FormValue("minDuration")
	maxScenes := r.FormValue("maxScenes")

	// Create scene detector
	detector := detector.NewSceneDetector(
		parseFloat(threshold, 0.3),
		parseFloat(minGap, 5.0),
		parseFloat(minDuration, 0.0),
		parseInt(maxScenes, 0),
	)

	// Detect scenes
	scenes, err := detector.DetectScenes(tempFile.Name())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to detect scenes: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to chapters
	chapters = make([]Chapter, len(scenes))
	for i, scene := range scenes {
		chapters[i] = Chapter{
			Timestamp: scene.Timestamp,
			Title:     fmt.Sprintf("Chapter %d", i+1),
		}
	}

	// Save chapters
	if err := writeChaptersToFile(chapters, "chapters.json"); err != nil {
		http.Error(w, "Failed to save chapters", http.StatusInternalServerError)
		return
	}

	// Return chapters
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chapters)

	// Play notification sound
	playNotificationSound()
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=chapters.json")
	json.NewEncoder(w).Encode(chapters)
}

func handleYouTube(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req YouTubeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.VideoID == "" {
		http.Error(w, "Video ID is required", http.StatusBadRequest)
		return
	}

	if len(req.Chapters) == 0 {
		http.Error(w, "At least one chapter is required", http.StatusBadRequest)
		return
	}

	// Convert our chapters to YouTube chapters
	ytChapters := make([]youtube.Chapter, len(req.Chapters))
	for i, chapter := range req.Chapters {
		ytChapters[i] = youtube.Chapter{
			Time:  time.Duration(chapter.Timestamp * float64(time.Second)),
			Title: chapter.Title,
		}
	}

	// Create YouTube service
	service, err := youtube.NewService("credentials.json")
	if err != nil {
		log.Printf("Failed to create YouTube service: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create YouTube service: %v", err), http.StatusInternalServerError)
		return
	}

	// Update YouTube video description
	if err := service.UpdateVideoChapters(req.VideoID, ytChapters, true); err != nil {
		log.Printf("Failed to update YouTube video: %v", err)
		http.Error(w, fmt.Sprintf("Failed to update YouTube video: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func parseFloat(s string, defaultValue float64) float64 {
	if s == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return value
}

// playNotificationSound plays a system notification sound depending on the OS
func playNotificationSound() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// For Windows, use PowerShell to play a system sound
		cmd = exec.Command("powershell", "-c", "(New-Object Media.SoundPlayer 'C:\\Windows\\Media\\notify.wav').PlaySync();")
	case "darwin":
		// For macOS
		cmd = exec.Command("afplay", "/System/Library/Sounds/Ping.aiff")
	case "linux":
		// For Linux, try paplay with a freedesktop sound
		cmd = exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/complete.oga")
	default:
		// Unsupported OS
		return
	}

	// Run the command without waiting for output
	_ = cmd.Start()
}
