package detector

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type SceneDetector struct {
	Threshold   float64
	MinGap      float64
	MinDuration float64
	MaxScenes   int
}

type Scene struct {
	Timestamp float64
	Frame     int64
	Score     float64
}

func NewSceneDetector(threshold, minGap, minDuration float64, maxScenes int) *SceneDetector {
	return &SceneDetector{
		Threshold:   threshold,
		MinGap:      minGap,
		MinDuration: minDuration,
		MaxScenes:   maxScenes,
	}
}

func (sd *SceneDetector) DetectScenes(videoPath string) ([]Scene, error) {
	fmt.Printf("Starting scene detection with threshold: %f, min gap: %f, min duration: %f\n",
		sd.Threshold, sd.MinGap, sd.MinDuration)

	// Check if ffmpeg is installed
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not found: %v", err)
	}

	// Get video duration
	duration, err := getVideoDuration(videoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get video duration: %v", err)
	}
	fmt.Printf("Video duration: %.2f seconds\n", duration)

	// Construct FFmpeg command for scene detection
	cmd := exec.Command(
		"ffmpeg",
		"-i", videoPath,
		"-vf", fmt.Sprintf("select='gt(scene,%f)',metadata=print:file=-", sd.Threshold),
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("FFmpeg output:\n%s\n", string(output))
		return nil, fmt.Errorf("ffmpeg command failed: %v", err)
	}

	// Parse FFmpeg output to extract timestamps
	var scenes []Scene
	var lastTimestamp float64
	startTime := time.Now()

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "pts_time:") {
			fields := strings.Fields(line)
			var timestamp float64
			var score float64

			for _, field := range fields {
				if strings.HasPrefix(field, "pts_time:") {
					timeStr := strings.TrimPrefix(field, "pts_time:")
					timestamp, _ = strconv.ParseFloat(timeStr, 64)
				}
				if strings.HasPrefix(field, "score:") {
					scoreStr := strings.TrimPrefix(field, "score:")
					score, _ = strconv.ParseFloat(scoreStr, 64)
				}
			}

			// Apply filters
			if timestamp > 0 {
				// Check minimum gap
				if len(scenes) > 0 && timestamp-lastTimestamp < sd.MinGap {
					continue
				}

				// Check minimum duration
				if timestamp < sd.MinDuration {
					continue
				}

				// Check maximum scenes
				if sd.MaxScenes > 0 && len(scenes) >= sd.MaxScenes {
					break
				}

				scenes = append(scenes, Scene{
					Timestamp: timestamp,
					Score:     score,
				})
				lastTimestamp = timestamp

				// Report progress
				elapsed := time.Since(startTime).Seconds()
				progress := (timestamp / duration) * 100
				fmt.Printf("\rProgress: %.1f%% (%.1f seconds elapsed)", progress, elapsed)
			}
		}
	}
	fmt.Println("\nScene detection completed")

	return scenes, nil
}

func getVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", videoPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, err
	}

	return duration, nil
}
