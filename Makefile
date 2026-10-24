VERSION ?= 1.0.0
LDFLAGS := -s -w -X cryptomon/internal/app.Version=$(VERSION)

.PHONY: build run test dist docker clean

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/cryptomon ./cmd/cryptomon

run: build
	./bin/cryptomon

test:
	go vet ./...
	go test ./...

# Binaries for Linux, macOS and Windows (amd64 + arm64) in dist/.
dist:
	VERSION=$(VERSION) ./scripts/build.sh

docker:
	docker compose up --build -d

clean:
	rm -rf bin dist
