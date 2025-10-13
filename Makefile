.PHONY: docker-build docker-up docker-down test-api test-all

docker-build:
	@echo "🐳 Building Docker image..."
	docker build \
		--build-arg TEMPORAL_VERSION=$${TEMPORAL_VERSION:-1.28.1} \
		-t temporal-frontend-apikey:$${TEMPORAL_VERSION:-1.28.1} \
		.
	@echo "✅ Docker image built: temporal-frontend-apikey:$${TEMPORAL_VERSION:-1.28.1}"

docker-up:
	@echo "🚀 Starting Temporal with API key authentication..."
	cd test-docker-compose && docker compose up -d
	sleep 15

docker-down:
	@echo "🧹 Removing Docker volumes..."
	cd test-docker-compose && docker compose down -v
	@echo "✅ All cleaned"

test-api:
	@echo "🧪 Running test suite..."
	@echo ""
	cd test-docker-compose && ./test-api.sh
	@echo ""
	@echo "All tests completed!"
	@clean

test-all: docker-build docker-up test-api docker-down
