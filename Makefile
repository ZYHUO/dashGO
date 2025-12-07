.PHONY: build run test clean docker

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=xboard

# Build flags
LDFLAGS=-ldflags "-s -w"

all: build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/server

run:
	$(GOCMD) run ./cmd/server -config configs/config.yaml

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

deps:
	$(GOMOD) download
	$(GOMOD) tidy

docker-build:
	docker build -t xboard-go .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd/server

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe ./cmd/server

build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/server

build-all: build-linux build-windows build-darwin

# Agent builds
agent:
	cd agent && $(GOBUILD) $(LDFLAGS) -o xboard-agent .

agent-linux-amd64:
	cd agent && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o xboard-agent-linux-amd64 .

agent-linux-arm64:
	cd agent && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o xboard-agent-linux-arm64 .

agent-darwin-amd64:
	cd agent && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o xboard-agent-darwin-amd64 .

agent-darwin-arm64:
	cd agent && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o xboard-agent-darwin-arm64 .

agent-all: agent-linux-amd64 agent-linux-arm64 agent-darwin-amd64 agent-darwin-arm64

# Database migrations
migrate:
	$(GOCMD) run ./cmd/migrate -action up

migrate-status:
	$(GOCMD) run ./cmd/migrate -action status

migrate-auto:
	$(GOCMD) run ./cmd/migrate -action auto

# Build migrate tool
build-migrate:
	$(GOBUILD) $(LDFLAGS) -o migrate ./cmd/migrate

# Release (build all binaries)
release: build-all agent-all build-migrate
	@echo "Release binaries built successfully"
