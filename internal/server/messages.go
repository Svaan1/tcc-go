package server

import (
	"github.com/svaan1/go-tcc/internal/protocols"
)

func (sv *Server) handleClientMessages(node *Node) {
	defer sv.closeConnection(node)

	for {
		select {
		case <-node.closedChan:
			return
		default:
			packet, err := protocols.ReceivePacket(node.conn)
			if err != nil {
				return
			}

			switch packet.Type {
			case protocols.ResourceUsageType:
				node.ResourceUsage = *protocols.NewResourceUsageFromPacketData(packet.Data)

			default:
				return
			}
		}
	}
}
