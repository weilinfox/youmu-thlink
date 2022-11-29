
all:
	export GOPATH=${HOME}/go
	go build -o ./build/broker ./broker/
	go build -o ./build/client ./client/

static:
	export GOPATH=${HOME}/go
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ./build/thlink-broker-amd64-linux ./broker/
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ./build/thlink-client-amd64-linux ./client/

loong64:
	export GOPATH=${HOME}/go
	GOOS=linux GOARCH=loong64 go build -o ./build/thlink-broker-loong64-linux ./broker/
	GOOS=linux GOARCH=loong64 go build -o ./build/thlink-client-loong64-linux ./client/

windows:
	export GOPATH=${HOME}/go
	GOOS=windows GOARCH=amd64 go build -o ./build/thlink-broker-amd64-windows.exe ./broker/
	GOOS=windows GOARCH=amd64 go build -o ./build/thlink-client-amd64-windows.exe ./client/

test:
	go test ./utils
	go test ./broker/lib

clean:
	rm -rf ./build
