package client

import (
	"context"
	"io"
	"log"

	pb "github.com/svaan1/tcc-go/internal/grpc/proto"
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
