install-plugin:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

protoc-gen-config:
	protoc -I ./pkg/proto/ --go_out=paths=source_relative:./pkg/proto ./pkg/proto/*.proto

build:
	go build -o ./bin/rainbow-iptv-source-filterd ./cmd/rainbow-iptv-source-filterd/main.go