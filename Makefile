
all:resource test bin

resource:
	go generate ./resource

bin:
	go build -o go-randgen cmd/go-randgen/*.go

test:
	go test -race -count=1 ./...

debug:
	go build -o go-randgen-debug -gcflags "-N -l" cmd/go-randgen/*.go


darwin: # cross compile to mac
	GOOS=darwin GOARCH=amd64 go build -o go-randgen-darwin cmd/go-randgen/*.go

.PHONY: all resource bin test debug darwin