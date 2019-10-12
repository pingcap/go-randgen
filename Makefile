
all:resource test bin

resource:
	go generate ./resource

bin:
	go build -o go-randgen cmd/go-randgen/*.go

test:
	go test -count=1 ./...

debug:
	go build -o go-randgen-debug -gcflags "-N -l" cmd/randgen/*.go

.PHONY: all resource bin test debug