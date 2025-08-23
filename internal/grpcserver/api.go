package grpcserver

import (
	"fmt"

	"context"

	"github.com/google/uuid"
	"github.com/svaan1/go-tcc/internal/app"
	pb "github.com/svaan1/go-tcc/internal/transcoding"
)

func (sv *Server) GetNodes() []*app.Node {
	return sv.App.ListNodes(context.TODO())
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

	// Select a node via app policy.
	id, err := sv.App.PickNodeForJob(context.TODO())
	if err != nil {
		return fmt.Errorf("no node available: %w", err)
	}

	sv.mu.RLock()
	conn, ok := sv.NodeConns[id]
	sv.mu.RUnlock()

	if !ok {
		return fmt.Errorf("selected node connection not found")
	}

	return conn.SendJobAssignmentRequest(job)
}
