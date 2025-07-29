package protocols

import (
	"net"

	"github.com/vmihailenco/msgpack/v5"
)

type Packet struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func UnmarshalPacket(data []byte) (*Packet, error) {
	var p Packet

	err := msgpack.Unmarshal(data, &p)

	if err != nil {
		return &Packet{}, err
	}

	return &p, nil
}

func MarshalPacket(p *Packet) ([]byte, error) {
	b, err := msgpack.Marshal(&p)

	if err != nil {
		return []byte{}, err
	}

	return b, nil
}

func ReceivePacket(conn *net.Conn) (*Packet, error) {
	buffer := make([]byte, 1024)
	n, err := (*conn).Read(buffer)
	if err != nil {
		return nil, err
	}

	packet, err := UnmarshalPacket(buffer[:n])
	if err != nil {
		return nil, err
	}

	return packet, nil
}

func SendPacket(conn *net.Conn, p Packet) error {
	msg, err := MarshalPacket(&p)
	if err != nil {
		return err
	}

	_, err = (*conn).Write(msg)
	if err != nil {
		return err
	}

	return nil
}
