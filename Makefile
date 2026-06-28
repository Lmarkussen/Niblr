.PHONY: build run test fmt clean release-linux

BINARY := niblr
VERSION ?= dev
LDFLAGS := -X main.version=$(VERSION)
DIST_DIR := dist
LINUX_ARCHIVE := $(DIST_DIR)/niblr-linux-amd64.tar.gz

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

run:
	go run .

test:
	go test ./...

fmt:
	gofmt -w .

clean:
	rm -f $(BINARY)
	rm -rf $(DIST_DIR)

release-linux: clean
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/niblr .
	tar -czf $(LINUX_ARCHIVE) -C $(DIST_DIR) niblr -C .. README.md LICENSE CHANGELOG.md
