package server

import (
	"net"
	"time"

	"github.com/google/uuid"
)

type ServerConfig struct {
	Network string
	Address string

	ResourceUsagePollingTimeout time.Duration
}

type Server struct {
	Config   ServerConfig
	Nodes    map[uuid.UUID]*Node
	listener *net.Listener
}

func New(address string) *Server {
	return &Server{
		Config: ServerConfig{
			Network: "tcp",
			Address: address,

			ResourceUsagePollingTimeout: 10 * time.Second,
		},
		Nodes:    map[uuid.UUID]*Node{},
		listener: nil,
	}
}

func (sv *Server) Listen() error {
	listener, err := net.Listen(sv.Config.Network, sv.Config.Address)
	if err != nil {
		return err
	}

	sv.listener = &listener

	go sv.handleConnections()

	return nil
}

func (sv *Server) GetNodes() []*Node {
	nodes := make([]*Node, 0, len(sv.Nodes))
	for _, v := range sv.Nodes {
		nodes = append(nodes, v)
	}

	return nodes
}
