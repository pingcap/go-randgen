
all:resource bin

resource:
	go generate ./resource

bin:
	go build -o go-randgen cmd/randgen/*.go

test:
	go test ./...

debug:
	go build -o go-randgen-debug -gcflags "-N -l" cmd/randgen/*.go

.PHONY: all resource bin test debug