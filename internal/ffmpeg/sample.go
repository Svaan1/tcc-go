package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

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
