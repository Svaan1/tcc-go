package client

import (
	"log"

	pb "github.com/svaan1/go-tcc/internal/transcoding"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) register(stream grpc.BidiStreamingClient[pb.NodeMessage, pb.OrchestratorMessage], name string, codecs []string) error {
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

	if err := stream.Send(registerMsg); err != nil {
		return err
	}

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
	c.stream = stream

	return nil
}
