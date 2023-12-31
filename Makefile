@run:
	@go run ./cmd/kade/main.go
build:
	@echo "Building binaries for common platforms"
	@rm -rf ./bin
	@mkdir -p ./bin
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ./bin/kade-linux-amd64 ./cmd/kade/main.go
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o ./bin/kade-windows-amd64.exe ./cmd/kade/main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o ./bin/kade-darwin-arm64 ./cmd/kade/main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o ./bin/kade-darwin-amd64 ./cmd/kade/main.go