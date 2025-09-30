package client

import (
	"context"
	"log"
	"time"

	"github.com/svaan1/tcc-go/internal/ffmpeg"
	pb "github.com/svaan1/tcc-go/internal/grpc/proto"
	"github.com/svaan1/tcc-go/internal/node"
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
	Config  ClientConfig
	Service *node.Service

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
		Service: node.NewService(),
		nodeID:  "",
		conn:    nil,
		stream:  nil,
	}
}

func (c *Client) Connect(ctx context.Context, name string, profiles []ffmpeg.EncodingProfile) error {
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

	if err := c.register(stream, name, profiles); err != nil {
		log.Printf("Failed to register %v", err)
		return err
	}

	log.Printf("Sucessfully registered as %s", c.nodeID)

	// Start message and resource handling goroutines
	go c.handleStream(ctx)
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

		if err := c.stream.Send(disconnectMsg); err != nil {
			return err
		}
	}

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

func (c *Client) register(stream grpc.BidiStreamingClient[pb.NodeMessage, pb.OrchestratorMessage], name string, profiles []ffmpeg.EncodingProfile) error {
	var encodingProfiles []*pb.EncodingProfile

	for _, profile := range profiles {
		encodingProfiles = append(encodingProfiles, &pb.EncodingProfile{
			Name:       profile.Name,
			Codec:      profile.Codec,
			Params:     profile.Params,
			EncodeTime: profile.EncodeTime,
			DecodeTime: profile.DecodeTime,
			Fps:        profile.FPS,
			Score:      profile.Score,
		})
	}

	registerMsg := &pb.NodeMessage{
		Base: &pb.MessageBase{
			MessageId: "register-" + name,
			Timestamp: timestamppb.Now(),
		},
		Payload: &pb.NodeMessage_RegisterRequest{
			RegisterRequest: &pb.RegisterRequest{
				Name:             name,
				EncodingProfiles: encodingProfiles,
			},
		},
	}

	if err := stream.Send(registerMsg); err != nil {
		return err
	}

	response, err := stream.Recv()
	if err != nil {
		return err
	}

	registerResponse := response.GetRegisterResponse()
	if registerResponse == nil || !registerResponse.Success {
		return err
	}

	c.nodeID = registerResponse.NodeId
	c.stream = stream

	return nil
}
