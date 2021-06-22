
all:resource test bin format

resource:
	go generate ./resource

bin:
	go build -o go-randgen cmd/go-randgen/*.go

test:
	# close cache by -count=1
	go test -race -coverprofile=cover.out -count=1 ./...
	go tool cover -html=cover.out -o coverage.html


debug:
	go build -o go-randgen-debug -gcflags "-N -l" cmd/go-randgen/*.go


darwin: # cross compile to mac
	GOOS=darwin GOARCH=amd64 go build -o go-randgen-darwin cmd/go-randgen/*.go

format:
	go fmt ./...

.PHONY: all resource bin test debug darwin format