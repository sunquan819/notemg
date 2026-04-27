.PHONY: all dev build frontend clean test lint

BINARY_NAME=notemg
FRONTEND_DIR=frontend
BUILD_DIR=build

all: frontend build

dev:
	@echo "Starting development server..."
	go run ./cmd/notemg serve --dev

frontend:
	@echo "Building frontend..."
	cd $(FRONTEND_DIR) && npm install && npm run build

frontend-dev:
	@echo "Starting frontend dev server..."
	cd $(FRONTEND_DIR) && npm run dev

build: frontend
	@echo "Building $(BINARY_NAME)..."
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/notemg

build-go:
	@echo "Building $(BINARY_NAME) (no frontend rebuild)..."
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/notemg

clean:
	rm -rf $(BUILD_DIR) $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/node_modules

test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

init:
	go run ./cmd/notemg init

.DEFAULT_GOAL := all
