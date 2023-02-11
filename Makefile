
VERSION:=0.0.12
BUILD_ARCH=$(shell go env GOARCH)
BUILD_OS=$(shell go env GOOS)

all:
	export GOPATH=${HOME}/go
	# go build -o ./build/thlink-broker ./broker/
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ./build/thlink-broker-v${VERSION}-${BUILD_ARCH}-${BUILD_OS} ./broker/
	go build -o ./build/thlink-client-v${VERSION}-${BUILD_ARCH}-${BUILD_OS} ./client/
	go build -o ./build/thlink-client-gtk-v${VERSION}-${BUILD_ARCH}-${BUILD_OS} ./client-gtk3/

test:
	go test ./utils
	go test ./broker/lib
	go test ./client/lib

clean:
	rm -rf ./build
