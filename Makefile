test:
	go test ./...

bin/projektove_github-linux-amd64:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o bin/projektove_github-linux-amd64 cmd/cli/main.go

.PHONY: test
