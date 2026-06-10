.PHONY: run build tidy setup

## run: start your development server
run:
	go run main.go

## build: compile your app to ./bin/app
build:
	@mkdir -p bin
	go build -o bin/app .

## tidy: clean up go.mod and go.sum
tidy:
	go mod tidy

## setup: pull the Ollama models used in this bootcamp
setup:
	@echo "Pulling models — this may take a few minutes on first run..."
	ollama pull llama3.2:3b
	ollama pull nomic-embed-text
	ollama pull llava:7b
	@echo "✅ Models ready"

## help: list available commands
help:
	@grep -E '^##' Makefile | sed 's/## //'