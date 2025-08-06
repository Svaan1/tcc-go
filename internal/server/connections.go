package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/svaan1/go-tcc/internal/protocols"
)

type Node struct {
	ID            uuid.UUID               `json:"id"`
	Name          string                  `json:"name"`
	Codecs        []string                `json:"codecs"`
	ResourceUsage protocols.ResourceUsage `json:"resourceUsage"`

	conn       *net.Conn
	logger     *log.Logger
	closedOnce sync.Once
	closedChan chan struct{}
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
			ID:     uuid.New(),
			Name:   register.Name,
			Codecs: register.Codecs,

			conn:       &conn,
			logger:     log.New(os.Stdout, fmt.Sprintf("[%s] ", register.Name), log.LstdFlags),
			closedChan: make(chan struct{}),
			ResourceUsage: protocols.ResourceUsage{
				Time: time.Now(),
			},
		}
		sv.Nodes[node.ID] = node

		go sv.handleClientMessages(node)
		go sv.handleResourceUsagePolling(node)

		node.logger.Printf("Node registered")
	}
}

func (sv *Server) closeConnection(node *Node) {
	node.closedOnce.Do(func() {
		(*node.conn).Close()
		delete(sv.Nodes, node.ID)
		close(node.closedChan)
		node.logger.Print("Node disconnected")
	})
}
