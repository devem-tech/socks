.PHONY: build
build:
	@GOOS=linux \
	GOARCH=amd64 \
	CGO_ENABLED=0 \
	go build \
		-ldflags="-s -w -extldflags '-static'" \
		-o bin/i-socks \
		cmd/main.go