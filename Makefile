# Bonk build configuration
VERSION ?= 0.1.0
API_KEY ?=
LDFLAGS := -s -w
ifdef API_KEY
	LDFLAGS += -X bonk/internal/llm.embeddedAPIKey=$(API_KEY)
endif

.PHONY: build build-all clean

# Local build
build:
	go build -ldflags "$(LDFLAGS)" -o bin/bonk ./cmd/bonk

# Build for distribution (all platforms)
build-all: clean
	@mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/bonk-darwin-arm64 ./cmd/bonk
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/bonk-darwin-amd64 ./cmd/bonk
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/bonk-linux-amd64 ./cmd/bonk
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/bonk-linux-arm64 ./cmd/bonk
	@echo "Built binaries in dist/"
	@ls -lh dist/

clean:
	rm -rf dist/
	rm -f bin/bonk
