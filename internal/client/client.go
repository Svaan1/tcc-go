package client

import (
	"context"
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

func New(address string) *Client {
	return &Client{
		Config: ClientConfig{
			Network: "tcp",
			Address: address,

			ResourceUsagePollingTickTime: 5 * time.Second,
		},
		conn: nil,
	}
}

func (c *Client) Connect(ctx context.Context, name string, codecs []string) error {
	ctx, cancel := context.WithCancel(ctx)

	var d net.Dialer
	conn, err := d.DialContext(ctx, c.Config.Network, c.Config.Address)
	if err != nil {
		cancel()
		return err
	}

	c.conn = &conn

	register := protocols.NewRegisterPacket(name, codecs)
	protocols.SendPacket(&conn, *register)

	go c.handleServerMessages(ctx, cancel)
	go c.handleResourceUsagePolling(ctx)

	return nil
}
