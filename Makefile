build:
	go build .
.PHONY: go-build

vendor:
	go mod vendor
.PHONY: vendor

install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
	export PATH="$PATH:$(go env GOPATH)/bin"
.PHONY: install

protos:
	protoc --go_out=./ --go-grpc_out=./ proto/goChat.proto
.PHONY: protos