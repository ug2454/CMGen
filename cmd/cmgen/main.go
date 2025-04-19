package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/yourusername/cmgen/internal/detector"
)

type Chapter struct {
	Timestamp float64 `json:"timestamp"`
	Title     string  `json:"title"`
}

var chapters []Chapter

func main() {
	// Define command-line flags
	webFlag := flag.Bool("web", false, "Start web server")
	threshold := flag.Float64("t", 0.3, "Scene detection threshold (0.0 to 1.0)")
	minGap := flag.Float64("g", 5.0, "Minimum gap between scenes in seconds")
	minDuration := flag.Float64("d", 0.0, "Minimum duration for a scene in seconds")
	maxScenes := flag.Int("m", 0, "Maximum number of scenes to detect (0 for unlimited)")
	port := flag.String("p", "3000", "Web server port")
	flag.Parse()

	if *webFlag {
		startWebServer(*port)
		return
	}

	// Get video file path from command-line arguments
	if len(flag.Args()) < 1 {
		fmt.Println("Usage: cmgen [options] <video_file>")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	videoPath := flag.Arg(0)
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		log.Fatalf("Video file not found: %s", videoPath)
	}

	// Create scene detector with specified parameters
	detector := detector.NewSceneDetector(*threshold, *minGap, *minDuration, *maxScenes)

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

func startWebServer(port string) {
	// API endpoints
	http.HandleFunc("/api/chapters", handleChapters)
	http.HandleFunc("/api/detect", handleDetect)
	http.HandleFunc("/api/export", handleExport)

	// Serve static files
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/build/index.html")
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/build/static"))))

	fmt.Printf("Starting web server on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
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
