package server

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/svaan1/tcc-go/internal/ffmpeg"
	pb "github.com/svaan1/tcc-go/internal/grpc/transcoding"
	"github.com/svaan1/tcc-go/internal/metrics"
)

func (sv *Server) Stream(stream pb.VideoTranscoding_StreamServer) error {
	log.Printf("New connection established")

	msg, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("error receiving initial message: %v", err)
	}

	register := msg.GetRegisterRequest()
	if register == nil {
		return fmt.Errorf("expected RegisterRequest but got different message type")
	}

	node, err := sv.Service.RegisterNode(context.TODO(), register.Name, make([]ffmpeg.EncodingProfile, 0))
	if err != nil {
		return fmt.Errorf("failed to register node: %v", err)
	}

	nodeConn := newNodeConn(node.ID, stream)

	sv.mu.Lock()
	sv.NodeConns[node.ID] = nodeConn
	sv.mu.Unlock()

	if err := nodeConn.SendRegisterResponse(); err != nil {
		return fmt.Errorf("error sending registration response to %s: %v", node.Name, err)
	}

	defer func() {
		sv.mu.Lock()
		sv.Service.UnregisterNode(context.TODO(), node.ID)
		delete(sv.NodeConns, node.ID)
		sv.mu.Unlock()

		log.Printf("Node %s (%s) disconnected", node.Name, node.ID.String())
	}()

	log.Printf("Starting message processing loop for node %s", node.Name)
	for {
		select {
		case <-nodeConn.closedChan:
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

			// log.Printf("Received message from node %s - Message ID: %s, Timestamp: %v",
			// 	node.Name, msg.Base.MessageId, msg.Base.Timestamp.AsTime())

			// ts := msg.Base.Timestamp.AsTime()

			switch payload := msg.Payload.(type) {
			case *pb.NodeMessage_ResourceUsageRequest:
				sv.Service.UpdateNodeMetrics(context.TODO(),
					payload.ResourceUsageRequest.NodeId,
					&metrics.ResourceUsage{
						CPUUsagePercent:    payload.ResourceUsageRequest.CpuPercent,
						MemoryUsagePercent: payload.ResourceUsageRequest.MemoryPercent,
						DiskUsagePercent:   payload.ResourceUsageRequest.DiskPercent,
					},
				)
			case *pb.NodeMessage_DisconnectRequest:
				nodeConn.SendDisconnectResponse(&pb.DisconnectResponse{
					Acknowledged: true,
				})

				sv.mu.Lock()
				close(nodeConn.closedChan)
				sv.mu.Unlock()
			}
		}
	}
}
