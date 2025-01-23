# Output directory
OUTPUT_DIR=releases

# Get the version from the source file
VERSION=0.2.1

.PHONY: build-all macos macos-arm64 linux clean print-version run-local-image-server run-local-image-client

print-version:
	@echo $(VERSION)

build-all: clean
	$(MAKE) -j4 macos macos-arm64 linux

macos: $(OUTPUT_DIR)
	@echo "Compiling for macOS (x86_64)..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(OUTPUT_DIR)/go-reversi-$(VERSION)-macos .

macos-arm64: $(OUTPUT_DIR)
	@echo "Compiling for macOS (Apple Silicon)..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(OUTPUT_DIR)/go-reversi-$(VERSION)-macos-arm64 .

linux: $(OUTPUT_DIR)
	@echo "Compiling for Linux (x86_64)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/go-reversi-$(VERSION)-linux-x86 .

clean:
	@echo "Cleaning up binaries..."
	rm -rf $(OUTPUT_DIR)/*
	touch $(OUTPUT_DIR)/.gitkeep

# You have to login with
# echo YOUR_GITHUB_TOKEN | docker login ghcr.io -u karintomania --password-stdin
publish:
	docker build -t go-reversi --build-arg VERSION=$(VERSION) . 
	docker tag go-reversi ghcr.io/karintomania/go-reversi:latest
	docker push ghcr.io/karintomania/go-reversi:latest

run-local-image-server:
	docker run --rm -it -p 4696:4696 go-reversi:latest -s -d

run-local-image-client:
	docker run --rm -it go-reversi:latest -url http://172.17.0.1 -d

run-published-image-server:
	docker run --rm -it -p 4696:4696 ghcr.io/karintomania/go-reversi:latest -s -d
