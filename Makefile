test: main.go
	@gofmt -w -s main.go
	@goimports -w main.go
	@go vet ./...
	@go test ./...