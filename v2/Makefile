.PHONY: gen build build_gui buildinstall dev dev_cli install_artefact help fix-wails-frontend-models pack staging

help:
	@echo "Available commands:"
	@echo "  dev    - Start development server (backs up existing config)"
	@echo "  reset  - Restore original config or remove temporary files"
	@echo "  clean  - Remove all configuration files and backups"
	@echo "  help   - Show this help message"
	@echo ""

generate:
	@echo "--- Generating update helper ---"
	@GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w' -trimpath -o ./pkg/updater/updatercopier/assets/update-helper-darwin-amd64 ./pkg/updater/updatercopier/cmd/main.go
	@GOOS=darwin GOARCH=arm64 go build -ldflags '-s -w' -trimpath -o ./pkg/updater/updatercopier/assets/update-helper-darwin-arm64 ./pkg/updater/updatercopier/cmd/main.go
	@GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -trimpath -o ./pkg/updater/updatercopier/assets/update-helper-linux-amd64 ./pkg/updater/updatercopier/cmd/main.go
	@GOOS=linux GOARCH=arm64 go build -ldflags '-s -w' -trimpath -o ./pkg/updater/updatercopier/assets/update-helper-linux-arm64 ./pkg/updater/updatercopier/cmd/main.go
	@GOOS=windows GOARCH=amd64 go build -ldflags '-s -w' -trimpath -o ./pkg/updater/updatercopier/assets/update-helper-windows-amd64.exe ./pkg/updater/updatercopier/cmd/main.go
	@GOOS=windows GOARCH=arm64 go build -ldflags '-s -w' -trimpath -o ./pkg/updater/updatercopier/assets/update-helper-windows-arm64.exe ./pkg/updater/updatercopier/cmd/main.go

%:
	@: