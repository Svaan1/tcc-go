package ffmpeg

import (
	"fmt"
	"os/exec"
)

type EncodingParams struct {
	InputPath  string
	OutputPath string
	VideoCodec string
	AudioCodec string
	Crf        string
	Preset     string
}

func Encode(req *EncodingParams) error {
	args := []string{
		"-i", req.InputPath,
	}

	args = append(args, "-c:v", req.VideoCodec)
	args = append(args, "-c:a", req.AudioCodec)
	args = append(args, "-crf", req.Crf)
	args = append(args, "-preset", req.Preset)

	args = append(args, req.OutputPath)

	cmd := exec.Command("ffmpeg", args...)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	// Here you could read from stdout and stderr to monitor progress
	// For now, just wait for completion
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg execution failed: %v", err)
	}

	return nil
}
