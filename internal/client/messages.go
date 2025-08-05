package client

import (
	"context"
	"log"

	"github.com/svaan1/go-tcc/internal/protocols"
)

func (c *Client) handleServerMessages(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		(*c.conn).Close()
		log.Println("Server closed connection")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			packet, err := protocols.ReceivePacket(c.conn)
			if err != nil {
				log.Println("Failed to receive packet, closing connection", err)
				defer cancel()
				return
			}

			switch packet.Type {

			case protocols.EncodeRequestType:
				protocols.NewEncodeRequestFromPacketData(packet.Data)

			}
		}
	}
}
