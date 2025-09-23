package client

import (
	"context"
	"log"

	"github.com/svaan1/tcc-go/internal/ffmpeg"
	pb "github.com/svaan1/tcc-go/internal/grpc/proto"
)

func (c *Client) handleJobAssignment(ctx context.Context, jobRequest *pb.JobAssignmentRequest) {
	_ = ctx
	log.Printf("Received job assignment: ID=%s, input=%s, output=%s",
		jobRequest.JobId, jobRequest.InputPath, jobRequest.OutputPath)

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
