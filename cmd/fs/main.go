package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/svaan1/tcc-go/internal/storage"
)

func main() {
	ctx := context.Background()

	bucket := "videos"
	fileName := "sample.mp4"
	filePath := "./samples/" + fileName

	store := storage.NewFileSystemStorage("./data")

	// upload
	file, _ := os.Open(filePath)
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	size := stat.Size()

	err = store.Upload(ctx, bucket, fileName, file, size, "application/octet-stream")
	if err != nil {
		log.Fatalf("Fail to upload file %v", err)
	}

	// download
	rc, err := store.Download(ctx, bucket, fileName)
	if err != nil {
		panic(err)
	}
	defer rc.Close()

	out, _ := os.Create(fileName)
	defer out.Close()
	_, _ = io.Copy(out, rc)

	fmt.Println("upload + download done")
}
