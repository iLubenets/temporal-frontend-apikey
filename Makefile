.PHONY: docker-build docker-up docker-down test-integration test test-all

docker-build:
	@echo "🐳 Building Docker image..."
	docker build \
		--build-arg TEMPORAL_VERSION=$${TEMPORAL_VERSION:-1.28.1} \
		-t temporal-frontend-apikey:$${TEMPORAL_VERSION:-1.28.1} \
		.
	@echo "✅ Docker image built: temporal-frontend-apikey:$${TEMPORAL_VERSION:-1.28.1}"

docker-up:
	@echo "🚀 Starting Temporal with API key authentication..."
	docker compose -f test-docker-compose/docker-compose.yaml up -d
	sleep 15

docker-down:
	@echo "🧹 Removing Docker volumes..."
	docker compose -f test-docker-compose/docker-compose.yaml down -v
	@echo "✅ All cleaned"

test-integration:
	@echo "🧪 Running test suite..."
	@echo ""
	./test-docker-compose/test-api.sh
	@echo ""
	@echo "All tests completed!"

test:
	@echo "Run go generate, tidy, lint, test.."
	@echo ""
	go generate ./...
	go mod tidy
	golangci-lint run ./...
	go test -v -timeout=100s -race ./...
	@echo ""
	@echo "All checks completed!"

test-all: test docker-build docker-up test-integration docker-down
