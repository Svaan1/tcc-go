package main

import (
	"log"

	"github.com/svaan1/tcc-go/internal/ffmpeg"
)

func main() {
	profile, err := ffmpeg.GenerateEncodingProfile("264-slow", "libx264", []string{}, 1, "./samples/sample.mp4")

	if err != nil {
		log.Printf("Failed to generate encoding profile %v", err)
	}

	log.Print(profile)
}
