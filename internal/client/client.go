package client

import (
	"context"
	"log"
	"time"

	pb "github.com/svaan1/go-tcc/internal/transcoding"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ClientConfig struct {
	Network string
	Address string

	ResourceUsagePollingTickTime time.Duration
}

type Client struct {
	Config ClientConfig

	nodeID string
	conn   *grpc.ClientConn
	stream grpc.BidiStreamingClient[pb.NodeMessage, pb.OrchestratorMessage]
}

func New(address string) *Client {
	return &Client{
		Config: ClientConfig{
			Network: "tcp",
			Address: address,

			ResourceUsagePollingTickTime: 5 * time.Second,
		},
		nodeID: "",
		conn:   nil,
		stream: nil,
	}
}

func (c *Client) Connect(ctx context.Context, name string, codecs []string) error {
	conn, err := grpc.NewClient(c.Config.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	c.conn = conn
	client := pb.NewVideoTranscodingClient(conn)

	stream, err := client.Stream(ctx)
	if err != nil {
		return err
	}

	c.stream = stream

	// Create register request message
	registerMsg := &pb.NodeMessage{
		Base: &pb.MessageBase{
			MessageId: "register-" + name,
			Timestamp: timestamppb.Now(),
		},
		Payload: &pb.NodeMessage_RegisterRequest{
			RegisterRequest: &pb.RegisterRequest{
				Name:   name,
				Codecs: codecs,
			},
		},
	}

	err = stream.Send(registerMsg)
	if err != nil {
		return err
	}

	// Wait for register response and get node ID
	response, err := stream.Recv()
	if err != nil {
		return err
	}

	registerResponse := response.GetRegisterResponse()
	if registerResponse == nil || !registerResponse.Success {
		log.Printf("Registration failed: %s", registerResponse.GetMessage())
		return err
	}

	c.nodeID = registerResponse.NodeId
	log.Printf("Successfully registered as node ID: %s", c.nodeID)

	// Start message and resource handling goroutines
	go c.handleServerMessages(ctx)
	go c.handleResourceUsagePolling(ctx)

	return nil
}

func (c *Client) Close() error {
	if c.nodeID != "" && c.stream != nil {
		// Send disconnect message
		disconnectMsg := &pb.NodeMessage{
			Base: &pb.MessageBase{
				MessageId: "disconnect-" + c.nodeID,
				Timestamp: timestamppb.Now(),
			},
			Payload: &pb.NodeMessage_DisconnectRequest{
				DisconnectRequest: &pb.DisconnectRequest{
					NodeId: c.nodeID,
					Reason: "Client shutdown",
				},
			},
		}

		err := c.stream.Send(disconnectMsg)
		if err != nil {
			log.Printf("Failed to send disconnect message: %v", err)
		}
	}

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
