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

func (sv *Server) GetQueue(ctx context.Context, req *pb.GetQueueRequest) (*pb.GetQueueResponse, error) {
	pendingJobs, activeJobs, err := sv.Service.GetQueueInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue info: %w", err)
	}

	pbPendingJobs := make([]*pb.JobInfo, 0, len(pendingJobs))
	for _, job := range pendingJobs {
		pbPendingJobs = append(pbPendingJobs, &pb.JobInfo{
			JobId:       job.ID.String(),
			InputPath:   job.Params.InputPath,
			OutputPath:  job.Params.OutputPath,
			ProfileName: job.Params.ProfileName,
			Status:      "pending",
			EnqueuedAt:  timestamppb.New(job.CreatedAt),
		})
	}

	pbProcessingJobs := make([]*pb.JobInfo, 0)
	for nodeID, jobs := range activeJobs {
		for _, job := range jobs {
			var status string
			switch job.Status {
			case 0:
				status = "pending"
			case 1:
				status = "assigned"
			case 2:
				status = "processing"
			case 3:
				status = "completed"
			case 4:
				status = "failed"
			default:
				status = "unknown"
			}

			jobInfo := &pb.JobInfo{
				JobId:          job.JobID.String(),
				Status:         status,
				AssignedNodeId: nodeID.String(),
			}

			if job.StartedAt != nil {
				jobInfo.StartedAt = timestamppb.New(*job.StartedAt)
			}
			if job.CompletedAt != nil {
				jobInfo.CompletedAt = timestamppb.New(*job.CompletedAt)
			}

			pbProcessingJobs = append(pbProcessingJobs, jobInfo)
		}
	}

	return &pb.GetQueueResponse{
		PendingJobs:     pbPendingJobs,
		ProcessingJobs:  pbProcessingJobs,
		TotalPending:    int32(len(pbPendingJobs)),
		TotalProcessing: int32(len(pbProcessingJobs)),
	}, nil
}

func (sv *Server) GetJobHistory(ctx context.Context, req *pb.GetJobHistoryRequest) (*pb.GetJobHistoryResponse, error) {
	jobs, err := sv.Service.GetJobHistory(ctx, req.StatusFilter, int(req.Limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get job history: %w", err)
	}

	pbJobs := make([]*pb.JobInfo, 0, len(jobs))
	for _, job := range jobs {
		var status string
		switch job.Status {
		case 0: // JobStatusPending
			status = "pending"
		case 1: // JobStatusAssigned
			status = "assigned"
		case 2: // JobStatusRunning
			status = "processing"
		case 3: // JobStatusCompleted
			status = "completed"
		case 4: // JobStatusFailed
			status = "failed"
		default:
			status = "unknown"
		}

		jobInfo := &pb.JobInfo{
			JobId:          job.JobID.String(),
			Status:         status,
			AssignedNodeId: job.NodeID.String(),
			StartedAt:      timestamppb.New(job.StartedAt),
		}

		if job.CompletedAt != nil {
			jobInfo.CompletedAt = timestamppb.New(*job.CompletedAt)
		}

		pbJobs = append(pbJobs, jobInfo)
	}

	return &pb.GetJobHistoryResponse{
		Jobs:       pbJobs,
		TotalCount: int32(len(pbJobs)),
	}, nil
}

func (sv *Server) ClearQueue(ctx context.Context, req *pb.ClearQueueRequest) (*pb.ClearQueueResponse, error) {
	count, err := sv.Service.ClearQueue(ctx)
	if err != nil {
		return &pb.ClearQueueResponse{
			Success:      false,
			ClearedCount: 0,
			Message:      fmt.Sprintf("Failed to clear queue: %v", err),
		}, nil
	}

	log.Printf("Queue cleared, removed %d jobs", count)

	return &pb.ClearQueueResponse{
		Success:      true,
		ClearedCount: int32(count),
		Message:      fmt.Sprintf("Successfully cleared %d jobs from queue", count),
	}, nil
}
