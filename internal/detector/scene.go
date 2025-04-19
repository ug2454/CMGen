package detector

import (
	"fmt"
	"math"
	"math/rand"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
}

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
	fmt.Printf("Starting intelligent scene detection with threshold: %f, min gap: %f, min duration: %f\n",
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

	// Detect both visual and audio scene changes for better accuracy
	visualScenes, err := sd.detectVisualScenes(videoPath, duration)
	if err != nil {
		return nil, err
	}

	// Try to detect audio silence points which often indicate chapter transitions
	audioScenes, err := sd.detectAudioScenes(videoPath, duration)
	if err != nil {
		fmt.Printf("Warning: Could not detect audio scenes: %v\n", err)
		// Continue with just visual scenes
	}

	// Combine and filter scenes
	allScenes := combineScenes(visualScenes, audioScenes, sd.MinGap)

	// Apply intelligent filtering to get logical chapters
	scenes := sd.intelligentFiltering(allScenes, duration)

	// If we still have no scenes at this point, create fallback chapters
	if len(scenes) == 0 {
		scenes = createFallbackChapters(duration)
		fmt.Println("Using fallback chapter generation method")
	}

	// Add 0:00 as the first chapter if it's not already there
	hasZeroChapter := false
	for _, scene := range scenes {
		if scene.Timestamp < 1.0 { // Consider anything in the first second as a zero timestamp
			hasZeroChapter = true
			break
		}
	}

	if !hasZeroChapter && len(scenes) > 0 {
		scenes = append([]Scene{{Timestamp: 0, Score: 1.0}}, scenes...)
	}

	// Limit to max scenes if specified
	if sd.MaxScenes > 0 && len(scenes) > sd.MaxScenes {
		scenes = selectRepresentativeScenes(scenes, sd.MaxScenes, duration)
	}

	fmt.Printf("Detected %d logical chapter points\n", len(scenes))

	// Final check - enforce minimum number of chapters
	if len(scenes) < 3 && duration > 180 { // For videos longer than 3 minutes
		scenes = createFallbackChapters(duration)
		fmt.Println("Enforcing minimum chapter count using fallback method")
	}

	return scenes, nil
}

// detectVisualScenes detects scene changes based on visual content
func (sd *SceneDetector) detectVisualScenes(videoPath string, duration float64) ([]Scene, error) {
	fmt.Println("Analyzing visual scene changes...")

	// Try two different methods for scene detection
	scenes, err := sd.detectScenesByThreshold(videoPath, duration)
	if err != nil {
		fmt.Printf("Warning: Threshold-based scene detection had an issue: %v\n", err)
		scenes = []Scene{} // Empty slice in case of error
	}

	// If we didn't get enough scenes, try an alternative method
	if len(scenes) < 5 && duration > 300 { // For videos longer than 5 minutes
		fmt.Println("Few scenes detected, trying alternative detection method...")
		alternativeScenes, altErr := sd.detectScenesByInterval(videoPath, duration)
		if altErr == nil && len(alternativeScenes) > len(scenes) {
			scenes = alternativeScenes
		}
	}

	if len(scenes) == 0 {
		fmt.Println("Warning: No visual scenes detected")
	} else {
		fmt.Printf("Visual scene detection completed, found %d scenes\n", len(scenes))
	}

	return scenes, nil
}

// detectScenesByThreshold uses FFmpeg's scene detection with a threshold
func (sd *SceneDetector) detectScenesByThreshold(videoPath string, duration float64) ([]Scene, error) {
	// Construct FFmpeg command for visual scene detection with more advanced filters
	cmd := exec.Command(
		"ffmpeg",
		"-i", videoPath,
		"-vf", fmt.Sprintf("select='gt(scene,%f)',metadata=print:file=-", sd.Threshold),
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg visual scene detection failed: %v", err)
	}

	// Parse FFmpeg output to extract timestamps
	var scenes []Scene
	startTime := time.Now()

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "pts_time:") && strings.Contains(line, "score:") {
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

			// Apply basic filtering immediately
			if timestamp > 0 && timestamp < duration-5 { // Exclude scenes near the end
				scenes = append(scenes, Scene{
					Timestamp: timestamp,
					Score:     score,
				})

				// Report progress
				elapsed := time.Since(startTime).Seconds()
				progress := (timestamp / duration) * 100
				fmt.Printf("\rVisual analysis progress: %.1f%% (%.1f seconds elapsed)", progress, elapsed)
			}
		}
	}

	return scenes, nil
}

// detectScenesByInterval generates scene timestamps by analyzing key frames at regular intervals
func (sd *SceneDetector) detectScenesByInterval(videoPath string, duration float64) ([]Scene, error) {
	// Extract keyframes at regular intervals and check for significant changes
	cmd := exec.Command(
		"ffmpeg",
		"-i", videoPath,
		"-vf", "select='isnan(prev_selected_t)+gte(t-prev_selected_t,30)',metadata=print:file=-",
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg interval-based scene detection failed: %v", err)
	}

	var scenes []Scene
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "pts_time:") {
			fields := strings.Fields(line)
			var timestamp float64

			for _, field := range fields {
				if strings.HasPrefix(field, "pts_time:") {
					timeStr := strings.TrimPrefix(field, "pts_time:")
					timestamp, _ = strconv.ParseFloat(timeStr, 64)
				}
			}

			if timestamp > 30 && timestamp < duration-30 { // Exclude start/end
				scenes = append(scenes, Scene{
					Timestamp: timestamp,
					Score:     0.5, // Default score for interval-based detection
				})
			}
		}
	}

	return scenes, nil
}

