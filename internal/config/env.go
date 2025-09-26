package config

import (
	"os"
)

var (
	ServerHostName = GetEnv("SERVER_HOSTNAME", "localhost")
	ServerPortGRPC = GetEnv("SERVER_PORT_GRPC", "8080")

	MinioHostName  = GetEnv("MINIO_HOSTNAME", "localhost:9000")
	MinioAccessKey = GetEnv("MINIO_ACCESS_KEY", "minioadmin")
	MInioSecretKey = GetEnv("MINIO_SECRET_KEY", "minioadmin")

	FileSystemStorageRoot = GetEnv("FILE_SYSTEM_STORAGE_ROOT", "./data")

	ClientName = GetEnv("NODE_NAME", "node")
)

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
