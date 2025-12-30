.PHONY: generate help

help:
	@echo "Available commands:"
	@echo "  generate    - Build artefacts for the demonstration"
	@echo ""

generate:
	@echo "--- Generating example artefacts ---"
	@echo "Building v1 native in current folder"
	@go build -ldflags '-s -w -X github.com/joy-dx/gophorth/examples/from-json-url/cmd.BuildID=1.0.0' -trimpath -o ./app-example .
	@echo "Building v2 darwin builds for update test"
	@GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w -X github.com/joy-dx/gophorth/examples/from-json-url/cmd.BuildID=2.0.0' -trimpath -o ./cmd/assets/app-example-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 go build -ldflags '-s -w -X github.com/joy-dx/gophorth/examples/from-json-url/cmd.BuildID=2.0.0' -trimpath -o ./cmd/assets/app-example-darwin-arm64 .
	@echo "Building v2 linux builds for update test"
	@GOOS=linux GOARCH=amd64 go build -ldflags '-s -w -X github.com/joy-dx/gophorth/examples/from-json-url/cmd.BuildID=2.0.0' -trimpath -o ./cmd/assets/app-example-linux-amd64 .
	@GOOS=linux GOARCH=arm64 go build -ldflags '-s -w -X github.com/joy-dx/gophorth/examples/from-json-url/cmd.BuildID=2.0.0' -trimpath -o ./cmd/assets/app-example-linux-arm64 .
	@echo "Generate release information (checksums, signatures, and version.json)"
	@./app-example prepare-assets --version 2.0.0