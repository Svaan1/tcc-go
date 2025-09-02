package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Codec struct {
	Name   string   `json:"name"`
	Params []string `json:"params"`
}

type BenchmarkResult struct {
	Codec      string  `json:"codec"`
	EncodeTime float64 `json:"encode_time"`
	DecodeTime float64 `json:"decode_time"`
	FPS        float64 `json:"fps"`
	Score      float64 `json:"score"` // lower is better
}

func BenchmarkCodec(codec Codec, weight float64, inputVideo string) (BenchmarkResult, error) {
	tempDir := os.TempDir()
	outputFile := filepath.Join(tempDir, fmt.Sprintf("out_%s_%d.mp4", codec.Name, time.Now().Unix()))
	defer os.Remove(outputFile)

	// Encode benchmark
	encodeArgs := append([]string{"-y", "-i", inputVideo}, codec.Params...)
	encodeArgs = append(encodeArgs, outputFile)
	encodeCmd := exec.Command("ffmpeg", encodeArgs...)

	start := time.Now()
	output, err := encodeCmd.CombinedOutput()
	encodeTime := time.Since(start).Seconds()

	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("encoding failed: %w", err)
	}

	fps, err := extractFPS(string(output))
	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("failed to extract fps %v", err)
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

	return BenchmarkResult{
		EncodeTime: encodeTime,
		DecodeTime: decodeTime,
		FPS:        fps,
		Score:      score,
	}, nil
}
