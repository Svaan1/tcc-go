package client

import (
	"net"
	"time"

	"github.com/svaan1/go-tcc/internal/protocols"
)

type ClientConfig struct {
	Network string
	Address string

	ResourceUsagePollingTickTime time.Duration
}

type Client struct {
	Config ClientConfig
	conn   *net.Conn
}

func New() *Client {
	return &Client{
		Config: ClientConfig{
			Network: "tcp",
			Address: "localhost:8081",

			ResourceUsagePollingTickTime: 5 * time.Second,
		},
		conn: nil,
	}
}

func (c *Client) Connect(name string, codecs []string) error {
	conn, err := net.Dial(c.Config.Network, c.Config.Address)
	if err != nil {
		return err
	}

	c.conn = &conn

	register := protocols.NewRegisterPacket(name, codecs)
	protocols.SendPacket(&conn, *register)

	go c.handleServerMessages()
	go c.handleResourceUsagePolling()

	return nil
}
