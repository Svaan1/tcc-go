package client

import (
	"log"

	"github.com/svaan1/go-tcc/internal/protocols"
)

func (c *Client) handleServerMessages() {
	defer func() {
		(*c.conn).Close()
		log.Println("Server closed connection")
	}()

	for {
		packet, err := protocols.ReceivePacket(c.conn)
		if err != nil {
			log.Println("Failed to receive packet, closing connection", err)
			return
		}

		switch packet.Type {

		case protocols.EncodeRequestType:
			protocols.NewEncodeRequestFromPacketData(packet.Data)

		}
	}
}
