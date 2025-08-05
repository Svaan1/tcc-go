package server

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/svaan1/go-tcc/internal/protocols"
)

type Node struct {
	ID         uuid.UUID
	Name       string
	Codecs     []string
	conn       *net.Conn
	closedOnce sync.Once
	closedChan chan struct{}

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

		node := &Node{
			ID:         uuid.New(),
			Name:       register.Name,
			Codecs:     register.Codecs,
			conn:       &conn,
			closedChan: make(chan struct{}),
			ResourceUsage: protocols.ResourceUsage{
				Time: time.Now(),
			},
		}
		sv.nodes[node.ID] = node

		go sv.handleClientMessages(node)
		go sv.handleResourceUsagePolling(node)

		log.Println("Node registered", node)
	}
}

func (sv *Server) closeConnection(node *Node) {
	node.closedOnce.Do(func() {
		(*node.conn).Close()
		delete(sv.nodes, node.ID)
		close(node.closedChan)
		log.Printf("Node disconnected %s", node.ID)
	})
}
