package ffmpeg

import (
	"fmt"
	"os/exec"

	pb "github.com/svaan1/go-tcc/internal/transcoding"
)

// Encode executes ffmpeg with the parameters from JobAssignmentRequest
func Encode(req *pb.JobAssignmentRequest) error {
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
