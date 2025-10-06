package node

import (
	"context"
	"fmt"
	"io"

	"github.com/svaan1/tcc-go/internal/config"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
	"github.com/svaan1/tcc-go/internal/storage"
)

var (
	inputBucket  = "input-videos"
	outputBucket = "output-videos"
)

type Service struct {
	storage  storage.Storage
	profiles map[string]ffmpeg.EncodingProfile
}

func NewService(profiles []ffmpeg.EncodingProfile) *Service {
	profileMap := make(map[string]ffmpeg.EncodingProfile)
	for _, p := range profiles {
		profileMap[p.Name] = p
	}

	return &Service{
		storage:  storage.NewMinioStorage(config.MinioHostName, config.MinioAccessKey, config.MinioSecretKey, false),
		profiles: profileMap,
	}
}

func (s *Service) AcceptJobAssignment(ctx context.Context, jobID, inputPath, outputPath, profileName string) (bool, string) {
	// Check if we have the requested profile
	if _, exists := s.profiles[profileName]; !exists {
		return false, fmt.Sprintf("Profile %s not available on this node", profileName)
	}

	// Additional checks could include:
	// - Available disk space
	// - Current workload
	// - Resource availability
	// - Queue size, etc.

	return true, ""
}

func (s *Service) HandleJobAssignment(ctx context.Context, inputObject, outputObject, profileName string) error {
	// Get the profile
	profile, exists := s.profiles[profileName]
	if !exists {
		return fmt.Errorf("profile %s not found", profileName)
	}

	// Download from MinIO
	inputReader, err := s.storage.Download(ctx, inputBucket, inputObject)
	if err != nil {
		return fmt.Errorf("download error: %v", err)
	}
	defer inputReader.Close()

	// Create pipe for output
	pr, pw := io.Pipe()

	// Process video with ffmpeg using the profile
	go func() {
		defer pw.Close()
		if err := ffmpeg.EncodeWithProfile(profile, inputReader, pw); err != nil {
			pw.CloseWithError(err)
		}
	}()

	// Upload to MinIO (reads from pipe as ffmpeg writes)
	if err := s.storage.Upload(ctx, outputBucket, outputObject, pr, -1, "video/mp4"); err != nil {
		return fmt.Errorf("upload error: %v", err)
	}

	return nil
}
