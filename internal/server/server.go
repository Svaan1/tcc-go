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
	listener *net.Listener
	nodes    map[uuid.UUID]*Node
}

func New() *Server {
	return &Server{
		Config: ServerConfig{
			Network: "tcp",
			Address: "localhost:8081",

			ResourceUsagePollingTimeout: 10 * time.Second,
		},
		listener: nil,
		nodes:    map[uuid.UUID]*Node{},
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
