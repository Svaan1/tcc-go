package ffmpeg

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

type EncodingParams struct {
	VideoCodec string
	AudioCodec string
	Crf        string
	Preset     string
}

func Encode(req EncodingParams, input io.Reader, output io.Writer) error {
	args := []string{
		"-i", "pipe:0",
	}

	args = append(args, "-c:v", req.VideoCodec)
	args = append(args, "-c:a", req.AudioCodec)
	args = append(args, "-crf", req.Crf)
	args = append(args, "-preset", req.Preset)
	args = append(args, "-f", "mp4")
	args = append(args, "-movflags", "frag_keyframe+empty_moov")
	args = append(args, "pipe:1")

	cmd := exec.Command("ffmpeg", args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Handle stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println("ffmpeg:", scanner.Text())
		}
	}()

	errChan := make(chan error, 2)

	// Stream input to ffmpeg stdin
	go func() {
		defer stdin.Close()
		buf := make([]byte, 32*1024)
		for {
			n, err := input.Read(buf)
			if n > 0 {
				if _, writeErr := stdin.Write(buf[:n]); writeErr != nil {
					errChan <- fmt.Errorf("write to ffmpeg stdin failed: %w", writeErr)
					return
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				errChan <- fmt.Errorf("read input failed: %w", err)
				return
			}
		}
		errChan <- nil
	}()

	// Stream ffmpeg stdout to output writer
	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				if _, writeErr := output.Write(buf[:n]); writeErr != nil {
					errChan <- fmt.Errorf("write output failed: %w", writeErr)
					return
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				errChan <- fmt.Errorf("read from ffmpeg stdout failed: %w", err)
				return
			}
		}
		errChan <- nil
	}()

	// Wait for both goroutines
	for range 2 {
		if err := <-errChan; err != nil {
			cmd.Process.Kill()
			return err
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w", err)
	}

	return nil
}
