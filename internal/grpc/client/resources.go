package client

import (
	"context"
	"log"
	"time"

	"github.com/svaan1/go-tcc/internal/app"
	pb "github.com/svaan1/go-tcc/internal/grpc/transcoding"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) handleResourceUsagePolling(ctx context.Context) {
	ticker := time.NewTicker(c.Config.ResourceUsagePollingTickTime)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

			resources, err := app.GetAvailableResources()
			if err != nil {
				log.Printf("Failed to fetch available resources: %v", err)
				continue
			}

			// Create resource usage message
			resourceMsg := &pb.NodeMessage{
				Base: &pb.MessageBase{
					MessageId: "resource-usage-" + c.nodeID,
					Timestamp: timestamppb.Now(),
				},
				Payload: &pb.NodeMessage_ResourceUsageRequest{
					ResourceUsageRequest: &pb.ResourceUsageRequest{
						NodeId:        c.nodeID,
						CpuPercent:    resources.CPUUsagePercent,
						MemoryPercent: resources.MemoryUsagePercent,
						DiskPercent:   resources.DiskUsagePercent,
					},
				},
			}

			err = c.stream.Send(resourceMsg)
			if err != nil {
				log.Printf("Failed to send resource usage: %v", err)
				continue
			}

			log.Printf("Sent resource usage: CPU=%.2f%%, Memory=%.2f%%, Disk=%.2f%%",
				resources.CPUUsagePercent, resources.MemoryUsagePercent, resources.DiskUsagePercent)
		}
	}
}
