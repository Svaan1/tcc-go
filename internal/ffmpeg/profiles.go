package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type EncodingProfile struct {
	// Info
	Name   string   `json:"name"`
	Codec  string   `json:"codec"`
	Params []string `json:"params"`

	// Scores
	EncodeTime float64 `json:"encode_time"`
	DecodeTime float64 `json:"decode_time"`
	FPS        float64 `json:"fps"`
	Score      float64 `json:"score"` // lower is better
}

var (
	// H.264 Profile - Web Streaming (1080p, balanced quality/speed)
	H264WebStreaming = EncodingProfile{
		Name:  "H264_1080p",
		Codec: "libx264",
		Params: []string{
			"-c:v", "libx264",
			"-preset", "medium",
			"-crf", "23",
			"-vf", "scale=1920:1080",
			"-b:v", "4M",
			"-maxrate", "4.5M",
			"-bufsize", "8M",
			"-c:a", "aac",
			"-b:a", "128k",
		},
	}

	// H.265/HEVC Profile - High Quality Archive (4K)
	HEVCHighQuality = EncodingProfile{
		Name:  "HEVC_4K",
		Codec: "libx265",
		Params: []string{
			"-c:v", "libx265",
			"-preset", "slow",
			"-crf", "20",
			"-vf", "scale=3840:2160",
			"-x265-params", "log-level=error",
			"-c:a", "aac",
			"-b:a", "192k",
		},
	}

	// VP9 Profile - YouTube/Web (1080p, two-pass)
	VP9WebOptimized = EncodingProfile{
		Name:  "VP9_1080p",
		Codec: "libvpx-vp9",
		Params: []string{
			"-c:v", "libvpx-vp9",
			"-b:v", "2M",
			"-vf", "scale=1920:1080",
			"-deadline", "good",
			"-cpu-used", "2",
			"-row-mt", "1",
			"-c:a", "libopus",
			"-b:a", "128k",
		},
	}

	// AV1 Profile - Next-Gen Streaming (1080p, high compression)
	AV1NextGen = EncodingProfile{
		Name:  "AV1_1080p",
		Codec: "libaom-av1",
		Params: []string{
			"-c:v", "libaom-av1",
			"-crf", "30",
			"-vf", "scale=1920:1080",
			"-b:v", "0",
			"-cpu-used", "4",
			"-row-mt", "1",
			"-tiles", "2x2",
			"-c:a", "libopus",
			"-b:a", "128k",
		},
	}
)

func GetAvailableProfiles() map[string]EncodingProfile {
	return map[string]EncodingProfile{
		"H264_1080p": H264WebStreaming,
		"HEVC_4K":    HEVCHighQuality,
		"VP9_1080p":  VP9WebOptimized,
		"AV1_1080p":  AV1NextGen,
	}
}

// This is a blank video sample and should be used with caution, a better benchmark
// is done with videos with a lot of action and high bitrate
func GenerateVideoSample(duration int, resolution string) (string, error) {
	tempDir := os.TempDir()
	testFile := filepath.Join(tempDir, fmt.Sprintf("test_%d.mp4", time.Now().Unix()))

	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi",
		"-i", fmt.Sprintf("testsrc2=duration=%d:size=%s:rate=30", duration, resolution),
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "23",
		testFile)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to generate test video: %w", err)
	}
	return testFile, nil
}

func ParseProfileNames(profilesStr string) []string {
	if profilesStr == "" {
		return []string{}
	}

	names := strings.Split(profilesStr, ";")
	result := make([]string, 0, len(names))

	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func GetProfilesByNames(names []string) ([]EncodingProfile, error) {
	availableProfiles := GetAvailableProfiles()
	profiles := make([]EncodingProfile, 0, len(names))

	for _, name := range names {
		profile, exists := availableProfiles[name]
		if !exists {
			return nil, fmt.Errorf("invalid profile name: %s. Available profiles: %v",
				name, getAvailableProfileNames())
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func getAvailableProfileNames() []string {
	profiles := GetAvailableProfiles()
	names := make([]string, 0, len(profiles))
	for name := range profiles {
		names = append(names, name)
	}
	return names
}

func BenchmarkProfiles(profiles []EncodingProfile, inputVideo string) ([]EncodingProfile, error) {
	benchmarked := make([]EncodingProfile, 0, len(profiles))

	for _, profile := range profiles {
		tempDir := os.TempDir()
		outputFile := filepath.Join(tempDir, fmt.Sprintf("bench_%s_%d.mp4", profile.Name, time.Now().Unix()))

		// Encode benchmark
		encodeArgs := append([]string{"-y", "-i", inputVideo}, profile.Params...)
		encodeArgs = append(encodeArgs, outputFile)
		encodeCmd := exec.Command("ffmpeg", encodeArgs...)

		start := time.Now()
		output, err := encodeCmd.CombinedOutput()
		encodeTime := time.Since(start).Seconds()

		if err != nil {
			os.Remove(outputFile)
			return nil, fmt.Errorf("encoding failed for profile %s: %w\nOutput: %s", profile.Name, err, string(output))
		}

		fps, err := extractFPS(string(output))
		if err != nil {
			os.Remove(outputFile)
			return nil, fmt.Errorf("failed to extract fps for profile %s: %w", profile.Name, err)
		}

		// Decode benchmark
		decodeCmd := exec.Command("ffmpeg", "-y", "-i", outputFile, "-f", "null", "-")
		start = time.Now()
		err = decodeCmd.Run()
		decodeTime := time.Since(start).Seconds()
		if err != nil {
			decodeTime = 999999
		}

		// Update scores
		score := encodeTime + decodeTime

		profile.EncodeTime = encodeTime
		profile.DecodeTime = decodeTime
		profile.FPS = fps
		profile.Score = score

		benchmarked = append(benchmarked, profile)

		os.Remove(outputFile)
	}

	return benchmarked, nil
}
