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

Resource aware scheduling

- Codec-specific performance
- Current workload/queue depth
- Network proximity
- Historical reliability


TODOS:1
- Get CPU usage on benchmark
- Predefine encoding profiles
- Return job result to orchestrator (success / fail)
- Make the job dequeue consider already running jobs in each node
- implement more strategies
- create initial tests and a script to run them with each strategy