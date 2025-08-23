package grpcserver

import (
	"fmt"

	"github.com/google/uuid"
	pb "github.com/svaan1/go-tcc/internal/transcoding"
)

func (sv *Server) GetNodes() []*Node {
	sv.mu.RLock()
	defer sv.mu.RUnlock()

	nodes := make([]*Node, 0, len(sv.Nodes))
	for _, v := range sv.Nodes {
		nodes = append(nodes, v)
	}

	return nodes
}

func (sv *Server) AssignJob(input, output, crf, preset, audioCodec, videoCodec string) error {
	job := &pb.JobAssignmentRequest{
		JobId:      uuid.NewString(),
		InputPath:  input,
		OutputPath: output,
		Crf:        crf,
		Preset:     preset,
		AudioCodec: audioCodec,
		VideoCodec: videoCodec,
	}

	// temporary, just to assign to the first node
	for _, node := range sv.Nodes {
		node.SendJobAssignmentRequest(job)
		return nil
	}

	return fmt.Errorf("no node available")
}
