package server

import (
	"log"

	"github.com/svaan1/go-tcc/internal/protocols"
)

func (sv *Server) handleClientMessages(node *Node) {
	defer sv.closeConnection(node)

	for {
		packet, err := protocols.ReceivePacket(node.Conn)
		if err != nil {
			log.Println("Failed to receive packet, closing connection", err)
			return
		}

		switch packet.Type {
		case protocols.ResourceUsageType:
			protocols.NewResourceUsageFromPacketData(packet.Data)

		default:
			log.Println("Received invalid packet type, closing connection", packet.Type)
			return
		}

	}
}
