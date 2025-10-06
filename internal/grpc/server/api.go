package server

import (
	"context"
	"fmt"
	"log"
	"sort"

	pb "github.com/svaan1/tcc-go/internal/grpc/proto"
	"github.com/svaan1/tcc-go/internal/orchestrator/jq"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (sv *Server) GetAllNodes(ctx context.Context, req *pb.GetAllNodesRequest) (*pb.GetAllNodesResponse, error) {
	nodes, err := sv.Service.ListNodes(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get nodes %v", err)
	}

	nodeInfos := make([]*pb.NodeInfo, len(nodes))
	for i, node := range nodes {
		var profiles []*pb.EncodingProfile

		for _, profile := range node.Profiles {
			profiles = append(profiles, &pb.EncodingProfile{
				Name:       profile.Name,
				Codec:      profile.Codec,
				Params:     profile.Params,
				EncodeTime: profile.EncodeTime,
				DecodeTime: profile.DecodeTime,
				Fps:        profile.FPS,
				Score:      profile.Score,
			})
		}

		nodeInfos[i] = &pb.NodeInfo{
			NodeId:           node.ID.String(),
			Name:             node.Name,
			EncodingProfiles: profiles,
			CpuPercent:       node.ResourceUsage.CPUUsagePercent,
			MemoryPercent:    node.ResourceUsage.MemoryUsagePercent,
			DiskPercent:      node.ResourceUsage.DiskUsagePercent,
			LastSeen:         timestamppb.New(node.HeartBeat),
		}
	}

	sort.SliceStable(nodeInfos, func(i, j int) bool {
		return nodeInfos[i].NodeId < nodeInfos[j].NodeId
	})

	return &pb.GetAllNodesResponse{
		Nodes:      nodeInfos,
		TotalCount: int32(len(nodes)),
	}, nil
}

func (sv *Server) EnqueueJob(ctx context.Context, req *pb.EnqueueJobRequest) (*pb.EnqueueJobResponse, error) {
	jobID, err := sv.Service.EnqueueJob(ctx, jq.JobParams{
		InputPath:   req.InputPath,
		OutputPath:  req.OutputPath,
		ProfileName: req.ProfileName,
	})

	if err != nil {
		return &pb.EnqueueJobResponse{
			JobId:   jobID.String(),
			Success: false,
			Message: fmt.Sprintf("Failed to enqueue job: %v", err),
		}, nil
	}

	log.Printf("Job %s enqueued", jobID)

	return &pb.EnqueueJobResponse{
		JobId:   jobID.String(),
		Success: true,
		Message: "Job successfully enqueued",
	}, nil
}
