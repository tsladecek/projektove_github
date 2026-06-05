VERSION ?= dev
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

test:
	go test ./...

bin/projektove_github-linux-amd64:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build $(LDFLAGS) -o bin/projektove_github-linux-amd64 cmd/cli/main.go

.PHONY: test
