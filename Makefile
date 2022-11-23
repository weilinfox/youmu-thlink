
all:
	export GOPATH=${HOME}/go
	go build -o ./build/broker ./broker/
	go build -o ./build/client ./client/

static:
	export GOPATH=${HOME}/go
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ./build/thlink-broker ./broker/
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ./build/thlink-client ./client/

windows:
	export GOPATH=${HOME}/go
	GOOS=windows go build -o ./build/thlink-broker.exe ./broker/
	GOOS=windows go build -o ./build/thlink-client.exe ./client/

clean:
	rm -rf ./build
