package main

import (
	"log"

	"github.com/svaan1/go-tcc/internal/ffmpeg"
)

func main() {
	sample, err := ffmpeg.GenerateVideoSample(50, "1920x1080")
	if err != nil {
		log.Fatalf("Failed to generate video sample %v", err)
	}

	codec := ffmpeg.Codec{
		Name: "x264",
		Params: []string{
			"-preset", "medium",
			"-crf", "23",
			"-tune", "film",
			"-profile:v", "high",
			"-level", "4.0",
			"-pix_fmt", "yuv420p",
			"-movflags", "+faststart",
		},
	}

	result, err := ffmpeg.BenchmarkCodec(codec, 1, sample)
	if err != nil {
		log.Fatalf("Failed to benchmark codec %s: %v", codec.Name, result)
	}

	log.Printf("Benchmark result for codec %s:\n%+v", codec.Name, result)
}
