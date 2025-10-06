package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/svaan1/tcc-go/internal/config"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
	"github.com/svaan1/tcc-go/internal/grpc/client"
)

func main() {
	profiles := getProfiles()

	address := net.JoinHostPort(config.ServerHostName, config.ServerPortGRPC)
	ctx, cancel := context.WithCancel(context.Background())
	clientConn := client.New(address, profiles)

	err := clientConn.Connect(ctx, config.ClientName, profiles)
	if err != nil {
		log.Fatalf("Failed to connect to server %v", err)
	}

	log.Printf("Connected to server at %s", address)

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("Received shutdown signal, disconnecting...")
		cancel()
		clientConn.Close()
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
		clientConn.Close()
	}
}

func getProfiles() []ffmpeg.EncodingProfile {
	profileNames := ffmpeg.ParseProfileNames(config.EncodingProfiles)
	if len(profileNames) == 0 {
		log.Fatal("No encoding profiles specified. Set ENCODING_PROFILES environment variable (e.g., 'H264_Web_Streaming_1080p;VP9_Web_Optimized_1080p')")
	}

	profiles, err := ffmpeg.GetProfilesByNames(profileNames)
	if err != nil {
		log.Fatalf("Failed to get encoding profiles: %v", err)
	}

	log.Printf("Selected profiles: %v", profileNames)

	log.Println("Generating test video for benchmarking...")

	testVideo, err := ffmpeg.GenerateVideoSample(10, "1920x1080")
	if err != nil {
		log.Fatalf("Failed to generate test video: %v", err)
	}
	defer os.Remove(testVideo)

	log.Printf("Test video generated: %s", testVideo)

	log.Println("Benchmarking encoding profiles...")

	benchmarkedProfiles, err := ffmpeg.BenchmarkProfiles(profiles, testVideo)
	if err != nil {
		log.Fatalf("Failed to benchmark profiles: %v", err)
	}

	log.Println("Benchmark Results:")
	for _, profile := range benchmarkedProfiles {
		log.Printf("  %s: Encode=%.2fs, Decode=%.2fs, FPS=%.2f, Score=%.2f",
			profile.Name, profile.EncodeTime, profile.DecodeTime, profile.FPS, profile.Score)
	}

	return benchmarkedProfiles
}
