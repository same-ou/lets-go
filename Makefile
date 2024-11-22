
setup:
	@echo "Setting up the environment..."
	docker compose up -d

build:
	@echo "Building the program..."
	go build ./cmd/web

run: build 
	@echo "Running the program..."
	./web

clean:
	@echo "Cleaning up..."
	rm -f web


.PHONY: build run clean setup