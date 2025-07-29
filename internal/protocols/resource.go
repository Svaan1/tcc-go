package protocols

import (
	"log"
	"time"
)

const ResourceUsageType string = "resource_usage"

type ResourceUsage struct {
	CPU    float64
	Memory float64
	Disk   float64
	Time   time.Time
}

func NewResourceUsagePacket(cpu float64, memory float64, disk float64, time time.Time) *Packet {
	return &Packet{
		Type: ResourceUsageType,
		Data: ResourceUsage{
			CPU:    cpu,
			Memory: memory,
			Disk:   disk,
			Time:   time,
		},
	}
}

func NewResourceUsageFromPacketData(data any) *ResourceUsage {
	m, ok := data.(map[string]any)

	if !ok {
		log.Println("Failed to create ResourceUsage from data:", data)
		return nil
	}

	return &ResourceUsage{
		CPU:    m["CPU"].(float64),
		Memory: m["Memory"].(float64),
		Disk:   m["Disk"].(float64),
		Time:   m["Time"].(time.Time),
	}
}
