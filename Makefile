# Variables for Docker image names and tags
DOCKER_SERVER_IMAGE ?= word-of-wisdom-server
DOCKER_CLIENT_IMAGE ?= word-of-wisdom-client
DOCKER_TAG ?= latest

# Install golangci-lint
.PHONY: install-lint
install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Linter: Run golangci-lint
.PHONY: lint
lint: install-lint
	golangci-lint run ./...

# Vet: Run go vet
.PHONY: vet
vet:
	go vet ./...

# Test: Run Go tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -race -v ./...

# Build server binary
.PHONY: build-server
build-server:
	go build -race -o bin/server cmd/server/main.go

# Build client binary
.PHONY: build-client
build-client:
	go build -o bin/client cmd/client/main.go

# Run server with environment variables from .env
.PHONY: run-server
run-server:
	source .env && go run cmd/server/main.go

# Run client with environment variables from .env
.PHONY: run-client
run-client:
	source .env && go run cmd/client/main.go

# Build the Docker image for the server
.PHONY: docker-build-server
docker-build-server:
	docker build -t $(DOCKER_SERVER_IMAGE):$(DOCKER_TAG) -f Dockerfile.server .

# Build the Docker image for the client
.PHONY: docker-build-client
docker-build-client:
	docker build -t $(DOCKER_CLIENT_IMAGE):$(DOCKER_TAG) -f Dockerfile.client .

# Build both server and client Docker images
.PHONY: docker-build-all
docker-build-all:
	$(MAKE) docker-build-server
	$(MAKE) docker-build-client

# Start the services with docker-compose and rebuild images if necessary
.PHONY: up
up:
	docker-compose up --build --remove-orphans

# Clean Docker containers and images
.PHONY: clean
clean:
	@echo "Cleaning Docker containers and images..."
	docker-compose down
	docker rmi $(DOCKER_SERVER_IMAGE):$(DOCKER_TAG) $(DOCKER_CLIENT_IMAGE):$(DOCKER_TAG) || true
