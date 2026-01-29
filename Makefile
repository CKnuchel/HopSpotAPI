.PHONY: build run test test-unit test-integration coverage mocks clean swagger

# Build
build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

# Testing
test:
	go test -v ./...

test-unit:
	go test -v -short ./...

test-integration:
	go test -v -run Integration ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Mocks generieren
mocks:
	mockery --all --with-expecter --output=./mocks

# Swagger docs
swagger:
	swag init -g cmd/server/main.go

# Cleanup
clean:
	rm -rf bin/
	rm -rf mocks/
	rm -f coverage.out coverage.html