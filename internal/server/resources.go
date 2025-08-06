package server

import (
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
				node.logger.Printf("Node resource usage timed out")
				sv.closeConnection(node)
				return
			}
		}
	}
}
