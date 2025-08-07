# Local development
server:
	go run cmd/server/main.go

client:
	go run cmd/client/main.go

# Docker builds
docker-server:
	docker build -f docker/server.Dockerfile -t go-tcc-server .

docker-client:
	docker build -f docker/client.Dockerfile -t go-tcc-client .

docker-build: docker-server docker-client

# Docker compose
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Clean up
docker-clean:
	docker-compose down -v --remove-orphans
	docker system prune -f

.PHONY: server client docker-server docker-client docker-build docker-up docker-down docker-logs docker-clean