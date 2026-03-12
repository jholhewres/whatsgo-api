VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BINARY  := whatsgo
MODULE  := github.com/jholhewres/whatsgo-api

.PHONY: build run dev tidy test clean frontend

frontend:
	cd web && npm install && npm run build
	rm -rf internal/webui/dist
	cp -r web/dist internal/webui/dist

build: frontend
	CGO_ENABLED=1 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY) ./cmd/whatsgo

build-go:
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
