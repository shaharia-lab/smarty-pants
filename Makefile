# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=smarty-pants
BINARY_UNIX=$(BINARY_NAME)_unix

# Version
VERSION=$(shell git describe --tags --always --dirty)

# Directories
BACKEND_DIR=backend
FRONTEND_DIR=frontend/smarty-pants

# Frontend parameters
NPM=npm

.PHONY: all backend-build backend-build-prod backend-test backend-clean backend-run backend-deps backend-build-linux backend-docker-build backend-migrate-up backend-migrate-down frontend-dev frontend-build frontend-start frontend-lint

all: backend-test backend-build frontend-build

# Backend commands
backend-build:
	cd $(BACKEND_DIR) && $(GOBUILD) -o $(BINARY_NAME) -v

backend-build-prod:
	cd $(BACKEND_DIR) && $(GOBUILD) -o $(BINARY_NAME) -v -ldflags "-X main.Version=$(VERSION)"

backend-test:
	cd $(BACKEND_DIR) && $(GOTEST) -v ./...

backend-clean:
	cd $(BACKEND_DIR) && $(GOCLEAN)
	cd $(BACKEND_DIR) && rm -f $(BINARY_NAME)
	cd $(BACKEND_DIR) && rm -f $(BINARY_UNIX)

backend-run:
	cd $(BACKEND_DIR) && $(GOBUILD) -o $(BINARY_NAME) -v ./...
	cd $(BACKEND_DIR) && ./$(BINARY_NAME)

backend-deps:
	cd $(BACKEND_DIR) && $(GOGET) -v -d ./...

backend-build-linux:
	cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v

backend-docker-build:
	cd $(BACKEND_DIR) && docker build -t $(BINARY_NAME):latest .

backend-migrate-up:
	cd $(BACKEND_DIR) && go run main.go migrate up

backend-migrate-down:
	cd $(BACKEND_DIR) && go run main.go migrate down

backend-lint:
	cd $(BACKEND_DIR) && golangci-lint run

# Frontend commands
frontend-dev:
	cd $(FRONTEND_DIR) && $(NPM) run dev

frontend-build:
	cd $(FRONTEND_DIR) && $(NPM) run build

frontend-start:
	cd $(FRONTEND_DIR) && $(NPM) run start

frontend-lint:
	cd $(FRONTEND_DIR) && $(NPM) run lint

frontend-test:
	cd $(FRONTEND_DIR) && $(NPM) run test

# Combined commands
dev: backend-run frontend-dev

build: backend-build frontend-build

clean: backend-clean
	cd $(FRONTEND_DIR) && $(NPM) run clean

test: backend-test
	cd $(FRONTEND_DIR) && $(NPM) run test