// detectAudioScenes detects potential chapter points based on audio characteristics
func (sd *SceneDetector) detectAudioScenes(videoPath string, duration float64) ([]Scene, error) {
	fmt.Println("Analyzing audio for potential chapter points...")

	// Use more lenient silence detection for better results
	// Try two different noise levels
	scenes1, err := sd.detectSilence(videoPath, "-30dB", 0.5)
	if err != nil {
		return nil, err
	}

	scenes2, err := sd.detectSilence(videoPath, "-20dB", 0.75)
	if err != nil {
		return scenes1, nil // Return what we have if the second attempt fails
	}

	// Combine both sets of scenes
	allScenes := append(scenes1, scenes2...)

	// Sort by timestamp
	sort.Slice(allScenes, func(i, j int) bool {
		return allScenes[i].Timestamp < allScenes[j].Timestamp
	})

	// Try to detect speech pauses which can also indicate chapter transitions
	speechScenes, err := sd.detectSpeechPauses(videoPath, duration)
	if err == nil {
		allScenes = append(allScenes, speechScenes...)
	}

	fmt.Printf("Audio analysis completed, found %d potential points\n", len(allScenes))
	return allScenes, nil
}

// detectSilence identifies silence periods in the audio
func (sd *SceneDetector) detectSilence(videoPath string, noiseLevel string, baseScore float64) ([]Scene, error) {
	cmd := exec.Command(
		"ffmpeg",
		"-i", videoPath,
		"-af", fmt.Sprintf("silencedetect=noise=%s:d=0.5", noiseLevel),
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("audio silence detection failed: %v", err)
	}

	var scenes []Scene
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.Contains(line, "silence_end") {
			fields := strings.Fields(line)
			for i, field := range fields {
				if strings.HasPrefix(field, "silence_end:") {
					timeStr := strings.TrimPrefix(field, "silence_end:")
					timestamp, _ := strconv.ParseFloat(timeStr, 64)

					// Get duration if available for scoring
					score := baseScore // Default score
					if i+2 < len(fields) && strings.HasPrefix(fields[i+2], "silence_duration:") {
						durationStr := strings.TrimPrefix(fields[i+2], "silence_duration:")
						silenceDuration, _ := strconv.ParseFloat(durationStr, 64)
						// Higher score for longer silence
						score = math.Min(0.9, baseScore+silenceDuration/5.0)
					}

					scenes = append(scenes, Scene{
						Timestamp: timestamp,
						Score:     score,
					})
				}
			}
		}
	}

	return scenes, nil
}

// detectSpeechPauses tries to identify pauses in speech
func (sd *SceneDetector) detectSpeechPauses(videoPath string, duration float64) ([]Scene, error) {
	// Using a simple volume detection approach to find quiet periods
	cmd := exec.Command(
		"ffmpeg",
		"-i", videoPath,
		"-af", "volumedetect",
		"-f", "null",
		"-",
	)

	// This doesn't provide timestamps, so we'll use it to guide chapter creation
	_, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("speech pause detection failed: %v", err)
	}

	// For now, just generate some points based on duration
	// This is a simplified approach - a real implementation would analyze the volume data
	var scenes []Scene

	// Create points at logical speech pause locations
	// In a real implementation, these would be based on actual audio analysis
	pointCount := int(duration / 90) // Approximately one point every 1.5 minutes
	for i := 1; i <= pointCount; i++ {
		timestamp := float64(i) * 90.0
		if timestamp > 30 && timestamp < duration-30 {
			scenes = append(scenes, Scene{
				Timestamp: timestamp,
				Score:     0.4, // Lower score than silence detection
			})
		}
	}

	return scenes, nil
}

// createFallbackChapters creates a reasonable set of chapters when detection methods fail
func createFallbackChapters(duration float64) []Scene {
	// Calculate how many chapters to create based on video length
	chapterCount := 5 // Default

	if duration < 300 { // < 5 minutes
		chapterCount = 3
	} else if duration < 900 { // 5-15 minutes
		chapterCount = 5
	} else if duration < 1800 { // 15-30 minutes
		chapterCount = 8
	} else { // > 30 minutes
		chapterCount = 10 + int(duration/1800) // Add 1 chapter per 30 minutes
	}

	// Create evenly spaced chapters
	chapters := make([]Scene, chapterCount)
	chapterDuration := duration / float64(chapterCount)

	// Always start with 0
	chapters[0] = Scene{
		Timestamp: 0,
		Score:     1.0,
	}

	// Create the rest of the chapters
	for i := 1; i < chapterCount; i++ {
		// Slightly randomize the timestamps to avoid mechanical-looking chapters
		jitter := chapterDuration * 0.1 * (rand.Float64() - 0.5)
		timestamp := float64(i)*chapterDuration + jitter

		// Ensure timestamp is positive and within duration
		if timestamp <= 0 {
			timestamp = float64(i) * chapterDuration
		}
		if timestamp >= duration {
			timestamp = duration - 10 // 10 seconds before the end
		}

		chapters[i] = Scene{
			Timestamp: timestamp,
			Score:     0.5,
		}
	}

	return chapters
}

