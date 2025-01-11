.PHONY: build run test clean

# Variables
BINARY_NAME=k8s-glance-backend
GO=go

# Build the application
build:
	$(GO) build -o bin/$(BINARY_NAME) cmd/server/main.go

# Run the application
run:
	$(GO) run cmd/server/main.go

# Run tests
test:
	$(GO) test -v ./...

# Clean build artifacts
clean:
	rm -f bin/$(BINARY_NAME)
	rm -f *.out

# Format code
fmt:
	$(GO) fmt ./...

# Run code linting
lint:
	golangci-lint run

# Generate swagger documentation
swagger:
	swag init -g cmd/server/main.go -o api/swagger

# Install dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Build and run in development mode
dev: build
	./bin/$(BINARY_NAME)