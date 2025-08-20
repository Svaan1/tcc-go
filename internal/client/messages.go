package client

import (
	"context"
	"io"
	"log"

	pb "github.com/svaan1/go-tcc/internal/transcoding"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) handleServerMessages(ctx context.Context) {
	defer func() {
		c.conn.Close()
		log.Println("Server closed connection")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.stream.Recv()
			if err == io.EOF {
				log.Println("Server closed stream")
				return
			}
			if err != nil {
				log.Printf("Failed to receive message: %v", err)
				return
			}

			switch payload := msg.Payload.(type) {
			case *pb.OrchestratorMessage_RegisterResponse:
				log.Printf("Register response: success=%t, message=%s",
					payload.RegisterResponse.Success,
					payload.RegisterResponse.Message)

			case *pb.OrchestratorMessage_ResourceUsageResponse:
				log.Printf("Resource usage response: success=%t, message=%s",
					payload.ResourceUsageResponse.Success,
					payload.ResourceUsageResponse.Message)

			case *pb.OrchestratorMessage_JobAssignmentRequest:
				c.handleJobAssignment(ctx, payload.JobAssignmentRequest)

			case *pb.OrchestratorMessage_DisconnectResponse:
				log.Printf("Disconnect response: acknowledged=%t",
					payload.DisconnectResponse.Acknowledged)
				return

			default:
				log.Printf("Unknown message type received")
			}
		}
	}
}

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
