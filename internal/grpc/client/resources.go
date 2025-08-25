package client

import (
	"context"
	"log"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	pb "github.com/svaan1/go-tcc/internal/transcoding"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func GetCPU() (float64, error) {
	stat, err := cpu.Percent(0, false)
	if err != nil {
		return 0, err
	}

	return stat[0], nil
}

func GetMemory() (float64, error) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}

	return stat.UsedPercent, nil
}

func GetDisk() (float64, error) {
	stat, err := disk.Usage("/")
	if err != nil {
		return 0, err
	}

	return stat.UsedPercent, nil
}

func (c *Client) handleResourceUsagePolling(ctx context.Context) {
	ticker := time.NewTicker(c.Config.ResourceUsagePollingTickTime)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cpu, err := GetCPU()
			if err != nil {
				log.Println("Failed to get CPU usage", err)
				continue
			}

			mem, err := GetMemory()
			if err != nil {
				log.Println("Failed to get Memory usage", err)
				continue
			}

			disk, err := GetDisk()
			if err != nil {
				log.Println("Failed to get Disk usage", err)
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
						CpuPercent:    cpu,
						MemoryPercent: mem,
						DiskPercent:   disk,
					},
				},
			}

			err = c.stream.Send(resourceMsg)
			if err != nil {
				log.Printf("Failed to send resource usage: %v", err)
				continue
			}

			log.Printf("Sent resource usage: CPU=%.2f%%, Memory=%.2f%%, Disk=%.2f%%",
				cpu, mem, disk)
		}
	}
}
