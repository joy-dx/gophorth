.PHONY: generate help

help:
	@echo "Available commands:"
	@echo "  generate    - Build artefacts for the demonstration"
	@echo ""

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

generate-linux:
	@echo "--- Generating example artefacts ---"
	@echo "Building v1 native in current folder"
	@wails build -ldflags '-s -w -X main.BuildID=1.0.0'
	@echo "Building v2 linux builds for update test"
	@GOOS=linux GOARCH=amd64 wails build -ldflags '-s -w -X main.BuildID=2.0.0'
	@GOOS=linux GOARCH=arm64 wails build -ldflags '-s -w -X main.BuildID=2.0.0'
