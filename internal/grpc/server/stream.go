package server

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
	pb "github.com/svaan1/tcc-go/internal/grpc/proto"
	"github.com/svaan1/tcc-go/internal/metrics"
)

func (sv *Server) Stream(stream pb.VideoTranscoding_StreamServer) error {
	log.Printf("New connection established")

	ctx := stream.Context()

	nodeID, nodeConn, err := sv.registerNode(ctx, stream)
	if err != nil {
		log.Printf("Failed to register node: %v", err)
		return err
	}

	defer sv.unregisterNode(ctx, nodeID)

	log.Printf("Starting message processing loop for node %s", nodeID)
	for {
		select {
		case <-nodeConn.closedChan:
			log.Printf("Node %s stream closing due to closed channel", nodeID)
			return nil
		default:
			msg, err := stream.Recv()
			if err == io.EOF {
				log.Printf("Stream closed by node %s", nodeID)
				return nil
			}
			if err != nil {
				log.Printf("Error receiving message from node %s: %v", nodeID, err)
				return err
			}

			switch payload := msg.Payload.(type) {
			case *pb.NodeMessage_ResourceUsageRequest:
				log.Printf("Node %s resource usage update: CPU=%.2f%%, Memory=%.2f%%, Disk=%.2f%%",
					nodeID,
					payload.ResourceUsageRequest.GetCpuPercent(),
					payload.ResourceUsageRequest.GetMemoryPercent(),
					payload.ResourceUsageRequest.GetDiskPercent(),
				)
				sv.Service.UpdateNodeMetrics(context.TODO(),
					payload.ResourceUsageRequest.NodeId,
					&metrics.ResourceUsage{
						CPUUsagePercent:    payload.ResourceUsageRequest.GetCpuPercent(),
						MemoryUsagePercent: payload.ResourceUsageRequest.GetMemoryPercent(),
						DiskUsagePercent:   payload.ResourceUsageRequest.GetDiskPercent(),
					},
				)
			case *pb.NodeMessage_JobAssignmentResponse:
				log.Printf("Node %s job assignment response for job %s: accepted=%t, message=%s",
					nodeID,
					payload.JobAssignmentResponse.GetJobId(),
					payload.JobAssignmentResponse.GetAccepted(),
					payload.JobAssignmentResponse.GetMessage(),
				)

				if !payload.JobAssignmentResponse.GetAccepted() {
					log.Printf("Node %s rejected job %s, marking for reassignment",
						nodeID,
						payload.JobAssignmentResponse.GetJobId(),
					)
					sv.Service.RejectJob(context.TODO(),
						payload.JobAssignmentResponse.GetJobId(),
						payload.JobAssignmentResponse.GetMessage(),
					)
				} else {
					log.Printf("Node %s accepted job %s, starting execution",
						nodeID,
						payload.JobAssignmentResponse.GetJobId(),
					)
				}
			case *pb.NodeMessage_JobCompletionRequest:
				log.Printf("Node %s job completion for job %s: success=%t, message=%s",
					nodeID,
					payload.JobCompletionRequest.GetJobId(),
					payload.JobCompletionRequest.GetSuccess(),
					payload.JobCompletionRequest.GetMessage(),
				)
				sv.Service.CompleteJob(context.TODO(),
					payload.JobCompletionRequest.GetJobId(),
					payload.JobCompletionRequest.GetSuccess(),
					payload.JobCompletionRequest.GetMessage(),
				)
			case *pb.NodeMessage_DisconnectRequest:
				log.Printf("Node %s requested disconnect", nodeID)
				nodeConn.SendDisconnectResponse(&pb.DisconnectResponse{Acknowledged: true})
				return nil
			default:
				log.Printf("Node %s sent unknown message type", nodeID)
			}
		}
	}
}

func (sv *Server) registerNode(ctx context.Context, stream pb.VideoTranscoding_StreamServer) (nodeID uuid.UUID, nodeConn *NodeConn, err error) {
	log.Println("Waiting for node registration request")
	msg, err := stream.Recv()
	if err != nil {
		log.Printf("Error receiving initial message: %v", err)
		return uuid.Nil, nil, fmt.Errorf("error receiving initial message: %v", err)
	}

	register := msg.GetRegisterRequest()
	if register == nil {
		log.Println("Received non-RegisterRequest message during registration")
		return uuid.Nil, nil, fmt.Errorf("expected RegisterRequest but got different message type")
	}

	log.Printf("Received registration request from node: %s with %d encoding profiles",
		register.Name, len(register.EncodingProfiles))

	// Parse the encoding profile structs
	var profiles []ffmpeg.EncodingProfile
	for _, profile := range register.EncodingProfiles {
		profiles = append(profiles, ffmpeg.EncodingProfile{
			Name:       profile.GetName(),
			Codec:      profile.GetCodec(),
			Params:     profile.GetParams(),
			EncodeTime: profile.GetEncodeTime(),
			DecodeTime: profile.GetDecodeTime(),
			FPS:        profile.GetFps(),
			Score:      profile.GetScore(),
		})
	}

	nodeID, err = sv.Service.RegisterNode(ctx, register.Name, profiles)
	if err != nil {
		log.Printf("Failed to register node %s: %v", register.Name, err)
		return uuid.Nil, nil, fmt.Errorf("failed to register node: %v", err)
	}

	log.Printf("Successfully registered node %s with ID %s", register.Name, nodeID)

	nodeConn = newNodeConn(nodeID, stream)

	sv.mu.Lock()
	sv.NodeConns[nodeID] = nodeConn
	sv.mu.Unlock()

	if err := nodeConn.SendRegisterResponse(); err != nil {
		log.Printf("Error sending registration response to %s: %v", nodeID, err)
		return uuid.Nil, nil, fmt.Errorf("error sending registration response to %s: %v", nodeID, err)
	}

	log.Printf("Sent registration response to node %s", nodeID)

	return nodeID, nodeConn, nil
}

func (sv *Server) unregisterNode(ctx context.Context, nodeID uuid.UUID) {
	log.Printf("Unregistering node %s", nodeID.String())
	sv.mu.Lock()
	sv.Service.UnregisterNode(ctx, nodeID)
	delete(sv.NodeConns, nodeID)
	sv.mu.Unlock()

	log.Printf("Node %s disconnected and removed from registry", nodeID.String())
}
