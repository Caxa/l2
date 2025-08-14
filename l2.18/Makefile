APP_NAME = calendar
PORT ?= 8081

.PHONY: all build run test race vet lint clean

all: build

build:
	@echo "🔨 Building $(APP_NAME)..."
	@go build -o $(APP_NAME) ./...

run: build
	@echo "🚀 Starting $(APP_NAME) on port $(PORT)..."
	@PORT=$(PORT) ./$(APP_NAME)

	@go test -v ./...

race:
	@go test -race -v ./...

vet:
	@go vet ./...


clean:
	@echo "🧹 Cleaning up..."
	@rm -f $(APP_NAME)
	@rm -f server.log
