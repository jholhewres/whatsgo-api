VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BINARY  := whatsgo
MODULE  := github.com/jholhewres/whatsgo-api

.PHONY: build run dev tidy test clean frontend check-deps

check-deps:
	@command -v go >/dev/null 2>&1 || { echo "Error: go is not installed"; exit 1; }
	@test -d cmd/whatsgo || { echo "Error: cmd/whatsgo directory not found. Make sure you are running from the project root."; exit 1; }
	@test -f go.mod || { echo "Error: go.mod not found. Make sure you are running from the project root."; exit 1; }

frontend:
	@test -d web || { echo "Error: web/ directory not found. Skipping frontend build."; exit 1; }
	cd web && npm install && npm run build
	rm -rf internal/webui/dist
	cp -r web/dist internal/webui/dist

build: check-deps frontend
	CGO_ENABLED=1 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY) ./cmd/whatsgo

build-go: check-deps
	CGO_ENABLED=1 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY) ./cmd/whatsgo

run: build
	./$(BINARY)

dev:
	go run ./cmd/whatsgo

dev-frontend:
	cd web && npm run dev

tidy:
	go mod tidy

test:
	go test ./... -v

clean:
	rm -f $(BINARY)

docker-build:
	docker build -t whatsgo-api:$(VERSION) .

docker-up:
	docker compose up -d

docker-down:
	docker compose down
