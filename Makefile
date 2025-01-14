# Output directory
OUTPUT_DIR=releases

# Get the version from the source file
VERSION=0.2.0

.PHONY: build-all macos macos-arm64 linux clean print-version

print-version:
	@echo $(VERSION)

build-all: clean
	$(MAKE) -j4 macos macos-arm64 linux

macos: $(OUTPUT_DIR)
	@echo "Compiling for macOS (x86_64)..."
	GOOS=darwin GOARCH=amd64 go build -o $(OUTPUT_DIR)/go-reversi-$(VERSION)-macos .

macos-arm64: $(OUTPUT_DIR)
	@echo "Compiling for macOS (Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 go build -o $(OUTPUT_DIR)/go-reversi-$(VERSION)-macos-arm64 .

linux: $(OUTPUT_DIR)
	@echo "Compiling for Linux (x86_64)..."
	GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/go-reversi-$(VERSION)-linux-x86 .

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
