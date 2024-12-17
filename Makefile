all: build
.PHONY: build run clean

build:
	go build -o bin/main cmd/*.go

run:
	go run cmd/main.go

clean:
	rm -rf bin
