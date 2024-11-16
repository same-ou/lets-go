.PHONY: run
run:
	@echo "Running the program..."
	go build ./cmd/web && ./web

.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -f web
.PHONY: setup
setup:
	@echo "Setting up the environment..."
	docker compose up -d
