.PHONY: build run clean test

BINARY_NAME=server
BIN_DIR=bin

build:
	@go build -race -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/server

run: build
	@./$(BIN_DIR)/$(BINARY_NAME)

clean:
	rm -rf $(BIN_DIR)/

test:
	go test -v -race ./...

.DEFAULT_GOAL := run