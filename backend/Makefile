.PHONY: run build tidy test setup

# Load backend/.env for make targets (token, port, etc.)
ifneq (,$(wildcard .env))
include .env
export
endif

setup:
	@test -f .env || cp .env.example .env
	@echo "Edit backend/.env and set GITHUB_TOKEN=ghp_your_token"

run: setup
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

tidy:
	go mod tidy

test:
	go test ./...
