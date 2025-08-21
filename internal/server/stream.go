package server

import (
	"io"
	"log"

	pb "github.com/svaan1/go-tcc/internal/transcoding"
)

func (sv *Server) Stream(stream pb.VideoTranscoding_StreamServer) error {
	log.Printf("New connection established")

	node, err := sv.RegisterNode(stream)
	if err != nil {
		return err
	}

	defer func() {
		sv.mu.Lock()
		delete(sv.Nodes, node.ID)
		sv.mu.Unlock()
		log.Printf("Node %s (%s) disconnected", node.Name, node.ID.String())
	}()

	log.Printf("Starting message processing loop for node %s", node.Name)
	for {
		select {
		case <-node.closedChan:
			log.Printf("Node %s stream closing due to closed channel", node.Name)
			return nil
		default:
			msg, err := stream.Recv()
			if err == io.EOF {
				log.Printf("Stream closed by node %s", node.Name)
				return nil
			}
			if err != nil {
				log.Printf("Error receiving message from node %s: %v", node.Name, err)
				return err
			}

			log.Printf("Received message from node %s - Message ID: %s, Timestamp: %v",
				node.Name, msg.Base.MessageId, msg.Base.Timestamp.AsTime())

			switch payload := msg.Payload.(type) {
			case *pb.NodeMessage_ResourceUsageRequest:
				node.SetResourceUsage(payload.ResourceUsageRequest, msg.Base.Timestamp.AsTime())
			case *pb.NodeMessage_JobAssignmentResponse:
				log.Printf("Job assignment response: job_id=%s, accepted=%t, message=%s",
					payload.JobAssignmentResponse.JobId,
					payload.JobAssignmentResponse.Accepted,
					payload.JobAssignmentResponse.Message)
			case *pb.NodeMessage_DisconnectRequest:
				log.Printf("Disconnect request: reason=%s",
					payload.DisconnectRequest.Reason)
			default:
				log.Printf("Unknown message type received")
			}
		}
	}
}
