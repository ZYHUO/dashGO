.PHONY: build run test clean docker

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=dashgo

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
	docker build -t dashgo .

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
	cd agent && $(GOBUILD) $(LDFLAGS) -o dashgo-agent .

agent-linux-amd64:
	cd agent && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dashgo-agent-linux-amd64 .

agent-linux-arm64:
	cd agent && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dashgo-agent-linux-arm64 .

agent-darwin-amd64:
	cd agent && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dashgo-agent-darwin-amd64 .

agent-darwin-arm64:
	cd agent && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dashgo-agent-darwin-arm64 .

agent-all: agent-linux-amd64 agent-linux-arm64 agent-darwin-amd64 agent-darwin-arm64

# Database migrations
migrate:
	@bash migrate.sh up

migrate-status:
	@bash migrate.sh status

migrate-auto:
	@bash migrate.sh auto

migrate-down:
	@bash migrate.sh down

migrate-reset:
	@bash migrate.sh reset

migrate-create:
	@bash migrate.sh create $(name)

# Build migrate tool
build-migrate:
	$(GOBUILD) $(LDFLAGS) -o migrate ./cmd/migrate

# Release (build all binaries)
release: build-all agent-all build-migrate
	@echo "Release binaries built successfully"

# Local installation
install-dev:
	@bash local-install.sh dev

install-prod:
	@bash local-install.sh prod

install-build:
	@bash local-install.sh build

install-migrate:
	@bash local-install.sh migrate

install-frontend:
	@bash local-install.sh frontend

# Frontend
frontend-install:
	cd web && npm install

frontend-dev:
	cd web && npm run dev

frontend-build:
	cd web && npm run build

# Development helpers
dev: deps
	@echo "Starting development environment..."
	@$(GOCMD) run ./cmd/server -config configs/config.yaml

dev-watch:
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	@air -c .air.toml

# Help
help:
	@echo "dashGO Makefile Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build              - Build server binary"
	@echo "  make build-all          - Build for all platforms"
	@echo "  make agent              - Build agent binary"
	@echo "  make agent-all          - Build agent for all platforms"
	@echo "  make release            - Build all release binaries"
	@echo ""
	@echo "Run Commands:"
	@echo "  make run                - Run server"
	@echo "  make dev                - Run in development mode"
	@echo "  make dev-watch          - Run with hot reload (requires air)"
	@echo ""
	@echo "Installation Commands:"
	@echo "  make install-dev        - Install development environment"
	@echo "  make install-prod       - Install production environment"
	@echo "  make install-build      - Build binaries"
	@echo "  make install-migrate    - Run database migrations"
	@echo "  make install-frontend   - Build frontend"
	@echo ""
	@echo "Frontend Commands:"
	@echo "  make frontend-install   - Install frontend dependencies"
	@echo "  make frontend-dev       - Run frontend dev server"
	@echo "  make frontend-build     - Build frontend for production"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-build       - Build Docker image"
	@echo "  make docker-run         - Start Docker containers"
	@echo "  make docker-stop        - Stop Docker containers"
	@echo ""
	@echo "Database Commands:"
	@echo "  make migrate            - Run migrations"
	@echo "  make migrate-status     - Check migration status"
	@echo "  make migrate-auto       - Auto migrate (dev only)"
	@echo "  make migrate-down       - Rollback migration"
	@echo "  make migrate-reset      - Reset database (dangerous!)"
	@echo "  make migrate-create name=xxx - Create new migration"
	@echo ""
	@echo "Other Commands:"
	@echo "  make test               - Run tests"
	@echo "  make clean              - Clean build files"
	@echo "  make deps               - Download dependencies"
	@echo "  make help               - Show this help message"
