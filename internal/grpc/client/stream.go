package client

import (
	"context"
	"io"
	"log"

	pb "github.com/svaan1/tcc-go/internal/grpc/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) handleStream(ctx context.Context) {
	defer func() {
		c.Close()
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

			case *pb.OrchestratorMessage_JobAssignmentRequest:
				jobRequest := payload.JobAssignmentRequest

				accepted, message := c.Service.AcceptJobAssignment(ctx, jobRequest.JobId, jobRequest.InputPath, jobRequest.OutputPath, jobRequest.ProfileName)

				jobAssignmentResponseMsg := &pb.NodeMessage{
					Base: &pb.MessageBase{
						MessageId: "job-assignment-response-" + jobRequest.JobId,
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

				if sendErr := c.stream.Send(jobAssignmentResponseMsg); sendErr != nil {
					log.Printf("Failed to send job assignment response: %v", sendErr)
					return
				}

				// Only process the job if it was accepted
				if !accepted {
					log.Printf("Job %s was rejected: %s", jobRequest.JobId, message)
					return
				}

				go func() {
					err := c.Service.HandleJobAssignment(ctx, jobRequest.InputPath, jobRequest.OutputPath, jobRequest.ProfileName)

					if err != nil {
						log.Printf("Failed to handle job assignment: %v", err)
					}

					jobCompletionMsg := &pb.NodeMessage{
						Base: &pb.MessageBase{
							MessageId: "job-completion-" + jobRequest.JobId,
							Timestamp: timestamppb.Now(),
						},
						Payload: &pb.NodeMessage_JobCompletionRequest{
							JobCompletionRequest: &pb.JobCompletionRequest{
								JobId:   jobRequest.JobId,
								Success: err == nil,
								Message: func() string {
									if err != nil {
										return err.Error()
									}
									return ""
								}(),
							},
						},
					}

					if err := c.stream.Send(jobCompletionMsg); err != nil {
						log.Printf("Failed to send job completion message: %v", err)
					}
				}()
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
