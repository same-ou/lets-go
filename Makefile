setup:
	@echo "Setting up the environment..."
	docker compose up -d

down:
	@echo "Shutting down the environment..."
	docker compose down

build:
	@echo "Building the program..."
	go build ./cmd/web

run: build 
	@echo "Running the program..."
	./web

clean:
	@echo "Cleaning up..."
	rm -f web

test:
	@echo "Unit Testing ..."
	go test -v ./cmd/web

.PHONY: build run clean setup test down