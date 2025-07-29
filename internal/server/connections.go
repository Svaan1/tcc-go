package server

import (
	"log"
	"net"

	"github.com/google/uuid"
	"github.com/svaan1/go-tcc/internal/protocols"
)

type Node struct {
	ID     uuid.UUID
	Conn   *net.Conn
	Name   string
	Codecs []string

	ResourceUsage protocols.ResourceUsage
}

func (sv *Server) handleConnections() {
	defer (*sv.listener).Close()

	for {
		conn, err := (*sv.listener).Accept()
		if err != nil {
			log.Println("Failed to accept connection", err)
			continue
		}

		packet, err := protocols.ReceivePacket(&conn)
		if err != nil {
			log.Println("Failed to receive packet", err)
			conn.Close()
			continue
		}

		if packet.Type != protocols.RegisterType {
			log.Println("Connection failed to send register packet")
			conn.Close()
			return
		}

		register := protocols.NewRegisterFromPacketData(packet.Data)
		if register == nil {
			log.Println("Failed to create register protocol from packet data", err)
			conn.Close()
			return
		}

		node := Node{
			ID:     uuid.New(),
			Conn:   &conn,
			Name:   register.Name,
			Codecs: register.Codecs,
		}
		sv.nodes[node.ID] = node

		go sv.handleClientMessages(&node)

		log.Println("Node registered", node)
	}
}

func (sv *Server) closeConnection(node *Node) {
	(*node.Conn).Close()
	delete(sv.nodes, node.ID)
	log.Printf("Node disconnected %s", node.ID)

}
