package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

func GenerateEncodingProfile(name, codec string, params []string, weight float64, inputVideo string) (EncodingProfile, error) {
	tempDir := os.TempDir()
	outputFile := filepath.Join(tempDir, fmt.Sprintf("out_%s_%d.mp4", name, time.Now().Unix()))
	defer os.Remove(outputFile)

	// Encode benchmark
	encodeArgs := append([]string{"-y", "-i", inputVideo}, params...)
	encodeArgs = append(encodeArgs, outputFile)
	encodeCmd := exec.Command("ffmpeg", encodeArgs...)

	start := time.Now()
	output, err := encodeCmd.CombinedOutput()
	encodeTime := time.Since(start).Seconds()

	if err != nil {
		return EncodingProfile{}, fmt.Errorf("encoding failed: %w", err)
	}

	fps, err := extractFPS(string(output))
	if err != nil {
		return EncodingProfile{}, fmt.Errorf("failed to extract fps %v", err)
	}

	// Decode benchmark
	decodeCmd := exec.Command("ffmpeg", "-y", "-i", outputFile, "-f", "null", "-")
	start = time.Now()
	err = decodeCmd.Run()
	decodeTime := time.Since(start).Seconds()
	if err != nil {
		decodeTime = 999999 // High penalty for decode failure
	}

	// Calculate score (encode + decode time, weighted)
	score := (encodeTime + decodeTime) / weight

	return EncodingProfile{
		Name:       name,
		Params:     params,
		Codec:      codec,
		EncodeTime: encodeTime,
		DecodeTime: decodeTime,
		FPS:        fps,
		Score:      score,
	}, nil
}
