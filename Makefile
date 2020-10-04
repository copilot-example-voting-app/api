all: build

build:
	go build -o ./bin/api ./cmd/api

test:
	go test -race -cover -count=1 ./...