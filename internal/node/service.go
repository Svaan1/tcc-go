package node

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/svaan1/tcc-go/internal/config"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
	"github.com/svaan1/tcc-go/internal/storage"
)

var (
	inputBucket  = "input-videos"
	outputBucket = "output-videos"
)

type Service struct {
	storage storage.Storage
}

func NewService() *Service {
	return &Service{
		storage: storage.NewMinioStorage(config.MinioHostName, config.MinioAccessKey, config.MinioSecretKey, false),
	}
}

// AcceptJobAssignment decides whether to accept or reject a job assignment
// Returns true if accepted, false if rejected, along with an optional message
func (s *Service) AcceptJobAssignment(ctx context.Context, jobID, inputPath, outputPath string) (bool, string) {
	// Add your business logic here to determine if the job should be accepted
	// For now, we'll accept all jobs
	// You could add checks for:
	// - Available disk space
	// - Current workload
	// - Resource availability
	// - Queue size, etc.

	return true, "Job accepted"
}

func (s *Service) HandleJobAssignment(ctx context.Context, inputObject, outputObject string, req ffmpeg.EncodingParams) error {
	// Download from MinIO
	inputReader, err := s.storage.Download(ctx, inputBucket, inputObject)
	if err != nil {
		return fmt.Errorf("download error: %v", err)
	}
	defer inputReader.Close()

	// Create pipe for output
	pr, pw := io.Pipe()

	// Process video with ffmpeg
	go func() {
		defer pw.Close()
		if err := ffmpeg.Encode(req, inputReader, pw); err != nil {
			pw.CloseWithError(err)
			log.Printf("ffmpeg error: %v", err)
		}
	}()

	// Upload to MinIO (reads from pipe as ffmpeg writes)
	if err := s.storage.Upload(ctx, outputBucket, outputObject, pr, -1, "video/mp4"); err != nil {
		fmt.Fprintf(os.Stderr, "Upload error: %v\n", err)
		os.Exit(1)
	}

	return nil
}
