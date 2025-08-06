# 定义默认版本号（如果未通过 Git 获取到则使用该值）
DEFAULT_VERSION := dev

# 从 Git 动态获取版本号
VERSION := $(shell git describe --abbrev=0 --tags 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo $(DEFAULT_VERSION))


install-plugin:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

protoc-gen-config:
	protoc -I ./pkg/proto/ --go_out=paths=source_relative:./pkg/proto ./pkg/proto/*.proto

build:
	go build -ldflags "-X main.version=$(VERSION)" -o ./bin/rainbow-iptv-source-filterd ./cmd/rainbow-iptv-source-filterd/main.go

clean-log:
	cd log && \
	rm -rf *.log && \
	cd -

