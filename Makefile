
all:resource bin

resource:
	go generate ./resource

bin:
	go build -o go-randgen cmd/randgen/*.go

test:
	go test ./...

.PHONY: resource bin test