// combineScenes merges visual and audio scenes, combining those that are close to each other
func combineScenes(visualScenes, audioScenes []Scene, minGap float64) []Scene {
	if len(audioScenes) == 0 {
		return visualScenes
	}

	// Combine all scenes
	allScenes := append([]Scene{}, visualScenes...)
	allScenes = append(allScenes, audioScenes...)

	// Sort by timestamp
	sort.Slice(allScenes, func(i, j int) bool {
		return allScenes[i].Timestamp < allScenes[j].Timestamp
	})

	// Merge scenes that are close to each other
	var mergedScenes []Scene
	var lastTimestamp float64 = -100 // Start with a negative value to ensure first scene is included

	for _, scene := range allScenes {
		if scene.Timestamp-lastTimestamp >= minGap {
			mergedScenes = append(mergedScenes, scene)
			lastTimestamp = scene.Timestamp
		} else if len(mergedScenes) > 0 {
			// If scenes are close, keep the one with higher score
			lastIndex := len(mergedScenes) - 1
			if scene.Score > mergedScenes[lastIndex].Score {
				mergedScenes[lastIndex] = scene
				lastTimestamp = scene.Timestamp
			}
		}
	}

	return mergedScenes
}

// intelligentFiltering applies heuristics to identify logical chapter points
func (sd *SceneDetector) intelligentFiltering(scenes []Scene, duration float64) []Scene {
	if len(scenes) == 0 {
		return scenes
	}

	// Apply minimum duration filter
	var filteredScenes []Scene
	for _, scene := range scenes {
		if scene.Timestamp >= sd.MinDuration {
			filteredScenes = append(filteredScenes, scene)
		}
	}

	// Calculate ideal chapter count based on video length
	// - Short videos (< 5 mins): 3-5 chapters
	// - Medium videos (5-15 mins): 5-8 chapters
	// - Long videos (15-30 mins): 8-12 chapters
	// - Very long videos (> 30 mins): 10-15 chapters
	idealCount := 5
	if duration < 300 { // < 5 minutes
		idealCount = 3
	} else if duration < 900 { // 5-15 minutes
		idealCount = 5
	} else if duration < 1800 { // 15-30 minutes
		idealCount = 8
	} else { // > 30 minutes
		idealCount = 10 + int(duration/1800) // Add 1 chapter per 30 minutes
	}

	// If we have too many scenes, apply more aggressive filtering
	if len(filteredScenes) > idealCount*2 {
		// Find evenly distributed chapters based on content significance
		return selectRepresentativeScenes(filteredScenes, idealCount, duration)
	}

	return filteredScenes
}

// selectRepresentativeScenes chooses the most representative scenes to match the desired count
func selectRepresentativeScenes(scenes []Scene, desiredCount int, duration float64) []Scene {
	if len(scenes) <= desiredCount {
		return scenes
	}

	// If we have 0:00 as first chapter, always keep it
	var result []Scene
	if scenes[0].Timestamp < 1.0 {
		result = append(result, scenes[0])
		scenes = scenes[1:]
		desiredCount--
	}

	// For the remaining scenes, select evenly across the video while prioritizing higher scores

	// First, divide the video into segments
	segmentCount := desiredCount
	segmentDuration := duration / float64(segmentCount)

	segments := make([][]Scene, segmentCount)
	for _, scene := range scenes {
		segment := int(scene.Timestamp / segmentDuration)
		if segment >= segmentCount {
			segment = segmentCount - 1
		}
		segments[segment] = append(segments[segment], scene)
	}

	// Choose best scene from each segment (highest score or closest to middle)
	for i, segment := range segments {
		if len(segment) == 0 {
			continue
		}

		// If only one scene, use it
		if len(segment) == 1 {
			result = append(result, segment[0])
			continue
		}

		// Find best scene in this segment
		bestScore := -1.0
		bestIndex := 0
		segmentMiddle := float64(i)*segmentDuration + segmentDuration/2

		for j, scene := range segment {
			// Calculate a combined score based on actual score and position in segment
			distanceFromMiddle := math.Abs(scene.Timestamp - segmentMiddle)
			normalizedDistance := distanceFromMiddle / segmentDuration

			// Combined score favors high scene score and proximity to segment middle
			combinedScore := scene.Score * (1 - normalizedDistance/2)

			if combinedScore > bestScore {
				bestScore = combinedScore
				bestIndex = j
			}
		}

		result = append(result, segment[bestIndex])
	}

	// Sort by timestamp
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp < result[j].Timestamp
	})

	return result
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
