.PHONY: generate help

help:
	@echo "Available commands:"
	@echo "  generate    - Build artefacts for the demonstration"
	@echo ""

generate:
	@echo "--- Generating example artefacts ---"
	@echo "Building v1 native in current folder"
	@go build -ldflags '-s -w -X github.com/joy-dx/gophorth/examples/from-github-release/cmd.BuildID=1.0.0' -trimpath -o ./app-example .