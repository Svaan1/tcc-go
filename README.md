TODO:

- run with -race flag to check race conditions, add a mu sync.Mutex at the node struct and do a node.mu.Lock() defer node.mu.Unlock() in every operation involving either reading or writing to this.

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protoc --go_out=. --go_opt=paths=source_relative \
 --go-grpc_out=. --go-grpc_opt=paths=source_relative \
 internal/grpc/transcoding/transcoding.proto

- client benchmark
- grafana
- queue + lb
