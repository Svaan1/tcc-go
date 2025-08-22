package server

import (
	"fmt"

	"github.com/google/uuid"
	pb "github.com/svaan1/go-tcc/internal/transcoding"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (sv *Server) GetNodes() []*Node {
	nodes := make([]*Node, 0, len(sv.Nodes))
	for _, v := range sv.Nodes {
		nodes = append(nodes, v)
	}

	return nodes
}

func (sv *Server) AssignJob(input, output, crf, preset, audioCodec, videoCodec string) error {
	job := pb.JobAssignmentRequest{
		JobId:      uuid.NewString(),
		InputPath:  input,
		OutputPath: output,
		Crf:        crf,
		Preset:     preset,
		AudioCodec: audioCodec,
		VideoCodec: videoCodec,
	}

	msg := pb.OrchestratorMessage{
		Base: &pb.MessageBase{
			MessageId: "job-assignment-" + job.JobId,
			Timestamp: timestamppb.Now(),
		},
		Payload: &pb.OrchestratorMessage_JobAssignmentRequest{
			JobAssignmentRequest: &job,
		},
	}

	for _, node := range sv.Nodes {
		node.stream.Send(&msg)
		return nil
	}

	return fmt.Errorf("no node available")
}
