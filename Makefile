
all:
	@export GOPATH=$HOME/go
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ./build/broker ./broker/
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ./build/client ./client/

clean:
	rm -rf ./build
