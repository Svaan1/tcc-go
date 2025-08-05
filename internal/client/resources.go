package client

import (
	"context"
	"log"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/svaan1/go-tcc/internal/protocols"
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
			}
			mem, err := GetMemory()
			if err != nil {
				log.Println("Failed to get Memory usage", err)
			}
			disk, err := GetDisk()
			if err != nil {
				log.Println("Failed to get Disk usage", err)
			}

			now := time.Now()

			resourceUsage := protocols.NewResourceUsagePacket(cpu, mem, disk, now)
			err = protocols.SendPacket(c.conn, *resourceUsage)
			if err != nil {
				log.Println("Failed to send resource usage packet", err)
				continue
			}

		}
	}
}
