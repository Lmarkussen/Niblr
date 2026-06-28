.PHONY: build run test fmt clean

BINARY := niblr

build:
	go build -o $(BINARY) .

run:
	go run .

test:
	go test ./...

fmt:
	gofmt -w .

clean:
	rm -f $(BINARY)
