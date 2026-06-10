.PHONY: run build tidy clean setup help

## run: start the development server
run:
	go run main.go

## build: compile to ./bin/app
build:
	@mkdir -p bin
	go build -o bin/app .
	@echo "✅ Built → bin/app"

## tidy: download dependencies and regenerate go.sum
tidy:
	go mod tidy

## clean: remove build artifacts
clean:
	rm -rf bin/

## setup: download Go dependencies and pull the Ollama model
setup:
	@echo "Downloading Go dependencies (goldmark)..."
	go mod tidy
	@echo "Pulling llama3.2:3b..."
	ollama pull llama3.2:3b
	@echo ""
	@echo "✅ Ready. Run: make run → http://localhost:8080"

## help: list all available commands
help:
	@grep -E '^##' Makefile | sed 's/## //'