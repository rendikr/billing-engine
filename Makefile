.PHONY: help test run clean coverage fmt lint

help:
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

deps:
	go mod download
	go mod tidy

test:
	go test -v ./...

coverage:
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

run:
	go run main.go

fmt:
	go fmt ./...

lint:
	golangci-lint run

clean:
	rm -f coverage.out coverage.html
	go clean

build:
	go build -o bin/billing-engine main.go

.DEFAULT_GOAL := help
