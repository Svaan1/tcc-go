package server

import (
	"log"
	"time"
)

func (sv *Server) handleResourceUsagePolling(node *Node) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-node.closedChan:
			return
		case <-ticker.C:
			if time.Since(node.ResourceUsage.Time) > sv.Config.ResourceUsagePollingTimeout {
				log.Printf("Node %d resource usage timed out", node.ID)
				sv.closeConnection(node)
				return
			}
		}
	}
}
