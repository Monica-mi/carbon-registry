all: build

build:
	goreleaser release --snapshot --rm-dist --skip-sign --skip-publish

clean:
	rm -rf dist

deps:
	glide install

udp:
	go run cmd/udp.go

start:
	go run cmd/registry.go configs/config.ini
