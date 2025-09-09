package config

import (
	"os"
)

var (
	ServerHostName = GetEnv("SERVER_HOSTNAME", "localhost")
	ServerPortGRPC = GetEnv("SERVER_PORT_GRPC", "8080")

	ClientName   = GetEnv("NODE_NAME", "node")
	ClientCodecs = GetEnv("CODECS", "x264;x265")
)

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
