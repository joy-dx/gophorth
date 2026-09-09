UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Darwin)
  OS := mac
else ifeq ($(UNAME_S),Linux)
  OS := linux
else
  OS := other
endif

ifeq ($(UNAME_M),x86_64)
  ARCH := amd64
else ifeq ($(UNAME_M),amd64)
  ARCH := amd64
else ifeq ($(UNAME_M),aarch64)
  ARCH := arm64
else ifeq ($(UNAME_M),arm64)
  ARCH := arm64
else
  ARCH := unknown
endif

.PHONY: generate generate-mac generate-linux \
        generate-linux-amd64 generate-linux-arm64 generate help

help:
	@echo "Available commands:"
	@echo "  generate    - Build artefacts for the demonstration"
	@echo ""

generate:
ifeq ($(OS),mac)
	$(MAKE) generate-mac
else ifeq ($(OS),linux)
	$(MAKE) generate-linux-$(ARCH)
else
	$(error Unsupported OS: $(UNAME_S))
endif

generate-mac:
	@echo "--- Generating example artefacts ---"
	@echo "Building v1 native in current folder"
	@wails build -ldflags '-s -w -X main.BuildID=1.0.0'
	@rm -rf ./wails-app.app
	@mv ./build/bin/wails-app.app .
	@echo "Building v2 darwin builds for update test"
	@GOOS=darwin GOARCH=amd64 wails build -ldflags '-s -w -X main.BuildID=2.0.0'
	@tar -cpzf ./wails-app-darwin-amd64.tar.gz -C ./build/bin wails-app.app
	@mv ./wails-app-darwin-amd64.tar.gz ./assets/wails-app-darwin-amd64.tar.gz
	@rm -R ./build/bin
	@GOOS=darwin GOARCH=arm64 wails build -ldflags '-s -w -X main.BuildID=2.0.0'
	@tar -cpzf ./wails-app-darwin-arm64.tar.gz -C ./build/bin wails-app.app
	@mv ./wails-app-darwin-arm64.tar.gz ./assets/wails-app-darwin-arm64.tar.gz
	@rm -R ./build/bin
	@go run cmd/prepare_assets.go 2.0.0


generate-linux-amd64:
	@echo "--- Generating example artefacts ---"
	@echo "Building v1 native in current folder"
	@wails build -ldflags '-s -w -X main.BuildID=1.0.0' --tags "webkit2_41"
	@rm -rf ./wails-app
	@mv ./build/bin/wails-app ./wails-app
	@echo "Building v2 linux builds for update test"
	@GOOS=linux GOARCH=amd64 wails build -ldflags '-s -w -X main.BuildID=2.0.0' --tags "webkit2_41"
	@mv ./build/bin/wails-app ./assets/wails-app-linux-amd64-webkit241
	@go run cmd/prepare_assets.go 2.0.0

generate-linux-arm64:
	@echo "--- Generating example artefacts ---"
	@echo "Building v1 native in current folder"
	@wails build -ldflags '-s -w -X main.BuildID=1.0.0 -X main.Variant=webkit241' --tags "webkit2_41"
	@rm -rf ./wails-app
	@mv ./build/bin/wails-app ./wails-app
	@echo "Building v2 linux builds for update test"
	@GOOS=linux GOARCH=arm64 wails build -ldflags '-s -w -X main.BuildID=2.0.0 -X main.Variant=webkit241' --tags "webkit2_41"
	@mv ./build/bin/wails-app ./assets/wails-app-linux-arm64-webkit241
	@go run cmd/prepare_assets.go 2.0.0