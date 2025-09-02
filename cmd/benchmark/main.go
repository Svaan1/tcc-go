package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CodecProfile struct {
	Name         string   `json:"name"`
	FFmpegParams []string `json:"ffmpeg_params"`
	WeightFactor float64  `json:"weight_factor"`
}

type BenchmarkResult struct {
	Codec      string  `json:"codec"`
	EncodeTime float64 `json:"encode_time"`
	DecodeTime float64 `json:"decode_time"`
	FPS        float64 `json:"fps"`
	Score      float64 `json:"score"` // Lower is better
}

type SpeedBenchmark struct {
	CodecProfiles map[string]CodecProfile    `json:"codec_profiles"`
	Results       map[string]BenchmarkResult `json:"results"`
}

func NewSpeedBenchmark() *SpeedBenchmark {
	return &SpeedBenchmark{
		CodecProfiles: make(map[string]CodecProfile),
		Results:       make(map[string]BenchmarkResult),
	}
}

func (sb *SpeedBenchmark) AddCodec(name string, params []string, weight float64) {
	sb.CodecProfiles[name] = CodecProfile{
		Name:         name,
		FFmpegParams: params,
		WeightFactor: weight,
	}
}

func (sb *SpeedBenchmark) GenerateTestVideo(duration int, resolution string) (string, error) {
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

func (sb *SpeedBenchmark) extractFPS(output string) float64 {
	re := regexp.MustCompile(`frame=\s*\d+\s+fps=(\d+\.?\d*)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		if fps, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return fps
		}
	}
	return 0.0
}

func (sb *SpeedBenchmark) BenchmarkCodec(codecName, inputVideo string) (BenchmarkResult, error) {
	profile, exists := sb.CodecProfiles[codecName]
	if !exists {
		return BenchmarkResult{}, fmt.Errorf("codec %s not found", codecName)
	}

	tempDir := os.TempDir()
	outputFile := filepath.Join(tempDir, fmt.Sprintf("out_%s_%d.mp4", codecName, time.Now().Unix()))
	defer os.Remove(outputFile)

	// Encode benchmark
	encodeArgs := append([]string{"-y", "-i", inputVideo}, profile.FFmpegParams...)
	encodeArgs = append(encodeArgs, outputFile)
	encodeCmd := exec.Command("ffmpeg", encodeArgs...)

	start := time.Now()
	output, err := encodeCmd.CombinedOutput()
	encodeTime := time.Since(start).Seconds()

	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("encoding failed: %w", err)
	}

	log.Print(string(output))

	fps := sb.extractFPS(string(output))

	// Decode benchmark
	decodeCmd := exec.Command("ffmpeg", "-y", "-i", outputFile, "-f", "null", "-")
	start = time.Now()
	err = decodeCmd.Run()
	decodeTime := time.Since(start).Seconds()
	if err != nil {
		decodeTime = 999999 // High penalty for decode failure
	}

	// Calculate score (encode + decode time, weighted)
	score := (encodeTime + decodeTime) / profile.WeightFactor

	return BenchmarkResult{
		Codec:      codecName,
		EncodeTime: encodeTime,
		DecodeTime: decodeTime,
		FPS:        fps,
		Score:      score,
	}, nil
}

func (sb *SpeedBenchmark) RunBenchmark(testDuration int, resolution string) error {
	if len(sb.CodecProfiles) == 0 {
		return fmt.Errorf("no codecs configured")
	}

	testVideo, err := sb.GenerateTestVideo(testDuration, resolution)
	if err != nil {
		return err
	}
	defer os.Remove(testVideo)

	for codecName := range sb.CodecProfiles {
		fmt.Printf("Testing %s...\n", codecName)
		result, err := sb.BenchmarkCodec(codecName, testVideo)
		if err != nil {
			log.Printf("Failed %s: %v", codecName, err)
			continue
		}
		sb.Results[codecName] = result
	}
	return nil
}

func (sb *SpeedBenchmark) PrintResults() {
	if len(sb.Results) == 0 {
		fmt.Println("No results")
		return
	}

	fmt.Printf("\n%-15s %-12s %-12s %-8s %-10s\n", "Codec", "Encode(s)", "Decode(s)", "FPS", "Score")
	fmt.Println(strings.Repeat("-", 60))

	for codec, result := range sb.Results {
		fmt.Printf("%-15s %-12.2f %-12.2f %-8.1f %-10.2f\n",
			codec, result.EncodeTime, result.DecodeTime, result.FPS, result.Score)
	}
}

func (sb *SpeedBenchmark) GetCodecScore(codec string) float64 {
	if result, exists := sb.Results[codec]; exists {
		return result.Score
	}
	return -1 // Codec not found
}

func (sb *SpeedBenchmark) SaveResults(filename string) error {
	data, err := json.MarshalIndent(sb.Results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (sb *SpeedBenchmark) ExportNodeInfo() map[string]interface{} {
	hostname, _ := os.Hostname()
	return map[string]interface{}{
		"hostname":  hostname,
		"codecs":    sb.CodecProfiles,
		"results":   sb.Results,
		"timestamp": time.Now().Unix(),
	}
}

func main() {
	benchmark := NewSpeedBenchmark()

	// User-defined codecs
	benchmark.AddCodec("x264_fast", []string{"-c:v", "libx264", "-preset", "fast"}, 1.0)
	benchmark.AddCodec("x264_medium", []string{"-c:v", "libx264", "-preset", "medium"}, 1.2)
	benchmark.AddCodec("nvenc_h264", []string{"-c:v", "h264_nvenc", "-preset", "fast"}, 0.8)
	benchmark.AddCodec("nvenc_h265", []string{"-c:v", "h265_nvenc", "-preset", "fast"}, 0.9)
	benchmark.AddCodec("amf_h264", []string{"-c:v", "h264_amf", "-quality", "speed"}, 0.7)

	// Run benchmark (10 second test, 1080p)
	if err := benchmark.RunBenchmark(100, "1920x1080"); err != nil {
		log.Fatal(err)
	}

	benchmark.PrintResults()
	fmt.Printf("\nCodec scores (lower = faster):\n")
	for codec := range benchmark.Results {
		score := benchmark.GetCodecScore(codec)
		fmt.Printf("%s: %.2f\n", codec, score)
	}

	// Export for distributed system
	nodeInfo := benchmark.ExportNodeInfo()
	nodeJSON, _ := json.MarshalIndent(nodeInfo, "", "  ")
	fmt.Printf("\nNode info:\n%s\n", string(nodeJSON))

	benchmark.SaveResults("speed_results.json")
}
