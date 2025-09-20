package server

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/google/uuid"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
	pb "github.com/svaan1/tcc-go/internal/grpc/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetAllNodes returns all registered nodes with their current resource usage
func (sv *Server) GetAllNodes(ctx context.Context, req *pb.GetAllNodesRequest) (*pb.GetAllNodesResponse, error) {
	nodes, err := sv.Service.ListNodes(context.TODO())

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

// EnqueueJob enqueues a new transcoding job to be assigned to an available node
func (sv *Server) EnqueueJob(ctx context.Context, req *pb.EnqueueJobRequest) (*pb.EnqueueJobResponse, error) {
	jobID := uuid.New().String()

	err := sv.Service.EnqueueJob(ctx, &ffmpeg.EncodingParams{
		InputPath:  req.InputPath,
		OutputPath: req.OutputPath,
		VideoCodec: req.VideoCodec,
	})

	if err != nil {
		return &pb.EnqueueJobResponse{
			JobId:   jobID,
			Success: false,
			Message: fmt.Sprintf("Failed to enqueue job: %v", err),
		}, nil
	}

	log.Printf("Job %s enqueued", jobID)

	return &pb.EnqueueJobResponse{
		JobId:   jobID,
		Success: true,
		Message: "Job successfully enqueued",
	}, nil
}
