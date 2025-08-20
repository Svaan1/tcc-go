package client

import (
	"context"
	"log"

	pb "github.com/svaan1/go-tcc/internal/transcoding"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) handleJobAssignment(ctx context.Context, jobRequest *pb.JobAssignmentRequest) {
	log.Printf("Received job assignment: ID=%s, input=%s, output=%s",
		jobRequest.JobId, jobRequest.InputPath, jobRequest.OutputPath)

	// For now, just accept all jobs
	// In a real implementation, you would check if the node can handle the job
	accepted := true
	message := "Job accepted"

	response := &pb.NodeMessage{
		Base: &pb.MessageBase{
			MessageId: "job-response-" + jobRequest.JobId,
			Timestamp: timestamppb.Now(),
		},
		Payload: &pb.NodeMessage_JobAssignmentResponse{
			JobAssignmentResponse: &pb.JobAssignmentResponse{
				JobId:    jobRequest.JobId,
				Accepted: accepted,
				Message:  message,
			},
		},
	}

	err := c.stream.Send(response)
	if err != nil {
		log.Printf("Failed to send job response: %v", err)
	}
}
