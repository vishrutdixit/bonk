# Bonk build configuration
VERSION ?= 0.1.0
API_KEY ?=
LDFLAGS := -s -w
ifdef API_KEY
	LDFLAGS += -X bonk/internal/llm.embeddedAPIKey=$(API_KEY)
endif

.PHONY: build build-all clean fmt release-tag

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

# Format code
fmt:
	gofmt -w cmd internal

# Create and push a release tag (triggers GitHub release workflow)
# Usage: make release-tag RELEASE_VERSION=v0.2.0
RELEASE_VERSION ?= v$(VERSION)
release-tag:
	@test -n "$(RELEASE_VERSION)" || (echo "RELEASE_VERSION is required (example: v0.2.0)" && exit 1)
	@case "$(RELEASE_VERSION)" in \
		v*) ;; \
		*) echo "RELEASE_VERSION must start with 'v' (example: v0.2.0)"; exit 1 ;; \
	esac
	git tag -a "$(RELEASE_VERSION)" -m "Release $(RELEASE_VERSION)"
	git push origin "$(RELEASE_VERSION)"
