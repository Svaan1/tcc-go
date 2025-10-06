package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/svaan1/tcc-go/internal/config"
	pb "github.com/svaan1/tcc-go/internal/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to the gRPC server
	address := net.JoinHostPort(config.ServerHostName, config.ServerPortGRPC)
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewVideoTranscodingClient(conn)
	ctx := context.Background()

	log.Printf("Connected to server at %s", address)

	// Enqueue 5 tasks with HEVC_4K profile
	log.Println("Enqueueing 5 tasks with HEVC_4K profile...")
	for i := range 5 {
		outputPath := fmt.Sprintf("output_hevc_4k_%d_%d.mp4", time.Now().Unix(), rand.Intn(10000))

		req := &pb.EnqueueJobRequest{
			InputPath:   "sample.mp4",
			OutputPath:  outputPath,
			ProfileName: "HEVC_4K",
		}

		resp, err := client.EnqueueJob(ctx, req)
		if err != nil {
			log.Printf("Failed to enqueue job %d (HEVC_4K): %v", i+1, err)
			continue
		}

		if resp.Success {
			log.Printf("✓ Job %d (HEVC_4K) enqueued successfully. Job ID: %s, Output: %s", i+1, resp.JobId, outputPath)
		} else {
			log.Printf("✗ Job %d (HEVC_4K) failed: %s", i+1, resp.Message)
		}
	}

	// Enqueue 5 tasks with H264_1080p profile
	log.Println("\nEnqueueing 5 tasks with H264_1080p profile...")
	for i := range 5 {
		outputPath := fmt.Sprintf("output_h264_1080p_%d_%d.mp4", time.Now().Unix(), rand.Intn(10000))

		req := &pb.EnqueueJobRequest{
			InputPath:   "sample.mp4",
			OutputPath:  outputPath,
			ProfileName: "H264_1080p",
		}

		resp, err := client.EnqueueJob(ctx, req)
		if err != nil {
			log.Printf("Failed to enqueue job %d (H264_1080p): %v", i+1, err)
			continue
		}

		if resp.Success {
			log.Printf("✓ Job %d (H264_1080p) enqueued successfully. Job ID: %s, Output: %s", i+1, resp.JobId, outputPath)
		} else {
			log.Printf("✗ Job %d (H264_1080p) failed: %s", i+1, resp.Message)
		}
	}

	log.Println("\nAll jobs enqueued successfully!")
}
