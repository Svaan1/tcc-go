package config

import "github.com/svaan1/go-tcc/pkg/utils"

var (
	ServerHostName = utils.GetEnv("SERVER_HOSTNAME", "localhost")
	ServerPortGRPC = utils.GetEnv("SERVER_PORT_GRPC", "8080")
	ServerPortHTTP = utils.GetEnv("SERVER_PORT_HTTP", "8081")

	ClientName   = utils.GetEnv("NODE_NAME", "node")
	ClientCodecs = utils.GetEnv("CODECS", "x264;x265")
)
