build:
	go build -o bin/banner-rotation.exe cmd/main.go

run:
	go run cmd/main.go

test:
	go test -v ./...

lint:
	golangci-lint run