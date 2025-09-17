package client

import (
	"context"
	"log"

	"github.com/svaan1/tcc-go/internal/ffmpeg"
	pb "github.com/svaan1/tcc-go/internal/grpc/transcoding"
)

func (c *Client) handleJobAssignment(ctx context.Context, jobRequest *pb.JobAssignmentRequest) {
	_ = ctx
	log.Printf("Received job assignment: ID=%s, input=%s, output=%s",
		jobRequest.JobId, jobRequest.InputPath, jobRequest.OutputPath)

	// For now, just accept all jobs
	// In a real implementation, you would check if the node can handle the job
	err := ffmpeg.Encode(&ffmpeg.EncodingParams{
		InputPath:  jobRequest.InputPath,
		OutputPath: jobRequest.OutputPath,
		VideoCodec: jobRequest.VideoCodec,
		AudioCodec: jobRequest.AudioCodec,
		Crf:        jobRequest.Crf,
		Preset:     jobRequest.Preset,
	})

	if err != nil {
		log.Printf("Failed to execute ffmpeg: %v", err)
	}
}
